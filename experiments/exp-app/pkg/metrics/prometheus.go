package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/network"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

// PrometheusSnapshot captures all Prometheus metrics at a point in time
type PrometheusSnapshot struct {
	Timestamp  time.Time                   `json:"timestamp"`
	Chaincodes map[string]ChaincodeMetrics `json:"chaincodes"` // carbon, interop
	Peers      map[string]PeerMetrics      `json:"peers"`      // mma, farmers, companies
	Orderers   []OrdererMetrics            `json:"orderers"`   // orderer instances
	ExpApp     *ExpAppMetrics              `json:"exp_app"`    // exp-app's own metrics from collector
}

// ChaincodeMetrics holds metrics for one chaincode across all orgs
type ChaincodeMetrics struct {
	Name       string                        `json:"name"`
	TxRequests map[string]map[string]float64 `json:"tx_requests"` // org -> function -> count
	TxDuration map[string]map[string]float64 `json:"tx_duration"` // org -> function -> sum (seconds)
}

// PeerMetrics holds Fabric peer metrics
type PeerMetrics struct {
	Organization            string  `json:"organization"`
	BlockHeight             float64 `json:"block_height"`
	TransactionCount        float64 `json:"transaction_count"`
	EndorsementDuration     float64 `json:"endorsement_duration_sum"`
	EndorsementACLDuration  float64 `json:"endorsement_acl_duration_sum"`
	GossipStateHeight       float64 `json:"gossip_state_height"`
	ChaincodeLaunchDuration float64 `json:"chaincode_launch_duration_sum"`
}

// OrdererMetrics holds Fabric orderer metrics
type OrdererMetrics struct {
	Name                     string  `json:"name"`
	CommittedBlockNumber     float64 `json:"committed_block_number"`
	BroadcastProcessedCount  float64 `json:"broadcast_processed_count"`
	BroadcastEnqueueDuration float64 `json:"broadcast_enqueue_duration_sum"`
	DeliverRequestsCompleted float64 `json:"deliver_requests_completed"`
}

// ExpAppMetrics holds exp-app's own application metrics from the collector
type ExpAppMetrics struct {
	TotalTxs        int64   `json:"total_txs"`
	SuccessfulTxs   int64   `json:"successful_txs"`
	FailedTxs       int64   `json:"failed_txs"`
	ThroughputTPS   float64 `json:"throughput_tps"`
	LatencyP50MS    float64 `json:"latency_p50_ms"`
	LatencyP95MS    float64 `json:"latency_p95_ms"`
	LatencyP99MS    float64 `json:"latency_p99_ms"`
	LatencyAvgMS    float64 `json:"latency_avg_ms"`
	LatencyMinMS    float64 `json:"latency_min_ms"`
	LatencyMaxMS    float64 `json:"latency_max_ms"`
	SuccessRatePerc float64 `json:"success_rate_perc"`
}

type chaincodeMetricValueKind int

const (
	chaincodeMetricValueCounterOrGauge chaincodeMetricValueKind = iota
	chaincodeMetricValueHistogramSum
	chaincodeMetricValueHistogramCount
)

// CollectPrometheusSnapshot queries all Prometheus endpoints and returns a snapshot
func CollectPrometheusSnapshot(ctx context.Context, profile *network.NetworkProfile, collector MetricsCollector) (*PrometheusSnapshot, error) {
	snapshot := &PrometheusSnapshot{
		Timestamp:  time.Now(),
		Chaincodes: make(map[string]ChaincodeMetrics),
		Peers:      make(map[string]PeerMetrics),
		Orderers:   make([]OrdererMetrics, 0),
	}

	// Collect chaincode metrics
	for ccName, ccConfig := range profile.Chaincodes {
		if !ccConfig.MetricsEnabled {
			continue
		}

		ccMetrics := ChaincodeMetrics{
			Name:       ccName,
			TxRequests: make(map[string]map[string]float64),
			TxDuration: make(map[string]map[string]float64),
		}

		for orgName, endpoint := range ccConfig.Metrics.Endpoints {
			metrics, err := fetchPrometheusMetrics(ctx, endpoint)
			if err != nil {
				return nil, fmt.Errorf("fetch chaincode %s metrics from %s: %w", ccName, orgName, err)
			}

			// Parse all function-level chaincode metrics.
			ccMetrics.TxRequests[orgName] = extractChaincodeMetrics(
				metrics,
				[]string{
					fmt.Sprintf("%s_chaincode_tx_requests_total", ccName),
					fmt.Sprintf("%s_chaincode_tx_requests", ccName),
				},
				chaincodeMetricValueCounterOrGauge,
			)

			ccMetrics.TxDuration[orgName] = extractChaincodeMetrics(
				metrics,
				[]string{
					fmt.Sprintf("%s_chaincode_tx_duration_seconds", ccName),
					fmt.Sprintf("%s_chaincode_tx_duration_seconds_sum", ccName),
				},
				chaincodeMetricValueHistogramSum,
			)

			if len(ccMetrics.TxRequests[orgName]) == 0 {
				// Fallback: derive request counts from histogram sample_count.
				ccMetrics.TxRequests[orgName] = extractChaincodeMetrics(
					metrics,
					[]string{
						fmt.Sprintf("%s_chaincode_tx_duration_seconds", ccName),
						fmt.Sprintf("%s_chaincode_tx_duration_seconds_count", ccName),
					},
					chaincodeMetricValueHistogramCount,
				)
			}
		}

		snapshot.Chaincodes[ccName] = ccMetrics
	}

	// Collect peer metrics
	for orgName, peerConfig := range profile.Peers {
		for _, peer := range peerConfig.Peers {
			if peer.MetricsEndpoint == "" {
				continue
			}

			metrics, err := fetchPrometheusMetrics(ctx, peer.MetricsEndpoint)
			if err != nil {
				return nil, fmt.Errorf("fetch peer metrics from %s: %w", peer.Name, err)
			}

			peerMetrics := PeerMetrics{
				Organization:            orgName,
				BlockHeight:             extractSingleMetric(metrics, "ledger_blockchain_height"),
				TransactionCount:        extractSingleMetric(metrics, "ledger_transaction_count"),
				EndorsementDuration:     extractSingleMetric(metrics, "endorser_proposal_duration_sum"),
				EndorsementACLDuration:  extractSingleMetric(metrics, "endorser_proposal_acl_check_duration_sum"),
				GossipStateHeight:       extractSingleMetric(metrics, "gossip_state_height"),
				ChaincodeLaunchDuration: extractSingleMetric(metrics, "chaincode_launch_duration_sum"),
			}

			snapshot.Peers[orgName] = peerMetrics
		}
	}

	// Collect orderer metrics
	for _, orderer := range profile.Orderers {
		if orderer.MetricsEndpoint == "" {
			continue
		}

		metrics, err := fetchPrometheusMetrics(ctx, orderer.MetricsEndpoint)
		if err != nil {
			return nil, fmt.Errorf("fetch orderer metrics from %s: %w", orderer.Name, err)
		}

		ordererMetrics := OrdererMetrics{
			Name:                     orderer.Name,
			CommittedBlockNumber:     extractSingleMetric(metrics, "consensus_etcdraft_committed_block_number"),
			BroadcastProcessedCount:  extractSingleMetric(metrics, "broadcast_processed_count"),
			BroadcastEnqueueDuration: extractSingleMetric(metrics, "broadcast_enqueue_duration_sum"),
			DeliverRequestsCompleted: extractSingleMetric(metrics, "deliver_requests_completed"),
		}

		snapshot.Orderers = append(snapshot.Orderers, ordererMetrics)
	}

	// Include exp-app's own metrics from the collector
	if collector != nil {
		appSnapshot := collector.GetSnapshot()
		snapshot.ExpApp = &ExpAppMetrics{
			TotalTxs:        appSnapshot.TotalTxs,
			SuccessfulTxs:   appSnapshot.SuccessfulTxs,
			FailedTxs:       appSnapshot.FailedTxs,
			ThroughputTPS:   appSnapshot.ThroughputTPS,
			LatencyP50MS:    appSnapshot.LatencyP50MS,
			LatencyP95MS:    appSnapshot.LatencyP95MS,
			LatencyP99MS:    appSnapshot.LatencyP99MS,
			LatencyAvgMS:    appSnapshot.LatencyAvgMS,
			LatencyMinMS:    appSnapshot.LatencyMinMS,
			LatencyMaxMS:    appSnapshot.LatencyMaxMS,
			SuccessRatePerc: appSnapshot.SuccessRatePerc,
		}
	}

	return snapshot, nil
}

// fetchPrometheusMetrics fetches metrics from a Prometheus endpoint
func fetchPrometheusMetrics(ctx context.Context, endpoint string) (map[string]*dto.MetricFamily, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// Parse Prometheus text format
	parser := expfmt.NewTextParser(model.UTF8Validation)
	metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse metrics: %w", err)
	}

	return metricFamilies, nil
}

// extractChaincodeMetrics extracts chaincode metrics grouped by transaction function name.
func extractChaincodeMetrics(metrics map[string]*dto.MetricFamily, metricNames []string, valueKind chaincodeMetricValueKind) map[string]float64 {
	result := make(map[string]float64)

	family := findFirstMetricFamily(metrics, metricNames)
	if family == nil {
		return result
	}

	for _, metric := range family.Metric {
		functionName := extractTxNameLabel(metric)
		if functionName == "" {
			continue
		}

		value := extractMetricValue(metric, valueKind)

		// Sum across result labels (ok/error) and any duplicated label sets.
		result[functionName] += value
	}

	return result
}

func findFirstMetricFamily(metrics map[string]*dto.MetricFamily, names []string) *dto.MetricFamily {
	for _, name := range names {
		if family, ok := metrics[name]; ok {
			return family
		}
	}
	return nil
}

func extractTxNameLabel(metric *dto.Metric) string {
	for _, label := range metric.Label {
		switch label.GetName() {
		case "tx_name", "function", "txName":
			return label.GetValue()
		}
	}
	return ""
}

func extractMetricValue(metric *dto.Metric, valueKind chaincodeMetricValueKind) float64 {
	switch valueKind {
	case chaincodeMetricValueHistogramSum:
		if metric.Histogram != nil {
			return metric.Histogram.GetSampleSum()
		}
		if metric.Counter != nil {
			return metric.Counter.GetValue()
		}
		if metric.Gauge != nil {
			return metric.Gauge.GetValue()
		}
		if metric.Untyped != nil {
			return metric.Untyped.GetValue()
		}
	case chaincodeMetricValueHistogramCount:
		if metric.Histogram != nil {
			return float64(metric.Histogram.GetSampleCount())
		}
		if metric.Counter != nil {
			return metric.Counter.GetValue()
		}
		if metric.Gauge != nil {
			return metric.Gauge.GetValue()
		}
		if metric.Untyped != nil {
			return metric.Untyped.GetValue()
		}
	default:
		if metric.Counter != nil {
			return metric.Counter.GetValue()
		}
		if metric.Gauge != nil {
			return metric.Gauge.GetValue()
		}
		if metric.Histogram != nil {
			return float64(metric.Histogram.GetSampleCount())
		}
		if metric.Untyped != nil {
			return metric.Untyped.GetValue()
		}
	}

	return 0
}

// extractSingleMetric extracts a single metric value
func extractSingleMetric(metrics map[string]*dto.MetricFamily, metricName string) float64 {
	family, ok := metrics[metricName]
	if !ok || len(family.Metric) == 0 {
		return 0
	}

	metric := family.Metric[0]

	// Extract value based on metric type
	if metric.Counter != nil {
		return metric.Counter.GetValue()
	} else if metric.Gauge != nil {
		return metric.Gauge.GetValue()
	} else if metric.Histogram != nil {
		return metric.Histogram.GetSampleSum()
	} else if metric.Untyped != nil {
		return metric.Untyped.GetValue()
	}

	return 0
}

// SavePrometheusSnapshot saves snapshot to JSON file
func SavePrometheusSnapshot(snapshot *PrometheusSnapshot, path string) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// LoadPrometheusSnapshot loads snapshot from JSON file
func LoadPrometheusSnapshot(path string) (*PrometheusSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var snapshot PrometheusSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}

	return &snapshot, nil
}

// CalculateDelta computes final - baseline for Prometheus snapshots
func CalculateDelta(baseline, final *PrometheusSnapshot) (*PrometheusSnapshot, error) {
	delta := &PrometheusSnapshot{
		Timestamp:  final.Timestamp,
		Chaincodes: make(map[string]ChaincodeMetrics),
		Peers:      make(map[string]PeerMetrics),
		Orderers:   make([]OrdererMetrics, 0),
		ExpApp:     final.ExpApp, // exp-app metrics are already deltas (collected during run)
	}

	// Delta chaincode metrics
	for ccName, finalCC := range final.Chaincodes {
		baselineCC, ok := baseline.Chaincodes[ccName]
		if !ok {
			delta.Chaincodes[ccName] = finalCC
			continue
		}

		deltaCC := ChaincodeMetrics{
			Name:       ccName,
			TxRequests: make(map[string]map[string]float64),
			TxDuration: make(map[string]map[string]float64),
		}

		for orgName, finalReq := range finalCC.TxRequests {
			deltaCC.TxRequests[orgName] = make(map[string]float64)
			baselineReq := baselineCC.TxRequests[orgName]

			for fn, finalVal := range finalReq {
				baselineVal := baselineReq[fn]
				deltaCC.TxRequests[orgName][fn] = finalVal - baselineVal
			}
		}

		for orgName, finalDur := range finalCC.TxDuration {
			deltaCC.TxDuration[orgName] = make(map[string]float64)
			baselineDur := baselineCC.TxDuration[orgName]

			for fn, finalVal := range finalDur {
				baselineVal := baselineDur[fn]
				deltaCC.TxDuration[orgName][fn] = finalVal - baselineVal
			}
		}

		delta.Chaincodes[ccName] = deltaCC
	}

	// Delta peer metrics
	for orgName, finalPeer := range final.Peers {
		baselinePeer, ok := baseline.Peers[orgName]
		if !ok {
			delta.Peers[orgName] = finalPeer
			continue
		}

		delta.Peers[orgName] = PeerMetrics{
			Organization:            orgName,
			BlockHeight:             finalPeer.BlockHeight - baselinePeer.BlockHeight,
			TransactionCount:        finalPeer.TransactionCount - baselinePeer.TransactionCount,
			EndorsementDuration:     finalPeer.EndorsementDuration - baselinePeer.EndorsementDuration,
			EndorsementACLDuration:  finalPeer.EndorsementACLDuration - baselinePeer.EndorsementACLDuration,
			GossipStateHeight:       finalPeer.GossipStateHeight - baselinePeer.GossipStateHeight,
			ChaincodeLaunchDuration: finalPeer.ChaincodeLaunchDuration - baselinePeer.ChaincodeLaunchDuration,
		}
	}

	// Delta orderer metrics
	for i, finalOrd := range final.Orderers {
		if i >= len(baseline.Orderers) {
			delta.Orderers = append(delta.Orderers, finalOrd)
			continue
		}

		baselineOrd := baseline.Orderers[i]
		delta.Orderers = append(delta.Orderers, OrdererMetrics{
			Name:                     finalOrd.Name,
			CommittedBlockNumber:     finalOrd.CommittedBlockNumber - baselineOrd.CommittedBlockNumber,
			BroadcastProcessedCount:  finalOrd.BroadcastProcessedCount - baselineOrd.BroadcastProcessedCount,
			BroadcastEnqueueDuration: finalOrd.BroadcastEnqueueDuration - baselineOrd.BroadcastEnqueueDuration,
			DeliverRequestsCompleted: finalOrd.DeliverRequestsCompleted - baselineOrd.DeliverRequestsCompleted,
		})
	}

	return delta, nil
}

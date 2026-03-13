package contract

// TODOHP: review metrics
import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type TxMetrics interface {
	Observe(txName string, ok bool, duration time.Duration)
}

type PrometheusTxMetrics struct {
	txRequestsTotal *prometheus.CounterVec
	txDuration      *prometheus.HistogramVec
}

var (
	promMetricsOnce sync.Once
	promMetrics     *PrometheusTxMetrics
)

func NewPrometheusTxMetrics() *PrometheusTxMetrics {
	promMetricsOnce.Do(func() {
		promMetrics = &PrometheusTxMetrics{
			txRequestsTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "carbon_chaincode",
					Name:      "tx_requests_total",
					Help:      "Total number of carbon chaincode transactions.",
				},
				[]string{"tx_name", "result"},
			),
			txDuration: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: "carbon_chaincode",
					Name:      "tx_duration_seconds",
					Help:      "Duration of carbon chaincode transactions.",
					Buckets:   prometheus.DefBuckets,
				},
				[]string{"tx_name", "result"},
			),
		}
	})
	return promMetrics
}

func (m *PrometheusTxMetrics) Observe(txName string, ok bool, d time.Duration) {
	result := "error"
	if ok {
		result = "ok"
	}
	m.txRequestsTotal.WithLabelValues(txName, result).Inc()
	m.txDuration.WithLabelValues(txName, result).Observe(d.Seconds())
}

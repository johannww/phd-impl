package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/workload"
)

func main() {
	// Parse flags
	peerAddr := flag.String("peer", "localhost:7051", "Peer address (host:port)")
	tlsCertPath := flag.String("tls-cert", "", "Path to peer TLS certificate")
	mspID := flag.String("msp-id", "Org1MSP", "MSP ID")
	userCertPath := flag.String("user-cert", "", "Path to user certificate")
	userKeyPath := flag.String("user-key", "", "Path to user private key")
	channelName := flag.String("channel", "carbon", "Channel name")
	chaincodeName := flag.String("chaincode", "carbon", "Chaincode name")
	duration := flag.Duration("duration", 5*time.Minute, "Test duration")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent transactions")
	outputJSON := flag.String("output-json", "results.json", "Output JSON file")
	outputCSV := flag.String("output-csv", "results.csv", "Output CSV file")
	metricsInterval := flag.Duration("metrics-interval", 5*time.Second, "Metrics print interval")

	flag.Parse()

	// Validate required flags
	if *tlsCertPath == "" || *userCertPath == "" || *userKeyPath == "" {
		fmt.Fprintf(os.Stderr, "Error: --tls-cert, --user-cert, and --user-key are required\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Create gateway client
	gatewayCfg := &gateway.GatewayConfig{
		PeerAddr:      *peerAddr,
		TLSCertPath:   *tlsCertPath,
		MspID:         *mspID,
		UserCertPath:  *userCertPath,
		UserKeyPath:   *userKeyPath,
		ChannelName:   *channelName,
		ChaincodeName: *chaincodeName,
	}

	log.Println("Connecting to gateway...")
	client, err := gateway.NewClientWrapper(gatewayCfg)
	if err != nil {
		log.Fatalf("Failed to create gateway client: %v", err)
	}
	defer client.Close()

	log.Println("Connected! Starting performance test...")

	// Create executor
	execCfg := &workload.ExecutorConfig{
		ConcurrencyLevel: *concurrency,
		Duration:         *duration,
		MetricsInterval:  *metricsInterval,
	}

	executor := workload.NewExecutor(client, execCfg)

	// Simple test transaction: CheckCredAttr
	testTx := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
		result, err := c.EvaluateTransaction("CheckCredAttr", "price_viewer")
		return string(result), err
	}

	// Run test
	ctx, cancel := context.WithTimeout(context.Background(), *duration+10*time.Second)
	defer cancel()

	if err := executor.Execute(ctx, testTx, "simple-read"); err != nil {
		log.Printf("Execution error: %v", err)
	}

	// Print final report
	reporter := metrics.NewReporter(executor.GetCollector())
	reporter.PrintFinalReport()

	// Export results
	log.Printf("Exporting results to %s and %s...", *outputJSON, *outputCSV)
	if err := reporter.ExportJSON(*outputJSON); err != nil {
		log.Printf("Failed to export JSON: %v", err)
	}
	if err := reporter.ExportCSV(*outputCSV); err != nil {
		log.Printf("Failed to export CSV: %v", err)
	}

	log.Println("Test complete!")
}

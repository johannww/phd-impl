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
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/network"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/workload"
)

func main() {
	// Parse flags
	profilePath := flag.String("profile", "", "Path to network profile JSON (required)")
	duration := flag.Duration("duration", 5*time.Minute, "Test duration")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent transactions")
	outputJSON := flag.String("output-json", "results.json", "Output JSON file")
	outputCSV := flag.String("output-csv", "results.csv", "Output CSV file")
	metricsInterval := flag.Duration("metrics-interval", 5*time.Second, "Metrics print interval")

	flag.Parse()

	// Validate required flags
	if *profilePath == "" {
		fmt.Fprintf(os.Stderr, "Error: --profile is required\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load network profile
	log.Printf("Loading network profile from %s...", *profilePath)
	profile, err := network.LoadJSON(*profilePath)
	if err != nil {
		log.Fatalf("Failed to load network profile: %v", err)
	}

	// Get first organization from profile
	var orgName string
	var orgConfig network.PeerConfig
	for name, config := range profile.Peers {
		orgName = name
		orgConfig = config
		break
	}

	if orgName == "" {
		log.Fatal("No organizations found in network profile")
	}

	log.Printf("Using organization: %s (MSP: %s)", orgName, orgConfig.MspID)

	// Get first peer from organization
	if len(orgConfig.Peers) == 0 {
		log.Fatal("No peers found in organization")
	}
	peerNode := orgConfig.Peers[0]
	peerAddr := peerNode.Address

	log.Printf("Using peer: %s at %s", peerNode.Name, peerAddr)

	// Get certificate paths
	certs := orgConfig.Certificates
	tlsCertPath := certs.TLSCACert
	userCertPath := certs.User1Cert
	userKeyPath := certs.User1Key
	channelName := profile.Network.ChannelName
	chaincodeName := profile.Chaincode.Name

	// Validate certificates exist
	for _, path := range []string{tlsCertPath, userCertPath, userKeyPath} {
		if _, err := os.Stat(path); err != nil {
			log.Fatalf("Certificate file not found: %s", path)
		}
	}

	// Create gateway client
	gatewayCfg := &gateway.GatewayConfig{
		PeerAddr:      peerAddr,
		TLSCertPath:   tlsCertPath,
		MspID:         orgConfig.MspID,
		UserCertPath:  userCertPath,
		UserKeyPath:   userKeyPath,
		ChannelName:   channelName,
		ChaincodeName: chaincodeName,
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

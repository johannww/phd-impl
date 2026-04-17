package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/network"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/scenarios"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/setup"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/workload"
)

func main() {
	// f, _ := os.Create("cpu.prof")
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// Parse flags
	profilePath := flag.String("profile", "", "Path to network profile JSON (required)")
	armTemplatePath := flag.String("arm-template", "../../tee_auction/azure/arm_template.json", "Path to ARM template JSON for CCE policy")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent transactions")
	tps := flag.Float64("tps", 0, "Target throughput in transactions per second (0 = unlimited)")
	burst := flag.Int("burst", 1, "Token bucket burst size for rate limiting")
	outputJSON := flag.String("output-json", "results.json", "Output JSON file")
	outputCSV := flag.String("output-csv", "results.csv", "Output CSV file")
	metricsInterval := flag.Duration("metrics-interval", 5*time.Second, "Metrics print interval")
	mintInterval := flag.Duration("mint-interval", 100*time.Millisecond, "Interval between credit minting rounds")
	buyBidInterval := flag.Duration("buy-bid-interval", 1*time.Second, "Interval between buy bid submissions")
	sellBidInterval := flag.Duration("sell-bid-interval", 2*time.Second, "Interval between sell bid submissions")
	auctionInterval := flag.Duration("auction-interval", 15*time.Second, "Interval between auction rounds")

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

	// Validate that at least one user exists
	if len(certs.Users) == 0 {
		log.Fatal("No user certificates found in organization")
	}

	tlsCertPath := certs.TLSCACert
	userCertPath := certs.Users[0].Cert
	userKeyPath := certs.Users[0].Key
	channelName := profile.Network.ChannelName

	// Get carbon chaincode config (primary chaincode for exp-app)
	carbonCC, ok := profile.Chaincodes["carbon"]
	if !ok {
		log.Fatal("Carbon chaincode not found in network profile")
	}
	chaincodeName := carbonCC.Name

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

	const (
		nPropsPerOrg    = 2
		nChunksPerProp  = 2
		quantityPerMint = int64(1000)
	)

	setupManager := setup.NewSetupManager(client, profile, *armTemplatePath)
	teeClient, err := setupManager.InitializeBETS(context.Background(), nPropsPerOrg, nChunksPerProp)
	if err != nil {
		log.Fatalf("Setup failed: %v", err)
	}

	// Create executor
	execCfg := &workload.ExecutorConfig{
		ConcurrencyLevel: *concurrency,
		Duration:         *duration,
		MetricsInterval:  *metricsInterval,
		TPS:              *tps,
		BurstSize:        *burst,
	}

	executor := workload.NewExecutor(client, execCfg)

	// Scenarios setup
	creditScenario := scenarios.NewCreditScenario(executor)
	biddingScenario := scenarios.NewBiddingScenario(executor)
	coupledScenario := scenarios.NewCoupledAuctionScenario(executor)
	interopScenario := scenarios.NewInteropScenario(executor)
	_ = interopScenario // Placeholder for future use

	// Set TEE client if it was initialized
	if teeClient != nil {
		log.Println("Using real TEE service for coupled auctions")
		coupledScenario.SetTEEClient(teeClient)
	} else {
		log.Println("TEE not available, using mock results for coupled auctions")
	}

	// Run full BETS workflow
	ctx, cancel := context.WithTimeout(context.Background(), *duration+30*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	log.Println("Starting Parallel Continuous Workload...")

	// 1. Continuous Minting (every 10 seconds)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Launcher: Continuous Minting started")
		if err := creditScenario.MintCreditsContinuous(ctx, client, *mintInterval, nPropsPerOrg, quantityPerMint); err != nil && err != context.Canceled {
			log.Printf("MintCreditsContinuous error: %v", err)
		}
	}()

	// 2. Continuous Buy Bidding
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Launcher: Continuous Buy Bidding started")
		if err := biddingScenario.CreateBuyBidsContinuous(ctx, client, *buyBidInterval); err != nil && err != context.Canceled {
			log.Printf("CreateBuyBidsContinuous error: %v", err)
		}
	}()

	// 3. Continuous Sell Bidding
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Launcher: Continuous Sell Bidding started")
		if err := biddingScenario.CreateSellBidsContinuous(ctx, client, *sellBidInterval); err != nil && err != context.Canceled {
			log.Printf("CreateSellBidsContinuous error: %v", err)
		}
	}()

	// 4. Periodic Auction Settlement (every 15 seconds)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Launcher: Periodic Auction started")
		if err := coupledScenario.PeriodicAuction(ctx, client, *auctionInterval); err != nil && err != context.Canceled {
			log.Printf("PeriodicAuction error: %v", err)
		}
	}()
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

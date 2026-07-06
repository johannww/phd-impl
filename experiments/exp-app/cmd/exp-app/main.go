package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/charts"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/network"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/scenarios"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/setup"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/tee"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/workload"
)

func main() {
	// f, _ := os.Create("cpu.prof")
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// Parse flags
	profilePath := flag.String("profile", "", "Path to network profile JSON (required)")
	duration := flag.Duration("duration", 120*time.Second, "Test duration")
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

	// Metrics collection flags
	enableMetrics := flag.Bool("enable-metrics", false, "Enable post-execution metrics collection and visualization")
	metricsOutputDir := flag.String("metrics-output", "monitoring-exports", "Directory for metrics exports")
	metricsFormats := flag.String("metrics-formats", "png,html,pdf,json", "Chart formats (comma-separated: png,html,pdf,json)")
	organizationFilter := flag.String("organization", strings.TrimSpace(os.Getenv("EXP_APP_ORGANIZATION")), "Organization to use from profile (default: first organization)")
	defaultUserCount := 0
	if envUserCount := strings.TrimSpace(os.Getenv("EXP_APP_USER_COUNT")); envUserCount != "" {
		parsedUserCount, parseErr := strconv.Atoi(envUserCount)
		if parseErr != nil {
			log.Fatalf("Invalid EXP_APP_USER_COUNT value %q: %v", envUserCount, parseErr)
		}
		defaultUserCount = parsedUserCount
	}
	userCountOverride := flag.Int("user-count", defaultUserCount, "Number of users to run per organization (0 = use profile user count)")
	defaultRunCoupled := true
	if envRunCoupled := strings.TrimSpace(os.Getenv("EXP_APP_RUN_COUPLED")); envRunCoupled != "" {
		defaultRunCoupled = strings.EqualFold(envRunCoupled, "true") || envRunCoupled == "1"
	}
	runCoupled := flag.Bool("run-coupled", defaultRunCoupled, "Run coupled auction periodic routine")

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

	// Get organization from profile (explicit if provided, deterministic first otherwise)
	var orgName string
	var orgConfig network.PeerConfig
	if *organizationFilter != "" {
		selected, ok := profile.Peers[*organizationFilter]
		if !ok {
			log.Fatalf("Organization %q not found in network profile", *organizationFilter)
		}
		orgName = *organizationFilter
		orgConfig = selected
	} else {
		orgNames := make([]string, 0, len(profile.Peers))
		for name := range profile.Peers {
			orgNames = append(orgNames, name)
		}
		sort.Strings(orgNames)
		if len(orgNames) > 0 {
			orgName = orgNames[0]
			orgConfig = profile.Peers[orgName]
		}
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

	effectiveUserCount := *userCountOverride
	if effectiveUserCount == 0 {
		effectiveUserCount = len(certs.Users)
	}
	if effectiveUserCount <= 0 {
		log.Fatal("No users available to run workload")
	}
	if effectiveUserCount > len(certs.Users) {
		log.Printf("Requested user count %d exceeds available users %d for org %s; capping", effectiveUserCount, len(certs.Users), orgName)
		effectiveUserCount = len(certs.Users)
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
	carbonChaincodeName := carbonCC.Name

	interopCC, ok := profile.Chaincodes["interop"]
	if !ok {
		log.Fatal("Interop chaincode not found in network profile")
	}
	interopChaincodeName := interopCC.Name

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
		ChaincodeName: carbonChaincodeName,
	}

	log.Println("Connecting to gateway...")
	client, err := gateway.NewClientWrapper(gatewayCfg)
	if err != nil {
		log.Fatalf("Failed to create gateway client: %v", err)
	}
	defer client.Close()

	log.Println("Connected! Starting performance test...")

	// Collect baseline Prometheus metrics if enabled
	var baselineSnapshot *metrics.PrometheusSnapshot
	if *enableMetrics {
		log.Println("Collecting baseline Prometheus metrics snapshot...")
		var err error
		baselineSnapshot, err = metrics.CollectPrometheusSnapshot(context.Background(), profile, nil)
		if err != nil {
			log.Printf("Warning: Failed to collect baseline metrics: %v", err)
			log.Println("Continuing without metrics collection...")
			*enableMetrics = false
		} else {
			// Save baseline snapshot
			if err := os.MkdirAll(*metricsOutputDir, 0755); err != nil {
				log.Printf("Warning: Failed to create metrics output dir: %v", err)
			}
			if err := metrics.SavePrometheusSnapshot(baselineSnapshot, fmt.Sprintf("%s/metrics-baseline.json", *metricsOutputDir)); err != nil {
				log.Printf("Warning: Failed to save baseline snapshot: %v", err)
			}
			log.Println("Baseline metrics collected successfully")
		}
	}

	const (
		nPropsPerIdentity = 2
		nChunksPerProp    = 2
		quantityPerMint   = int64(1000)
	)

	var teeClient *tee.Client
	if *runCoupled && profile.TEEAuction.Enabled {
		candidate := tee.NewClient(fmt.Sprintf("https://%s", profile.TEEAuction.Address), true)
		if err := candidate.Ping(); err != nil {
			log.Printf("TEE ping failed, using mock results for coupled auctions: %v", err)
		} else {
			teeClient = candidate
		}
	}

	setupPropertyIDsByUser := make([][]uint64, effectiveUserCount)
	log.Println("Running identity-scoped setup for all users")

	for userIdx := 0; userIdx < effectiveUserCount; userIdx++ {
		user := certs.Users[userIdx]
		setupGatewayCfg := &gateway.GatewayConfig{
			PeerAddr:      peerAddr,
			TLSCertPath:   tlsCertPath,
			MspID:         orgConfig.MspID,
			UserCertPath:  user.Cert,
			UserKeyPath:   user.Key,
			ChannelName:   channelName,
			ChaincodeName: carbonChaincodeName,
		}

		setupClient, setupClientErr := gateway.NewClientWrapper(setupGatewayCfg)
		if setupClientErr != nil {
			log.Fatalf("Failed to create setup gateway client for user %d: %v", userIdx+1, setupClientErr)
		}

		setupManager := setup.NewSetupManager(setupClient, profile, "")
		setupResult, setupErr := setupManager.InitializeBETS(
			context.Background(),
			nPropsPerIdentity,
			nChunksPerProp,
			effectiveUserCount,
			setup.IdentityAssignment{Organization: orgName, UserIndex: userIdx},
		)
		if setupErr != nil {
			_ = setupClient.Close()
			log.Fatalf("Setup failed for user %d: %v", userIdx+1, setupErr)
		}
		if setupResult == nil || len(setupResult.PropertyIDs) == 0 {
			_ = setupClient.Close()
			log.Fatalf("Setup returned no property IDs for user %d", userIdx+1)
		}
		log.Printf("Setup assigned properties for user %d: %v", userIdx+1, setupResult.PropertyIDs)
		setupPropertyIDsByUser[userIdx] = append([]uint64(nil), setupResult.PropertyIDs...)

		_ = setupClient.Close()
	}

	if teeClient != nil {
		log.Println("Using real TEE service for coupled auctions")
	} else if !(*runCoupled) {
		log.Println("TEE not available, but this instance is not configured to run coupled auctions")
	} else {
		log.Println("TEE not available, using mock results for coupled auctions")
	}

	// Run full BETS workflow
	ctx, cancel := context.WithTimeout(context.Background(), *duration+30*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	log.Printf("Starting parallel workload for org=%s users=%d", orgName, effectiveUserCount)

	type userRuntime struct {
		idx         int
		id          string
		carbon      *gateway.ClientWrapper
		interop     *gateway.ClientWrapper
		executor    *workload.Executor
		credit      *scenarios.CreditScenario
		bidding     *scenarios.BiddingScenario
		interopFlow *scenarios.InteropScenario
		coupled     *scenarios.CoupledAuctionScenario
		propertyIDs []uint64
		ownsCoupled bool
	}

	runtimes := make([]*userRuntime, 0, effectiveUserCount)
	for userIdx := 0; userIdx < effectiveUserCount; userIdx++ {
		user := certs.Users[userIdx]
		propertyIDs := append([]uint64(nil), setupPropertyIDsByUser[userIdx]...)
		if len(propertyIDs) == 0 {
			log.Fatalf("No setup property IDs available for user %d", userIdx+1)
		}
		userGatewayCfg := &gateway.GatewayConfig{
			PeerAddr:      peerAddr,
			TLSCertPath:   tlsCertPath,
			MspID:         orgConfig.MspID,
			UserCertPath:  user.Cert,
			UserKeyPath:   user.Key,
			ChannelName:   channelName,
			ChaincodeName: carbonChaincodeName,
		}

		carbonClient, carbonClientErr := gateway.NewClientWrapper(userGatewayCfg)
		if carbonClientErr != nil {
			log.Fatalf("Failed to create gateway client for user %d: %v", userIdx+1, carbonClientErr)
		}

		interopGatewayCfg := &gateway.GatewayConfig{
			PeerAddr:      peerAddr,
			TLSCertPath:   tlsCertPath,
			MspID:         orgConfig.MspID,
			UserCertPath:  user.Cert,
			UserKeyPath:   user.Key,
			ChannelName:   channelName,
			ChaincodeName: interopChaincodeName,
		}

		interopClient, interopClientErr := gateway.NewClientWrapper(interopGatewayCfg)
		if interopClientErr != nil {
			_ = carbonClient.Close()
			log.Fatalf("Failed to create interop gateway client for user %d: %v", userIdx+1, interopClientErr)
		}

		runtimeExecCfg := &workload.ExecutorConfig{
			ConcurrencyLevel: *concurrency,
			Duration:         *duration,
			MetricsInterval:  *metricsInterval,
			TPS:              *tps,
			BurstSize:        *burst,
		}
		exec := workload.NewExecutor(carbonClient, runtimeExecCfg)
		runtimeCoupled := scenarios.NewCoupledAuctionScenario(exec)
		if teeClient != nil {
			runtimeCoupled.SetTEEClient(teeClient)
		}
		sharedBuckets := scenarios.NewSharedCreditBuckets()

		runtimes = append(runtimes, &userRuntime{
			idx:         userIdx + 1,
			id:          carbonClient.GetIdentityID(),
			carbon:      carbonClient,
			interop:     interopClient,
			executor:    exec,
			credit:      scenarios.NewCreditScenario(exec),
			bidding:     scenarios.NewBiddingScenario(exec, sharedBuckets),
			interopFlow: scenarios.NewInteropScenario(exec, sharedBuckets),
			coupled:     runtimeCoupled,
			propertyIDs: propertyIDs,
			ownsCoupled: userIdx == 0 && *runCoupled,
		})
	}
	defer func() {
		for _, rt := range runtimes {
			_ = rt.carbon.Close()
			_ = rt.interop.Close()
		}
	}()

	runtimeCollectors := make([]metrics.MetricsCollector, 0, len(runtimes))
	for _, rt := range runtimes {
		runtimeCollectors = append(runtimeCollectors, rt.executor.GetCollector())
	}
	combinedCollector := metrics.NewCombinedCollector(runtimeCollectors...)

	for _, rt := range runtimes {
		log.Printf("Launcher: user[%d]=%s minting started", rt.idx, rt.id)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := rt.credit.MintCreditsContinuous(ctx, rt.carbon, *mintInterval, rt.propertyIDs, quantityPerMint); err != nil && err != context.Canceled {
				log.Printf("MintCreditsContinuous user %d error: %v", rt.idx, err)
			}
		}()

		log.Printf("Launcher: user[%d]=%s buy bidding started", rt.idx, rt.id)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := rt.bidding.CreateBuyBidsContinuous(ctx, rt.carbon, *buyBidInterval); err != nil && err != context.Canceled {
				log.Printf("CreateBuyBidsContinuous user %d error: %v", rt.idx, err)
			}
		}()

		log.Printf("Launcher: user[%d]=%s sell bidding started", rt.idx, rt.id)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := rt.bidding.CreateSellBidsContinuous(ctx, rt.carbon, *sellBidInterval); err != nil && err != context.Canceled {
				log.Printf("CreateSellBidsContinuous user %d error: %v", rt.idx, err)
			}
		}()

		log.Printf("Launcher: user[%d]=%s interop workflow started", rt.idx, rt.id)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := rt.interopFlow.HTLCWorkflow(ctx, rt.carbon, rt.interop, 100); err != nil && err != context.Canceled {
				log.Printf("HTLCWorkflow user %d error: %v", rt.idx, err)
			}
		}()

		if rt.ownsCoupled {
			log.Printf("Launcher: user[%d]=%s periodic auction started", rt.idx, rt.id)
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := rt.coupled.PeriodicAuction(ctx, rt.carbon, *auctionInterval); err != nil && err != context.Canceled {
					log.Printf("PeriodicAuction user %d error: %v", rt.idx, err)
				}
			}()
		}
	}

	// Wait for duration to expire
	log.Printf("Workload running for %v...", *duration)
	select {
	case <-time.After(*duration):
		log.Println("Test duration reached, stopping...")
		cancel()
	case <-ctx.Done():
		log.Println("Context done, stopping...")
	}

	wg.Wait()

	// Stop the metrics collector worker pool
	log.Println("Stopping metrics collectors...")
	for _, rt := range runtimes {
		rt.executor.GetCollector().Stop()
	}

	aggregateReporter := metrics.NewReporter(combinedCollector)
	aggregateReporter.PrintFinalReport()

	if err := aggregateReporter.ExportJSON(*outputJSON); err != nil {
		log.Printf("Failed to export aggregate JSON: %v", err)
	}
	if err := aggregateReporter.ExportCSV(*outputCSV); err != nil {
		log.Printf("Failed to export aggregate CSV: %v", err)
	}

	// Export one report set per user
	for _, rt := range runtimes {
		jsonPath := fmt.Sprintf("%s.user-%02d", *outputJSON, rt.idx)
		csvPath := fmt.Sprintf("%s.user-%02d", *outputCSV, rt.idx)
		rep := metrics.NewReporter(rt.executor.GetCollector())
		if err := rep.ExportJSON(jsonPath); err != nil {
			log.Printf("Failed to export JSON for user %d: %v", rt.idx, err)
		}
		if err := rep.ExportCSV(csvPath); err != nil {
			log.Printf("Failed to export CSV for user %d: %v", rt.idx, err)
		}
	}

	// Collect final Prometheus metrics and generate charts if enabled
	if *enableMetrics && baselineSnapshot != nil {
		log.Println("Collecting final Prometheus metrics snapshot...")
		finalSnapshot, err := metrics.CollectPrometheusSnapshot(context.Background(), profile, combinedCollector)
		if err != nil {
			log.Printf("Warning: Failed to collect final metrics: %v", err)
		} else {
			// Save final snapshot
			if err := metrics.SavePrometheusSnapshot(finalSnapshot, fmt.Sprintf("%s/metrics-final.json", *metricsOutputDir)); err != nil {
				log.Printf("Warning: Failed to save final snapshot: %v", err)
			}

			// Calculate delta
			log.Println("Calculating metrics delta...")
			deltaSnapshot, err := metrics.CalculateDelta(baselineSnapshot, finalSnapshot)
			if err != nil {
				log.Printf("Warning: Failed to calculate delta: %v", err)
			} else {
				// Save delta snapshot
				if err := metrics.SavePrometheusSnapshot(deltaSnapshot, fmt.Sprintf("%s/metrics-delta.json", *metricsOutputDir)); err != nil {
					log.Printf("Warning: Failed to save delta snapshot: %v", err)
				}

				// Parse formats
				formatStrings := strings.Split(*metricsFormats, ",")
				formats := make([]charts.ChartFormat, 0, len(formatStrings))
				for _, f := range formatStrings {
					formats = append(formats, charts.ChartFormat(strings.TrimSpace(f)))
				}

				// Generate charts
				log.Println("Generating charts...")
				if err := charts.GenerateAllCharts(deltaSnapshot, *metricsOutputDir, formats); err != nil {
					log.Printf("Warning: Failed to generate charts: %v", err)
				} else {
					chartsDir := fmt.Sprintf("%s/charts", *metricsOutputDir)

					// Generate PDF report if requested
					for _, format := range formats {
						if format == charts.FormatPDF {
							log.Println("Generating PDF report...")
							if err := charts.GeneratePDFReport(chartsDir, fmt.Sprintf("%s/report.pdf", *metricsOutputDir)); err != nil {
								log.Printf("Warning: Failed to generate PDF report: %v", err)
							}
						}
						if format == charts.FormatHTML {
							log.Println("Generating HTML report...")
							if err := charts.GenerateHTMLReport(deltaSnapshot, chartsDir, fmt.Sprintf("%s/report.html", *metricsOutputDir)); err != nil {
								log.Printf("Warning: Failed to generate HTML report: %v", err)
							}
						}
						if format == charts.FormatJSON {
							log.Println("Generating JSON report...")
							if err := charts.GenerateJSONReport(deltaSnapshot, fmt.Sprintf("%s/metrics-report.json", *metricsOutputDir)); err != nil {
								log.Printf("Warning: Failed to generate JSON report: %v", err)
							}
						}
					}

					log.Printf("Metrics collection complete. Results saved to: %s", *metricsOutputDir)
				}
			}
		}
	}

	log.Println("Test complete!")

}

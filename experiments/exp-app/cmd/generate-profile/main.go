package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/network"
)

func main() {
	// Parse flags
	deployDir := flag.String("deploy-dir", "", "Path to deployment directory (experiments/deploy)")
	minikubeIP := flag.String("minikube-ip", "127.0.0.1", "Minikube IP address")
	outputFile := flag.String("output", "network-profile.json", "Output profile file")
	verbose := flag.Bool("verbose", false, "Verbose output")

	flag.Parse()

	if *deployDir == "" {
		fmt.Fprintf(os.Stderr, "Error: --deploy-dir is required\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Verify deploy directory exists
	if _, err := os.Stat(*deployDir); os.IsNotExist(err) {
		log.Fatalf("deployment directory not found: %s", *deployDir)
	}

	if *verbose {
		log.Printf("Generating network profile from: %s", *deployDir)
		log.Printf("Using Minikube IP: %s", *minikubeIP)
	}

	// Create generator
	gen := network.NewGenerator(*deployDir, *minikubeIP)

	// Generate profile
	profile, err := gen.Generate()
	if err != nil {
		log.Fatalf("failed to generate profile: %v", err)
	}

	if *verbose {
		log.Printf("Generated profile:")
		log.Printf("  Channel: %s", profile.Network.ChannelName)
		log.Printf("  Chaincode: %s@%s", profile.Chaincode.Name, profile.Chaincode.Version)
		log.Printf("  Organizations: %d", len(profile.Peers))
		log.Printf("  Orderers: %d", len(profile.Orderers))
		log.Printf("  SICAR: %v", profile.SICAR.Enabled)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(*outputFile)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	// Save profile
	if err := profile.SaveJSON(*outputFile); err != nil {
		log.Fatalf("failed to save profile: %v", err)
	}

	log.Printf("Network profile saved to: %s", *outputFile)
}

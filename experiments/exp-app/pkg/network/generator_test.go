package network

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratorWithDeploymentArtifacts(t *testing.T) {
	// Use actual deployment directory
	deployDir := filepath.Join(os.Getenv("HOME"), "prj", "dtr", "impl", "experiments", "deploy")
	if _, err := os.Stat(deployDir); os.IsNotExist(err) {
		t.Skipf("deployment directory not found: %s", deployDir)
	}

	gen := NewGenerator(deployDir, "127.0.0.1", "", false, "fabric-experiments")
	profile, err := gen.Generate()
	if err != nil {
		t.Fatalf("failed to generate profile: %v", err)
	}

	// Verify basic structure
	if profile.Network.ChannelName != "carbon" {
		t.Errorf("expected channel name 'carbon', got %q", profile.Network.ChannelName)
	}

	// Verify organizations were loaded
	if _, ok := profile.Peers["mma"]; !ok {
		t.Error("expected 'mma' organization in peers")
	}
	if _, ok := profile.Peers["farmers"]; !ok {
		t.Error("expected 'farmers' organization in peers")
	}
	if _, ok := profile.Peers["companies"]; !ok {
		t.Error("expected 'companies' organization in peers")
	}

	// Verify orderers were loaded
	if len(profile.Orderers) == 0 {
		t.Error("expected at least one orderer config")
	}

	// Verify SICAR configuration
	if !profile.SICAR.Enabled {
		t.Error("expected SICAR to be enabled")
	}
	if profile.SICAR.Certificate == "" {
		t.Error("expected SICAR certificate path to be set")
	}
	if profile.SICAR.Endpoint == "" {
		t.Error("expected SICAR endpoint to be set")
	}

	// Verify Data API configuration
	if !profile.DataAPI.Enabled {
		t.Error("expected Data API to be enabled")
	}
	if profile.DataAPI.Port != 30443 {
		t.Errorf("expected Data API port 30443, got %d", profile.DataAPI.Port)
	}

	// Verify peer configurations have certificates
	for orgName, peerCfg := range profile.Peers {
		if len(peerCfg.Peers) == 0 {
			t.Errorf("expected at least one peer for %s organization", orgName)
		}
		if peerCfg.Certificates.TLSCACert == "" {
			t.Errorf("expected TLS CA cert for %s organization", orgName)
		}
		if peerCfg.Certificates.AdminCert == "" {
			t.Errorf("expected admin cert for %s organization", orgName)
		}
		if peerCfg.Certificates.AdminKey == "" {
			t.Errorf("expected admin key for %s organization", orgName)
		}
	}

	t.Logf("successfully generated profile with %d peer orgs and %d orderers",
		len(profile.Peers), len(profile.Orderers))

	// Verify chaincodes were generated
	if len(profile.Chaincodes) == 0 {
		t.Error("expected at least one chaincode configuration")
	}

	// Verify carbon chaincode configuration
	carbonCC, ok := profile.Chaincodes["carbon"]
	if !ok {
		t.Error("expected carbon chaincode configuration")
	} else {
		if carbonCC.Name != "carbon" {
			t.Errorf("expected carbon chaincode name 'carbon', got %q", carbonCC.Name)
		}
		if carbonCC.Channel != "carbon" {
			t.Errorf("expected carbon chaincode channel 'carbon', got %q", carbonCC.Channel)
		}
		if !carbonCC.MetricsEnabled {
			t.Error("expected carbon chaincode metrics to be enabled")
		}

		// Verify carbon metrics
		if carbonCC.Metrics.Name != "carbon" {
			t.Errorf("expected carbon metrics name 'carbon', got %q", carbonCC.Metrics.Name)
		}
		if carbonCC.Metrics.Port != 9443 {
			t.Errorf("expected carbon metrics port 9443, got %d", carbonCC.Metrics.Port)
		}
		if carbonCC.Metrics.NodePortBase != 30100 {
			t.Errorf("expected carbon metrics NodePortBase 30100, got %d", carbonCC.Metrics.NodePortBase)
		}

		// Verify endpoints for each organization
		for _, orgName := range []string{"mma", "farmers", "companies"} {
			if _, ok := carbonCC.Metrics.Endpoints[orgName]; !ok {
				t.Errorf("expected external endpoint for %s organization in carbon metrics", orgName)
			}
			if _, ok := carbonCC.Metrics.InternalEndpoints[orgName]; !ok {
				t.Errorf("expected internal endpoint for %s organization in carbon metrics", orgName)
			}
		}

		t.Logf("carbon metrics endpoints: %v", carbonCC.Metrics.Endpoints)
		t.Logf("carbon metrics internal endpoints: %v", carbonCC.Metrics.InternalEndpoints)
	}

	// Verify interop chaincode configuration
	interopCC, ok := profile.Chaincodes["interop"]
	if !ok {
		t.Error("expected interop chaincode configuration")
	} else {
		if interopCC.Name != "interop" {
			t.Errorf("expected interop chaincode name 'interop', got %q", interopCC.Name)
		}
		if interopCC.Metrics.NodePortBase != 30200 {
			t.Errorf("expected interop metrics NodePortBase 30200, got %d", interopCC.Metrics.NodePortBase)
		}
	}
}

func TestGeneratorSaveLoadJSON(t *testing.T) {
	deployDir := filepath.Join(os.Getenv("HOME"), "prj", "dtr", "impl", "experiments", "deploy")
	if _, err := os.Stat(deployDir); os.IsNotExist(err) {
		t.Skipf("deployment directory not found: %s", deployDir)
	}

	// Generate profile
	gen := NewGenerator(deployDir, "127.0.0.1", "", false, "fabric-experiments")
	profile, err := gen.Generate()
	if err != nil {
		t.Fatalf("failed to generate profile: %v", err)
	}

	// Save to temp file
	tmpFile := filepath.Join(t.TempDir(), "network-profile.json")
	if err := profile.SaveJSON(tmpFile); err != nil {
		t.Fatalf("failed to save profile: %v", err)
	}

	// Load from file
	loaded, err := LoadJSON(tmpFile)
	if err != nil {
		t.Fatalf("failed to load profile: %v", err)
	}

	// Verify loaded profile matches original
	if loaded.Network.ChannelName != profile.Network.ChannelName {
		t.Error("loaded profile channel name mismatch")
	}
	if len(loaded.Chaincodes) != len(profile.Chaincodes) {
		t.Error("loaded profile chaincodes count mismatch")
	}
	if len(loaded.Peers) != len(profile.Peers) {
		t.Error("loaded profile peers count mismatch")
	}
	if len(loaded.Orderers) != len(profile.Orderers) {
		t.Error("loaded profile orderers count mismatch")
	}
}

func TestGeneratorInClusterMode(t *testing.T) {
	deployDir := filepath.Join(os.Getenv("HOME"), "prj", "dtr", "impl", "experiments", "deploy")
	if _, err := os.Stat(deployDir); os.IsNotExist(err) {
		t.Skipf("deployment directory not found: %s", deployDir)
	}

	// Generate profile in in-cluster mode
	gen := NewGenerator(deployDir, "127.0.0.1", "", true, "fabric-experiments")
	profile, err := gen.Generate()
	if err != nil {
		t.Fatalf("failed to generate profile: %v", err)
	}

	// Verify peers use Kubernetes DNS addresses
	for orgName, peerCfg := range profile.Peers {
		for i, peer := range peerCfg.Peers {
			expectedServiceName := orgName + "-peer-" + string(rune('0'+i))
			expectedAddress := expectedServiceName + ".fabric-experiments.svc.cluster.local:7051"
			if peer.Address != expectedAddress {
				t.Errorf("expected peer address %q, got %q", expectedAddress, peer.Address)
			}

			// Verify metrics endpoint uses Kubernetes DNS
			expectedMetricsEndpoint := "http://" + expectedServiceName + ".fabric-experiments.svc.cluster.local:9443/metrics"
			if peer.MetricsEndpoint != expectedMetricsEndpoint {
				t.Errorf("expected metrics endpoint %q, got %q", expectedMetricsEndpoint, peer.MetricsEndpoint)
			}
		}

		// Verify certificate paths use /workspace
		if peerCfg.Certificates.TLSCACert[:10] != "/workspace" {
			t.Errorf("expected certificate path to start with /workspace, got %q", peerCfg.Certificates.TLSCACert)
		}
	}

	// Verify orderers use Kubernetes DNS addresses
	for _, orderer := range profile.Orderers {
		if orderer.Address[:len(orderer.Organization)] != orderer.Organization {
			t.Errorf("expected orderer address to start with org name, got %q", orderer.Address)
		}
		if !contains(orderer.Address, ".fabric-experiments.svc.cluster.local:") {
			t.Errorf("expected orderer address to use Kubernetes DNS, got %q", orderer.Address)
		}
	}

	// Verify data-api uses Kubernetes DNS
	expectedDataAPIAddress := "data-api.fabric-experiments.svc.cluster.local:8443"
	if profile.DataAPI.Address != expectedDataAPIAddress {
		t.Errorf("expected data-api address %q, got %q", expectedDataAPIAddress, profile.DataAPI.Address)
	}

	t.Logf("successfully generated in-cluster profile with Kubernetes DNS addresses")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

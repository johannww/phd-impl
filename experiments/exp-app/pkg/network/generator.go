package network

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Generator generates network profiles from deployment artifacts
type Generator struct {
	deployDir  string
	varsDir    string
	valuesFile string
	minikubeIP string
	teeIP      string
}

// NewGenerator creates a new network profile generator
func NewGenerator(deployDir, minikubeIP, teeIP string) *Generator {
	return &Generator{
		deployDir:  deployDir,
		varsDir:    filepath.Join(deployDir, "vars"),
		valuesFile: filepath.Join(deployDir, "helm", "values.yaml"),
		minikubeIP: minikubeIP,
		teeIP:      teeIP,
	}
}

// HelmValuesYAML represents the Helm values structure
type HelmValuesYAML struct {
	Network struct {
		ChannelName   string `yaml:"channelName"`
		Organizations []struct {
			Name         string   `yaml:"name"`
			Peers        int      `yaml:"peers"`
			Orderers     int      `yaml:"orderers"`
			NodePortBase int      `yaml:"nodePortBase"`
			Collections  []string `yaml:"collections"`
		} `yaml:"organizations"`
	} `yaml:"network"`
}

// Generate creates a network profile from deployment artifacts
func (g *Generator) Generate() (*NetworkProfile, error) {
	profile := NewNetworkProfile()

	// Load values.yaml
	values, err := g.loadValues()
	if err != nil {
		return nil, fmt.Errorf("load values: %w", err)
	}

	// Set basic network info
	profile.Network.ChannelName = values.Network.ChannelName
	profile.Network.Version = "1.0"
	profile.Network.CreatedAt = time.Now().Format(time.RFC3339)
	profile.Chaincode.Channel = values.Network.ChannelName
	profile.Chaincode.Name = "carbon"
	profile.Chaincode.Version = "1.0"

	// Process organizations
	for _, org := range values.Network.Organizations {
		if org.Peers > 0 {
			peerCfg, err := g.generatePeerConfig(org.Name, org.Peers, org.NodePortBase, org.Collections)
			if err != nil {
				return nil, fmt.Errorf("generate peer config for %s: %w", org.Name, err)
			}
			profile.Peers[org.Name] = *peerCfg
		}

		if org.Orderers > 0 {
			orderers, err := g.generateOrdererConfigs(org.Name, org.Orderers, org.NodePortBase)
			if err != nil {
				return nil, fmt.Errorf("generate orderer config for %s: %w", org.Name, err)
			}
			profile.Orderers = append(profile.Orderers, orderers...)
		}
	}

	// Configure Data API (SICAR)
	profile.DataAPI = g.generateDataAPIConfig()

	// Configure SICAR
	profile.SICAR = g.generateSICARConfig()

	// Configure TEE Auction container
	profile.TEEAuction = g.generateTEEAuctionConfig()

	return profile, nil
}

// loadValues loads and parses the Helm values.yaml
func (g *Generator) loadValues() (*HelmValuesYAML, error) {
	data, err := os.ReadFile(g.valuesFile)
	if err != nil {
		return nil, fmt.Errorf("read values.yaml: %w", err)
	}

	var values HelmValuesYAML
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("parse values.yaml: %w", err)
	}

	return &values, nil
}

// generatePeerConfig generates configuration for a peer organization
func (g *Generator) generatePeerConfig(name string, peerCount, nodePortBase int, collections []string) (*PeerConfig, error) {
	cfg := &PeerConfig{
		Organization: name,
		MspID:        fmt.Sprintf("%sMSP", capitalizeFirst(name)),
		Peers:        make([]PeerNode, peerCount),
		Collections:  collections,
	}

	// Generate peer nodes
	for i := 0; i < peerCount; i++ {
		peerName := fmt.Sprintf("peer%d.%s", i, name)
		externalPort := nodePortBase + i
		cfg.Peers[i] = PeerNode{
			Name:         peerName,
			Address:      fmt.Sprintf("%s:%d", g.minikubeIP, externalPort),
			PortInternal: 7051,
			PortExternal: externalPort,
		}
	}

	// Load certificates
	orgPath := filepath.Join(g.varsDir, "organizations", "peerOrganizations", name)
	certs := PeerCerts{
		TLSCACert: filepath.Join(orgPath, "peers", fmt.Sprintf("peer0.%s", name), "tls", "ca.crt"),
		AdminCert: filepath.Join(orgPath, "users", fmt.Sprintf("Admin@%s", name), "msp", "signcerts", fmt.Sprintf("Admin@%s-cert.pem", name)),
		AdminKey:  filepath.Join(orgPath, "users", fmt.Sprintf("Admin@%s", name), "msp", "keystore", "priv_sk"),
		User1Cert: filepath.Join(orgPath, "users", fmt.Sprintf("User1@%s", name), "msp", "signcerts", fmt.Sprintf("User1@%s-cert.pem", name)),
		User1Key:  filepath.Join(orgPath, "users", fmt.Sprintf("User1@%s", name), "msp", "keystore", "priv_sk"),
	}

	// Verify certificates exist
	if err := g.verifyCertificates(&certs); err != nil {
		return nil, fmt.Errorf("verify certificates: %w", err)
	}

	cfg.Certificates = certs
	return cfg, nil
}

// generateOrdererConfigs generates configuration for orderers
func (g *Generator) generateOrdererConfigs(orgName string, ordererCount, nodePortBase int) ([]OrdererConfig, error) {
	orderers := make([]OrdererConfig, ordererCount)

	for i := 0; i < ordererCount; i++ {
		ordererName := fmt.Sprintf("orderer%d.%s", i, orgName)
		externalPort := nodePortBase + i

		orderers[i] = OrdererConfig{
			Organization: orgName,
			MspID:        "OrdererMSP",
			Name:         ordererName,
			Address:      fmt.Sprintf("%s:%d", g.minikubeIP, externalPort),
			PortInternal: 7050,
			PortExternal: externalPort,
			TLSCACert:    filepath.Join(g.varsDir, "organizations", "ordererOrganizations", orgName, "orderers", ordererName, "tls", "ca.crt"),
		}
	}

	return orderers, nil
}

// generateDataAPIConfig generates Data API (SICAR) configuration
func (g *Generator) generateDataAPIConfig() DataAPIConfig {
	return DataAPIConfig{
		Enabled:  true,
		Hostname: g.minikubeIP,
		Port:     30443,
		Address:  fmt.Sprintf("%s:30443", g.minikubeIP),
		TLSCert:  filepath.Join(g.varsDir, "sicar", "server.crt"),
		TLSKey:   filepath.Join(g.varsDir, "sicar", "server.key"),
	}
}

// generateSICARConfig generates SICAR configuration
func (g *Generator) generateSICARConfig() SICARConfig {
	return SICARConfig{
		Enabled:     true,
		Certificate: filepath.Join(g.varsDir, "sicar", "server.crt"),
		PrivateKey:  filepath.Join(g.varsDir, "sicar", "server.key"),
		Endpoint:    fmt.Sprintf("%s:30443", g.minikubeIP),
		DataPath:    filepath.Join(g.varsDir, "organizations", "sicar", "sicar.json"),
	}
}

// generateTEEAuctionConfig generates TEE auction container configuration.
// When teeIP is empty the service is marked as disabled.
func (g *Generator) generateTEEAuctionConfig() TEEAuctionConfig {
	if g.teeIP == "" {
		return TEEAuctionConfig{Enabled: false}
	}
	const teePort = 8080
	return TEEAuctionConfig{
		Enabled: true,
		IP:      g.teeIP,
		Port:    teePort,
		Address: fmt.Sprintf("%s:%d", g.teeIP, teePort),
	}
}

// verifyCertificates checks if certificate files exist
func (g *Generator) verifyCertificates(certs *PeerCerts) error {
	paths := []string{
		certs.TLSCACert,
		certs.AdminCert,
		certs.AdminKey,
		certs.User1Cert,
		certs.User1Key,
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("certificate not found: %s", path)
		}
	}

	return nil
}

// SaveJSON saves the profile as JSON file
func (p *NetworkProfile) SaveJSON(filepath string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// LoadJSON loads a profile from JSON file
func LoadJSON(filepath string) (*NetworkProfile, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var profile NetworkProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}

	return &profile, nil
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

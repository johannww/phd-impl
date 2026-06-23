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
	inCluster  bool
	namespace  string
}

// NewGenerator creates a new network profile generator
func NewGenerator(deployDir, minikubeIP, teeIP string, inCluster bool, namespace string) *Generator {
	return &Generator{
		deployDir:  deployDir,
		varsDir:    filepath.Join(deployDir, "vars"),
		valuesFile: filepath.Join(deployDir, "helm", "values.yaml"),
		minikubeIP: minikubeIP,
		teeIP:      teeIP,
		inCluster:  inCluster,
		namespace:  namespace,
	}
}

// HelmValuesYAML represents the Helm values structure
type HelmValuesYAML struct {
	Network struct {
		ChannelName   string `yaml:"channelName"`
		UserCount     int    `yaml:"userCount"`
		Organizations []struct {
			Name         string   `yaml:"name"`
			Peers        int      `yaml:"peers"`
			Orderers     int      `yaml:"orderers"`
			NodePortBase int      `yaml:"nodePortBase"`
			Collections  []string `yaml:"collections"`
		} `yaml:"organizations"`
	} `yaml:"network"`
	ChaincodeService struct {
		Enabled    bool `yaml:"enabled"`
		Chaincodes []struct {
			Name                string `yaml:"name"`
			MetricsPort         int    `yaml:"metricsPort"`
			MetricsNodePortBase int    `yaml:"metricsNodePortBase"`
		} `yaml:"chaincodes"`
	} `yaml:"chaincodeService"`
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

	// Process organizations
	userCount := values.Network.UserCount
	if userCount == 0 {
		userCount = 1 // Default to 1 user if not specified
	}

	for _, org := range values.Network.Organizations {
		if org.Peers > 0 {
			peerCfg, err := g.generatePeerConfig(org.Name, org.Peers, org.NodePortBase, org.Collections, userCount)
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

	// Configure Chaincode Metrics
	if err := g.generateChaincodeMetrics(profile, values); err != nil {
		return nil, fmt.Errorf("generate chaincode metrics: %w", err)
	}

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
func (g *Generator) generatePeerConfig(name string, peerCount, nodePortBase int, collections []string, userCount int) (*PeerConfig, error) {
	cfg := &PeerConfig{
		Organization: name,
		MspID:        fmt.Sprintf("%sMSP", capitalizeFirst(name)),
		Peers:        make([]PeerNode, peerCount),
		Collections:  collections,
	}

	// Generate peer nodes
	for i := 0; i < peerCount; i++ {
		peerName := fmt.Sprintf("peer%d.%s", i, name)
		serviceName := fmt.Sprintf("%s-peer-%d", name, i)
		externalPort := nodePortBase + i
		metricsNodePort := nodePortBase + i + 1000 // Metrics NodePort = base + peer index + 1000

		cfg.Peers[i] = PeerNode{
			Name:            peerName,
			Address:         g.buildServiceAddress(serviceName, 7051, externalPort),
			PortInternal:    7051,
			PortExternal:    externalPort,
			MetricsPort:     9443, // Standard Fabric peer metrics port
			MetricsNodePort: metricsNodePort,
			MetricsEndpoint: g.buildMetricsEndpoint(serviceName, 9443, metricsNodePort),
		}
	}

	// Load certificates
	certBasePath := g.getCertBasePath()
	orgPath := filepath.Join(certBasePath, "organizations", "peerOrganizations", name)

	// Generate user certificates (User1, User2, ..., UserN)
	users := make([]UserCert, userCount)
	for i := 0; i < userCount; i++ {
		userNum := i + 1
		userName := fmt.Sprintf("User%d@%s", userNum, name)
		users[i] = UserCert{
			Cert: filepath.Join(orgPath, "users", userName, "msp", "signcerts", fmt.Sprintf("%s-cert.pem", userName)),
			Key:  filepath.Join(orgPath, "users", userName, "msp", "keystore", "priv_sk"),
		}
	}

	certs := PeerCerts{
		TLSCACert: filepath.Join(orgPath, "peers", fmt.Sprintf("peer0.%s", name), "tls", "ca.crt"),
		AdminCert: filepath.Join(orgPath, "users", fmt.Sprintf("Admin@%s", name), "msp", "signcerts", fmt.Sprintf("Admin@%s-cert.pem", name)),
		AdminKey:  filepath.Join(orgPath, "users", fmt.Sprintf("Admin@%s", name), "msp", "keystore", "priv_sk"),
		Users:     users,
	}

	// Verify certificates exist (skip in-cluster mode as certs are in PVC)
	if !g.inCluster {
		if err := g.verifyCertificates(&certs); err != nil {
			return nil, fmt.Errorf("verify certificates: %w", err)
		}
	}

	cfg.Certificates = certs
	return cfg, nil
}

// generateOrdererConfigs generates configuration for orderers
func (g *Generator) generateOrdererConfigs(orgName string, ordererCount, nodePortBase int) ([]OrdererConfig, error) {
	orderers := make([]OrdererConfig, ordererCount)
	certBasePath := g.getCertBasePath()

	for i := 0; i < ordererCount; i++ {
		ordererName := fmt.Sprintf("orderer%d.%s", i, orgName)
		serviceName := fmt.Sprintf("%s-orderer-%d", orgName, i)
		externalPort := nodePortBase + i
		metricsNodePort := nodePortBase + i + 1000 // Metrics NodePort = base + orderer index + 1000

		orderers[i] = OrdererConfig{
			Organization:    orgName,
			MspID:           "OrdererMSP",
			Name:            ordererName,
			Address:         g.buildServiceAddress(serviceName, 7050, externalPort),
			PortInternal:    7050,
			PortExternal:    externalPort,
			TLSCACert:       filepath.Join(certBasePath, "organizations", "ordererOrganizations", orgName, "orderers", ordererName, "tls", "ca.crt"),
			MetricsPort:     8443, // Standard Fabric orderer metrics port
			MetricsNodePort: metricsNodePort,
			MetricsEndpoint: g.buildMetricsEndpoint(serviceName, 8443, metricsNodePort),
		}
	}

	return orderers, nil
}

// generateDataAPIConfig generates Data API (SICAR) configuration
func (g *Generator) generateDataAPIConfig() DataAPIConfig {
	certBasePath := g.getCertBasePath()
	serviceName := "data-api"
	nodePort := 30443

	return DataAPIConfig{
		Enabled:  true,
		Hostname: g.minikubeIP,
		Port:     nodePort,
		Address:  g.buildServiceAddress(serviceName, 8443, nodePort),
		TLSCert:  filepath.Join(certBasePath, "sicar", "server.crt"),
		TLSKey:   filepath.Join(certBasePath, "sicar", "server.key"),
	}
}

// generateSICARConfig generates SICAR configuration
func (g *Generator) generateSICARConfig() SICARConfig {
	certBasePath := g.getCertBasePath()
	serviceName := "data-api"
	nodePort := 30443

	return SICARConfig{
		Enabled:     true,
		Certificate: filepath.Join(certBasePath, "sicar", "server.crt"),
		PrivateKey:  filepath.Join(certBasePath, "sicar", "server.key"),
		Endpoint:    g.buildServiceAddress(serviceName, 8443, nodePort),
		DataPath:    filepath.Join(certBasePath, "organizations", "sicar", "sicar.json"),
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
	}

	// Add all user certificates and keys
	for _, user := range certs.Users {
		paths = append(paths, user.Cert, user.Key)
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("certificate not found: %s", path)
		}
	}

	return nil
}

// generateChaincodeMetrics configures Prometheus metrics endpoints for chaincodes
func (g *Generator) generateChaincodeMetrics(profile *NetworkProfile, values *HelmValuesYAML) error {
	// Check if chaincode configurations exist
	if len(values.ChaincodeService.Chaincodes) == 0 {
		return nil
	}

	// Get peer organizations (organizations with peers > 0)
	peerOrgs := make([]string, 0)
	orgIndexMap := make(map[string]int) // map org name to index for NodePort calculation
	for idx, org := range values.Network.Organizations {
		if org.Peers > 0 {
			peerOrgs = append(peerOrgs, org.Name)
			orgIndexMap[org.Name] = idx
		}
	}

	// Generate configuration for each chaincode
	for _, cc := range values.ChaincodeService.Chaincodes {
		if cc.MetricsPort == 0 {
			cc.MetricsPort = 9443 // Default metrics port
		}

		metricsConfig := MetricsConfig{
			Name:              cc.Name,
			Port:              cc.MetricsPort,
			NodePortBase:      cc.MetricsNodePortBase,
			Endpoints:         make(map[string]string),
			InternalEndpoints: make(map[string]string),
		}

		// Generate endpoints for each peer organization
		for _, orgName := range peerOrgs {
			orgIdx := orgIndexMap[orgName]
			serviceName := fmt.Sprintf("%s-%s", cc.Name, orgName)

			// Internal cluster endpoint (always populated)
			internalEndpoint := fmt.Sprintf("http://%s:%d/metrics", serviceName, cc.MetricsPort)
			metricsConfig.InternalEndpoints[orgName] = internalEndpoint

			// External endpoint - use helper for deployment mode awareness
			if cc.MetricsNodePortBase > 0 {
				nodePort := cc.MetricsNodePortBase + orgIdx
				externalEndpoint := g.buildMetricsEndpoint(serviceName, cc.MetricsPort, nodePort)
				metricsConfig.Endpoints[orgName] = externalEndpoint
			}
		}

		// Create chaincode config
		chaincodeConfig := ChaincodeConfig{
			Channel:        values.Network.ChannelName,
			Name:           cc.Name,
			Version:        "1.0", // TODO: Get version from values if available
			MetricsEnabled: cc.MetricsNodePortBase > 0 || len(metricsConfig.InternalEndpoints) > 0,
			Metrics:        metricsConfig,
		}

		profile.Chaincodes[cc.Name] = chaincodeConfig
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

// buildServiceAddress builds the appropriate service address based on deployment mode
// For in-cluster: {service}.{namespace}.svc.cluster.local:{port}
// For external: {minikubeIP}:{nodePort}
func (g *Generator) buildServiceAddress(serviceName string, internalPort, nodePort int) string {
	if g.inCluster {
		return fmt.Sprintf("%s.%s.svc.cluster.local:%d", serviceName, g.namespace, internalPort)
	}
	return fmt.Sprintf("%s:%d", g.minikubeIP, nodePort)
}

// buildMetricsEndpoint builds the Prometheus metrics endpoint based on deployment mode
func (g *Generator) buildMetricsEndpoint(serviceName string, metricsPort, metricsNodePort int) string {
	if g.inCluster {
		return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d/metrics", serviceName, g.namespace, metricsPort)
	}
	return fmt.Sprintf("http://%s:%d/metrics", g.minikubeIP, metricsNodePort)
}

// getCertBasePath returns the base path for certificates
// For in-cluster: /workspace (mounted from organizations PVC)
// For external: local vars directory
func (g *Generator) getCertBasePath() string {
	if g.inCluster {
		return "/workspace"
	}
	return g.varsDir
}

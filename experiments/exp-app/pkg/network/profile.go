package network

// NetworkProfile contains all network configuration for exp-app
type NetworkProfile struct {
	Network    NetworkConfig         `json:"network"`
	Peers      map[string]PeerConfig `json:"peers"` // key: org_name, value: peer config
	Orderers   []OrdererConfig       `json:"orderers"`
	DataAPI    DataAPIConfig         `json:"data_api"`
	SICAR      SICARConfig           `json:"sicar"`
	Chaincode  ChaincodeConfig       `json:"chaincode"`
	TEEAuction TEEAuctionConfig      `json:"tee_auction"`
}

// NetworkConfig basic network information
type NetworkConfig struct {
	ChannelName string `json:"channel_name"`
	Version     string `json:"version"`
	CreatedAt   string `json:"created_at"`
}

// PeerConfig represents a peer organization
type PeerConfig struct {
	Organization string     `json:"organization"`
	MspID        string     `json:"msp_id"`
	Peers        []PeerNode `json:"peers"`
	Certificates PeerCerts  `json:"certificates"`
	Collections  []string   `json:"collections"`
}

// PeerNode represents a single peer
type PeerNode struct {
	Name         string `json:"name"`          // e.g., "peer0.mma"
	Address      string `json:"address"`       // e.g., "localhost:7051"
	PortInternal int    `json:"port_internal"` // Internal port
	PortExternal int    `json:"port_external"` // NodePort
}

// PeerCerts contains paths to peer certificates
type PeerCerts struct {
	TLSCACert    string `json:"tls_ca_cert"`    // Path to TLS CA cert
	AdminCert    string `json:"admin_cert"`     // Path to admin certificate
	AdminKey     string `json:"admin_key"`      // Path to admin private key
	User1Cert    string `json:"user1_cert"`     // Path to User1 certificate
	User1Key     string `json:"user1_key"`      // Path to User1 private key
	AdminTLSCert string `json:"admin_tls_cert"` // Path to admin TLS cert
	AdminTLSKey  string `json:"admin_tls_key"`  // Path to admin TLS key
}

// OrdererConfig represents an orderer
type OrdererConfig struct {
	Organization string `json:"organization"`
	MspID        string `json:"msp_id"`
	Name         string `json:"name"` // e.g., "orderer0.orderer"
	Address      string `json:"address"`
	PortInternal int    `json:"port_internal"`
	PortExternal int    `json:"port_external"`
	TLSCACert    string `json:"tls_ca_cert"`
}

// DataAPIConfig represents the SICAR data API
type DataAPIConfig struct {
	Enabled  bool   `json:"enabled"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	Address  string `json:"address"` // full address like "localhost:30443"
	TLSCert  string `json:"tls_cert"`
	TLSKey   string `json:"tls_key"`
}

// SICARConfig represents SICAR certificate registration
type SICARConfig struct {
	Enabled     bool   `json:"enabled"`
	Certificate string `json:"certificate"` // Path to SICAR cert
	PrivateKey  string `json:"private_key"` // Path to SICAR private key
	Endpoint    string `json:"endpoint"`    // e.g., "localhost:30443"
	DataPath    string `json:"data_path"`   // Path to sicar.json
}

// ChaincodeConfig contains chaincode information
type ChaincodeConfig struct {
	Channel string `json:"channel"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// TEEAuctionConfig holds the address of the confidential container running the TEE auction service
type TEEAuctionConfig struct {
	Enabled bool   `json:"enabled"`
	IP      string `json:"ip"`      // Public IP assigned by Azure Container Instances
	Port    int    `json:"port"`    // HTTPS port (default 8080)
	Address string `json:"address"` // Full address, e.g. "1.2.3.4:8080"
}

// NewNetworkProfile creates a new empty network profile
func NewNetworkProfile() *NetworkProfile {
	return &NetworkProfile{
		Network:    NetworkConfig{},
		Peers:      make(map[string]PeerConfig),
		Orderers:   make([]OrdererConfig, 0),
		DataAPI:    DataAPIConfig{},
		SICAR:      SICARConfig{},
		Chaincode:  ChaincodeConfig{},
		TEEAuction: TEEAuctionConfig{},
	}
}

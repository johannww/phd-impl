package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/network"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/tee"
)

// TEESetupManager handles TEE service initialization
type TEESetupManager struct {
	client  *gateway.ClientWrapper
	profile *network.NetworkProfile
}

// NewTEESetupManager creates a new TEE setup manager
func NewTEESetupManager(client *gateway.ClientWrapper, profile *network.NetworkProfile) *TEESetupManager {
	return &TEESetupManager{
		client:  client,
		profile: profile,
	}
}

// SetupTEE initializes the TEE service and publishes policies and reports to the chaincode
func (m *TEESetupManager) SetupTEE(ctx context.Context) (*tee.Client, error) {
	if !m.profile.TEEAuction.Enabled {
		log.Println("TEE auction service is disabled")
		return nil, nil
	}

	log.Printf("TEE auction service enabled at: %s", m.profile.TEEAuction.Address)

	// Step 1: Create TEE client
	teeClient := tee.NewClient(
		fmt.Sprintf("https://%s", m.profile.TEEAuction.Address),
		true, // Skip TLS verification (self-signed cert)
	)

	// Step 2: Test connection
	if err := teeClient.Ping(); err != nil {
		log.Printf("WARNING: Failed to ping TEE service: %v", err)
		log.Println("TEE setup aborted, will fall back to mock results")
		return nil, fmt.Errorf("TEE service unreachable: %w", err)
	}
	log.Println("TEE service is reachable")

	// Step 3: Fetch TEE attestation report
	log.Println("Fetching TEE attestation report...")
	reportJSONBytes, err := teeClient.GetReport()

	// Step 4: Publish expected CCE policy
	log.Println("Publishing expected TEE CCE policy...")
	ccePolicy, err := m.getCCEPolicy()
	if err != nil {
		log.Printf("WARNING: Failed to load CCE policy: %v", err)
		return teeClient, fmt.Errorf("failed to load CCE policy: %w", err)
	}
	if err := m.publishCCEPolicy(ctx, ccePolicy); err != nil {
		log.Printf("WARNING: Failed to publish CCE policy: %v", err)
		return teeClient, fmt.Errorf("failed to publish CCE policy: %w", err)
	}
	log.Println("CCE policy published successfully")

	// Step 5: Publish initial TEE report
	log.Println("Publishing initial TEE report...")
	if err := m.publishTEEReport(ctx, reportJSONBytes); err != nil {
		log.Printf("WARNING: Failed to publish TEE report: %v", err)
		return teeClient, fmt.Errorf("failed to publish TEE report: %w", err)
	}
	log.Println("TEE report published successfully")

	log.Println("TEE setup completed successfully")
	return teeClient, nil
}

// getCCEPolicy returns the expected CCE policy from the ARM template
func (m *TEESetupManager) getCCEPolicy() (string, error) {
	// Read the ARM template JSON file
	data, err := os.ReadFile(m.armTemplatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read ARM template from %s: %w", m.armTemplatePath, err)
	}

	// Parse the JSON to extract ccePolicy
	var template struct {
		Resources []struct {
			Properties struct {
				ConfidentialComputeProperties struct {
					CCEPolicy string `json:"ccePolicy"`
				} `json:"confidentialComputeProperties"`
			} `json:"properties"`
		} `json:"resources"`
	}

	if err := json.Unmarshal(data, &template); err != nil {
		return "", fmt.Errorf("failed to parse ARM template JSON: %w", err)
	}

	if len(template.Resources) == 0 {
		return "", fmt.Errorf("no resources found in ARM template")
	}

	ccePolicy := template.Resources[0].Properties.ConfidentialComputeProperties.CCEPolicy
	if ccePolicy == "" {
		return "", fmt.Errorf("ccePolicy not found in ARM template")
	}

	return ccePolicy, nil
}

// publishCCEPolicy publishes the expected CCE policy to the chaincode
func (m *TEESetupManager) publishCCEPolicy(ctx context.Context, base64Policy string) error {
	_, err := m.client.SubmitTransaction("PublishExpectedTEECCEPolicy", base64Policy)
	return err
}

// publishTEEReport publishes the initial TEE report to the chaincode
func (m *TEESetupManager) publishTEEReport(ctx context.Context, reportJSON []byte) error {
	// Validate that it's valid JSON
	var testJSON map[string]interface{}
	if err := json.Unmarshal(reportJSON, &testJSON); err != nil {
		return fmt.Errorf("invalid JSON in TEE report: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := m.client.SubmitTransactionWithContext(ctx, "PublishInitialTEEReport", string(reportJSON))
	return err
}

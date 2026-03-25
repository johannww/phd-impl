package sicar

import (
	"fmt"
	"os"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/network"
)

// Bootstrapper handles SICAR certificate registration
type Bootstrapper struct {
	client  *gateway.ClientWrapper
	profile *network.NetworkProfile
}

// NewBootstrapper creates a new SICAR bootstrapper
func NewBootstrapper(client *gateway.ClientWrapper, profile *network.NetworkProfile) *Bootstrapper {
	return &Bootstrapper{
		client:  client,
		profile: profile,
	}
}

// RegisterSICARCertificate registers the SICAR certificate as a trusted data provider
func (b *Bootstrapper) RegisterSICARCertificate() error {
	if !b.profile.SICAR.Enabled {
		return fmt.Errorf("SICAR is not enabled in network profile")
	}

	// Read SICAR certificate
	certPEM, err := os.ReadFile(b.profile.SICAR.Certificate)
	if err != nil {
		return fmt.Errorf("failed to read SICAR certificate: %w", err)
	}

	// Call AddTrustedProvider on carbon chaincode
	// This requires the caller to have the "trusted_database_registrator" attribute
	result, err := b.client.SubmitTransaction(
		"AddTrustedProvider",
		"SICAR",
		b.profile.DataAPI.Address,
		string(certPEM),
	)
	if err != nil {
		return fmt.Errorf("failed to register SICAR: %w", err)
	}

	if len(result) > 0 {
		fmt.Printf("SICAR registration result: %s\n", string(result))
	}

	return nil
}

// CheckSICARStatus checks if SICAR is already registered
func (b *Bootstrapper) CheckSICARStatus() (bool, error) {
	// This would query the chaincode to check if SICAR is registered
	// For now, we'll return not implemented
	return false, fmt.Errorf("CheckSICARStatus not yet implemented")
}

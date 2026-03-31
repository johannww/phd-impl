package setup

import (
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/network"
)

// SetupManager handles initialization of blockchain state before performance tests
type SetupManager struct {
	client  *gateway.ClientWrapper
	profile *network.NetworkProfile
}

// NewSetupManager creates a new setup manager
func NewSetupManager(client *gateway.ClientWrapper, profile *network.NetworkProfile) *SetupManager {
	return &SetupManager{
		client:  client,
		profile: profile,
	}
}

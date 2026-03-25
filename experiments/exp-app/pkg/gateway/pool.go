package gateway

import (
	"fmt"
	"math/rand"
	"sync/atomic"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/network"
)

// PeerPool manages multiple peer connections for load balancing
type PeerPool struct {
	clients  []*ClientWrapper
	current  atomic.Int32
	strategy LoadBalancingStrategy
	orgName  string
}

// LoadBalancingStrategy defines how to select peers
type LoadBalancingStrategy string

const (
	RoundRobin LoadBalancingStrategy = "round_robin"
	Random     LoadBalancingStrategy = "random"
)

// NewPeerPool creates a new peer pool for an organization
func NewPeerPool(
	orgName string,
	peerCfg *network.PeerConfig,
	strategy LoadBalancingStrategy,
) (*PeerPool, error) {
	if len(peerCfg.Peers) == 0 {
		return nil, fmt.Errorf("no peers configured for organization %s", orgName)
	}

	pool := &PeerPool{
		clients:  make([]*ClientWrapper, 0, len(peerCfg.Peers)),
		strategy: strategy,
		orgName:  orgName,
	}

	// Create a client for each peer
	for _, peerNode := range peerCfg.Peers {
		cfg := &GatewayConfig{
			PeerAddr:      peerNode.Address,
			TLSCertPath:   peerCfg.Certificates.TLSCACert,
			MspID:         peerCfg.MspID,
			UserCertPath:  peerCfg.Certificates.User1Cert,
			UserKeyPath:   peerCfg.Certificates.User1Key,
			ChannelName:   "", // Will be set when connecting
			ChaincodeName: "", // Will be set when connecting
		}

		client, err := NewClientWrapper(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create client for %s: %w", peerNode.Address, err)
		}

		pool.clients = append(pool.clients, client)
	}

	return pool, nil
}

// GetClient returns a peer client using the configured load balancing strategy
func (p *PeerPool) GetClient() *ClientWrapper {
	if len(p.clients) == 1 {
		return p.clients[0]
	}

	switch p.strategy {
	case RoundRobin:
		idx := p.current.Add(1) % int32(len(p.clients))
		return p.clients[idx]
	case Random:
		idx := rand.Intn(len(p.clients))
		return p.clients[idx]
	default:
		return p.clients[0]
	}
}

// Close closes all peer connections
func (p *PeerPool) Close() error {
	for _, client := range p.clients {
		if err := client.Close(); err != nil {
			return fmt.Errorf("failed to close client: %w", err)
		}
	}
	return nil
}

// GetOrganizationPools creates peer pools for multiple organizations
func GetOrganizationPools(
	profile *network.NetworkProfile,
	strategy LoadBalancingStrategy,
) (map[string]*PeerPool, error) {
	pools := make(map[string]*PeerPool)

	for orgName, peerCfg := range profile.Peers {
		pool, err := NewPeerPool(orgName, &peerCfg, strategy)
		if err != nil {
			return nil, fmt.Errorf("failed to create pool for %s: %w", orgName, err)
		}
		pools[orgName] = pool
	}

	return pools, nil
}

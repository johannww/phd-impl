package setup

import (
	"fmt"
	"sort"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/network"
)

// SetupManager handles initialization of blockchain state before performance tests
type SetupManager struct {
	client          *gateway.ClientWrapper
	profile         *network.NetworkProfile
	armTemplatePath string
}

// NewSetupManager creates a new setup manager
func NewSetupManager(client *gateway.ClientWrapper, profile *network.NetworkProfile, armTemplatePath string) *SetupManager {
	return &SetupManager{
		client:          client,
		profile:         profile,
		armTemplatePath: armTemplatePath,
	}
}

// resolveIdentitySlot computes the global identity slot index for (organization, userIndex)
// considering all peer organizations and usersPerOrg.
func (s *SetupManager) resolveIdentitySlot(usersPerOrg int, assignment IdentityAssignment) (slot int, total int, err error) {
	if usersPerOrg <= 0 {
		return -1, 0, fmt.Errorf("usersPerOrg must be > 0")
	}

	if _, ok := s.profile.Peers[assignment.Organization]; !ok {
		return -1, 0, fmt.Errorf("organization %q not found in profile", assignment.Organization)
	}

	orgNames := make([]string, 0, len(s.profile.Peers))
	for orgName := range s.profile.Peers {
		orgNames = append(orgNames, orgName)
	}
	sort.Strings(orgNames)

	slot = -1
	total = 0
	for _, orgName := range orgNames {
		orgCfg := s.profile.Peers[orgName]
		activeUsers := usersPerOrg
		if activeUsers > len(orgCfg.Certificates.Users) {
			activeUsers = len(orgCfg.Certificates.Users)
		}

		for userIdx := 0; userIdx < activeUsers; userIdx++ {
			if orgName == assignment.Organization && userIdx == assignment.UserIndex {
				slot = total
			}
			total++
		}
	}

	if slot < 0 {
		return -1, total, fmt.Errorf(
			"identity slot not found for org=%s userIndex=%d with usersPerOrg=%d",
			assignment.Organization,
			assignment.UserIndex,
			usersPerOrg,
		)
	}

	return slot, total, nil
}

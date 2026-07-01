package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/network"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/setup"
)

func main() {
	profilePath := flag.String("profile", "", "Path to network profile JSON (required)")
	defaultArmTemplatePath := strings.TrimSpace(os.Getenv("EXP_APP_ARM_TEMPLATE"))
	if defaultArmTemplatePath == "" {
		defaultArmTemplatePath = "../../tee_auction/azure/arm_template.json"
	}
	armTemplatePath := flag.String("arm-template", defaultArmTemplatePath, "Path to ARM template JSON for CCE policy")
	organizationFilter := flag.String("organization", strings.TrimSpace(os.Getenv("EXP_APP_ORGANIZATION")), "Organization to use from profile (default: first organization)")
	userIndex := flag.Int("user-index", 0, "User index to use for setup identity")

	flag.Parse()

	if *profilePath == "" {
		fmt.Fprintf(os.Stderr, "Error: --profile is required\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	profile, err := network.LoadJSON(*profilePath)
	if err != nil {
		log.Fatalf("Failed to load network profile: %v", err)
	}

	orgName, orgConfig := resolveOrganization(profile, *organizationFilter)
	if orgName == "" {
		log.Fatal("No organizations found in network profile")
	}
	if len(orgConfig.Peers) == 0 {
		log.Fatalf("No peers found in organization %s", orgName)
	}

	if *userIndex < 0 || *userIndex >= len(orgConfig.Certificates.Users) {
		log.Fatalf("user-index %d out of range for org %s (available users: %d)", *userIndex, orgName, len(orgConfig.Certificates.Users))
	}

	peerAddr := orgConfig.Peers[0].Address
	tlsCertPath := orgConfig.Certificates.TLSCACert
	userCertPath := orgConfig.Certificates.Users[*userIndex].Cert
	userKeyPath := orgConfig.Certificates.Users[*userIndex].Key
	channelName := profile.Network.ChannelName

	gatewayCfg := &gateway.GatewayConfig{
		PeerAddr:      peerAddr,
		TLSCertPath:   tlsCertPath,
		MspID:         orgConfig.MspID,
		UserCertPath:  userCertPath,
		UserKeyPath:   userKeyPath,
		ChannelName:   channelName,
		ChaincodeName: "carbon",
	}

	client, err := gateway.NewClientWrapper(gatewayCfg)
	if err != nil {
		log.Fatalf("Failed to connect to Fabric gateway: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	setupMgr := setup.NewSetupManager(client, profile, *armTemplatePath)
	log.Printf("Running global setup (org=%s, userIndex=%d, armTemplate=%s)", orgName, *userIndex, *armTemplatePath)
	teeClient, err := setupMgr.RunGlobalSetup(context.Background())
	if err != nil {
		log.Fatalf("Global setup failed: %v", err)
	}

	if teeClient != nil {
		log.Println("Global setup finished with TEE enabled")
	} else {
		log.Println("Global setup finished without TEE")
	}
}

func resolveOrganization(profile *network.NetworkProfile, organizationFilter string) (string, network.PeerConfig) {
	if organizationFilter != "" {
		selected, ok := profile.Peers[organizationFilter]
		if !ok {
			log.Fatalf("Organization %q not found in network profile", organizationFilter)
		}
		return organizationFilter, selected
	}

	orgNames := make([]string, 0, len(profile.Peers))
	for name := range profile.Peers {
		orgNames = append(orgNames, name)
	}
	sort.Strings(orgNames)
	if len(orgNames) == 0 {
		return "", network.PeerConfig{}
	}

	return orgNames[0], profile.Peers[orgNames[0]]
}

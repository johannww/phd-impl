package registry

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"google.golang.org/protobuf/proto"
)

const REGISTRY_SUMMARY_PREFIX = "registrySummary"
const ACTIVE_STATUS = "ACTIVE"
const INACTIVE_STATUS = "INACTIVE"

// RegistrySummary represents a generic environmental registry response.
// This is the canonical format used by the chaincode regardless of the data source.
type RegistrySummary struct {
	RegistryPropID  string    `json:"registryPropId"`  // e.g., SICAR property ID
	Status          string    `json:"status"`          // e.g., "ACTIVE", "INACTIVE"
	LastUpdate      time.Time `json:"lastUpdate"`      // last cadastral update date from the registry
	TotalArea       float64   `json:"totalArea"`       // in hectares
	LegalForestArea float64   `json:"legalForestArea"` // APP + Legal Reserve
	VerifiedForest  float64   `json:"verifiedForest"`  // Actual forest area
}

var _ state.WorldStateManager = (*RegistrySummary)(nil)

func (r *RegistrySummary) ToProto() proto.Message {
	lastUpdate := ""
	if !r.LastUpdate.IsZero() {
		lastUpdate = utils.UnixMillisNowFromGoTime(r.LastUpdate)
	}
	return &pb.RegistrySummary{
		RegistryPropId:  r.RegistryPropID,
		Status:          r.Status,
		LastUpdate:      lastUpdate,
		TotalArea:       r.TotalArea,
		LegalForestArea: r.LegalForestArea,
		VerifiedForest:  r.VerifiedForest,
	}
}

func (r *RegistrySummary) FromProto(m proto.Message) error {
	pr, ok := m.(*pb.RegistrySummary)
	if !ok {
		return fmt.Errorf("unexpected proto message type for RegistrySummary")
	}
	r.RegistryPropID = pr.RegistryPropId
	r.Status = pr.Status
	if pr.LastUpdate != "" {
		t, err := utils.ParseHexTimestamp(pr.LastUpdate)
		if err != nil {
			return fmt.Errorf("could not parse LastUpdate %q: %v", pr.LastUpdate, err)
		}
		r.LastUpdate = t
	}
	r.TotalArea = pr.TotalArea
	r.LegalForestArea = pr.LegalForestArea
	r.VerifiedForest = pr.VerifiedForest
	return nil
}

func (r *RegistrySummary) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetStateWithCompositeKey(stub, REGISTRY_SUMMARY_PREFIX, keyAttributes, r)
}

func (r *RegistrySummary) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutStateWithCompositeKey(stub, REGISTRY_SUMMARY_PREFIX, r.GetID(), r)
}

func (r *RegistrySummary) GetID() *[][]string {
	return &[][]string{{r.RegistryPropID}}
}

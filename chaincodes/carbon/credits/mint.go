package credits

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/data/registry"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/state"
)

const (
	MINT_CREDIT_PREFIX = "mintCredit"
)

// MintCredit represents a carbon credit that has been minted and
// it is associated to mint multiplier and mint timestamp.
type MintCredit struct {
	Credit
	MintMult      int64  `json:"mintMult"`
	MintTimeStamp string `json:"mintTimestamp"`
}

var _ state.WorldStateManager = (*MintCredit)(nil)

func (mc *MintCredit) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := state.GetStateWithCompositeKey(stub, MINT_CREDIT_PREFIX, keyAttributes, mc)
	if err != nil {
		return err
	}

	mc.Chunk = &properties.PropertyChunk{}
	if err := mc.Chunk.FromWorldState(stub, mc.ChunkID); err != nil {
		return fmt.Errorf("could not put property chunk in state: %v", err)
	}
	return nil
}

func (mc *MintCredit) ToWorldState(stub shim.ChaincodeStubInterface) error {
	copyMc := *mc      // create a copy to avoid modifying the original object
	copyMc.Chunk = nil // avoid storing the chunk in the world state, as it is already stored in the property chunk
	if err := state.PutStateWithCompositeKey(stub, MINT_CREDIT_PREFIX, copyMc.GetID(), &copyMc); err != nil {
		return fmt.Errorf("could not put mint credit in state: %v", err)
	}

	return nil
}
func (mc *MintCredit) GetID() *[][]string {
	creditId := (*mc.Credit.GetID())[0]
	creditId = append(creditId, mc.MintTimeStamp)
	return &[][]string{creditId}
}

func mintCreditInternal(
	stub shim.ChaincodeStubInterface,
	property *properties.Property,
	summary *registry.RegistrySummary,
	chunk *properties.PropertyChunk,
	quantity int64,
	timestampRFC3339 string,
) (*MintCredit, error) {
	activePolicies, err := policies.GetActivePolicies(stub)
	if err != nil {
		return nil, fmt.Errorf("could not get active policies: %v", err)
	}

	if len(activePolicies) == 0 {
		return nil, fmt.Errorf("no active policies found")
	}

	pApplier := policies.NewPolicyApplier()
	pInput := &policies.PolicyInput{
		Chunk: chunk,
	}

	mintMult, err := pApplier.MintIndependentMult(pInput, activePolicies)
	if err != nil {
		return nil, fmt.Errorf("could not get mint multiplier from active policies: %v", err)
	}

	credit := &MintCredit{
		Credit: Credit{
			OwnerID: property.OwnerID,
			ChunkID: []string{
				strconv.FormatUint(property.ID, 10),
				strconv.FormatFloat(chunk.Coordinates[0].Latitude, 'f', 6, 64),
				strconv.FormatFloat(chunk.Coordinates[0].Longitude, 'f', 6, 64),
			},
			Quantity: quantity,
		},
		MintMult:      mintMult,
		MintTimeStamp: timestampRFC3339,
	}
	if err := credit.ToWorldState(stub); err != nil {
		return nil, err
	}
	return credit, nil
}

// MintQuantityCreditForChunk mints a credit with a specified quantity and applies multipliers.
func MintQuantityCreditForChunk(
	stub shim.ChaincodeStubInterface,
	propertyID []string,
	chunkID []string,
	quantity int64,
	timestampRFC3339 string,
) (*MintCredit, error) {
	if cid.AssertAttributeValue(stub, identities.CreditMinter, "true") != nil {
		return nil, fmt.Errorf("caller is not a minter")
	}

	// Load property to get registry info
	property := &properties.Property{}
	if err := property.FromWorldState(stub, propertyID); err != nil {
		return nil, fmt.Errorf("could not get property from world state: %v", err)
	}

	var summary *registry.RegistrySummary
	if property.RegistryProvider != "" {
		summary = &registry.RegistrySummary{}
		if err := summary.FromWorldState(stub, []string{property.RegistryID}); err != nil {
			return nil, fmt.Errorf("failed to get registry summary from world state: %v", err)
		}

		if summary.Status != "Ativo" {
			return nil, fmt.Errorf("property is not active in the registry: %s", summary.Status)
		}
	}

	// Load chunk
	chunk := &properties.PropertyChunk{}
	if err := chunk.FromWorldState(stub, chunkID); err != nil {
		return nil, fmt.Errorf("could not get property chunk from world state: %v", err)
	}

	return mintCreditInternal(stub, property, summary, chunk, quantity, timestampRFC3339)
}

// MintEstimatedCreditsForProperty mints credits for all chunks of a property by estimating the quantity for each.
func MintEstimatedCreditsForProperty(
	stub shim.ChaincodeStubInterface,
	propertyID []string,
	intervalStartRFC3339 string,
	intervalEndRFC3339 string,
) ([]*MintCredit, error) {
	if cid.AssertAttributeValue(stub, identities.CreditMinter, "true") != nil {
		return nil, fmt.Errorf("caller is not a minter")
	}

	intervalStart, err := time.Parse(time.RFC3339, intervalStartRFC3339)
	if err != nil {
		return nil, fmt.Errorf("invalid interval start timestamp: %v", err)
	}
	intervalEnd, err := time.Parse(time.RFC3339, intervalEndRFC3339)
	if err != nil {
		return nil, fmt.Errorf("invalid interval end timestamp: %v", err)
	}

	// Load property (this also loads chunks)
	property := &properties.Property{}
	if err := property.FromWorldState(stub, propertyID); err != nil {
		return nil, fmt.Errorf("could not get property from world state: %v", err)
	}

	var summary *registry.RegistrySummary
	if property.RegistryProvider != "" {
		summary = &registry.RegistrySummary{}
		if err := summary.FromWorldState(stub, []string{property.RegistryID}); err != nil {
			return nil, fmt.Errorf("failed to get registry summary from world state: %v", err)
		}

		if summary.Status != "Ativo" {
			return nil, fmt.Errorf("property is not active in the registry: %s", summary.Status)
		}
	}

	estimator := &policies.Estimator{}
	mintedCredits := make([]*MintCredit, 0, len(property.Chunks))

	for _, chunk := range property.Chunks {
		// Estimate quantity for each chunk
		quantity, err := estimator.Estimate(chunk, intervalStart, intervalEnd)
		if err != nil {
			return nil, fmt.Errorf("could not estimate credit quantity for chunk: %v", err)
		}

		mc, err := mintCreditInternal(stub, property, summary, chunk, quantity, intervalEndRFC3339)
		if err != nil {
			return nil, fmt.Errorf("could not mint credit for chunk: %v", err)
		}
		mintedCredits = append(mintedCredits, mc)
	}

	return mintedCredits, nil
}

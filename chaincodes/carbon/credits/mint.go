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
	MINT_CREDIT_PREFIX        = "mintCredit"
	ZEROED_MINT_CREDIT_PREFIX = "zeroedMintCredit"
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
	// Try to get from active credits first
	err := state.GetStateWithCompositeKey(stub, MINT_CREDIT_PREFIX, keyAttributes, mc)
	if err != nil {
		// Try to get from zeroed credits
		err = state.GetStateWithCompositeKey(stub, ZEROED_MINT_CREDIT_PREFIX, keyAttributes, mc)
		if err != nil {
			return err
		}
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

	id := copyMc.GetID()

	if mc.Quantity > 0 {
		// Save to active credits
		if err := state.PutStateWithCompositeKey(stub, MINT_CREDIT_PREFIX, id, &copyMc); err != nil {
			return fmt.Errorf("could not put mint credit in state: %v", err)
		}
		// Ensure it's not in zeroed credits
		_ = state.DeleteStateWithCompositeKey(stub, ZEROED_MINT_CREDIT_PREFIX, id)
	} else {
		// Save to zeroed credits
		if err := state.PutStateWithCompositeKey(stub, ZEROED_MINT_CREDIT_PREFIX, id, &copyMc); err != nil {
			return fmt.Errorf("could not put zeroed mint credit in state: %v", err)
		}
		// Remove from active credits
		_ = state.DeleteStateWithCompositeKey(stub, MINT_CREDIT_PREFIX, id)
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

	// Apply multiplier to quantity
	effectiveQuantity := quantity + (quantity * mintMult / policies.MULTIPLIER_SCALE)

	credit := &MintCredit{
		Credit: Credit{
			OwnerID: property.OwnerID,
			ChunkID: []string{
				strconv.FormatUint(property.ID, 10),
				strconv.FormatFloat(chunk.Coordinates[0].Latitude, 'f', 6, 64),
				strconv.FormatFloat(chunk.Coordinates[0].Longitude, 'f', 6, 64),
			},
			Quantity: effectiveQuantity,
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

		if summary.Status != registry.ACTIVE_STATUS {
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

		if summary.Status != registry.ACTIVE_STATUS {
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

// GetAvailableCreditsByOwner returns all MintCredits owned by the specified owner.
func GetAvailableCreditsByOwner(stub shim.ChaincodeStubInterface, ownerID string) ([]*MintCredit, error) {
	resultsIterator, err := stub.GetStateByPartialCompositeKey(MINT_CREDIT_PREFIX, []string{ownerID})
	if err != nil {
		return nil, fmt.Errorf("could not get credits by owner: %v", err)
	}
	defer resultsIterator.Close()

	var credits []*MintCredit
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		mc := &MintCredit{}
		err = state.UnmarshalStateAs(response.Value, mc)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal mint credit: %v", err)
		}
		credits = append(credits, mc)
	}

	return credits, nil
}

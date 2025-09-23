package credits

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
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

func MintCreditForChunk(
	stub shim.ChaincodeStubInterface,
	ownerID string,
	chunkID []string,
	quantity int64,
	RFC339Timestamp string,
	mintMult int64,
) (*MintCredit, error) {
	if cid.AssertAttributeValue(stub, identities.CreditMinter, "true") != nil {
		return nil, fmt.Errorf("caller is not a minter")
	}

	credit := &MintCredit{
		Credit: Credit{
			OwnerID:  ownerID,
			ChunkID:  chunkID,
			Quantity: quantity,
		},
		MintMult:      mintMult,
		MintTimeStamp: RFC339Timestamp,
	}
	if err := credit.ToWorldState(stub); err != nil {
		return nil, err
	}
	return credit, nil
}

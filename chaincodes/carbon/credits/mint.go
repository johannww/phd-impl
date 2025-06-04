package credits

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	MINT_CREDIT_PREFIX = "mintCredit"
)

// MintCredit represents a carbon credit that has been minted and
// it is associated to mint multiplier and mint timestamp.
type MintCredit struct {
	Credit
	MintMult      float64 `json:"mintMult"`
	MintTimeStamp string  `json:"mintTimestamp"`
}

var _ state.WorldStateManager = (*MintCredit)(nil)

func (mc *MintCredit) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := state.GetStateWithCompositeKey(stub, string(MINT_CREDIT_PREFIX), keyAttributes, mc)
	if err != nil {
		return err
	}
	return nil
}

func (mc *MintCredit) ToWorldState(stub shim.ChaincodeStubInterface) error {
	if err := state.PutStateWithCompositeKey(stub, string(MINT_CREDIT_PREFIX), mc.GetID(), mc); err != nil {
		return fmt.Errorf("could not put sellbid in state: %v", err)
	}

	return nil
}
func (mc *MintCredit) GetID() *[][]string {
	creditId := (*mc.Credit.GetID())[0]
	creditId = append(creditId, mc.MintTimeStamp)
	return &[][]string{creditId}
}

// TODO: finish this
// func MintCreditForChunk(stub shim.ChaincodeStubInterface, chunkID string, mintMult float64) (*MintCredit, error) {
// 	orsnatoie
// 	credit := &MintCredit{
// 		Credit: Credit{
// 			Chunk.PropertyID
// 		},
// 		MintMult:      mintMult,
// 		MintTimeStamp: uint64(stub.GetTxTimestamp().GetSeconds()),
// 	}
// 	if err := credit.ToWorldState(stub); err != nil {
// 		return nil, err
// 	}
// 	return credit, nil
// }

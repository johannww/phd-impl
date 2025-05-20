package credits

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

// MintCredit represents a carbon credit that has been minted and
// it is associated to mint multiplier and mint timestamp.
type MintCredit struct {
	Credit        Credit  `json:"credit"`
	MintMult      float64 `json:"mintMult"`
	MintTimeStamp string  `json:"mintTimestamp"`
}

var _ state.WorldStateManager = (*MintCredit)(nil)

func (mc *MintCredit) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	panic("not implemented") // TODO: Implement
}
func (mc *MintCredit) ToWorldState(stub shim.ChaincodeStubInterface) error {
	panic("not implemented") // TODO: Implement
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

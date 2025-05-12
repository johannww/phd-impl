package credits

import "github.com/hyperledger/fabric-chaincode-go/shim"

// MintCredit represents a carbon credit that has been minted and
// it is associated to mint multiplier and mint timestamp.
type MintCredit struct {
	Credit        Credit  `json:"credit"`
	MintMult      float64 `json:"mintMult"`
	MintTimeStamp uint64  `json:"mintTimestamp"`
}

func (mc *MintCredit) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	panic("not implemented") // TODO: Implement
}
func (mc *MintCredit) ToWorldState(stub shim.ChaincodeStubInterface) error {
	panic("not implemented") // TODO: Implement
}
func (mc *MintCredit) GetID() *[][]string {
	return mc.Credit.GetID()
}

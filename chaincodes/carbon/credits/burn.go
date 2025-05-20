package credits

import "github.com/hyperledger/fabric-chaincode-go/v2/shim"

// BurnCredit represents a minted carbon credit to be burned.
// it is associated to burn multiplier and burn timestamp.
type BurnCredit struct {
	MintCredit    MintCredit `json:"mintCredit"`
	BurnMult      float64    `json:"burnMult"`
	BurnTimeStamp uint64     `json:"burnTimestamp"`
}

func (bc *BurnCredit) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	panic("not implemented") // TODO: Implement
}
func (bc *BurnCredit) ToWorldState(stub shim.ChaincodeStubInterface) error {
	panic("not implemented") // TODO: Implement
}
func (bc *BurnCredit) GetID() *[][]string {
	return bc.MintCredit.GetID()
}

// TODO: implement
func Burn(stub shim.ChaincodeStubInterface) error {
	return nil
}

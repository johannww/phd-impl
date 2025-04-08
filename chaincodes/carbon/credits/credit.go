package credits

import "github.com/hyperledger/fabric-chaincode-go/shim"

const (
	CREDIT_PREFIX = "credit"
)

// Credit represents a carbon unit minted for a property chunk
// at a specific time.
// TODO: enhance this struct
type Credit struct {
	ID       string        `json:"id"`
	Owner    string        `json:"owner"`
	Property Property      `json:"property"`
	Chunk    PropertyChunk `json:"chunk"`
}

func (c *Credit) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) (_ error) {
	panic("not implemented") // TODO: Implement
}

func (c *Credit) ToWorldState(stub shim.ChaincodeStubInterface) (_ error) {
	panic("not implemented") // TODO: Implement
}

func (c *Credit) GetID() (_ []string) {
	panic("not implemented") // TODO: Implement
}

package credits

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	prop "github.com/johannww/phd-impl/chaincodes/carbon/properties"
)

const (
	CREDIT_PREFIX = "credit"
)

// Credit represents a carbon unit minted for a property chunk
// at a specific time.
// TODO: enhance this struct
type Credit struct {
	OwnerID  string         `json:"owner"`
	Property *prop.Property `json:"property"`
	Chunk    *prop.Property `json:"chunk"`
}

// TODO:
func (c *Credit) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) (_ error) {
	panic("not implemented") // TODO: Implement
}

// TODO:
func (c *Credit) ToWorldState(stub shim.ChaincodeStubInterface) (_ error) {
	panic("not implemented") // TODO: Implement
}

// TODO:
func (c *Credit) GetID() (_ []string) {
	creditId := []string{c.OwnerID}
	creditId = append(creditId, c.Property.GetID()...)
	creditId = append(creditId, c.Chunk.GetID()...)
	return creditId
}

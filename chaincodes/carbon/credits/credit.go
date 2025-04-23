package credits

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	prop "github.com/johannww/phd-impl/chaincodes/carbon/properties"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	CREDIT_PREFIX = "credit"
)

// Credit represents a carbon unit minted for a property chunk
// at a specific time.
// TODO: enhance this struct
type Credit struct {
	OwnerID string              `json:"owner"`
	Chunk   *prop.PropertyChunk `json:"chunk"`
}

var _ state.WorldStateManager = (*Credit)(nil)

// TODO:
func (c *Credit) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) (_ error) {
	panic("not implemented") // TODO: Implement
}

// TODO:
func (c *Credit) ToWorldState(stub shim.ChaincodeStubInterface) (_ error) {
	panic("not implemented") // TODO: Implement
}

// TODO:
func (c *Credit) GetID() *[][]string {
	creditId := []string{c.OwnerID}
	creditId = append(creditId, (*c.Chunk.GetID())[0]...)
	return &[][]string{creditId}
}

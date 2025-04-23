package properties

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
	v "github.com/johannww/phd-impl/chaincodes/carbon/vegetation"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// PropertyChunk represents a chunk of a property.
// It exists because properties might have heterogeneous chunks.
// It points to the property because otherwise---if in a slice in the
// property struct---it could generate MVCC_READ_CONFLICT errors.
// See: https://github.com/hyperledger/fabric/issues/3748
type PropertyChunk struct {
	PropertyID       uint64              `json:"propertyId"`
	ChunkID          uint64              `json:"chunkId"`
	Coordinates      []Coordinates       `json:"coordinates"`
	VegetationsProps []v.VegetationProps `json:"vegetationsProps"`
}

var _ state.WorldStateManager = (*PropertyChunk)(nil)

func (propertychunk *PropertyChunk) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	panic("not implemented") // TODO: Implement
}

func (propertychunk *PropertyChunk) ToWorldState(stub shim.ChaincodeStubInterface) error {
	panic("not implemented") // TODO: Implement
}

func (propertychunk *PropertyChunk) GetID() *[][]string {
	panic("not implemented") // TODO: Implement
}

type Property struct {
	ID     uint64          `json:"id"`
	Chunks []PropertyChunk `json:"-"`
}

var _ state.WorldStateManager = (*Property)(nil)

func (property *Property) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	panic("not implemented") // TODO: Implement
}

func (property *Property) ToWorldState(stub shim.ChaincodeStubInterface) error {
	panic("not implemented") // TODO: Implement
}

func (property *Property) GetID() *[][]string {
	panic("not implemented") // TODO: Implement
}

// mock

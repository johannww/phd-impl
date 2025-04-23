package properties

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
	v "github.com/johannww/phd-impl/chaincodes/carbon/vegetation"
)

const (
	PROPERTY_PREFIX       = "property"
	PROPERTY_CHUNK_PREFIX = "propertyChunk"
	COORDINATES_PREFIX    = "coords"
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

// TODO: review how chunks should be loaded
type Property struct {
	ID uint64 `json:"id"`
	// Chunks will not be marshalled to the world state via
	// this struct. Instead, it will be marshalled via the
	// PropertyChunk struct.
	Chunks *[]PropertyChunk `json:"chunks"`
}

var _ state.WorldStateManager = (*Property)(nil)

func (property *Property) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	panic("not implemented") // TODO: Implement
}

func (property *Property) ToWorldState(stub shim.ChaincodeStubInterface) error {
	chunks := property.Chunks
	property.Chunks = nil // do not marshal chunks in the property struct

	err := state.PutStateWithCompositeKey(stub, PROPERTY_PREFIX, property.GetID(), property)

	for _, chunk := range *chunks {
		chunk.ToWorldState(stub)
	}

	// reset property chunks
	property.Chunks = chunks
	return err
}

func (property *Property) GetID() *[][]string {
	return &[][]string{{string(property.ID)}}
}

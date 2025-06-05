package properties

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	PROPERTY_PREFIX = "property"
)

type Property struct {
	OwnerID string `json:"ownerId"`
	ID      uint64 `json:"id"`
	// Chunks will not be marshalled to the world state via
	// this struct. Instead, it will be marshalled via the
	// PropertyChunk struct.
	Chunks []*PropertyChunk `json:"chunks"`
}

var _ state.WorldStateManager = (*Property)(nil)

func (property *Property) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := state.GetStateWithCompositeKey(stub, PROPERTY_PREFIX, keyAttributes, property)
	if err != nil {
		return fmt.Errorf("could not get property from world state: %v", err)
	}

	property.Chunks, err = state.GetStatesByPartialCompositeKey[PropertyChunk](stub, PROPERTY_CHUNK_PREFIX, keyAttributes)
	if err != nil {
		return fmt.Errorf("could not get property chunks from world state: %v", err)
	}

	return nil

}

func (property *Property) ToWorldState(stub shim.ChaincodeStubInterface) error {
	chunks := property.Chunks
	property.Chunks = nil // do not marshal chunks in the property struct

	err := state.PutStateWithCompositeKey(stub, PROPERTY_PREFIX, property.GetID(), property)

	for _, chunk := range chunks {
		chunk.ToWorldState(stub)
	}

	// reset property chunks
	property.Chunks = chunks
	return err
}

func (property *Property) GetID() *[][]string {
	return &[][]string{{
		property.OwnerID,
		strconv.FormatUint(property.ID, 10),
	}}
}

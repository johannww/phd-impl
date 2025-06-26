package properties

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
	v "github.com/johannww/phd-impl/chaincodes/carbon/vegetation"
)

const (
	PROPERTY_CHUNK_PREFIX = "propertyChunk"
	COORDINATES_PREFIX    = "coords"
)

// Coordinate represents a geographical coordinate in the floating point format.
type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// PropertyChunk represents a chunk of a property.
// It exists because properties might have heterogeneous chunks.
// It points to the property because otherwise---if in a slice in the
// property struct---it could generate MVCC_READ_CONFLICT errors.
// See: https://github.com/hyperledger/fabric/issues/3748
type PropertyChunk struct {
	PropertyID       uint64                `json:"propertyId"`
	Coordinates      []Coordinate          `json:"coordinates"`
	VegetationsProps []*v.VegetationProps  `json:"vegetationsProps"`
	ValidationProps  *data.ValidationProps `json:"validationProps"`
}

var _ state.WorldStateManager = (*PropertyChunk)(nil)

func (propertychunk *PropertyChunk) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	if len(keyAttributes) != 2 {
		return fmt.Errorf("invalid number of key attributes: %d", len(keyAttributes))
	}

	err := state.GetStateWithCompositeKey(stub, PROPERTY_CHUNK_PREFIX, keyAttributes, propertychunk)
	if err != nil {
		return fmt.Errorf("could not get property chunk from world state: %v", err)
	}

	return nil
}

func (propertychunk *PropertyChunk) ToWorldState(stub shim.ChaincodeStubInterface) error {
	err := state.PutStateWithCompositeKey(stub, PROPERTY_CHUNK_PREFIX, propertychunk.GetID(), propertychunk)
	if err != nil {
		return err
	}

	return nil
}

func (propertychunk *PropertyChunk) GetID() *[][]string {
	// WARN: This assumes that only this chunk has this coordinate
	firstCoordinate := propertychunk.Coordinates[0]

	return &[][]string{{
		strconv.FormatUint(propertychunk.PropertyID, 10),
		strconv.FormatFloat(firstCoordinate.Latitude, 'f', 6, 64),
		strconv.FormatFloat(firstCoordinate.Longitude, 'f', 6, 64),
	}}
}

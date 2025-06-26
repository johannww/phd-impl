package metheorological

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const TEMPERATURE_PREFIX = "temperature"

// Temperature represents a temperature reading at a specific geographical coordinate and time.
type Temperature struct {
	Coordinate  properties.Coordinate `json:"coordinate"`  // Geographical coordinate in floating point format
	Datetime    string                `json:"datetime"`    // RFC3339 utc format
	Temperature float64               `json:"temperature"` // Temperature in degrees Celsius
}

var _ state.WorldStateManager = (*Temperature)(nil)

// FromWorldState implements state.WorldStateManager.
func (t *Temperature) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	if len(keyAttributes) != 3 {
		return fmt.Errorf("invalid number of key attributes: %d", len(keyAttributes))
	}

	err := state.GetStateWithCompositeKey(stub, TEMPERATURE_PREFIX, keyAttributes, t)
	if err != nil {
		return fmt.Errorf("could not get property chunk from world state: %v", err)
	}

	return nil
}

// ToWorldState implements state.WorldStateManager.
func (t *Temperature) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutStateWithCompositeKey(stub, TEMPERATURE_PREFIX, t.GetID(), t)
}

// GetID implements state.WorldStateManager.
func (t *Temperature) GetID() *[][]string {
	// TODOHP: should i use a georeferenced database?
	latitudeStr := strconv.FormatFloat(t.Coordinate.Latitude, 'f', 6, 64)
	longitudeStr := strconv.FormatFloat(t.Coordinate.Longitude, 'f', 6, 64)
	return &[][]string{
		{latitudeStr, longitudeStr, t.Datetime},
		{t.Datetime, latitudeStr, longitudeStr},
	}
}

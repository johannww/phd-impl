package metheorological

import (
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const WIND_PREFIX = "wind"

// TODO: Save it on a georeferenced database
type Wind struct {
	Coordinate properties.Coordinate `json:"coordinate"` // Geographical coordinate in floating point format
	Datetime   string                `json:"datetime"`   // RFC3339 utc format
	Speed      float64               `json:"speed"`      // Wind speed in meters per second
	Direction  float64               `json:"direction"`  // Wind direction in degrees
}

var _ state.WorldStateManager = (*Wind)(nil)

// FromWorldState implements state.WorldStateManager.
func (w *Wind) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	if len(keyAttributes) != 3 {
		return nil
	}

	err := state.GetStateWithCompositeKey(stub, WIND_PREFIX, keyAttributes, w)
	if err != nil {
		return err
	}

	return nil
}

// GetID implements state.WorldStateManager.
func (w *Wind) GetID() *[][]string {
	latitudeStr := strconv.FormatFloat(w.Coordinate.Latitude, 'f', 6, 64)
	longitudeStr := strconv.FormatFloat(w.Coordinate.Longitude, 'f', 6, 64)
	return &[][]string{
		{latitudeStr, longitudeStr, w.Datetime},
		{w.Datetime, latitudeStr, longitudeStr},
	}
}

// ToWorldState implements state.WorldStateManager.
func (w *Wind) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutStateWithCompositeKey(stub, WIND_PREFIX, w.GetID(), w)
}

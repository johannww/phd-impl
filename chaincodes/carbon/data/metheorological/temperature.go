package metheorological

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"google.golang.org/protobuf/proto"
)

const TEMPERATURE_PREFIX = "temperature"

// Temperature represents a temperature reading at a specific geographical coordinate and time.
type Temperature struct {
	Coordinate  utils.Coordinate `json:"coordinate"`  // Geographical coordinate in floating point format
	Datetime    string           `json:"datetime"`    // RFC3339 utc format
	Temperature float64          `json:"temperature"` // Temperature in degrees Celsius
}

var _ state.WorldStateManager = (*Temperature)(nil)

func (t *Temperature) ToProto() proto.Message {
	return &pb.Temperature{
		Coordinate: &pb.Coordinate{
			Latitude:  t.Coordinate.Latitude,
			Longitude: t.Coordinate.Longitude,
		},
		Datetime:    t.Datetime,
		Temperature: t.Temperature,
	}
}

func (t *Temperature) FromProto(m proto.Message) error {
	pt, ok := m.(*pb.Temperature)
	if !ok {
		return fmt.Errorf("unexpected proto message type for Temperature")
	}
	if pt.Coordinate != nil {
		t.Coordinate = utils.Coordinate{
			Latitude:  pt.Coordinate.Latitude,
			Longitude: pt.Coordinate.Longitude,
		}
	} else {
		t.Coordinate = utils.Coordinate{}
	}
	t.Datetime = pt.Datetime
	t.Temperature = pt.Temperature
	return nil
}

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
	// I could use a mongo db with geospatial indexing and save readings hashes on chain
	latitudeStr := strconv.FormatFloat(t.Coordinate.Latitude, 'f', 6, 64)
	longitudeStr := strconv.FormatFloat(t.Coordinate.Longitude, 'f', 6, 64)
	return &[][]string{
		{latitudeStr, longitudeStr, t.Datetime},
		{t.Datetime, latitudeStr, longitudeStr},
	}
}

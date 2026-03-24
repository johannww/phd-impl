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

const WIND_PREFIX = "wind"

// TODO: Save it on a georeferenced database
type Wind struct {
	Coordinate utils.Coordinate `json:"coordinate"` // Geographical coordinate in floating point format
	Datetime   string           `json:"datetime"`   // RFC3339 utc format
	Speed      float64          `json:"speed"`      // Wind speed in meters per second
	Direction  float64          `json:"direction"`  // Wind direction in degrees
}

var _ state.WorldStateManager = (*Wind)(nil)

func (w *Wind) ToProto() proto.Message {
	return &pb.Wind{
		Coordinate: &pb.Coordinate{
			Latitude:  w.Coordinate.Latitude,
			Longitude: w.Coordinate.Longitude,
		},
		Datetime:  w.Datetime,
		Speed:     w.Speed,
		Direction: w.Direction,
	}
}

func (w *Wind) FromProto(m proto.Message) error {
	pw, ok := m.(*pb.Wind)
	if !ok {
		return fmt.Errorf("unexpected proto message type for Wind")
	}
	if pw.Coordinate != nil {
		w.Coordinate = utils.Coordinate{Latitude: pw.Coordinate.Latitude, Longitude: pw.Coordinate.Longitude}
	} else {
		w.Coordinate = utils.Coordinate{}
	}
	w.Datetime = pw.Datetime
	w.Speed = pw.Speed
	w.Direction = pw.Direction
	return nil
}

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

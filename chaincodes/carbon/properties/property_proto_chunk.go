package properties

import (
	"fmt"

	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/vegetation"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"google.golang.org/protobuf/proto"
)

// ToProto/FromProto implement explicit conversion for PropertyChunk
func (propertychunk *PropertyChunk) ToProto() proto.Message {
	coords := make([]*pb.Coordinate, 0, len(propertychunk.Coordinates))
	for _, c := range propertychunk.Coordinates {
		coords = append(coords, &pb.Coordinate{Latitude: c.Latitude, Longitude: c.Longitude})
	}

	var vegProps *pb.VegetationProps
	if propertychunk.VegetationsProps != nil {
		vegProps = &pb.VegetationProps{
			ForestType:    pb.ForestType(propertychunk.VegetationsProps.ForestType),
			ForestDensity: propertychunk.VegetationsProps.ForestDensity,
			CropType:      pb.CropType(propertychunk.VegetationsProps.CropType),
		}
	}

	var valProps *pb.ValidationProps
	if propertychunk.ValidationProps != nil {
		methods := make([]pb.ValidationMethod, 0, len(propertychunk.ValidationProps.Methods))
		for _, m := range propertychunk.ValidationProps.Methods {
			methods = append(methods, pb.ValidationMethod(m))
		}
		valProps = &pb.ValidationProps{Methods: methods}
	}

	return &pb.PropertyChunk{
		PropertyId:       propertychunk.PropertyID,
		Coordinates:      coords,
		VegetationsProps: vegProps,
		ValidationProps:  valProps,
	}
}

func (propertychunk *PropertyChunk) FromProto(m proto.Message) error {
	ppc, ok := m.(*pb.PropertyChunk)
	if !ok {
		return fmt.Errorf("unexpected proto message type for PropertyChunk")
	}
	if ppc == nil {
		return nil
	}
	propertychunk.PropertyID = ppc.PropertyId
	coords := make([]utils.Coordinate, 0, len(ppc.Coordinates))
	for _, c := range ppc.Coordinates {
		coords = append(coords, utils.Coordinate{Latitude: c.Latitude, Longitude: c.Longitude})
	}
	propertychunk.Coordinates = coords

	if ppc.VegetationsProps != nil {
		propertychunk.VegetationsProps = &vegetation.VegetationProps{
			ForestType:    vegetation.ForestType(ppc.VegetationsProps.ForestType),
			ForestDensity: ppc.VegetationsProps.ForestDensity,
			CropType:      vegetation.CropType(ppc.VegetationsProps.CropType),
		}
	} else {
		propertychunk.VegetationsProps = nil
	}

	if ppc.ValidationProps != nil {
		methods := make([]data.ValidationMethod, 0, len(ppc.ValidationProps.Methods))
		for _, m := range ppc.ValidationProps.Methods {
			methods = append(methods, data.ValidationMethod(m))
		}
		propertychunk.ValidationProps = &data.ValidationProps{Methods: methods}
	} else {
		propertychunk.ValidationProps = nil
	}

	return nil
}

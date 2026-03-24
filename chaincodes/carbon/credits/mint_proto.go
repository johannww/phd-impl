package credits

import (
	"fmt"

	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	"github.com/johannww/phd-impl/chaincodes/carbon/vegetation"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"google.golang.org/protobuf/proto"
)

// ToProto/FromProto implement explicit conversion to/from generated proto
func (mc *MintCredit) ToProto() proto.Message {
	var pbCredit *pb.Credit
	if mc.Credit.OwnerID != "" || len(mc.Credit.ChunkID) > 0 || mc.Credit.Quantity != 0 {
		var pbChunk *pb.PropertyChunk
		if mc.Chunk != nil {
			// Map PropertyChunk -> pb.PropertyChunk
			coords := make([]*pb.Coordinate, 0, len(mc.Chunk.Coordinates))
			for _, c := range mc.Chunk.Coordinates {
				coords = append(coords, &pb.Coordinate{Latitude: c.Latitude, Longitude: c.Longitude})
			}
			var vegProps *pb.VegetationProps
			if mc.Chunk.VegetationsProps != nil {
				vegProps = &pb.VegetationProps{
					ForestType:    pb.ForestType(mc.Chunk.VegetationsProps.ForestType),
					ForestDensity: mc.Chunk.VegetationsProps.ForestDensity,
					CropType:      pb.CropType(mc.Chunk.VegetationsProps.CropType),
				}
			}
			var valProps *pb.ValidationProps
			if mc.Chunk.ValidationProps != nil {
				methods := make([]pb.ValidationMethod, 0, len(mc.Chunk.ValidationProps.Methods))
				for _, m := range mc.Chunk.ValidationProps.Methods {
					methods = append(methods, pb.ValidationMethod(m))
				}
				valProps = &pb.ValidationProps{Methods: methods}
			}
			pbChunk = &pb.PropertyChunk{
				PropertyId:       mc.Chunk.PropertyID,
				Coordinates:      coords,
				VegetationsProps: vegProps,
				ValidationProps:  valProps,
			}
		}
		pbCredit = &pb.Credit{
			Owner:    mc.Credit.OwnerID,
			ChunkId:  mc.Credit.ChunkID,
			Chunk:    pbChunk,
			Quantity: mc.Credit.Quantity,
		}
	}

	var pbMintChunk *pb.MintCredit
	if pbCredit != nil {
		pbMintChunk = &pb.MintCredit{
			Credit:        pbCredit,
			MintMult:      mc.MintMult,
			MintTimestamp: mc.MintTimeStamp,
		}
	}

	return pbMintChunk
}

func (mc *MintCredit) FromProto(m proto.Message) error {
	pmc, ok := m.(*pb.MintCredit)
	if !ok {
		return fmt.Errorf("unexpected proto message type for MintCredit")
	}
	if pmc == nil {
		return nil
	}
	if pmc.Credit != nil {
		mc.Credit = Credit{
			OwnerID:  pmc.Credit.Owner,
			ChunkID:  pmc.Credit.ChunkId,
			Quantity: pmc.Credit.Quantity,
		}
		if pmc.Credit.Chunk != nil {
			// Map pb.PropertyChunk -> PropertyChunk
			coords := make([]utils.Coordinate, 0, len(pmc.Credit.Chunk.Coordinates))
			for _, c := range pmc.Credit.Chunk.Coordinates {
				coords = append(coords, utils.Coordinate{Latitude: c.Latitude, Longitude: c.Longitude})
			}
			var vegProps *vegetation.VegetationProps
			if pmc.Credit.Chunk.VegetationsProps != nil {
				vegProps = &vegetation.VegetationProps{
					ForestType:    vegetation.ForestType(pmc.Credit.Chunk.VegetationsProps.ForestType),
					ForestDensity: pmc.Credit.Chunk.VegetationsProps.ForestDensity,
					CropType:      vegetation.CropType(pmc.Credit.Chunk.VegetationsProps.CropType),
				}
			}
			var valProps *data.ValidationProps
			if pmc.Credit.Chunk.ValidationProps != nil {
				methods := make([]data.ValidationMethod, 0, len(pmc.Credit.Chunk.ValidationProps.Methods))
				for _, m := range pmc.Credit.Chunk.ValidationProps.Methods {
					methods = append(methods, data.ValidationMethod(m))
				}
				valProps = &data.ValidationProps{Methods: methods}
			}
			mc.Chunk = &properties.PropertyChunk{
				PropertyID:       pmc.Credit.Chunk.PropertyId,
				Coordinates:      coords,
				VegetationsProps: vegProps,
				ValidationProps:  valProps,
			}
		} else {
			mc.Chunk = nil
		}
	}
	mc.MintMult = pmc.MintMult
	mc.MintTimeStamp = pmc.MintTimestamp
	return nil
}

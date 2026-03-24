package properties

import (
	"fmt"

	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"google.golang.org/protobuf/proto"
)

// ToProto/FromProto implement explicit conversion for Property
func (property *Property) ToProto() proto.Message {
	var chunks []*pb.PropertyChunk
	for _, c := range property.Chunks {
		if c == nil {
			continue
		}
		pc := c.ToProto().(*pb.PropertyChunk)
		chunks = append(chunks, pc)
	}
	return &pb.Property{
		OwnerId:          property.OwnerID,
		Id:               property.ID,
		RegistryId:       property.RegistryID,
		RegistryProvider: property.RegistryProvider,
		Chunks:           chunks,
	}
}

func (property *Property) FromProto(m proto.Message) error {
	pp, ok := m.(*pb.Property)
	if !ok {
		return fmt.Errorf("unexpected proto message type for Property")
	}
	property.OwnerID = pp.OwnerId
	property.ID = pp.Id
	property.RegistryID = pp.RegistryId
	property.RegistryProvider = pp.RegistryProvider
	chunks := make([]*PropertyChunk, 0, len(pp.Chunks))
	for _, c := range pp.Chunks {
		pc := &PropertyChunk{}
		if err := pc.FromProto(c); err != nil {
			return fmt.Errorf("could not convert chunk proto: %v", err)
		}
		chunks = append(chunks, pc)
	}
	property.Chunks = chunks
	return nil
}

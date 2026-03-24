package serializer

import (
	"fmt"
	"google.golang.org/protobuf/proto"
)

type ProtoSerializer struct{}

func NewProtoSerializer() *ProtoSerializer { return &ProtoSerializer{} }

func (p *ProtoSerializer) Marshal(v ProtoConvertible) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("nil value")
	}
	msg := v.ToProto()
	if msg == nil {
		return nil, fmt.Errorf("ToProto returned nil")
	}
	return proto.Marshal(msg)
}

func (p *ProtoSerializer) Unmarshal(data []byte, v ProtoConvertible) error {
	if v == nil {
		return fmt.Errorf("nil out")
	}
	if len(data) == 0 {
		return fmt.Errorf("empty data")
	}
	msg := v.ToProto()
	if msg == nil {
		return fmt.Errorf("ToProto returned nil message for unmarshal target")
	}
	if err := proto.Unmarshal(data, msg); err != nil {
		return err
	}
	return v.FromProto(msg)
}

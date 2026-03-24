package serializer

import "google.golang.org/protobuf/proto"

// ProtoConvertible describes a domain object that can convert to/from a proto
type ProtoConvertible interface {
	ToProto() proto.Message
	FromProto(proto.Message) error
}

// Serializer defines methods to marshal/unmarshal domain objects to bytes
type Serializer interface {
	Marshal(v ProtoConvertible) ([]byte, error)
	Unmarshal(data []byte, v ProtoConvertible) error
}

// ErrNotProtoCompatible is returned when ProtoSerializer cannot handle a type
type ErrNotProtoCompatible struct{ t string }

func (e ErrNotProtoCompatible) Error() string { return "not proto compatible: " + e.t }

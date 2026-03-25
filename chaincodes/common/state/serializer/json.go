package serializer

import (
	"bytes"
	"encoding/json"
)

type JSONSerializer struct{}

func NewJSONSerializer() *JSONSerializer { return &JSONSerializer{} }

func (j *JSONSerializer) Marshal(v ProtoConvertible) ([]byte, error) {
	// For JSON we marshal the domain object itself (not its proto representation).
	return json.Marshal(v)
}

// Unmarshal expects v to implement ProtoConvertible. We prefer explicit
// ProtoConvertible usage in this package so callers must provide a type that
// can convert to/from a proto message. For JSON-backed types the FromProto
// implementation should perform JSON unmarshalling from the bytes into the
// receiver.
func (j *JSONSerializer) Unmarshal(data []byte, v ProtoConvertible) error {
	if v == nil {
		return &ErrNotProtoCompatible{"nil target"}
	}
	// For JSON-backed types we expect the receiver to be a pointer to the
	// concrete domain struct that also implements ProtoConvertible. Passing
	// that pointer to json.Unmarshal works because json.Unmarshal accepts an
	// interface{} and will update the underlying value.
	return json.Unmarshal(data, v)
}

// StrictUnmarshal unmarshals data while disallowing unknown/extra fields.
// This is useful for distinguishing between different message types that
// might have overlapping field names.
func (j *JSONSerializer) StrictUnmarshal(data []byte, v ProtoConvertible) error {
	if v == nil {
		return &ErrNotProtoCompatible{"nil target"}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

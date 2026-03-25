package serializer

import (
	"bytes"
	"encoding/json"
)

type JSONSerializer struct{}

func NewJSONSerializer() *JSONSerializer { return &JSONSerializer{} }

func (j *JSONSerializer) Marshal(v ProtoConvertible) ([]byte, error) {
	return json.Marshal(v)
}

func (j *JSONSerializer) Unmarshal(data []byte, v ProtoConvertible) error {
	if v == nil {
		return &ErrNotProtoCompatible{"nil target"}
	}
	return json.Unmarshal(data, v)
}

// StrictUnmarshal rejects unknown/extra fields to distinguish between result types
func (j *JSONSerializer) StrictUnmarshal(data []byte, v ProtoConvertible) error {
	if v == nil {
		return &ErrNotProtoCompatible{"nil target"}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

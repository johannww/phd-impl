package state

import (
	"encoding/json"
	"fmt"
	"github.com/johannww/phd-impl/chaincodes/common/state/serializer"
)

// StateSaver holds a Serializer and provides marshal/unmarshal helpers
type StateSaver struct {
	serializer serializer.Serializer
}

func (s *StateSaver) Marshal(v ProtoConvertible) ([]byte, error) {
	return s.serializer.Marshal(v)
}

func (s *StateSaver) Unmarshal(data []byte, v any) error {
	pc, ok := v.(ProtoConvertible)
	if !ok {
		return fmt.Errorf("unmarshal target does not implement ProtoConvertible")
	}
	return s.serializer.Unmarshal(data, pc)
}

// NewStateSaver constructs a StateSaver with the provided serializer
func NewStateSaver(s serializer.Serializer) *StateSaver {
	return &StateSaver{serializer: s}
}

// currentSerializer is the package-level serializer used by chaincode_utils and
// batch_recover helpers. It is initialized to NewJSONSerializer() for backward
// compatibility. Call SetSerializer() to override at init time (e.g., in main or tests).
//
// Usage:
//   - Default JSON serialization: No action needed, system uses JSON by default
//   - Switch to Proto serialization: Call SetSerializer(serializer.NewProtoSerializer()) in init()
//   - Custom serializer: Implement the serializer.Serializer interface and pass to SetSerializer()
var currentSerializer serializer.Serializer = serializer.NewJSONSerializer()

// SetSerializer sets the package-level serializer used by chaincode_utils and
// batch_recover helpers. This should be called once at init time (e.g., in main
// or test setup) before any marshaling/unmarshaling occurs. If s is nil, this
// function is a no-op to prevent nil-pointer dereferences.
//
// Example:
//
//	func init() {
//	    // Switch from default JSON to Protocol Buffers
//	    state.SetSerializer(serializer.NewProtoSerializer())
//	}
func SetSerializer(s serializer.Serializer) {
	if s != nil {
		currentSerializer = s
	}
}

// GetSerializer returns the current package-level serializer used by chaincode_utils
// and batch_recover helpers. Never returns nil.
//
// This is useful for introspecting the current serialization format or for
// manually serializing values in custom code.
func GetSerializer() serializer.Serializer {
	return currentSerializer
}

// UnmarshalStateAs unmarshals state bytes into a ProtoConvertible target using the current serializer.
// Falls back to JSON unmarshaling for types that don't implement ProtoConvertible (e.g., []string).
func UnmarshalStateAs(data []byte, target any) error {
	pc, ok := target.(ProtoConvertible)
	if !ok {
		return unmarshalJSON(data, target)
	}
	return GetSerializer().Unmarshal(data, pc)
}

// UnmarshalStateAsStrict unmarshals state bytes using strict validation.
// For JSON: rejects unknown fields. For proto: identical to Unmarshal (validates field numbers).
func UnmarshalStateAsStrict(data []byte, target any) error {
	pc, ok := target.(ProtoConvertible)
	if !ok {
		return unmarshalJSON(data, target)
	}
	return GetSerializer().StrictUnmarshal(data, pc)
}

func unmarshalJSON(data []byte, target any) error {
	var err error
	switch t := target.(type) {
	case *[]string:
		// Special case for []string used in secondary indexes
		err = json.Unmarshal(data, t)
	default:
		// For other types, try JSON unmarshaling as last resort
		err = json.Unmarshal(data, target)
	}
	return err
}

// MarshalStateAs marshals a ProtoConvertible value using the current serializer.
func MarshalStateAs(value any) ([]byte, error) {
	pc, ok := value.(ProtoConvertible)
	if !ok {
		return nil, fmt.Errorf("marshal value does not implement ProtoConvertible")
	}
	return GetSerializer().Marshal(pc)
}

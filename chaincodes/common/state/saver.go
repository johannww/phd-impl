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

// defaultManager is the package-level StateManager used by the facade helpers.
// It is initialized once and not mutated at runtime to avoid races and mutexes.
var defaultManager = NewStateManager(NewStateSaver(serializer.NewJSONSerializer()))

// DefaultManager returns the package-level StateManager. Callers that need
// different serializers should construct their own StateManager via
// NewStateManager.
func DefaultManager() *StateManager { return defaultManager }

// currentSerializer is the package-level serializer used by chaincode_utils and
// batch_recover helpers. It is initialized to NewJSONSerializer() for backward
// compatibility. Call SetSerializer() to override at init time (e.g., in main or tests).
var currentSerializer serializer.Serializer = serializer.NewJSONSerializer()

// SetSerializer sets the package-level serializer used by chaincode_utils and
// batch_recover helpers. This should be called once at init time (e.g., in main
// or test setup) before any marshaling/unmarshaling occurs. If s is nil, this
// function is a no-op to prevent nil-pointer dereferences.
func SetSerializer(s serializer.Serializer) {
	if s != nil {
		currentSerializer = s
	}
}

// GetSerializer returns the current package-level serializer used by chaincode_utils
// and batch_recover helpers. Never returns nil.
func GetSerializer() serializer.Serializer {
	return currentSerializer
}

// UnmarshalStateAs unmarshals state bytes into an arbitrary target that implements
// ProtoConvertible. This is useful for generic code that needs to unmarshal into
// a type parameter without compile-time constraint validation. Returns an error if
// the target does not implement ProtoConvertible.
func UnmarshalStateAs(data []byte, target any) error {
	pc, ok := target.(ProtoConvertible)
	if !ok {
		// Fallback to JSON unmarshaling for types that don't implement ProtoConvertible
		// This is useful for secondary index operations that store []string keys
		return unmarshalJSON(data, target)
	}
	return GetSerializer().Unmarshal(data, pc)
}

// unmarshalJSON unmarshals JSON bytes into target. Used as fallback for non-ProtoConvertible types.
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

// MarshalStateAs marshals an arbitrary value that implements ProtoConvertible to bytes.
// This is useful for generic code that needs to marshal from a type parameter without
// compile-time constraint validation. Returns an error if the value does not implement
// ProtoConvertible.
func MarshalStateAs(value any) ([]byte, error) {
	pc, ok := value.(ProtoConvertible)
	if !ok {
		return nil, fmt.Errorf("marshal value does not implement ProtoConvertible")
	}
	return GetSerializer().Marshal(pc)
}

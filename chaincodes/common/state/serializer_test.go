package state

import (
	"testing"

	"github.com/johannww/phd-impl/chaincodes/common/state/serializer"
	"github.com/stretchr/testify/require"
)

// TestSerializerSwitching verifies that SetSerializer correctly switches the active serializer
func TestSerializerSwitching(t *testing.T) {
	// Save original serializer to restore after test
	originalSerializer := GetSerializer()
	defer SetSerializer(originalSerializer)

	// Verify initial serializer is JSON
	initialSerializer := GetSerializer()
	require.NotNil(t, initialSerializer, "GetSerializer should never return nil")
	_, ok := initialSerializer.(*serializer.JSONSerializer)
	require.True(t, ok, "Initial serializer should be JSONSerializer")

	// Switch to ProtoSerializer
	protoSerializer := serializer.NewProtoSerializer()
	SetSerializer(protoSerializer)

	// Verify current serializer is now ProtoSerializer
	currentSerializer := GetSerializer()
	require.NotNil(t, currentSerializer, "GetSerializer should never return nil")
	_, ok = currentSerializer.(*serializer.ProtoSerializer)
	require.True(t, ok, "Current serializer should be ProtoSerializer after SetSerializer call")

	// Verify that setting nil is a no-op (doesn't change serializer)
	SetSerializer(nil)
	currentSerializer = GetSerializer()
	_, ok = currentSerializer.(*serializer.ProtoSerializer)
	require.True(t, ok, "SetSerializer(nil) should not change the current serializer")
}

// TestUnmarshalStateAsWithProtoConvertible verifies that UnmarshalStateAs works with ProtoConvertible types
func TestUnmarshalStateAsWithProtoConvertible(t *testing.T) {
	originalSerializer := GetSerializer()
	defer SetSerializer(originalSerializer)

	// Create a mock object that implements ProtoConvertible
	mockObj := &MockObjectWithSecondaryIndex{
		MockAttr: "test_attr",
		MockPvt:  "test_private",
	}

	// Marshal the object using JSON serializer (default)
	jsonSerializer := serializer.NewJSONSerializer()
	data, err := jsonSerializer.Marshal(mockObj)
	require.NoError(t, err, "Failed to marshal mock object")

	// Unmarshal using UnmarshalStateAs
	unmarshalledObj := &MockObjectWithSecondaryIndex{}
	err = UnmarshalStateAs(data, unmarshalledObj)
	require.NoError(t, err, "Failed to unmarshal state")
	require.Equal(t, mockObj.MockAttr, unmarshalledObj.MockAttr, "MockAttr should match after unmarshal")
	require.Equal(t, mockObj.MockPvt, unmarshalledObj.MockPvt, "MockPvt should match after unmarshal")
}

// TestUnmarshalStateAsFallbackToJSON verifies that UnmarshalStateAs falls back to JSON for non-ProtoConvertible types
func TestUnmarshalStateAsFallbackToJSON(t *testing.T) {
	originalSerializer := GetSerializer()
	defer SetSerializer(originalSerializer)

	// Test fallback for []string (used in secondary indexes)
	testStrings := []string{"index1", "index2", "index3"}
	jsonBytes := []byte(`["index1","index2","index3"]`)

	unmarshalledStrings := []string{}
	err := UnmarshalStateAs(jsonBytes, &unmarshalledStrings)
	require.NoError(t, err, "Failed to unmarshal string slice")
	require.Equal(t, testStrings, unmarshalledStrings, "String slice should match after unmarshal")
}

// TestMarshalStateAsWithProtoConvertible verifies that MarshalStateAs works with ProtoConvertible types
func TestMarshalStateAsWithProtoConvertible(t *testing.T) {
	mockObj := &MockObjectWithSecondaryIndex{
		MockAttr: "test_attr",
		MockPvt:  "test_private",
	}

	// Marshal using MarshalStateAs
	data, err := MarshalStateAs(mockObj)
	require.NoError(t, err, "Failed to marshal state")
	require.NotNil(t, data, "Marshaled data should not be nil")

	// Verify we can unmarshal it back
	unmarshalledObj := &MockObjectWithSecondaryIndex{}
	err = UnmarshalStateAs(data, unmarshalledObj)
	require.NoError(t, err, "Failed to unmarshal state")
	require.Equal(t, mockObj.MockAttr, unmarshalledObj.MockAttr, "MockAttr should match after round-trip")
}

// TestMarshalStateAsRejectsNonProtoConvertible verifies that MarshalStateAs rejects non-ProtoConvertible types
func TestMarshalStateAsRejectsNonProtoConvertible(t *testing.T) {
	// Try to marshal a plain string (not ProtoConvertible)
	data, err := MarshalStateAs("not proto convertible")
	require.Error(t, err, "MarshalStateAs should reject non-ProtoConvertible types")
	require.Nil(t, data, "Marshaled data should be nil on error")
	require.Contains(t, err.Error(), "does not implement ProtoConvertible", "Error message should be descriptive")
}

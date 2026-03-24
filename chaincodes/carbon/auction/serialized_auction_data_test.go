package auction

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestSerializedAuctionDataProtoRoundTrip verifies that SerializedAuctionData can be converted to proto and back
func TestSerializedAuctionDataProtoRoundTrip(t *testing.T) {
	testAuctionDataBytes := []byte{1, 2, 3, 4, 5}
	testHashBytes := []byte{0xaa, 0xbb, 0xcc, 0xdd}

	originalSad := &SerializedAuctionData{
		AuctionDataBytes: testAuctionDataBytes,
		Sum:              testHashBytes,
	}

	pbSad := originalSad.ToProto()
	require.NotNil(t, pbSad, "ToProto should return a non-nil pb.SerializedAuctionData")

	convertedSad := &SerializedAuctionData{}
	err := convertedSad.FromProto(pbSad)
	require.NoError(t, err, "FromProto should not error")

	require.Equal(t, originalSad.AuctionDataBytes, convertedSad.AuctionDataBytes, "AuctionDataBytes should match after round-trip")
	require.Equal(t, originalSad.Sum, convertedSad.Sum, "Sum should match after round-trip")
}

// TestSerializedAuctionDataProtoWithNilFields verifies SerializedAuctionData handles nil fields
func TestSerializedAuctionDataProtoWithNilFields(t *testing.T) {
	originalSad := &SerializedAuctionData{
		AuctionDataBytes: nil,
		Sum:              nil,
	}

	pbSad := originalSad.ToProto()
	require.NotNil(t, pbSad, "ToProto should return a non-nil pb.SerializedAuctionData")

	convertedSad := &SerializedAuctionData{}
	err := convertedSad.FromProto(pbSad)
	require.NoError(t, err, "FromProto should not error")

	require.Empty(t, convertedSad.AuctionDataBytes, "AuctionDataBytes should be empty after round-trip")
	require.Empty(t, convertedSad.Sum, "Sum should be empty after round-trip")
}

// TestSerializedAuctionDataProtoMessageType verifies that ToProto returns correct proto type
func TestSerializedAuctionDataProtoMessageType(t *testing.T) {
	sad := &SerializedAuctionData{
		AuctionDataBytes: []byte("test data"),
		Sum:              []byte("test hash"),
	}

	pbSad := sad.ToProto()
	require.NotNil(t, pbSad, "ToProto should not return nil")

	data, err := proto.Marshal(pbSad)
	require.NoError(t, err, "Proto message should be marshalable")
	require.NotEmpty(t, data, "Marshaled proto should not be empty")
}

package credits

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestBurnCreditProtoRoundTrip verifies that BurnCredit can be converted to proto and back
func TestBurnCreditProtoRoundTrip(t *testing.T) {
	// Create a test BurnCredit with nested MintCredit
	originalBc := &BurnCredit{
		MintCreditID: []string{"owner1", "chunk1", "chunk2"},
		MintCredit: &MintCredit{
			Credit: Credit{
				OwnerID:  "owner1",
				ChunkID:  []string{"chunk1", "chunk2"},
				Quantity: 500,
			},
			MintMult:      100,
			MintTimeStamp: "2024-03-24T10:00:00Z",
		},
		BurnQuantity:  250,
		BurnMult:      50,
		BurnTimeStamp: "2024-03-24T11:00:00Z",
		Adjusted:      true,
	}

	// Convert to proto
	pbBc := originalBc.ToProto()
	require.NotNil(t, pbBc, "ToProto should return a non-nil pb.BurnCredit")

	// Convert back from proto
	convertedBc := &BurnCredit{}
	err := convertedBc.FromProto(pbBc)
	require.NoError(t, err, "FromProto should not error")

	// Verify round-trip preserves data
	require.Equal(t, originalBc.BurnQuantity, convertedBc.BurnQuantity, "BurnQuantity should match after round-trip")
	require.Equal(t, originalBc.BurnMult, convertedBc.BurnMult, "BurnMult should match after round-trip")
	require.Equal(t, originalBc.BurnTimeStamp, convertedBc.BurnTimeStamp, "BurnTimeStamp should match after round-trip")
	require.Equal(t, originalBc.Adjusted, convertedBc.Adjusted, "Adjusted should match after round-trip")
	require.NotNil(t, convertedBc.MintCredit, "MintCredit should not be nil after round-trip")
	require.Equal(t, originalBc.MintCredit.MintMult, convertedBc.MintCredit.MintMult, "MintMult should match after round-trip")
}

// TestBurnCreditProtoWithNilMintCredit verifies that BurnCredit handles nil MintCredit correctly
func TestBurnCreditProtoWithNilMintCredit(t *testing.T) {
	// Create a minimal BurnCredit without MintCredit
	originalBc := &BurnCredit{
		MintCreditID:  []string{"owner1", "chunk1"},
		BurnQuantity:  100,
		BurnMult:      25,
		BurnTimeStamp: "2024-03-24T11:00:00Z",
		Adjusted:      false,
	}

	// Convert to proto
	pbBc := originalBc.ToProto()
	require.NotNil(t, pbBc, "ToProto should return a non-nil pb.BurnCredit")

	// Convert back from proto
	convertedBc := &BurnCredit{}
	err := convertedBc.FromProto(pbBc)
	require.NoError(t, err, "FromProto should not error")

	// Verify round-trip preserves data and nil fields
	require.Equal(t, originalBc.BurnQuantity, convertedBc.BurnQuantity, "BurnQuantity should match after round-trip")
	require.Nil(t, convertedBc.MintCredit, "MintCredit should remain nil after round-trip")
	require.Equal(t, originalBc.Adjusted, convertedBc.Adjusted, "Adjusted should match after round-trip")
}

// TestBurnCreditProtoMessageType verifies that ToProto returns correct proto type
func TestBurnCreditProtoMessageType(t *testing.T) {
	bc := &BurnCredit{
		BurnQuantity: 42,
	}

	pbBc := bc.ToProto()
	require.NotNil(t, pbBc, "ToProto should not return nil")

	// Verify the proto message can be marshaled
	data, err := proto.Marshal(pbBc)
	require.NoError(t, err, "Proto message should be marshalable")
	require.NotEmpty(t, data, "Marshaled proto should not be empty")
}

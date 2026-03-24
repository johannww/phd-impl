package bids

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestMatchedBidProtoRoundTrip verifies that MatchedBid can be converted to proto and back
func TestMatchedBidProtoRoundTrip(t *testing.T) {
	// Create a test MatchedBid with nested BuyBid and SellBid
	originalMb := &MatchedBid{
		BuyBid: &BuyBid{
			BuyerID:     "buyer1",
			Timestamp:   "2024-03-24T10:00:00Z",
			AskQuantity: 100,
			PrivateQuantity: &PrivateQuantity{
				AskQuantity: 50,
				BidID:       []string{"bid1", "bid2"},
			},
		},
		SellBid: &SellBid{
			SellerID:  "seller1",
			Timestamp: "2024-03-24T10:00:00Z",
			Quantity:  100,
			CreditID:  []string{"credit1"},
			PrivatePrice: &PrivatePrice{
				Price: 95,
				BidID: []string{"matched1"},
			},
		},
		Quantity: 75,
		PrivatePrice: &PrivatePrice{
			Price: 95,
			BidID: []string{"matched1"},
		},
		PrivateMultiplier: &PrivateMultiplier{
			MatchingID: []string{"matched1"},
			Scale:      1000,
			Value:      2,
		},
	}

	// Convert to proto
	pbMb := originalMb.ToProto()
	require.NotNil(t, pbMb, "ToProto should return a non-nil pb.MatchedBid")

	// Convert back from proto
	convertedMb := &MatchedBid{}
	err := convertedMb.FromProto(pbMb)
	require.NoError(t, err, "FromProto should not error")

	// Verify round-trip preserves data
	require.Equal(t, originalMb.Quantity, convertedMb.Quantity, "Quantity should match after round-trip")
	require.NotNil(t, convertedMb.BuyBid, "BuyBid should not be nil after round-trip")
	require.NotNil(t, convertedMb.SellBid, "SellBid should not be nil after round-trip")
	require.Equal(t, originalMb.BuyBid.BuyerID, convertedMb.BuyBid.BuyerID, "BuyerID should match after round-trip")
	require.Equal(t, originalMb.SellBid.SellerID, convertedMb.SellBid.SellerID, "SellerID should match after round-trip")
	require.Equal(t, originalMb.BuyBid.AskQuantity, convertedMb.BuyBid.AskQuantity, "AskQuantity should match after round-trip")
}

// TestMatchedBidProtoWithNilFields verifies that MatchedBid handles nil fields correctly
func TestMatchedBidProtoWithNilFields(t *testing.T) {
	// Create a minimal MatchedBid without optional fields
	originalMb := &MatchedBid{
		BuyBid: &BuyBid{
			BuyerID:     "buyer1",
			Timestamp:   "2024-03-24T10:00:00Z",
			AskQuantity: 100,
		},
		SellBid: &SellBid{
			SellerID:  "seller1",
			Timestamp: "2024-03-24T10:00:00Z",
			Quantity:  100,
			CreditID:  []string{"credit1"},
		},
		Quantity: 75,
	}

	// Convert to proto
	pbMb := originalMb.ToProto()
	require.NotNil(t, pbMb, "ToProto should return a non-nil pb.MatchedBid")

	// Convert back from proto
	convertedMb := &MatchedBid{}
	err := convertedMb.FromProto(pbMb)
	require.NoError(t, err, "FromProto should not error")

	// Verify round-trip preserves data and nil fields
	require.Equal(t, originalMb.Quantity, convertedMb.Quantity, "Quantity should match after round-trip")
	require.Nil(t, convertedMb.PrivatePrice, "PrivatePrice should remain nil after round-trip")
	require.Nil(t, convertedMb.PrivateMultiplier, "PrivateMultiplier should remain nil after round-trip")
}

// TestMatchedBidProtoMessageType verifies that ToProto returns correct proto type
func TestMatchedBidProtoMessageType(t *testing.T) {
	mb := &MatchedBid{
		Quantity: 42,
	}

	pbMb := mb.ToProto()
	require.NotNil(t, pbMb, "ToProto should not return nil")

	// Verify the proto message can be marshaled
	data, err := proto.Marshal(pbMb)
	require.NoError(t, err, "Proto message should be marshalable")
	require.NotEmpty(t, data, "Marshaled proto should not be empty")
}

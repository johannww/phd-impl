package auction

import (
	"testing"

	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestAuctionDataProtoRoundTrip verifies that AuctionData can be converted to proto and back
func TestAuctionDataProtoRoundTrip(t *testing.T) {
	// Create a test AuctionData with nested bids and companies
	originalAd := &AuctionData{
		AuctionID: 123,
		SellBids: []*bids.SellBid{
			{
				SellerID:  "seller1",
				Timestamp: "2024-03-24T10:00:00Z",
				Quantity:  100,
				CreditID:  []string{"credit1"},
				PrivatePrice: &bids.PrivatePrice{
					Price: 95,
					BidID: []string{"matched1"},
				},
			},
		},
		BuyBids: []*bids.BuyBid{
			{
				BuyerID:     "buyer1",
				Timestamp:   "2024-03-24T10:00:00Z",
				AskQuantity: 100,
				PrivateQuantity: &bids.PrivateQuantity{
					AskQuantity: 50,
					BidID:       []string{"bid1"},
				},
			},
		},
		ActivePolicies: []policies.Name{policies.Name("temperature"), policies.Name("wind")},
		CompaniesPvt: map[string]*companies.Company{
			"buyer1": {
				ID: "company1",
				Coordinate: &utils.Coordinate{
					Latitude:  -23.5505,
					Longitude: -46.6333,
				},
			},
		},
		Coupled: true,
	}

	// Convert to proto
	pbAd := originalAd.ToProto()
	require.NotNil(t, pbAd, "ToProto should return a non-nil pb.AuctionData")

	// Convert back from proto
	convertedAd := &AuctionData{}
	err := convertedAd.FromProto(pbAd)
	require.NoError(t, err, "FromProto should not error")

	// Verify round-trip preserves data
	require.Equal(t, originalAd.AuctionID, convertedAd.AuctionID, "AuctionID should match after round-trip")
	require.Equal(t, len(originalAd.SellBids), len(convertedAd.SellBids), "SellBids length should match")
	require.Equal(t, len(originalAd.BuyBids), len(convertedAd.BuyBids), "BuyBids length should match")
	require.Equal(t, originalAd.Coupled, convertedAd.Coupled, "Coupled should match")
	require.Equal(t, len(originalAd.ActivePolicies), len(convertedAd.ActivePolicies), "ActivePolicies length should match")
	require.Equal(t, len(originalAd.CompaniesPvt), len(convertedAd.CompaniesPvt), "CompaniesPvt size should match")

	// Verify nested bid data
	require.Equal(t, originalAd.SellBids[0].SellerID, convertedAd.SellBids[0].SellerID, "SellerID should match")
	require.Equal(t, originalAd.BuyBids[0].BuyerID, convertedAd.BuyBids[0].BuyerID, "BuyerID should match")

	// Verify companies data
	company, ok := convertedAd.CompaniesPvt["buyer1"]
	require.True(t, ok, "Company should be present in map")
	require.Equal(t, "company1", company.ID, "Company ID should match")
}

// TestAuctionDataProtoWithMinimalData verifies minimal AuctionData round-trip
func TestAuctionDataProtoWithMinimalData(t *testing.T) {
	// Create a minimal AuctionData
	originalAd := &AuctionData{
		AuctionID:      1,
		SellBids:       []*bids.SellBid{},
		BuyBids:        []*bids.BuyBid{},
		ActivePolicies: []policies.Name{},
		CompaniesPvt:   map[string]*companies.Company{},
		Coupled:        false,
	}

	// Convert to proto
	pbAd := originalAd.ToProto()
	require.NotNil(t, pbAd, "ToProto should return a non-nil pb.AuctionData")

	// Convert back from proto
	convertedAd := &AuctionData{}
	err := convertedAd.FromProto(pbAd)
	require.NoError(t, err, "FromProto should not error")

	// Verify round-trip
	require.Equal(t, originalAd.AuctionID, convertedAd.AuctionID, "AuctionID should match")
	require.Equal(t, 0, len(convertedAd.SellBids), "SellBids should be empty")
	require.Equal(t, 0, len(convertedAd.BuyBids), "BuyBids should be empty")
	require.Equal(t, false, convertedAd.Coupled, "Coupled should be false")
}

// TestAuctionDataProtoMessageType verifies that ToProto returns correct proto type
func TestAuctionDataProtoMessageType(t *testing.T) {
	ad := &AuctionData{
		AuctionID: 42,
	}

	pbAd := ad.ToProto()
	require.NotNil(t, pbAd, "ToProto should not return nil")

	// Verify the proto message can be marshaled
	data, err := proto.Marshal(pbAd)
	require.NoError(t, err, "Proto message should be marshalable")
	require.NotEmpty(t, data, "Marshaled proto should not be empty")
}

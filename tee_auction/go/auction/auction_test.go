package auction

import (
	"crypto/ed25519"
	"encoding/json"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	cc_auction "github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"github.com/johannww/phd-impl/chaincodes/common/state/serializer"
	"github.com/johannww/phd-impl/tee_auction/go/report"
	"github.com/stretchr/testify/require"
)

type MockHardwareReportFetcher struct {
	FakeReport []byte
}

func (m *MockHardwareReportFetcher) FetchReport(reportUserData [report.USER_DATA_SIZE]byte) ([]byte, error) {
	return m.FakeReport, nil
}

func TestRunTEEAuctionJSON(t *testing.T) {
	state.SetSerializer(serializer.NewJSONSerializer())

	fetcher := &MockHardwareReportFetcher{
		FakeReport: []byte("fake-report"),
	}

	// Load test data
	testDataBytes, err := os.ReadFile("testdata/testdata_tee.json")
	require.NoError(t, err)

	var serializedAD cc_auction.SerializedAuctionData
	err = json.Unmarshal(testDataBytes, &serializedAD)
	require.NoError(t, err)

	// Generate a keypair for the app
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	// Run the auction
	resultPub, resultPvt, err := RunTEEAuction(&serializedAD, priv, fetcher)
	require.NoError(t, err)

	// Verify the results
	require.NotNil(t, resultPub)
	require.NotNil(t, resultPvt)

	require.Equal(t, serializedAD.Sum, resultPub.ReceivedHash)
	require.Equal(t, []byte("fake-report"), resultPub.AmdReportBytes)
	require.NotEmpty(t, resultPub.AppSignature)

	// Verify AppSignature
	receivedResultAndBytesPub := append(resultPub.ResultBytes, resultPub.ReceivedHash...)
	valid := ed25519.Verify(pub, receivedResultAndBytesPub, resultPub.AppSignature)
	require.True(t, valid, "AppSignature should be valid")

	receivedResultAndBytesPvt := append(resultPvt.ResultBytes, resultPvt.ReceivedHash...)
	validPvt := ed25519.Verify(pub, receivedResultAndBytesPvt, resultPvt.AppSignature)
	require.True(t, validPvt, "AppSignature should be valid for private result")
}

func TestRunTEEAuctionProto(t *testing.T) {
	state.SetSerializer(serializer.NewProtoSerializer())

	fetcher := &MockHardwareReportFetcher{
		FakeReport: []byte("fake-report"),
	}

	testDataBytes, err := os.ReadFile("testdata/testdata_tee.pb")
	require.NoError(t, err)

	var pbSerializedAD pb.SerializedAuctionData
	err = proto.Unmarshal(testDataBytes, &pbSerializedAD)
	require.NoError(t, err)
	var serializedAD cc_auction.SerializedAuctionData
	err = serializedAD.FromProto(&pbSerializedAD)
	require.NoError(t, err)

	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	resultPub, resultPvt, err := RunTEEAuction(&serializedAD, priv, fetcher)
	require.NoError(t, err)

	require.NotNil(t, resultPub)
	require.NotNil(t, resultPvt)

	require.Equal(t, serializedAD.Sum, resultPub.ReceivedHash)
	require.Equal(t, []byte("fake-report"), resultPub.AmdReportBytes)
	require.NotEmpty(t, resultPub.AppSignature)

	receivedResultAndBytesPub := append(resultPub.ResultBytes, resultPub.ReceivedHash...)
	valid := ed25519.Verify(pub, receivedResultAndBytesPub, resultPub.AppSignature)
	require.True(t, valid, "AppSignature should be valid")

	receivedResultAndBytesPvt := append(resultPvt.ResultBytes, resultPvt.ReceivedHash...)
	validPvt := ed25519.Verify(pub, receivedResultAndBytesPvt, resultPvt.AppSignature)
	require.True(t, validPvt, "AppSignature should be valid for private result")
}

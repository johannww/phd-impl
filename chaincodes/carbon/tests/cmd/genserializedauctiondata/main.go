package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	utils_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/utils"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/state/mocks"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	pflag.Int("owners", 10, "number of owners")
	pflag.Int("chunks", 3, "number of chunks")
	pflag.Int("companies", 5, "number of companies")
	pflag.String("start", "2023-01-01T00:00:00Z", "start timestamp")
	pflag.String("end", "2023-01-01T00:30:00Z", "end timestamp")
	pflag.Duration("interval", 30*time.Second, "issue interval")
	pflag.StringP("output", "o", "testdata.json", "output file path")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	filePath := viper.GetString("output")

	testData := utils_test.GenDataWithBids(
		viper.GetInt("owners"),
		viper.GetInt("chunks"),
		viper.GetInt("companies"),
		viper.GetString("start"),
		viper.GetString("end"),
		viper.GetDuration("interval"),
	)

	testData.Policies = []policies.Name{policies.DISTANCE, policies.WIND_DIRECTION}

	stub := mocks.NewMockStub("carbon", nil)
	stub.MockTransactionStart("tx1")
	stub.Creator = (*testData.Identities)[identities.PolicySetter]
	testData.SaveToWorldState(stub)
	stub.MockTransactionEnd("tx1")

	stub.MockTransactionStart("set-auction-type")
	var auctionType auction.AuctionType = auction.AUCTION_COUPLED
	err := auctionType.ToWorldState(stub)
	stub.MockTransactionEnd("set-auction-type")
	panicOnError(err)

	stub.Creator = (*testData.Identities)[identities.PriceViewer]
	stub.MockTransactionStart("commit-data")
	auctionData := &auction.AuctionData{}
	auctionID, err := auction.IncrementAuctionID(stub)
	err = auctionData.RetrieveData(stub, testData.BidIssueLastTs)
	panicOnError(err)
	auctionData.AuctionID = auctionID
	serializedAD, err := auctionData.ToSerializedAuctionData()
	panicOnError(err)
	err = serializedAD.CommitmentToWorldState(stub, testData.BidIssueLastTs)
	panicOnError(err)
	stub.MockTransactionEnd("commit-data")

	stub.MockTransactionStart("retrieve-data")
	retrievedAD := &auction.AuctionData{}
	err = retrievedAD.RetrieveData(stub, testData.BidIssueLastTs)
	panicOnError(err)
	serializedAD, err = retrievedAD.ToSerializedAuctionData()
	panicOnError(err)
	serializedAD.CommitmentFromWorldState(stub, testData.BidIssueLastTs)
	panicOnError(err)
	stub.MockTransactionEnd("retrieve-data")

	bytes, err := json.MarshalIndent(serializedAD, "", "  ")
	panicOnError(err)

	err = os.WriteFile(filePath, bytes, os.ModePerm)
	panicOnError(err)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

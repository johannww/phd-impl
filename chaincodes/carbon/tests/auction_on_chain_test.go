package carbon_tests

import (
	"testing"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon"
	mocks "github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	utils_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/utils"
)

// TODO: Finish the test
func TestOnChainIndependentAuction(t *testing.T) {
	testData := utils_test.GenData(
		10, 3, 5,
		"2023-01-01T00:00:00Z",
		"2023-01-01T00:30:00Z", 30*time.Second)
	stub := mocks.NewMockStub("carbon", &carbon.Carbon{})

	stub.MockTransactionStart("tx1")
	testData.SaveToWorldState(stub)
	stub.MockTransactionEnd("tx1")

	// t.Fail()

}

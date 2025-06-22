package carbon_tests

import (
	"regexp"
	"testing"

	mocks "github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	"github.com/johannww/phd-impl/chaincodes/carbon/utils"
)

func TestUTCTimestamp(t *testing.T) {
	stub := mocks.NewMockStub("carbon", nil)
	stub.MockTransactionStart("tx1")
	ts, err := stub.GetTxTimestamp()
	if err != nil {
		t.Fatalf("Error getting timestamp: %v", err)
	}

	tsStr := utils.TimestampRFC3339UtcString(ts)

	rfc3339Regex := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`
	matched, err := regexp.MatchString(rfc3339Regex, tsStr)
	if err != nil || !matched {
		t.Fatalf("tsStr is not a valid RFC3339 string: %s", tsStr)
	}
}

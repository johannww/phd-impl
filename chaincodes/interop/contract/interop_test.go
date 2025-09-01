package contract

import (
	"testing"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	carbon "github.com/johannww/phd-impl/chaincodes/carbon/contract"
	credits "github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	"github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	"github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"

	"github.com/johannww/phd-impl/chaincodes/interop/lock"
	"github.com/stretchr/testify/require"
)

func TestLockedCredit(t *testing.T) {
	stub := mocks.NewMockStub("interop", nil)
	carbonCC, err := contractapi.NewChaincode(carbon.NewCarbonContract())
	require.NoError(t, err)
	carbonStub := mocks.NewMockStub(CARBON_CC_NAME, carbonCC)

	stub.MockPeerChaincode(CARBON_CC_NAME, carbonStub, "")

	carbonStub.MockTransactionStart("tx1")
	toBeLocked := createCreditOnCarbonCC(carbonStub, t)
	carbonStub.MockTransactionEnd("tx1")

	carbonStub.MockTransactionStart("tx2")
	err = credits.LockCredit(carbonStub, (*toBeLocked.GetID())[0], 0)
	require.NoError(t, err)
	carbonStub.MockTransactionEnd("tx2")

	stub.MockTransactionStart("tx1")
	mockIds := setup_test.SetupIdentities(stub)
	stub.Creator = mockIds[setup_test.REGULAR_ID]
	isLocked, err := lock.CreditIsLocked(stub, CARBON_CC_NAME, (*toBeLocked.GetID())[0])
	require.NoError(t, err)
	require.True(t, isLocked)
	stub.MockTransactionEnd("tx1")
}

func createCreditOnCarbonCC(carbonStub *mocks.MockStub,
	t *testing.T) *credits.MintCredit {
	mintedForChunk := &properties.PropertyChunk{
		PropertyID:  0,
		Coordinates: []properties.Coordinate{{Latitude: 1.0, Longitude: 2.0}},
	}
	err := mintedForChunk.ToWorldState(carbonStub)
	require.NoError(t, err)

	toBeLocked := &credits.MintCredit{
		Credit: credits.Credit{
			OwnerID:  "owner1",
			Quantity: 100,
			Chunk:    mintedForChunk,
			ChunkID:  (*mintedForChunk.GetID())[0],
		},
	}

	err = toBeLocked.ToWorldState(carbonStub)
	require.NoError(t, err)
	return toBeLocked
}

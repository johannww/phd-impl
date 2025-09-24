package contract

import (
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	carbon "github.com/johannww/phd-impl/chaincodes/carbon/contract"
	credits "github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	"github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	setup_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	carbon_utils "github.com/johannww/phd-impl/chaincodes/carbon/utils"

	"github.com/johannww/phd-impl/chaincodes/interop/lock"
	"github.com/johannww/phd-impl/chaincodes/interop/util"
	"github.com/stretchr/testify/require"
)

func TestLockedCredit(t *testing.T) {
	stub := mocks.NewMockStub("interop", nil)
	dstChainID := "mockDstChain"
	carbonCC, err := contractapi.NewChaincode(carbon.NewCarbonContract())
	require.NoError(t, err)
	carbonStub := mocks.NewMockStub(util.CARBON_CC_NAME, carbonCC)
	mockIds := setup_test.SetupIdentities(stub)

	stub.MockPeerChaincode(util.CARBON_CC_NAME, carbonStub, "")

	carbonStub.MockTransactionStart("tx1")
	toBeLocked := createCreditOnCarbonCC(carbonStub, t, mockIds)
	carbonStub.MockTransactionEnd("tx1")

	carbonStub.MockTransactionStart("tx2")
	lockID, err := credits.LockCredit(carbonStub, (*toBeLocked.GetID())[0], 0, dstChainID)
	require.NoError(t, err)
	carbonStub.MockTransactionEnd("tx2")

	stub.MockTransactionStart("tx1")
	stub.Creator = mockIds[identities.InteropRelayerAttr]
	isLocked, err := lock.CreditIsLocked(stub, util.CARBON_CC_NAME,
		(*toBeLocked.GetID())[0], lockID)
	require.NoError(t, err)
	require.True(t, isLocked)
	isLockedForChain, err := lock.CreditIsLockedForChainID(stub, util.CARBON_CC_NAME,
		(*toBeLocked.GetID())[0], lockID, dstChainID)
	require.NoError(t, err)
	require.True(t, isLockedForChain)
	stub.MockTransactionEnd("tx1")

	stub.MockTransactionStart("tx2")
	carbonStub.Creator = mockIds[identities.InteropRelayerAttr]
	stub.Creator = mockIds[identities.InteropRelayerAttr]
	err = lock.UnlockCredit(stub, util.CARBON_CC_NAME, (*toBeLocked.GetID())[0], lockID)
	require.NoError(t, err)
	stub.MockTransactionEnd("tx1")
}

func createCreditOnCarbonCC(carbonStub *mocks.MockStub,
	t *testing.T,
	mockIds map[string][]byte,
) *credits.MintCredit {
	carbonStub.Creator = mockIds[setup_test.REGULAR_ID]
	ownerID, err := cid.GetID(carbonStub)
	require.NoError(t, err)

	mintedForChunk := &properties.PropertyChunk{
		PropertyID:  0,
		Coordinates: []carbon_utils.Coordinate{{Latitude: 1.0, Longitude: 2.0}},
	}
	err = mintedForChunk.ToWorldState(carbonStub)
	require.NoError(t, err)

	toBeLocked := &credits.MintCredit{
		Credit: credits.Credit{
			OwnerID:  ownerID,
			Quantity: 100,
			Chunk:    mintedForChunk,
			ChunkID:  (*mintedForChunk.GetID())[0],
		},
	}

	err = toBeLocked.ToWorldState(carbonStub)
	require.NoError(t, err)
	return toBeLocked
}

package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	carbon "github.com/johannww/phd-impl/chaincodes/carbon/contract"
	credits "github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	setup_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/state/mocks"
	carbon_utils "github.com/johannww/phd-impl/chaincodes/common/utils"

	"github.com/johannww/phd-impl/chaincodes/interop/htlc"
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
	carbonStub.Creator = mockIds[setup_test.CREDIT_OWNER_ID]
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

func TestHTLCClaimFlow(t *testing.T) {
	stub := mocks.NewMockStub("interop", nil)
	carbonCC, err := contractapi.NewChaincode(carbon.NewCarbonContract())
	require.NoError(t, err)
	carbonStub := mocks.NewMockStub(util.CARBON_CC_NAME, carbonCC)
	stub.MockPeerChaincode(util.CARBON_CC_NAME, carbonStub, "")
	mockIds := setup_test.SetupIdentities(stub)

	carbonStub.MockTransactionStart("tx1")
	toBeLocked := createCreditOnCarbonCC(carbonStub, t, mockIds)
	carbonStub.MockTransactionEnd("tx1")

	carbonStub.MockTransactionStart("tx2")
	lockID, err := credits.LockCredit(carbonStub, (*toBeLocked.GetID())[0], 0, "mockDstChain")
	require.NoError(t, err)
	carbonStub.MockTransactionEnd("tx2")

	secret := "interop-htlc-secret"
	h := sha256.Sum256([]byte(secret))
	secretHash := hex.EncodeToString(h[:])

	stub.MockTransactionStart("tx3")
	stub.Creator = mockIds[setup_test.CREDIT_OWNER_ID]
	buyerID, err := cid.GetID(stub)
	require.NoError(t, err)
	htlcID, err := htlc.CreateHTLC(stub, util.CARBON_CC_NAME, (*toBeLocked.GetID())[0], "mockDstChain", &htlc.HTLC{
		SecretHash: secretHash,
		LockID:     lockID,
		BuyerID:    buyerID,
		Amount:     100,
		ValidUntil: time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339),
	})
	require.NoError(t, err)
	require.Equal(t, secretHash, htlcID)

	claimed, err := htlc.IsHTLCClaimed(stub, lockID)
	require.NoError(t, err)
	require.False(t, claimed)

	hashFromState, err := htlc.GetHTLCHashByLockID(stub, lockID)
	require.NoError(t, err)
	require.Equal(t, secretHash, hashFromState)

	_, err = htlc.GetHTLCPreimageByLockID(stub, lockID)
	require.Error(t, err)

	claimed, err = htlc.ClaimHTLC(stub, lockID, secret)
	require.NoError(t, err)
	require.True(t, claimed)

	isClaimedAfter, err := htlc.IsHTLCClaimed(stub, lockID)
	require.NoError(t, err)
	require.True(t, isClaimedAfter)

	preimage, err := htlc.GetHTLCPreimageByLockID(stub, lockID)
	require.NoError(t, err)
	require.Equal(t, secret, preimage)
	stub.MockTransactionEnd("tx3")
}

func TestHTLCUnlockExpiredFlow(t *testing.T) {
	stub := mocks.NewMockStub("interop", nil)
	carbonCC, err := contractapi.NewChaincode(carbon.NewCarbonContract())
	require.NoError(t, err)
	carbonStub := mocks.NewMockStub(util.CARBON_CC_NAME, carbonCC)
	stub.MockPeerChaincode(util.CARBON_CC_NAME, carbonStub, "")
	mockIds := setup_test.SetupIdentities(stub)

	carbonStub.MockTransactionStart("tx1")
	toBeLocked := createCreditOnCarbonCC(carbonStub, t, mockIds)
	carbonStub.MockTransactionEnd("tx1")

	carbonStub.MockTransactionStart("tx2")
	lockID, err := credits.LockCredit(carbonStub, (*toBeLocked.GetID())[0], 0, "mockDstChain")
	require.NoError(t, err)
	carbonStub.MockTransactionEnd("tx2")

	stub.MockTransactionStart("tx3")
	stub.Creator = mockIds[setup_test.CREDIT_OWNER_ID]
	buyerID, err := cid.GetID(stub)
	require.NoError(t, err)
	_, err = htlc.CreateHTLC(stub, util.CARBON_CC_NAME, (*toBeLocked.GetID())[0], "mockDstChain", &htlc.HTLC{
		SecretHash: "expired-secret-hash",
		LockID:     lockID,
		BuyerID:    buyerID,
		Amount:     100,
		ValidUntil: "2000-01-01T00:00:00Z",
	})
	require.NoError(t, err)

	stub.Creator = mockIds[identities.InteropRelayerAttr]
	carbonStub.Creator = mockIds[identities.InteropRelayerAttr]
	unlocked, err := htlc.UnlockHTLCExpired(stub, util.CARBON_CC_NAME, lockID)
	require.NoError(t, err)
	require.True(t, unlocked)

	isLocked, err := lock.CreditIsLocked(stub, util.CARBON_CC_NAME, (*toBeLocked.GetID())[0], lockID)
	require.NoError(t, err)
	require.False(t, isLocked)

	_, err = htlc.IsHTLCClaimed(stub, lockID)
	require.Error(t, err)
	stub.MockTransactionEnd("tx3")
}

func TestCreateHTLCDeniesRelayerWhenNotOwner(t *testing.T) {
	stub := mocks.NewMockStub("interop", nil)
	carbonCC, err := contractapi.NewChaincode(carbon.NewCarbonContract())
	require.NoError(t, err)
	carbonStub := mocks.NewMockStub(util.CARBON_CC_NAME, carbonCC)
	stub.MockPeerChaincode(util.CARBON_CC_NAME, carbonStub, "")
	mockIds := setup_test.SetupIdentities(stub)

	carbonStub.MockTransactionStart("tx1")
	carbonStub.Creator = mockIds[setup_test.CREDIT_OWNER_ID]
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
	carbonStub.MockTransactionEnd("tx1")

	carbonStub.MockTransactionStart("tx2")
	carbonStub.Creator = mockIds[setup_test.CREDIT_OWNER_ID]
	lockID, err := credits.LockCredit(carbonStub, (*toBeLocked.GetID())[0], 0, "mockDstChain")
	require.NoError(t, err)
	carbonStub.MockTransactionEnd("tx2")

	stub.MockTransactionStart("tx3")
	stub.Creator = mockIds[identities.InteropRelayerAttr]
	_, err = htlc.CreateHTLC(stub, util.CARBON_CC_NAME, (*toBeLocked.GetID())[0], "mockDstChain", &htlc.HTLC{
		SecretHash: "denied-secret-hash",
		LockID:     lockID,
		BuyerID:    "buyer-relayer",
		Amount:     100,
		ValidUntil: time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339),
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "only the credit owner can create HTLC")
	stub.MockTransactionEnd("tx3")
}

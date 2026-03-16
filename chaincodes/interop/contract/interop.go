package contract

import (
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/johannww/phd-impl/chaincodes/interop/htlc"
	"github.com/johannww/phd-impl/chaincodes/interop/lock"
	"github.com/johannww/phd-impl/chaincodes/interop/util"
)

type InteropContract struct {
	contractapi.Contract
}

func NewInteropContract() *InteropContract {
	return &InteropContract{}
}

func (c *InteropContract) InitLedger() {}

// LockCredit delegates credit lock operation to the carbon chaincode.
func (c *InteropContract) LockCredit(
	ctx contractapi.TransactionContextInterface,
	creditID []string,
	quantity int64,
	destChainID string,
) (string, error) {
	return lock.LockCredit(ctx.GetStub(), util.CARBON_CC_NAME, creditID, quantity, destChainID)
}

func (c *InteropContract) CreditIsLocked(
	ctx contractapi.TransactionContextInterface,
	creditID []string,
	lockID string,
) (bool, error) {
	return lock.CreditIsLocked(ctx.GetStub(), util.CARBON_CC_NAME, creditID, lockID)
}

func (c *InteropContract) CreditIsLockedForChainID(
	ctx contractapi.TransactionContextInterface,
	creditID []string,
	lockID string,
	destChainID string,
) (bool, error) {
	return lock.CreditIsLockedForChainID(ctx.GetStub(), util.CARBON_CC_NAME, creditID, lockID, destChainID)
}

func (c *InteropContract) UnlockCredit(
	ctx contractapi.TransactionContextInterface,
	creditID []string,
	lockID string,
) (bool, error) {
	if err := lock.UnlockCredit(ctx.GetStub(), util.CARBON_CC_NAME, creditID, lockID); err != nil {
		return false, err
	}
	return true, nil
}

func (c *InteropContract) CreateHTLC(
	ctx contractapi.TransactionContextInterface,
	creditID []string,
	lockID string,
	secretHash string,
	recipientID string,
	validUntil string,
	destChainID string,
	amount int,
) (string, error) {
	record := &htlc.HTLC{
		SecretHash: secretHash,
		LockID:     lockID,
		BuyerID:    recipientID,
		Amount:     amount,
		ValidUntil: validUntil,
	}
	return htlc.CreateHTLC(ctx.GetStub(), util.CARBON_CC_NAME, creditID, destChainID, record)
}

func (c *InteropContract) ClaimHTLC(
	ctx contractapi.TransactionContextInterface,
	lockID string,
	secret string,
) (bool, error) {
	return htlc.ClaimHTLC(ctx.GetStub(), lockID, secret)
}

func (c *InteropContract) IsHTLCClaimed(
	ctx contractapi.TransactionContextInterface,
	lockID string,
) (bool, error) {
	return htlc.IsHTLCClaimed(ctx.GetStub(), lockID)
}

func (c *InteropContract) GetHTLCHashByLockID(
	ctx contractapi.TransactionContextInterface,
	lockID string,
) (string, error) {
	return htlc.GetHTLCHashByLockID(ctx.GetStub(), lockID)
}

func (c *InteropContract) GetHTLCPreimageByLockID(
	ctx contractapi.TransactionContextInterface,
	lockID string,
) (string, error) {
	return htlc.GetHTLCPreimageByLockID(ctx.GetStub(), lockID)
}

func (c *InteropContract) UnlockHTLCExpired(
	ctx contractapi.TransactionContextInterface,
	lockID string,
) (bool, error) {
	return htlc.UnlockHTLCExpired(ctx.GetStub(), util.CARBON_CC_NAME, lockID)
}

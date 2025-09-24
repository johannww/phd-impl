package contract

import (
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
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

// TODOHP:  continue here
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
	return lock.CreditIsLocked(ctx.GetStub(), util.CARBON_CC_NAME, creditID, lockID)
}

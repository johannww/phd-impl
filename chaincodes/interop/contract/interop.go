package contract

import (
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/johannww/phd-impl/chaincodes/interop/lock"
)

const (
	CARBON_CC_NAME = "carbon"
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
	return lock.CreditIsLocked(ctx.GetStub(), CARBON_CC_NAME, creditID, lockID)
}

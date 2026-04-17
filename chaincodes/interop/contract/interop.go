package contract

import (
	"time"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/johannww/phd-impl/chaincodes/interop/htlc"
	"github.com/johannww/phd-impl/chaincodes/interop/lock"
	"github.com/johannww/phd-impl/chaincodes/interop/util"
)

type InteropContract struct {
	contractapi.Contract
	metrics TxMetrics
}

func NewInteropContract() *InteropContract {
	return &InteropContract{
		metrics: NewPrometheusTxMetrics(),
	}
}

func (c *InteropContract) InitLedger() {}

func (c *InteropContract) withMetricsErr(txName string, fn func() error) error {
	start := time.Now()
	err := fn()
	if c.metrics != nil {
		c.metrics.Observe(txName, err == nil, time.Since(start))
	}
	return err
}

func (c *InteropContract) withMetricsStringResult(txName string, fn func() (string, error)) (out string, err error) {
	start := time.Now()
	defer func() {
		if c.metrics != nil {
			c.metrics.Observe(txName, err == nil, time.Since(start))
		}
	}()
	return fn()
}

func (c *InteropContract) withMetricsBoolResult(txName string, fn func() (bool, error)) (out bool, err error) {
	start := time.Now()
	defer func() {
		if c.metrics != nil {
			c.metrics.Observe(txName, err == nil, time.Since(start))
		}
	}()
	return fn()
}

// LockCredit delegates credit lock operation to the carbon chaincode.
func (c *InteropContract) LockCredit(
	ctx contractapi.TransactionContextInterface,
	creditID []string,
	quantity int64,
	destChainID string,
) (string, error) {
	return c.withMetricsStringResult("LockCredit", func() (string, error) {
		return lock.LockCredit(ctx.GetStub(), util.CARBON_CC_NAME, creditID, quantity, destChainID)
	})
}

func (c *InteropContract) CreditIsLocked(
	ctx contractapi.TransactionContextInterface,
	creditID []string,
	lockID string,
) (bool, error) {
	return c.withMetricsBoolResult("CreditIsLocked", func() (bool, error) {
		return lock.CreditIsLocked(ctx.GetStub(), util.CARBON_CC_NAME, creditID, lockID)
	})
}

func (c *InteropContract) CreditIsLockedForChainID(
	ctx contractapi.TransactionContextInterface,
	creditID []string,
	lockID string,
	destChainID string,
) (bool, error) {
	return c.withMetricsBoolResult("CreditIsLockedForChainID", func() (bool, error) {
		return lock.CreditIsLockedForChainID(ctx.GetStub(), util.CARBON_CC_NAME, creditID, lockID, destChainID)
	})
}

func (c *InteropContract) UnlockCredit(
	ctx contractapi.TransactionContextInterface,
	creditID []string,
	lockID string,
) (bool, error) {
	return c.withMetricsBoolResult("UnlockCredit", func() (bool, error) {
		if err := lock.UnlockCredit(ctx.GetStub(), util.CARBON_CC_NAME, creditID, lockID); err != nil {
			return false, err
		}
		return true, nil
	})
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
	return c.withMetricsStringResult("CreateHTLC", func() (string, error) {
		record := &htlc.HTLC{
			SecretHash: secretHash,
			LockID:     lockID,
			BuyerID:    recipientID,
			Amount:     amount,
			ValidUntil: validUntil,
		}
		return htlc.CreateHTLC(ctx.GetStub(), util.CARBON_CC_NAME, creditID, destChainID, record)
	})
}

func (c *InteropContract) ClaimHTLC(
	ctx contractapi.TransactionContextInterface,
	lockID string,
	secret string,
) (bool, error) {
	return c.withMetricsBoolResult("ClaimHTLC", func() (bool, error) {
		return htlc.ClaimHTLC(ctx.GetStub(), lockID, secret)
	})
}

func (c *InteropContract) IsHTLCClaimed(
	ctx contractapi.TransactionContextInterface,
	lockID string,
) (bool, error) {
	return c.withMetricsBoolResult("IsHTLCClaimed", func() (bool, error) {
		return htlc.IsHTLCClaimed(ctx.GetStub(), lockID)
	})
}

func (c *InteropContract) GetHTLCHashByLockID(
	ctx contractapi.TransactionContextInterface,
	lockID string,
) (string, error) {
	return c.withMetricsStringResult("GetHTLCHashByLockID", func() (string, error) {
		return htlc.GetHTLCHashByLockID(ctx.GetStub(), lockID)
	})
}

func (c *InteropContract) GetHTLCPreimageByLockID(
	ctx contractapi.TransactionContextInterface,
	lockID string,
) (string, error) {
	return c.withMetricsStringResult("GetHTLCPreimageByLockID", func() (string, error) {
		return htlc.GetHTLCPreimageByLockID(ctx.GetStub(), lockID)
	})
}

func (c *InteropContract) UnlockHTLCExpired(
	ctx contractapi.TransactionContextInterface,
	lockID string,
) (bool, error) {
	return c.withMetricsBoolResult("UnlockHTLCExpired", func() (bool, error) {
		return htlc.UnlockHTLCExpired(ctx.GetStub(), util.CARBON_CC_NAME, lockID)
	})
}

package contract

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/data/registry"
	"github.com/johannww/phd-impl/chaincodes/carbon/payment"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	"github.com/johannww/phd-impl/chaincodes/carbon/semaphore"
	"github.com/johannww/phd-impl/chaincodes/carbon/tee"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"github.com/johannww/phd-impl/chaincodes/common/state/serializer"
	tee_auction "github.com/johannww/phd-impl/tee_auction/go/auction"
)

func init() {
	state.SetSerializer(serializer.NewProtoSerializer())
}

const (
	SERIALIZED_RESULT_TEE_PVT_KEY = "serializedResultPvt"
)

type CarbonContract struct {
	contractapi.Contract
	pApplier *policies.PolicyApplierImpl
	// TODOHP: review metrics
	metrics TxMetrics
}

func NewCarbonContract() *CarbonContract {
	carbonContract := new(CarbonContract)
	carbonContract.pApplier = policies.NewPolicyApplier()
	carbonContract.metrics = NewPrometheusTxMetrics()
	return carbonContract
}

func (c *CarbonContract) withMetricsErr(txName string, fn func() error) error {
	start := time.Now()
	err := fn()
	if c.metrics != nil {
		c.metrics.Observe(txName, err == nil, time.Since(start))
	}
	return err
}

func (c *CarbonContract) withMetricsStringResult(txName string, fn func() (string, error)) (out string, err error) {
	start := time.Now()
	defer func() {
		if c.metrics != nil {
			c.metrics.Observe(txName, err == nil, time.Since(start))
		}
	}()
	return fn()
}

func (c *CarbonContract) withMetricsSerializedAuctionDataResult(
	txName string,
	fn func() (*auction.SerializedAuctionData, error),
) (out *auction.SerializedAuctionData, err error) {
	start := time.Now()
	defer func() {
		if c.metrics != nil {
			c.metrics.Observe(txName, err == nil, time.Since(start))
		}
	}()
	return fn()
}

func (c *CarbonContract) withMetricsMatchedBidsResult(
	txName string,
	fn func() ([]*bids.MatchedBid, error),
) (out []*bids.MatchedBid, err error) {
	start := time.Now()
	defer func() {
		if c.metrics != nil {
			c.metrics.Observe(txName, err == nil, time.Since(start))
		}
	}()
	return fn()
}

func (c *CarbonContract) withMetricsBoolValue(txName string, fn func() bool) bool {
	start := time.Now()
	out := fn()
	if c.metrics != nil {
		c.metrics.Observe(txName, true, time.Since(start))
	}
	return out
}

func (c *CarbonContract) withMetricsStringValue(txName string, fn func() string) string {
	start := time.Now()
	out := fn()
	if c.metrics != nil {
		c.metrics.Observe(txName, true, time.Since(start))
	}
	return out
}

// CreateBuyBidPublicQuantity expects `quantity` and transient `price` to be
// provided as scaled integers (see common.QUANTITY_SCALE and
// bids.PRICE_SCALE).
func (c *CarbonContract) CreateBuyBidPublicQuantity(ctx contractapi.TransactionContextInterface, quantity int64) error {
	return c.withMetricsErr("CreateBuyBidPublicQuantity", func() error {
		stub := ctx.GetStub()
		err := bids.PublishBuyBidWithPublicQuanitity(stub, quantity)
		return err
	})
}

// CreateBuyBidPrivateQuantity expects transient `price` and private
// `quantity` to be provided as scaled integers (see bids.PRICE_SCALE and
// common.QUANTITY_SCALE).
func (c *CarbonContract) CreateBuyBidPrivateQuantity(ctx contractapi.TransactionContextInterface) error {
	return c.withMetricsErr("CreateBuyBidPrivateQuantity", func() error {
		stub := ctx.GetStub()
		err := bids.PublishBuyBidWithPrivateQuantity(stub)
		return err
	})
}

// CreateSellBid expects `quantity` and transient `price` to be provided as
// scaled integers (see common.QUANTITY_SCALE and bids.PRICE_SCALE).
func (c *CarbonContract) CreateSellBidFromCredit(
	ctx contractapi.TransactionContextInterface,
	quantity int64,
	creditID []string,
) error {
	return c.withMetricsErr("CreateSellBidFromCredit", func() error {
		return bids.PublishSellBidFromCredit(ctx.GetStub(), quantity, creditID)
	})
}

func (c *CarbonContract) CreateSellBidFromWallet(
	ctx contractapi.TransactionContextInterface,
	quantity int64,
	creditID []string,
) error {
	return c.withMetricsErr("CreateSellBidFromWallet", func() error {
		return bids.PublishSellBidFromWallet(ctx.GetStub(), quantity)
	})
}

// TODO: implement
func (c *CarbonContract) PublishData(ctx contractapi.TransactionContextInterface) error {
	return c.withMetricsErr("PublishData", func() error { return nil })
}

// MintQuantityCreditForChunk expects `quantity` to be a scaled integer
// according to common.QUANTITY_SCALE.
func (c *CarbonContract) MintQuantityCreditForChunk(
	ctx contractapi.TransactionContextInterface,
	propertyID []string,
	chunkID []string,
	quantity int64,
	timestampRFC3339 string,
) (*credits.MintCredit, error) {
	var mc *credits.MintCredit
	err := c.withMetricsErr("MintQuantityCreditForChunk", func() error {
		var err error
		mc, err = credits.MintQuantityCreditForChunk(ctx.GetStub(), propertyID, chunkID, quantity, timestampRFC3339)
		return err
	})
	return mc, err
}

func (c *CarbonContract) MintQuantityCreditsForProperty(
	ctx contractapi.TransactionContextInterface,
	propertyID []string,
	quantity int64,
	timestampRFC3339 string,
) ([]*credits.MintCredit, error) {
	var mcs []*credits.MintCredit
	err := c.withMetricsErr("MintQuantityCreditsForProperty", func() error {
		var err error
		mcs, err = credits.MintQuantityCreditsForProperty(ctx.GetStub(), propertyID, quantity, timestampRFC3339)
		return err
	})
	return mcs, err
}

// TransferFromMintToWallet deducts quantity from a MintCredit and credits the caller's fungible credit wallet.
// `mintCreditID` is the composite key attributes for the MintCredit and `quantity` is a scaled integer.
func (c *CarbonContract) TransferFromMintToWallet(ctx contractapi.TransactionContextInterface, mintCreditID []string, quantity int64) error {
	return c.withMetricsErr("TransferFromMintToWallet", func() error {
		return credits.TransferFromMintToWallet(ctx.GetStub(), mintCreditID, quantity)
	})
}

func (c *CarbonContract) MintEstimatedCreditsForProperty(
	ctx contractapi.TransactionContextInterface,
	propertyID []string,
	intervalStartRFC3339 string,
	intervalEndRFC3339 string,
) ([]*credits.MintCredit, error) {
	var mcs []*credits.MintCredit
	err := c.withMetricsErr("MintEstimatedCreditsForProperty", func() error {
		var err error
		mcs, err = credits.MintEstimatedCreditsForProperty(ctx.GetStub(), propertyID, intervalStartRFC3339, intervalEndRFC3339)
		return err
	})
	return mcs, err
}

// BurnNominalQuantity expects `burnQuantity` to be a scaled integer
// according to common.QUANTITY_SCALE.
func (c *CarbonContract) BurnNominalQuantity(ctx contractapi.TransactionContextInterface, mintCreditID []string, burnQuantity int64) (*credits.BurnCredit, error) {
	var bc *credits.BurnCredit
	err := c.withMetricsErr("BurnNominalQuantity", func() error {
		var err error
		bc, err = credits.BurnNominalQuantity(ctx.GetStub(), mintCreditID, burnQuantity)
		return err
	})
	return bc, err
}

func (c *CarbonContract) ApplyBurnMultipliers(ctx contractapi.TransactionContextInterface, burnCreditID []string) error {
	return c.withMetricsErr("ApplyBurnMultipliers", func() error {
		return credits.ApplyBurnMultipliers(ctx.GetStub(), burnCreditID)
	})
}

// LockAuctionSemaphore locks the auction semaphore to prevent concurrent auctions.
// Only identities with the TEEConfigurer attribute are authorized.
func (c *CarbonContract) LockAuctionSemaphore(ctx contractapi.TransactionContextInterface) error {
	return c.withMetricsErr("LockAuctionSemaphore", func() error {
		return semaphore.LockAuction(ctx.GetStub())
	})
}

// UnlockAuctionSemaphore releases the auction semaphore.
// Only identities with the TEEConfigurer attribute are authorized.
func (c *CarbonContract) UnlockAuctionSemaphore(ctx contractapi.TransactionContextInterface) error {
	return c.withMetricsErr("UnlockAuctionSemaphore", func() error {
		return semaphore.UnlockAuction(ctx.GetStub())
	})
}

func (c *CarbonContract) SetAuctionType(
	ctx contractapi.TransactionContextInterface,
	auctionType auction.AuctionType,
) error {
	return c.withMetricsErr("SetAuctionType", func() error {
		return auctionType.ToWorldState(ctx.GetStub())
	})
}

func (c *CarbonContract) PublishExpectedTEECCEPolicy(ctx contractapi.TransactionContextInterface, base64CcePolicy string) error {
	return c.withMetricsErr("PublishExpectedTEECCEPolicy", func() error {
		return tee.ExpectedCCEPolicyToWorldState(ctx.GetStub(), base64CcePolicy)
	})
}

// PublishInitialTEEReport stores the initial TEE report containing the
// confidential container's public key for communication and verification
func (c *CarbonContract) PublishInitialTEEReport(ctx contractapi.TransactionContextInterface, reportJsonBytes []byte) error {
	return c.withMetricsErr("PublishInitialTEEReport", func() error {
		return tee.InitialReportToWorldState(ctx.GetStub(), reportJsonBytes)
	})
}

// CommitDataForTEEAuction commits the auction data to the world state.
// The end timestamp is taken from the auction lock timestamp set by
// LockAuctionSemaphore — callers no longer supply it.
// Since it is a write operation, private data is not shared with the world state.
// To retrieve the auction data, use RetrieveDataForTEEAuction instead.
func (c *CarbonContract) CommitDataForTEEAuction(ctx contractapi.TransactionContextInterface) error {
	return c.withMetricsErr("CommitDataForTEEAuction", func() error {
		if cid.AssertAttributeValue(ctx.GetStub(), identities.PriceViewer, "true") != nil {
			return fmt.Errorf("caller does not have the %s attribute, "+
				"which is required to commit auction data", identities.PriceViewer)
		}

		endRFC339Timestamp, err := semaphore.GetLockTimestamp(ctx.GetStub())
		if err != nil {
			return err
		}
		if endRFC339Timestamp == "" {
			return fmt.Errorf("cannot commit data: auction semaphore must be locked first")
		}

		auctionID, err := auction.IncrementAuctionID(ctx.GetStub())
		if err != nil {
			return fmt.Errorf("could not increment auction ID: %v", err)
		}

		auctionData := &auction.AuctionData{}
		err = auctionData.RetrieveData(ctx.GetStub(), endRFC339Timestamp)
		if err != nil {
			return fmt.Errorf("could not retrieve auction data: %v", err)
		}

		auctionData.AuctionID = auctionID

		serializedAD, err := auctionData.ToSerializedAuctionData()
		if err != nil {
			return fmt.Errorf("could not serialize auction data: %v", err)
		}

		err = serializedAD.CommitmentToWorldState(ctx.GetStub(), endRFC339Timestamp)
		return err
	})
}

// RetrieveDataForTEEAuction retrieves the auction data from the world state.
// The end timestamp is read from the auction semaphore — callers no longer supply it.
// WARN: CALL THIS AS READ-ONLY OPERATION not to expose private data.
func (c *CarbonContract) RetrieveDataForTEEAuction(ctx contractapi.TransactionContextInterface) (*auction.SerializedAuctionData, error) {
	return c.withMetricsSerializedAuctionDataResult("RetrieveDataForTEEAuction", func() (*auction.SerializedAuctionData, error) {
		endRFC339Timestamp, err := semaphore.GetLockTimestamp(ctx.GetStub())
		if err != nil {
			return nil, err
		}
		if endRFC339Timestamp == "" {
			return nil, fmt.Errorf("cannot retrieve auction data: auction semaphore is not locked")
		}

		auctionData := &auction.AuctionData{}
		err = auctionData.RetrieveData(ctx.GetStub(), endRFC339Timestamp)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve auction data: %v", err)
		}

		serializedAD, err := auctionData.ToSerializedAuctionData()
		if err != nil {
			return nil, fmt.Errorf("could not serialize auction data: %v", err)
		}

		// Verify the commitment to the world state
		err = serializedAD.CommitmentFromWorldState(ctx.GetStub(), endRFC339Timestamp)
		if !serializedAD.ValidateHash() {
			return nil, fmt.Errorf("auction data hash does not match the commitment in the world state")
		}

		return serializedAD, err
	})
}

func (c *CarbonContract) PublishTEEAuctionResults(
	ctx contractapi.TransactionContextInterface,
	serializedResultsPub *tee_auction.SerializedAuctionResultTEE,
) error {
	return c.withMetricsErr("PublishTEEAuctionResults", func() error {
		transient, err := ctx.GetStub().GetTransient()
		if err != nil {
			return fmt.Errorf("could not get transient: %v", err)
		}
		serializedResultPvtBytes := transient[SERIALIZED_RESULT_TEE_PVT_KEY]
		if len(serializedResultPvtBytes) == 0 {
			return fmt.Errorf("serialized result pvt not found in transient")
		}

		var serializedResultsPvt tee_auction.SerializedAuctionResultTEE
		err = json.Unmarshal(serializedResultPvtBytes, &serializedResultsPvt)
		if err != nil {
			return fmt.Errorf("could not unmarshal serialized result pvt: %v", err)
		}

		err1 := tee.VerifyTEEResult(ctx.GetStub(), serializedResultsPub)
		err2 := tee.VerifyTEEResult(ctx.GetStub(), &serializedResultsPvt)
		if err1 != nil || err2 != nil {
			return fmt.Errorf("could not verify TEE auction result: %v, %v", err1, err2)
		}

		locked, err := semaphore.IsLocked(ctx.GetStub())
		if err != nil {
			return err
		}
		if !locked {
			return fmt.Errorf("cannot publish results: auction is not locked")
		}

		err = auction.ProcessOffChainAuctionResult(ctx.GetStub(),
			serializedResultsPub.ResultBytes,
			serializedResultsPvt.ResultBytes)
		if err != nil {
			return fmt.Errorf("could not process off-chain auction result: %v", err)
		}

		return nil
	})
}

func (c *CarbonContract) CheckCredAttr(ctx contractapi.TransactionContextInterface, attrName string) (string, error) {
	return c.withMetricsStringResult("CheckCredAttr", func() (string, error) {
		stub := ctx.GetStub()
		attrValue, found, err := cid.GetAttributeValue(stub, attrName)
		if err != nil {
			return "", err
		}

		if !found {
			return "", fmt.Errorf("Attribute '%s' not found", attrName)
		}

		return attrValue, nil
	})
}

// SetActivePolicies sets the list of active policies in the world state
func (c *CarbonContract) SetActivePolicies(ctx contractapi.TransactionContextInterface, activePolicies []policies.Name) error {
	return c.withMetricsErr("SetActivePolicies", func() error {
		stub := ctx.GetStub()
		err := c.pApplier.SetActivePolicies(stub, activePolicies)
		return err
	})
}

// AppendActivePolicy adds a new policy to the list of active policies in the world state
func (c *CarbonContract) AppendActivePolicy(ctx contractapi.TransactionContextInterface, policy policies.Name) error {
	return c.withMetricsErr("AppendActivePolicy", func() error {
		stub := ctx.GetStub()
		err := c.pApplier.AppendActivePolicy(stub, policies.Name(policy))
		return err
	})
}

// DeleteActivePolicy removes a policy from the list of active policies in the world state
func (c *CarbonContract) DeleteActivePolicy(ctx contractapi.TransactionContextInterface, policy policies.Name) error {
	return c.withMetricsErr("DeleteActivePolicy", func() error {
		stub := ctx.GetStub()
		err := policies.DeleteActivePolicy(stub, policies.Name(policy))
		return err
	})
}

func (c *CarbonContract) RegisterProperty(ctx contractapi.TransactionContextInterface, property properties.Property) error {
	return c.withMetricsErr("RegisterProperty", func() error {
		return properties.RegisterProperty(ctx.GetStub(), &property)
	})
}

func (c *CarbonContract) RegisterCompany(ctx contractapi.TransactionContextInterface, company companies.Company) error {
	return c.withMetricsErr("RegisterCompany", func() error {
		return companies.RegisterCompany(ctx.GetStub(), &company)
	})
}

func (c *CarbonContract) UpdateSellerAndBuyerVirtualTokenWallets(ctx contractapi.TransactionContextInterface, ownerID string, quantity int64) error {
	return c.withMetricsErr("UpdateSellerAndBuyerVirtualTokenWallets", func() error {
		return payment.UpdateVirtualTokenWallet(ctx.GetStub(), ownerID, quantity)
	})
}

func (c *CarbonContract) MintVirtualToken(ctx contractapi.TransactionContextInterface, ownerID string, quantity int64) (*payment.VirtualTokenWallet, error) {
	var wallet *payment.VirtualTokenWallet
	err := c.withMetricsErr("MintVirtualToken", func() error {
		var err error
		wallet, err = payment.MintVirtualToken(ctx.GetStub(), ownerID, quantity)
		return err
	})
	return wallet, err
}

// GetActivePolicies retrieves the list of caller's matched bids.
// It uses cid.GetID function to determine the id
func (c *CarbonContract) GetCallerMatchedBids(ctx contractapi.TransactionContextInterface) ([]*bids.MatchedBid, error) {
	return c.withMetricsMatchedBidsResult("GetCallerMatchedBids", func() ([]*bids.MatchedBid, error) {
		return bids.GetCallerMatchedBids(ctx.GetStub())
	})
}

func (c *CarbonContract) CreditIsLocked(ctx contractapi.TransactionContextInterface, creditID []string, lockID string) bool {
	return c.withMetricsBoolValue("CreditIsLocked", func() bool {
		return credits.CreditIsLocked(ctx.GetStub(), creditID, lockID)
	})
}

func (c *CarbonContract) ChainIDCreditIsLockedFor(ctx contractapi.TransactionContextInterface, creditID []string, lockID string) string {
	return c.withMetricsStringValue("ChainIDCreditIsLockedFor", func() string {
		return credits.ChainIDCreditIsLockedFor(ctx.GetStub(), creditID, lockID)
	})
}

func (c *CarbonContract) LockCredit(ctx contractapi.TransactionContextInterface,
	creditID []string,
	quantity int64,
	destChainID string,
) (lockID string, err error) {
	return c.withMetricsStringResult("LockCredit", func() (string, error) {
		return credits.LockCredit(ctx.GetStub(), creditID, quantity, destChainID)
	})
}

func (c *CarbonContract) UnlockCredit(ctx contractapi.TransactionContextInterface, creditID []string, lockID string) error {
	return c.withMetricsErr("UnlockCredit", func() error {
		return credits.UnlockCredit(ctx.GetStub(), creditID, lockID)
	})
}

func (c *CarbonContract) AddTrustedProvider(
	ctx contractapi.TransactionContextInterface,
	name string,
	baseURL string,
	rootCAPEM string,
) error {
	return c.withMetricsErr("AddTrustedProvider", func() error {
		return registry.AddTrustedProvider(ctx.GetStub(), name, baseURL, []byte(rootCAPEM))
	})
}

func (c *CarbonContract) RefreshRegistryDataForProperty(
	ctx contractapi.TransactionContextInterface,
	providerName string,
	registryPropID string,
) (*registry.RegistrySummary, error) {
	var summary *registry.RegistrySummary
	err := c.withMetricsErr("RefreshRegistryData", func() error {
		var err error
		summary, err = registry.RefreshRegistryData(ctx.GetStub(), providerName, registryPropID)
		return err
	})
	return summary, err
}

// GetAvailableCreditsByOwner returns all credits owned by the specified owner.
func (c *CarbonContract) GetAvailableCreditsByOwner(ctx contractapi.TransactionContextInterface, ownerID string) ([]*credits.MintCredit, error) {
	var mcs []*credits.MintCredit
	err := c.withMetricsErr("GetCreditsByOwner", func() error {
		var err error
		mcs, err = credits.GetAvailableCreditsByOwner(ctx.GetStub(), ownerID)
		return err
	})
	return mcs, err
}

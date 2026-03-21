package credits

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/state"
)

const (
	CREDIT_WALLET_PREFIX = "creditWallet"
)

// TODO: Credit wallet will generate MVCC read conflict if multiple transactions
// mint credits and add to the wallet. Evaluate that later.
// CreditWallet is for networks with fungible credits
type CreditWallet struct {
	OwnerID  string `json:"owner"`
	Quantity int64  `json:"quantity"`
}

var _ state.WorldStateManager = (*CreditWallet)(nil)

func (cw *CreditWallet) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	// NOTE: We are using the same private collection as the one used for the private price
	// err := state.GetPvtDataWithCompositeKey(stub, CREDIT_WALLET_PREFIX, keyAttributes, state.BIDS_PVT_DATA_COLLECTION, cw)
	err := state.GetStateWithCompositeKey(stub, CREDIT_WALLET_PREFIX, keyAttributes, cw)
	if err != nil {
		return err
	}
	return nil
}

func (cw *CreditWallet) ToWorldState(stub shim.ChaincodeStubInterface) error {
	// NOTE: We are using the same private collection as the one used for the private price
	// return state.PutPvtDataWithCompositeKey(
	// 	stub, CREDIT_WALLET_PREFIX,
	// 	(*cw.GetID())[0],
	// 	state.BIDS_PVT_DATA_COLLECTION, cw)
	return state.PutStateWithCompositeKey(
		stub, CREDIT_WALLET_PREFIX,
		cw.GetID(), cw)
}

func (cw *CreditWallet) GetID() *[][]string {
	return &[][]string{{cw.OwnerID}}
}

// TransferFromMintToWallet deducts quantity from the specified MintCredit and
// adds it to the specified owner's CreditWallet.
func TransferFromMintToWallet(
	stub shim.ChaincodeStubInterface,
	mintCreditID []string,
	quantity int64) error {
	mc := &MintCredit{}
	if err := mc.FromWorldState(stub, mintCreditID); err != nil {
		return fmt.Errorf("could not load mint credit: %v", err)
	}

	// ensure callerID is the owner of the mint credit
	callerID := identities.GetID(stub)
	if mc.OwnerID != callerID {
		return fmt.Errorf("caller %s is not the owner of the mint credit %v", callerID, mintCreditID)
	}

	if mc.Quantity < quantity {
		return fmt.Errorf("mint credit has insufficient quantity: have %d, need %d", mc.Quantity, quantity)
	}

	mc.Quantity -= quantity
	if err := mc.ToWorldState(stub); err != nil {
		return fmt.Errorf("could not persist updated mint credit: %v", err)
	}

	ownerID := identities.GetID(stub)
	cw := &CreditWallet{OwnerID: ownerID}
	if err := cw.FromWorldState(stub, []string{ownerID}); err != nil {
		cw = &CreditWallet{OwnerID: ownerID, Quantity: 0}
	}
	cw.Quantity += quantity
	if err := cw.ToWorldState(stub); err != nil {
		return fmt.Errorf("could not persist credit wallet: %v", err)
	}

	return nil
}

package credits

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
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

package payment

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	VIRTUAL_TOKEN_WALLET_PREFIX = "virtualTokenWallet"
)

// VirtualTokenWallet is a wallet that stores tokens corresponding to a
// fiat currency. It is used to pay for the carbon credits.
type VirtualTokenWallet struct {
	OwnerID  string  `json:"owner"`
	Quantity float64 `json:"quantity"`
}

var _ state.WorldStateManager = (*VirtualTokenWallet)(nil)

func MintVirtualToken(stub shim.ChaincodeStubInterface, ownerID string, quantity float64) (*VirtualTokenWallet, error) {
	tokenWallet := &VirtualTokenWallet{OwnerID: ownerID}
	err := tokenWallet.FromWorldState(stub, []string{ownerID})
	if cid.AssertAttributeValue(stub, identities.VirtualTokenMinter, "true") != nil {
		return nil, fmt.Errorf("only identities with the attribute %s can mint virtual tokens", identities.VirtualTokenMinter)
	}

	// tokenWallet.FromWorldState
	if err != nil {
		return &VirtualTokenWallet{
			OwnerID:  ownerID,
			Quantity: quantity,
		}, nil
	}

	tokenWallet.Quantity += quantity
	err = tokenWallet.ToWorldState(stub)
	if err != nil {
		return nil, err
	}
	return tokenWallet, nil
}

func (vtw *VirtualTokenWallet) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetPvtDataWithCompositeKey(stub, VIRTUAL_TOKEN_WALLET_PREFIX, keyAttributes, state.BIDS_PVT_DATA_COLLECTION, vtw)
}

func (vtw *VirtualTokenWallet) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutPvtDataWithCompositeKey(stub, VIRTUAL_TOKEN_WALLET_PREFIX, (*vtw.GetID())[0], state.BIDS_PVT_DATA_COLLECTION, vtw)
}

func (vtw *VirtualTokenWallet) GetID() *[][]string {
	return &[][]string{{vtw.OwnerID}}
}

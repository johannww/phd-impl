package payment

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"google.golang.org/protobuf/proto"
)

const (
	VIRTUAL_TOKEN_WALLET_PREFIX = "virtualTokenWallet"
)

// VirtualTokenWallet is a wallet that stores tokens corresponding to a
// fiat currency. It is used to pay for the carbon credits.
type VirtualTokenWallet struct {
	OwnerID  string `json:"owner"`
	Quantity int64  `json:"quantity"`
}

var _ state.WorldStateManager = (*VirtualTokenWallet)(nil)

func (vtw *VirtualTokenWallet) ToProto() proto.Message {
	return &pb.VirtualTokenWallet{
		Owner:    vtw.OwnerID,
		Quantity: vtw.Quantity,
	}
}

func (vtw *VirtualTokenWallet) FromProto(m proto.Message) error {
	pv, ok := m.(*pb.VirtualTokenWallet)
	if !ok {
		return fmt.Errorf("unexpected proto message type for VirtualTokenWallet")
	}
	vtw.OwnerID = pv.Owner
	vtw.Quantity = pv.Quantity
	return nil
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

func MintVirtualToken(stub shim.ChaincodeStubInterface, ownerID string, quantity int64) (*VirtualTokenWallet, error) {
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

// UpdateVirtualTokenWallet updates the balance of a virtual token wallet.
// If the wallet doesn't exist, it creates a new one.
func UpdateVirtualTokenWallet(stub shim.ChaincodeStubInterface, ownerID string, quantity int64) error {
	wallet := &VirtualTokenWallet{OwnerID: ownerID}
	// Try to get existing wallet
	err := wallet.FromWorldState(stub, []string{ownerID})
	if err != nil {
		// Wallet doesn't exist, start with 0
		wallet.Quantity = 0
	}

	wallet.Quantity += quantity
	if wallet.Quantity < 0 {
		return fmt.Errorf("insufficient funds for owner %s: current balance %d, update amount %d", ownerID, wallet.Quantity-quantity, quantity)
	}

	return wallet.ToWorldState(stub)
}

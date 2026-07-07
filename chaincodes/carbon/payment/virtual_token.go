package payment

import (
	"fmt"
	"strconv"

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
	Number   int64  `json:"number"`
}

var _ state.WorldStateManager = (*VirtualTokenWallet)(nil)

func (vtw *VirtualTokenWallet) ToProto() proto.Message {
	return &pb.VirtualTokenWallet{
		Owner:    vtw.OwnerID,
		Quantity: vtw.Quantity,
		Number:   vtw.Number,
	}
}

func (vtw *VirtualTokenWallet) FromProto(m proto.Message) error {
	pv, ok := m.(*pb.VirtualTokenWallet)
	if !ok {
		return fmt.Errorf("unexpected proto message type for VirtualTokenWallet")
	}
	vtw.OwnerID = pv.Owner
	vtw.Quantity = pv.Quantity
	vtw.Number = pv.Number
	return nil
}

func (vtw *VirtualTokenWallet) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetPvtDataWithCompositeKey(stub, VIRTUAL_TOKEN_WALLET_PREFIX, keyAttributes, state.BIDS_PVT_DATA_COLLECTION, vtw)
}

func (vtw *VirtualTokenWallet) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutPvtDataWithCompositeKey(stub, VIRTUAL_TOKEN_WALLET_PREFIX, (*vtw.GetID())[0], state.BIDS_PVT_DATA_COLLECTION, vtw)
}

func (vtw *VirtualTokenWallet) GetID() *[][]string {
	return &[][]string{{vtw.OwnerID, strconv.FormatInt(vtw.Number, 10)}}
}

func MintVirtualToken(stub shim.ChaincodeStubInterface, ownerID string, quantity int64) (*VirtualTokenWallet, error) {
	return MintVirtualTokenForWalletID(stub, ownerID, 0, quantity)
}

func MintVirtualTokenForWalletID(stub shim.ChaincodeStubInterface, ownerID string, walletNumber int64, quantity int64) (*VirtualTokenWallet, error) {
	tokenWallet := &VirtualTokenWallet{OwnerID: ownerID, Number: walletNumber}
	if cid.AssertAttributeValue(stub, identities.PaymentCompanyAttr, "true") != nil {
		return nil, fmt.Errorf("only identities with the attribute %s can mint virtual tokens", identities.PaymentCompanyAttr)
	}

	err := tokenWallet.FromWorldState(stub, (*tokenWallet.GetID())[0])

	if err != nil {
		tokenWallet.Quantity = 0
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
	return UpdateVirtualTokenWalletForWalletID(stub, ownerID, 0, quantity)
}

// UpdateVirtualTokenWalletForWalletID updates the balance of a virtual token wallet.
// If the wallet doesn't exist, it creates a new one.
func UpdateVirtualTokenWalletForWalletID(stub shim.ChaincodeStubInterface, ownerID string, walletNumber int64, quantity int64) error {
	wallet := &VirtualTokenWallet{OwnerID: ownerID, Number: walletNumber}
	// Try to get existing wallet
	err := wallet.FromWorldState(stub, (*wallet.GetID())[0])
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

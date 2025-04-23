package payment

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	PVT_DATA_COLLECTION      = "privateDataCollection"
	PVT_VIRTUAL_TOKEn_PREFIX = "virtualToken"
)

type VirtualTokenWallet struct {
	OwnerID  string  `json:"owner"`
	Quantity float64 `json:"quantity"`
}

var _ state.WorldStateManager = (*VirtualTokenWallet)(nil)

func MintVirtualToken(stub shim.ChaincodeStubInterface, ownerID string, quantity float64) (*VirtualTokenWallet, error) {
	tokenWallet := &VirtualTokenWallet{OwnerID: ownerID}
	err := tokenWallet.FromWorldState(stub, []string{ownerID})
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

func (virtualtokenwallet *VirtualTokenWallet) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	panic("not implemented") // TODO: Implement
}

func (virtualtokenwallet *VirtualTokenWallet) ToWorldState(stub shim.ChaincodeStubInterface) error {
	panic("not implemented") // TODO: Implement
}

func (virtualtokenwallet *VirtualTokenWallet) GetID() *[][]string {
	panic("not implemented") // TODO: Implement
}

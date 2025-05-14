package payment

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	PVT_VIRTUAL_TOKEN_PREFIX = "virtualToken"
)

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

func (virtualtokenwallet *VirtualTokenWallet) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	panic("not implemented") // TODO: Implement
}

func (virtualtokenwallet *VirtualTokenWallet) ToWorldState(stub shim.ChaincodeStubInterface) error {
	panic("not implemented") // TODO: Implement
}

func (virtualtokenwallet *VirtualTokenWallet) GetID() *[][]string {
	panic("not implemented") // TODO: Implement
}

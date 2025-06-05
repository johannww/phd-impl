package utils_test

import (
	"strings"

	"encoding/json"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/payment"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
)

// TestData holds a list as an identity map
// The map key is a string and the value is generic interface{}
type TestData struct {
	Identities    *setup.MockIdentities
	Properties    []*properties.Property
	MintCredits   []*credits.MintCredit
	CreditWallets []*credits.CreditWallet
	TokenWallets  []*payment.VirtualTokenWallet
	SellBids      []*bids.SellBid
	BuyBids       []*bids.BuyBid
}

func (data *TestData) SaveToWorldState(stub shim.ChaincodeStubInterface) {
	saveToWorldState(stub, data.Properties)
	saveToWorldState(stub, data.MintCredits)
	saveToWorldState(stub, data.CreditWallets)
	saveToWorldState(stub, data.TokenWallets)
	saveToWorldState(stub, data.SellBids)
	saveToWorldState(stub, data.BuyBids)
}

func (data *TestData) CompaniesIdentities() (companiesIds []string) {
	for ownerId := range *data.Identities {
		if strings.Contains(ownerId, COMPANY_PREFIX) {
			companiesIds = append(companiesIds, ownerId)
		}
	}
	return companiesIds
}

func (data *TestData) String() string {
	bytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return "error marshalling TestData: " + err.Error()
	}
	return string(bytes)
}
func saveToWorldState[T state.WorldStateManager](stub shim.ChaincodeStubInterface, data []T) {
	for _, item := range data {
		if err := item.ToWorldState(stub); err != nil {
			panic(err)
		}
	}
}

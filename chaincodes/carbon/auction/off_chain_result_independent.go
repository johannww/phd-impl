package auction

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

func processIndependentAuctionResult(stub shim.ChaincodeStubInterface,
	resultPub *OffChainIndepAuctionResult, resultPvt *OffChainIndepAuctionResult,
) error {
	result, err := MergeIndependentPublicPrivateResults(resultPub, resultPvt)
	if err != nil {
		return fmt.Errorf("could not merge independent public and private results: %v", err)
	}

	_ = result // TODO: do something with the result

	return fmt.Errorf("independent auction result processing not implemented yet")
}

package carbon

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

const (
	TEE_CONTAINER_HASH_PREFIX = "teeContainerHash"
)

type Auction struct{}

// TODO:
func (a *Auction) calculateCommitment(
	buyBids []*BuyBid,
	sellBids []*SellBid,
	// privateData
) error {
	return nil
}

// TODO: finish this
func (a *Auction) PublishTEEContainerHash(stub shim.ChaincodeStubInterface, containerHash string) error {
	err := cid.AssertAttributeValue(stub, "auction.containerPublisher", "true")
	if err != nil {
		return fmt.Errorf("this identity cannot publish the auction's container hash: %v", err)
	}

	transTs, err := stub.GetTxTimestamp()
	putStateWithCompositeKey[string](stub, TEE_CONTAINER_HASH_PREFIX, []string{transTs.String()}, containerHash)
	if err != nil {
		return err
	}

	return nil
}

// TODO: finish this
func (a *Auction) verifyTEEProof(proof []byte) bool {
	return true
}

// PublishTEEProof
// Currently this supports Microsoft Confidential Containers, which
// use the AMD SEV-SNP attestation protocol.
func (a *Auction) PublishTEEProof(stub shim.ChaincodeStubInterface, auctionID string) error {

	// Get commitment

	// Verify

	return nil
}

package carbon

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

const (
	TEE_CONTAINER_HASH_PREFIX  = "teeContainerHash"
	PVT_DATA_COMMITMENT_PREFIX = "pvtDataCommitment"
)

type Auction struct{}

// TODO: add more fields relevant to the auction
func (a *Auction) calculateCommitment(
	buyBids []*BuyBid,
	sellBids []*SellBid,
	privatePrice []*PrivatePrice,
) (*[32]byte, error) {
	buyBidsBytes, err := json.Marshal(buyBids)
	if err != nil {
		return nil, fmt.Errorf("could not marshal buy bids: %v", err)
	}

	sellBidsBytes, err := json.Marshal(sellBids)
	if err != nil {
		return nil, fmt.Errorf("could not marshal sell bids: %v", err)
	}

	privatePriceBytes, err := json.Marshal(privatePrice)
	if err != nil {
		return nil, fmt.Errorf("could not marshal private price: %v", err)
	}

	auctionDataBytes := append(buyBidsBytes, sellBidsBytes...)
	auctionDataBytes = append(auctionDataBytes, privatePriceBytes...)

	commitment := sha256.Sum256(auctionDataBytes)

	return &commitment, nil
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

package auction

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	ccstate "github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	TEE_CONTAINER_HASH_PREFIX  = "teeContainerHash"
	PVT_DATA_COMMITMENT_PREFIX = "pvtDataCommitment"
)

type Auction struct{}

// TODO: add more fields relevant to the auction
func (a *Auction) calculateCommitment(
	buyBids []*bids.BuyBid,
	sellBids []*bids.SellBid,
	privatePrice []*bids.PrivatePrice,
	txTimestamp string,
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

	txTimestampBytes := []byte(txTimestamp)
	auctionDataBytes = append(auctionDataBytes, txTimestampBytes...)

	commitment := sha256.Sum256(auctionDataBytes)

	return &commitment, nil
}

// TODO: finish this
func (a *Auction) PublishTEEContainerHash(stub shim.ChaincodeStubInterface, containerHash string) error {
	err := cid.AssertAttributeValue(stub, "auction.containerPublisher", "true")
	if err != nil {
		return fmt.Errorf("this identity cannot publish the auction's container hash: %v", err)
	}

	txTimestamp, err := stub.GetTxTimestamp()
	ccstate.PutStateWithCompositeKey[string](
		stub, TEE_CONTAINER_HASH_PREFIX,
		&[][]string{{txTimestamp.String()}},
		containerHash)
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

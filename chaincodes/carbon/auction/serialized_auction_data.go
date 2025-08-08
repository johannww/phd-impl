package auction

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

type SerializedAuctionData struct {
	AuctionDataBytes []byte // JSON Serialized AuctionData
	Sum              []byte // SHA256 sum of bytes of above fields
}

func (s *SerializedAuctionData) CommitmentToWorldState(stub shim.ChaincodeStubInterface, endRFC339Timestamp string) error {
	err := s.CalculateHash()
	if err != nil {
		return fmt.Errorf("could not calculate sum: %v", err)
	}

	auctionCommitment := &AuctionCommitment{
		EndTimestamp: endRFC339Timestamp,
		Hash:         s.Sum,
	}
	err = auctionCommitment.ToWorldState(stub)
	return err
}

func (s *SerializedAuctionData) CommitmentFromWorldState(stub shim.ChaincodeStubInterface, endRFC339Timestamp string) error {
	auctionCommitment := &AuctionCommitment{}
	err := auctionCommitment.FromWorldState(stub, []string{endRFC339Timestamp})
	s.Sum = auctionCommitment.Hash
	return err
}

func (s *SerializedAuctionData) CalculateHash() error {
	sum, err := s.calculateHash()
	if err != nil {
		return fmt.Errorf("could not calculate auction data hash: %v", err)
	}

	s.Sum = sum
	return nil
}

func (s *SerializedAuctionData) calculateHash() ([]byte, error) {
	hash := sha256.New()
	_, err := hash.Write(s.AuctionDataBytes)
	if err != nil {
		return nil, fmt.Errorf("could not write sell bid bytes to hash: %v", err)
	}

	return hash.Sum(nil), nil
}

func (s *SerializedAuctionData) ValidateHash() bool {
	if s.Sum == nil {
		return false
	}

	calculatedSum, err := s.calculateHash()
	return err == nil && bytes.Equal(s.Sum, calculatedSum)
}

func (s *SerializedAuctionData) ToAuctionData() (*AuctionData, error) {
	auctionData := &AuctionData{}

	err := json.Unmarshal(s.AuctionDataBytes, auctionData)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal sell bids: %v", err)
	}

	return auctionData, nil
}

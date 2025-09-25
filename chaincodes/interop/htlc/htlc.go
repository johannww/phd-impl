package htlc

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	carbon_utils "github.com/johannww/phd-impl/chaincodes/carbon/utils"
	"github.com/johannww/phd-impl/chaincodes/interop/lock"
	"github.com/johannww/phd-impl/chaincodes/interop/util"
)

type HTLC struct {
	SecretHash string `json:"secretHash"`
	// Secret is the preimage of the hash, only set when claimed
	Secret string `json:"secret,omitempty"`
	LockID string `json:"lockID"`
	// BuyerID is the string representation of the receiver's identity on
	// the destination chain
	BuyerID string `json:"receiverId"`
	// SellerWallet is the wallet address of the seller on the public chain
	SellerWallet string `json:"paymentReceiverWallet,omitempty"`
	Amount       int    `json:"amount"`
	Claimed      bool   `json:"claimed"`
	// ValidUntil is the timestamp (RFC3339) until which the HTLC is valid
	ValidUntil string `json:"validUntil"`
}

func (h *HTLC) UnlockCreditsAfterValidUntil(stub shim.ChaincodeStubInterface) error {

	var err error
	protoTs, _ := stub.GetTxTimestamp()
	transactionTs := carbon_utils.TimestampRFC3339UtcString(protoTs)

	htlcLock := &HtlcLock{
		LockID: h.LockID,
		HTLCID: h.SecretHash,
	}

	if err = htlcLock.FromWorldState(stub, (*htlcLock.GetID())[0]); err != nil {
		return fmt.Errorf("failed to get HTLC lock from world state: %v", err)
	}

	if transactionTs <= h.ValidUntil {
		return fmt.Errorf("HTLC is still valid until %s, current time is %s",
			h.ValidUntil, transactionTs)
	}

	if err = htlcLock.DeleteFromWorldState(stub); err != nil {
		return fmt.Errorf("failed to get HTLC lock from world state: %v", err)
	}

	return nil
}

// TODOHP: finish implementation
func NewHTLC(stub shim.ChaincodeStubInterface,
	secretHash, lockID, receiverID, validUntil, destChainID string,
	amount int) (*HTLC, error) {
	lock.CreditIsLockedForChainID(stub, util.CARBON_CC_NAME, nil, lockID, "")

	// get locked credit info

	// check owner is creator of HTLC

	// prevent locked credit from being used in another HTLC
	htlcLock := &HtlcLock{
		LockID: lockID,
		HTLCID: secretHash,
	}
	err := htlcLock.ToWorldState(stub)
	if err != nil {
		return nil, err
	}

	// create HTLC
	return &HTLC{
		SecretHash: secretHash,
		LockID:     lockID,
		BuyerID:    receiverID,
		Amount:     amount,
		ValidUntil: validUntil,
	}, nil
}

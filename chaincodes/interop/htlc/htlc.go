package htlc

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	ccstate "github.com/johannww/phd-impl/chaincodes/common/state"
	carbon_utils "github.com/johannww/phd-impl/chaincodes/common/utils"
	"github.com/johannww/phd-impl/chaincodes/interop/lock"
	"google.golang.org/protobuf/proto"
)

const HTLC_PREFIX = "htlc"

type HTLC struct {
	CreditID   []string `json:"creditID"`
	SecretHash string   `json:"secretHash"`
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

var _ ccstate.WorldStateManager = (*HTLC)(nil)

func (h *HTLC) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return ccstate.GetStateWithCompositeKey(stub, HTLC_PREFIX, keyAttributes, h)
}

func (h *HTLC) ToWorldState(stub shim.ChaincodeStubInterface) error {
	if h.LockID == "" {
		return fmt.Errorf("lockID is required")
	}
	if h.SecretHash == "" {
		return fmt.Errorf("secretHash is required")
	}
	if h.BuyerID == "" {
		return fmt.Errorf("buyerID is required")
	}
	if h.ValidUntil == "" {
		return fmt.Errorf("validUntil is required")
	}
	if _, err := time.Parse(time.RFC3339, h.ValidUntil); err != nil {
		return fmt.Errorf("validUntil must be RFC3339: %v", err)
	}
	return ccstate.PutStateWithCompositeKey(stub, HTLC_PREFIX, h.GetID(), h)
}

func (h *HTLC) DeleteFromWorldState(stub shim.ChaincodeStubInterface) error {
	return ccstate.DeleteStateWithCompositeKey(stub, HTLC_PREFIX, h.GetID())
}

func (h *HTLC) GetID() *[][]string {
	return &[][]string{{h.LockID}}
}

func (h *HTLC) UnlockCreditsAfterValidUntil(stub shim.ChaincodeStubInterface) error {
	protoTs, _ := stub.GetTxTimestamp()
	now := protoTs.AsTime().UTC()

	validUntil, err := time.Parse(time.RFC3339, h.ValidUntil)
	if err != nil {
		return fmt.Errorf("failed to parse validUntil timestamp %s: %v", h.ValidUntil, err)
	}

	if !now.After(validUntil) {
		return fmt.Errorf("HTLC is still valid until %s, current time is %s",
			h.ValidUntil, now.UTC().Format(time.RFC3339))
	}

	return nil
}

func CreateHTLC(
	stub shim.ChaincodeStubInterface,
	carbonCCName string,
	creditID []string,
	destChainID string,
	h *HTLC,
) (string, error) {
	if h == nil {
		return "", fmt.Errorf("htlc payload cannot be nil")
	}
	if h.LockID == "" || h.SecretHash == "" || h.BuyerID == "" || h.ValidUntil == "" {
		return "", fmt.Errorf("lockID, secretHash, buyerID and validUntil are required")
	}
	if _, err := time.Parse(time.RFC3339, h.ValidUntil); err != nil {
		return "", fmt.Errorf("validUntil must be RFC3339: %v", err)
	}
	if len(creditID) == 0 {
		return "", fmt.Errorf("creditID is required")
	}
	callerID := identities.GetID(stub)
	if callerID != creditID[0] { // creditID first field is the owner ID
		return "", fmt.Errorf("only the credit owner can create HTLC")
	}

	isLocked, err := lock.CreditIsLocked(stub, carbonCCName, creditID, h.LockID)
	if err != nil {
		return "", err
	}
	if !isLocked {
		return "", fmt.Errorf("credit is not locked for lockID %s", h.LockID)
	}

	isLockedForChain, err := lock.CreditIsLockedForChainID(stub, carbonCCName, creditID, h.LockID, destChainID)
	if err != nil {
		return "", err
	}
	if !isLockedForChain {
		return "", fmt.Errorf("credit lock %s is not associated with destination chain %s", h.LockID, destChainID)
	}

	existing := &HTLC{LockID: h.LockID}
	if err := existing.FromWorldState(stub, (*existing.GetID())[0]); err == nil {
		return "", fmt.Errorf("htlc for lock %s already exists", h.LockID)
	}

	h.CreditID = append([]string(nil), creditID...)
	h.Claimed = false

	if err := h.ToWorldState(stub); err != nil {
		return "", err
	}

	return h.SecretHash, nil
}

func ClaimHTLC(stub shim.ChaincodeStubInterface, lockID string, secret string) (bool, error) {
	record := &HTLC{LockID: lockID}
	err := record.FromWorldState(stub, (*record.GetID())[0])
	if err != nil {
		return false, err
	}
	if record.Claimed {
		return false, fmt.Errorf("htlc for lock %s is already claimed", lockID)
	}
	if !isMatchingSecret(secret, record.SecretHash) {
		return false, fmt.Errorf("provided secret does not match secret hash")
	}

	record.Secret = secret
	record.Claimed = true
	if err := record.ToWorldState(stub); err != nil {
		return false, err
	}
	return true, nil
}

func IsHTLCClaimed(stub shim.ChaincodeStubInterface, lockID string) (bool, error) {
	record := &HTLC{LockID: lockID}
	err := record.FromWorldState(stub, (*record.GetID())[0])
	if err != nil {
		return false, err
	}
	return record.Claimed, nil
}

func GetHTLCHashByLockID(stub shim.ChaincodeStubInterface, lockID string) (string, error) {
	record := &HTLC{LockID: lockID}
	err := record.FromWorldState(stub, (*record.GetID())[0])
	if err != nil {
		return "", err
	}
	return record.SecretHash, nil
}

func GetHTLCPreimageByLockID(stub shim.ChaincodeStubInterface, lockID string) (string, error) {
	record := &HTLC{LockID: lockID}
	err := record.FromWorldState(stub, (*record.GetID())[0])
	if err != nil {
		return "", err
	}
	if !record.Claimed || record.Secret == "" {
		return "", fmt.Errorf("htlc for lock %s has not been claimed", lockID)
	}
	return record.Secret, nil
}

func UnlockHTLCExpired(stub shim.ChaincodeStubInterface, carbonCCName string, lockID string) (bool, error) {
	record := &HTLC{LockID: lockID}
	err := record.FromWorldState(stub, (*record.GetID())[0])
	if err != nil {
		return false, err
	}
	if record.Claimed {
		return false, fmt.Errorf("cannot unlock claimed htlc")
	}

	if err := record.UnlockCreditsAfterValidUntil(stub); err != nil {
		return false, err
	}

	if err := lock.UnlockCredit(stub, carbonCCName, record.CreditID, lockID); err != nil {
		return false, err
	}
	if err := record.DeleteFromWorldState(stub); err != nil {
		return false, err
	}
	return true, nil
}

func isMatchingSecret(secret string, expectedHash string) bool {
	b := sha256.Sum256([]byte(secret))
	shaHex := hex.EncodeToString(b[:])
	if shaHex == expectedHash {
		return true
	}
	shaBase64 := base64.StdEncoding.EncodeToString(b[:])
	return shaBase64 == expectedHash
}

func (h *HTLC) ToProto() proto.Message {
	if h == nil {
		return nil
	}
	return &pb.HTLC{
		CreditID:              append([]string(nil), h.CreditID...),
		SecretHash:            h.SecretHash,
		Secret:                h.Secret,
		LockID:                h.LockID,
		ReceiverId:            h.BuyerID,
		PaymentReceiverWallet: h.SellerWallet,
		Amount:                int64(h.Amount),
		Claimed:               h.Claimed,
		ValidUntil:            h.ValidUntil,
	}
}

// FromProto populates HTLC from its protobuf representation.
func (h *HTLC) FromProto(m proto.Message) error {
	pbH, ok := m.(*pb.HTLC)
	if !ok {
		return fmt.Errorf("unexpected proto message type for HTLC")
	}
	if pbH == nil {
		// clear receiver
		*h = HTLC{}
		return nil
	}

	h.CreditID = append([]string(nil), pbH.GetCreditID()...)
	h.SecretHash = pbH.GetSecretHash()
	h.Secret = pbH.GetSecret()
	h.LockID = pbH.GetLockID()
	h.BuyerID = pbH.GetReceiverId()
	h.SellerWallet = pbH.GetPaymentReceiverWallet()
	h.Amount = int(pbH.GetAmount())
	h.Claimed = pbH.GetClaimed()
	h.ValidUntil = pbH.GetValidUntil()

	return nil
}

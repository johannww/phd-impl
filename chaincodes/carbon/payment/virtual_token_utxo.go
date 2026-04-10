package payment

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"google.golang.org/protobuf/proto"
)

const (
	VIRTUAL_TOKEN_UTXO_PREFIX = "virtualTokenUTXO"
)

// VirtualTokenUTXO represents an unspent transaction output for virtual tokens.
// This eliminates MVCC conflicts during auction settlement by creating
// parallel UTXOs instead of updating a centralized wallet.
// UTXOs are deleted (not marked as spent) when aggregated into a wallet.
// The txID is the auction ID, allowing one UTXO per owner per auction.
type VirtualTokenUTXO struct {
	OwnerID string `json:"ownerID"`
	Amount  int64  `json:"amount"`
	TxID    string `json:"txID"` // Auction ID or transaction identifier
}

// UTXOListResult wraps the result of UTXO queries, containing the list of UTXOs and total balance.
// This is needed because Fabric contract API only supports max 2 return values.
type UTXOListResult struct {
	UTXOs        []*VirtualTokenUTXO `json:"utxos"`
	TotalBalance int64               `json:"totalBalance"`
}

var _ state.WorldStateManager = (*VirtualTokenUTXO)(nil)

func (u *VirtualTokenUTXO) ToProto() proto.Message {
	return &pb.VirtualTokenUTXO{
		OwnerID: u.OwnerID,
		Amount:  u.Amount,
		TxID:    u.TxID,
	}
}

func (u *VirtualTokenUTXO) FromProto(m proto.Message) error {
	pv, ok := m.(*pb.VirtualTokenUTXO)
	if !ok {
		return fmt.Errorf("unexpected proto message type for VirtualTokenUTXO")
	}
	u.OwnerID = pv.OwnerID
	u.Amount = pv.Amount
	u.TxID = pv.TxID
	return nil
}

func (u *VirtualTokenUTXO) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetStateWithCompositeKey(stub, VIRTUAL_TOKEN_UTXO_PREFIX, keyAttributes, u)
}

func (u *VirtualTokenUTXO) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutStateWithCompositeKey(stub, VIRTUAL_TOKEN_UTXO_PREFIX, u.GetID(), u)
}

func (u *VirtualTokenUTXO) GetID() *[][]string {
	// Composite key: [ownerID, txID] allows efficient queries by owner
	return &[][]string{{u.OwnerID, u.TxID}}
}

// CreatePaymentUTXO creates a UTXO representing payment to a seller.
// This is used during auction settlement to avoid wallet MVCC conflicts.
// The txID should be the auction ID.
func CreatePaymentUTXO(stub shim.ChaincodeStubInterface, ownerID string, amount int64, txID string) error {
	if amount <= 0 {
		return fmt.Errorf("UTXO amount must be positive, got %d", amount)
	}

	utxo := &VirtualTokenUTXO{
		OwnerID: ownerID,
		Amount:  amount,
		TxID:    txID,
	}

	return utxo.ToWorldState(stub)
}

// CreateRefundUTXO creates a UTXO representing a refund to a buyer.
// This is used during auction settlement when the clearing price is lower than the bid price.
// The txID should be the auction ID.
func CreateRefundUTXO(stub shim.ChaincodeStubInterface, ownerID string, amount int64, txID string) error {
	if amount <= 0 {
		return fmt.Errorf("UTXO amount must be positive, got %d", amount)
	}

	utxo := &VirtualTokenUTXO{
		OwnerID: ownerID,
		Amount:  amount,
		TxID:    txID,
	}

	return utxo.ToWorldState(stub)
}

// GetUnspentUTXOsForOwner retrieves all UTXOs for a given owner.
// Returns the list of UTXOs and the total balance.
func GetUnspentUTXOsForOwner(stub shim.ChaincodeStubInterface, ownerID string) ([]*VirtualTokenUTXO, int64, error) {
	// Query using partial composite key: prefix + ownerID
	iterator, err := stub.GetStateByPartialCompositeKey(VIRTUAL_TOKEN_UTXO_PREFIX, []string{ownerID})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get UTXO iterator for owner %s: %v", ownerID, err)
	}
	defer iterator.Close()

	var utxos []*VirtualTokenUTXO
	var totalBalance int64

	for iterator.HasNext() {
		item, err := iterator.Next()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to iterate UTXOs: %v", err)
		}

		utxo := &VirtualTokenUTXO{}
		if err := state.UnmarshalStateAs(item.Value, utxo); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal UTXO: %v", err)
		}

		utxos = append(utxos, utxo)
		totalBalance += utxo.Amount
	}

	return utxos, totalBalance, nil
}

// AggregateUTXOsIntoWallet collects all UTXOs for the specified owner
// and adds their total to the owner's VirtualTokenWallet.
// The UTXOs are deleted from world state after aggregation.
func AggregateUTXOsIntoWallet(stub shim.ChaincodeStubInterface, ownerID string) (int64, error) {
	// Get all UTXOs for the owner
	utxos, totalAmount, err := GetUnspentUTXOsForOwner(stub, ownerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get UTXOs for owner %s: %v", ownerID, err)
	}

	if len(utxos) == 0 {
		return 0, nil // No UTXOs to aggregate
	}

	// Update the wallet balance
	err = UpdateVirtualTokenWallet(stub, ownerID, totalAmount)
	if err != nil {
		return 0, fmt.Errorf("failed to update wallet for owner %s: %v", ownerID, err)
	}

	// Delete all UTXOs from world state
	for _, utxo := range utxos {
		compositeKey, err := stub.CreateCompositeKey(VIRTUAL_TOKEN_UTXO_PREFIX, []string{utxo.OwnerID, utxo.TxID})
		if err != nil {
			return 0, fmt.Errorf("failed to create composite key for UTXO %s: %v", utxo.TxID, err)
		}
		if err := stub.DelState(compositeKey); err != nil {
			return 0, fmt.Errorf("failed to delete UTXO %s: %v", utxo.TxID, err)
		}
	}

	return totalAmount, nil
}

// GetTotalBalanceForOwner returns the total balance for an owner,
// including both wallet balance and unspent UTXOs.
func GetTotalBalanceForOwner(stub shim.ChaincodeStubInterface, ownerID string) (int64, error) {
	// Get wallet balance
	wallet := &VirtualTokenWallet{OwnerID: ownerID}
	walletBalance := int64(0)
	err := wallet.FromWorldState(stub, []string{ownerID})
	if err == nil {
		walletBalance = wallet.Quantity
	}
	// If wallet doesn't exist, walletBalance remains 0

	// Get UTXO balance
	_, utxoBalance, err := GetUnspentUTXOsForOwner(stub, ownerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get UTXO balance: %v", err)
	}

	return walletBalance + utxoBalance, nil
}

package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"google.golang.org/protobuf/proto"
)

func readTransactionProtoFromStdin() []byte {
	transactionProtoHexBytes := []byte{}
	stdinReadBuffer := make([]byte, 10000)
	var n int
	var err error
	for {
		n, err = os.Stdin.Read(stdinReadBuffer)
		if err != nil {
			fmt.Println("error occured")
			break
		}
		if n == 0 {
			break
		}
		transactionProtoHexBytes = append(transactionProtoHexBytes, stdinReadBuffer[:n]...)
		fmt.Printf("Read %d bytes from stdin\n", n)
	}
	transactionProtoBytes, err := hex.DecodeString(string(transactionProtoHexBytes[:len(transactionProtoHexBytes)-1]))
	panicOnError(err, "Failed to decode hex string from stdin")
	fmt.Println(len(transactionProtoBytes), "bytes read from stdin after hex decoding")

	return transactionProtoBytes
}

func main() {
	transactionProtoBytes := readTransactionProtoFromStdin()
	// fmt.Println(string(transactionProtoBytes))

	transactionProto := &peer.ProcessedTransaction{}
	err := proto.Unmarshal(transactionProtoBytes, transactionProto)
	if err != nil {
		fmt.Printf("Failed to unmarshal transaction proto: %v\n", err)
		return
	}

	fmt.Println(string(transactionProto.ProtoReflect().GetUnknown()))

	transactionPayload := &common.Payload{}
	err = proto.Unmarshal(transactionProto.TransactionEnvelope.Payload, transactionPayload)
	if err != nil {
		fmt.Printf("Failed to unmarshal transaction payload: %v\n", err)
		return
	}
	fmt.Println(string(transactionPayload.ProtoReflect().GetUnknown()))

	transaction := &peer.Transaction{}
	err = proto.Unmarshal(transactionPayload.Data, transaction)
	if err != nil {
		fmt.Printf("Failed to unmarshal transaction data: %v\n", err)
		return
	}
	fmt.Println(string(transaction.ProtoReflect().GetUnknown()))

	_ = transaction
	_ = json.Decoder{}
	fmt.Println(len(transaction.Actions), "actions in transaction")

	for _, action := range transaction.Actions {
		chaincodeActionPayload := &peer.ChaincodeActionPayload{}
		err = proto.Unmarshal(action.Payload, chaincodeActionPayload)
		if err != nil {
			fmt.Printf("Failed to unmarshal action payload: %v\n", err)
			return
		}
		fmt.Println(string(chaincodeActionPayload.ProtoReflect().GetUnknown()))
		// fmt.Println(string(chaincodeActionPayload.Action.ProposalResponsePayload))

		responsePayload := peer.ProposalResponsePayload{}
		err = proto.Unmarshal(chaincodeActionPayload.Action.ProposalResponsePayload, &responsePayload)
		if err != nil {
			fmt.Printf("Failed to unmarshal proposal response payload: %v\n", err)
			return
		}
		fmt.Println(string(responsePayload.ProtoReflect().GetUnknown()))

		chaincodeAction := &peer.ChaincodeAction{}
		err = proto.Unmarshal(responsePayload.Extension, chaincodeAction)
		if err != nil {
			fmt.Printf("Failed to unmarshal chaincode action: %v\n", err)
			return
		}
		fmt.Println(string(chaincodeAction.ProtoReflect().GetUnknown()))

		_ = rwsetutil.TxRwSet{}
		kvSet := &rwsetutil.TxRwSet{}
		err = kvSet.FromProtoBytes(chaincodeAction.Results)
		if err != nil {
			fmt.Printf("Failed to unmarshal kv read-write: %v\n", err)
			return
		}

		for _, nsRWSet := range kvSet.NsRwSets {
			fmt.Printf("Namespace: %s\n", nsRWSet.NameSpace)
			for _, rwSet := range nsRWSet.KvRwSet.Writes {
				fmt.Printf("Write - Key: %s\n, Value: %s\n", rwSet.Key, rwSet.Value)
			}
			for _, rwSet := range nsRWSet.KvRwSet.Reads {
				fmt.Printf("Read - Key: %s\n, Version: %d\n", rwSet.Key, rwSet.Version)
			}
		}
	}

}

func panicOnError(err error, message string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %v", message, err))
	}
}

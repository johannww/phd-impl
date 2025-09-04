package utils

import (
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func TimestampRFC3339UtcString(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	// RFC3339     = "2006-01-02T15:04:05Z07:00"
	// "AsTime" returns UTC time.
	// We also use environment variable TZ=UTC to make sure
	return ts.AsTime().Format(time.RFC3339)
}

func UnixNowFromTransactionTimestamp(stub shim.ChaincodeStubInterface) int64 {
	txTime, err := stub.GetTxTimestamp()
	if err != nil {
		panic(err)
	}
	return txTime.AsTime().Unix()
}

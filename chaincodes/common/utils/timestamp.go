package utils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const RFC3339WithMillis = "2006-01-02T15:04:05.000Z07:00"

// TimestampRFC3339WithMillisUtcString is good for debbugging and logging
func TimestampRFC3339WithMillisUtcString(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	// RFC3339     = "2006-01-02T15:04:05Z07:00"
	// RFC339WithMillis = "2006-01-02T15:04:05.000Z07:00"
	// "AsTime" returns UTC time.
	// We also use environment variable TZ=UTC to make sure
	return ts.AsTime().Format(RFC3339WithMillis)
}

func UnixNowFromTransactionTimestamp(stub shim.ChaincodeStubInterface) int64 {
	txTime, err := stub.GetTxTimestamp()
	if err != nil {
		panic(err)
	}
	return txTime.AsTime().Unix()
}

// UnixMillisNowFromStub returns the current timestamp in milliseconds
// as a hex string, which is useful for generating time-based IDs that sort chronologically.
func UnixMillisNowFromStub(stub shim.ChaincodeStubInterface) string {
	txTime, err := stub.GetTxTimestamp()
	if err != nil {
		panic(err)
	}
	unixMillis := txTime.AsTime().UnixMilli()
	return fmt.Sprintf("%016x", unixMillis)
}

func UnixMillisNowFromTransactionTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	unixMillis := ts.AsTime().UnixMilli()
	return fmt.Sprintf("%016x", unixMillis)
}

func UnixMillisNowFromGoTime(t time.Time) string {
	unixMillis := t.UnixMilli()
	return fmt.Sprintf("%016x", unixMillis)
}

func UnixMillisNowFromRFC3339String(rfc3339Str string) (string, error) {
	t, err := time.Parse(time.RFC3339, rfc3339Str)
	if err != nil {
		return "", err
	}
	return UnixMillisNowFromGoTime(t), nil
}

func ParseHexTimestamp(hexStr string) (time.Time, error) {
	unixMillis, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.UnixMilli(unixMillis), nil
}

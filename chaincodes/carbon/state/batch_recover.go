package state

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

const PAGE_SIZE = 30000
const maxUnicodeRuneValue = utf8.MaxRune

// TODO: implement
func GetStatesByRange(stub shim.ChaincodeStubInterface, objectType string, key string) ([][]byte, error) {
	return nil, nil
}

func readIteratorStates(stateIterator shim.StateQueryIteratorInterface) ([][]byte, string, error) {
	statesInRange := [][]byte{}
	i := 0
	lastKey := ""
	for stateIterator.HasNext() {
		kv, err := stateIterator.Next()
		if err != nil {
			if strings.Contains(err.Error(), "invalid iterator state") {
				// the error indicates that there is more keys to be fetched
				return statesInRange, lastKey, nil
			}
			return nil, "", fmt.Errorf("could not get next state by range: %v", err)
		}

		statesInRange = append(statesInRange, kv.GetValue())
		lastKey = kv.GetKey()
		i++
	}

	return statesInRange, "", nil
}

// NOTE: This is currently not used. It is kept because I have implemented it
func readIteratorStatesPagination(stateIterator shim.StateQueryIteratorInterface, metadata *peer.QueryResponseMetadata) ([][]byte, error) {
	statesInRange := make([][]byte, metadata.GetFetchedRecordsCount())
	i := 0
	for stateIterator.HasNext() {
		kv, err := stateIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("could not get next state by range: %v", err)
		}

		statesInRange[i] = kv.GetValue()
		i++
	}

	return statesInRange, nil
}

func getRangeCompositeKeys(stub shim.ChaincodeStubInterface, objectType string, startPrefixes, endPrefixes []string) (string, string, error) {
	startKey, err := stub.CreateCompositeKey(objectType, startPrefixes)
	if err != nil {
		return "", "", fmt.Errorf("could not create composite key for: %v", err)
	}
	endKey, err := stub.CreateCompositeKey(objectType, endPrefixes)
	if err != nil {
		return "", "", fmt.Errorf("could not create composite key for: %v", err)
	}
	return endKey, startKey, nil
}

func GetStatesByRangeCompositeKey(stub shim.ChaincodeStubInterface, objectType string, startPrefixes, endPrefixes []string) ([][]byte, error) {
	endKey, startKey, err := getRangeCompositeKeys(stub, objectType, startPrefixes, endPrefixes)
	if err != nil {
		return nil, fmt.Errorf("could not create composite key for: %v", err)
	}

	states := [][]byte{}
	for {
		stateIterator, err := stub.GetStateByRange(startKey, endKey)
		if err != nil {
			return nil, fmt.Errorf("could not get state by range for start key %s and end key %s: %v", startKey, endKey, err)
		}
		defer stateIterator.Close()

		statesInRange, lastKey, err := readIteratorStates(stateIterator)
		if err != nil {
			return nil, fmt.Errorf("could not read iterator states: %v", err)
		}
		states = append(states, statesInRange...)

		// If lastKey is not empty, we should read again.
		// GetStateByRange is capped by "totalQueryLimit"
		if lastKey != "" {
			startKey = lastKey + string(maxUnicodeRuneValue)
		} else {
			break
		}

	}
	return states, nil
}

// NOTE: This is currently not used. It is kept because I have implemented it
func GetStatesByRangeCompositeKeyReadOnly(stub shim.ChaincodeStubInterface, objectType string, startPrefixes, endPrefixes []string) ([][]byte, error) {
	endKey, startKey, err := getRangeCompositeKeys(stub, objectType, startPrefixes, endPrefixes)
	if err != nil {
		return nil, fmt.Errorf("could not create composite key for: %v", err)
	}

	bookmark := ""
	states := [][]byte{}
	for {
		stateIterator, metadata, err := stub.GetStateByRangeWithPagination(
			startKey,
			endKey,
			PAGE_SIZE, bookmark)
		if err != nil {
			return nil, fmt.Errorf("could not get state by range for start key %s and end key %s: %v", startKey, endKey, err)
		}
		defer stateIterator.Close()

		if metadata.FetchedRecordsCount == 0 {
			break
		}

		statesInRange, err := readIteratorStatesPagination(stateIterator, metadata)
		if err != nil {
			return nil, fmt.Errorf("could not read iterator states: %v", err)
		}
		states = append(states, statesInRange...)

		bookmark = metadata.Bookmark
		if bookmark == "" {
			break
		}

	}

	return states, nil
}

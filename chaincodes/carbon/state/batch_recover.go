package state

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

const PAGE_SIZE = 30000
const maxUnicodeRuneValue = utf8.MaxRune

// TODO: implement
func GetStatesByRange(stub shim.ChaincodeStubInterface, objectType string, key string) ([][]byte, error) {
	return nil, nil
}

func GetStatesByRangeCompositeKey[T any](stub shim.ChaincodeStubInterface, objectType string, startPrefixes, endPrefixes []string) ([]*T, error) {
	endKey, startKey, err := getRangeCompositeKeys(stub, objectType, startPrefixes, endPrefixes)
	if err != nil {
		return nil, fmt.Errorf("could not create composite key for: %v", err)
	}

	states, err := getStatesByRange[T](stub, startKey, endKey)
	return states, nil
}

func GetStatesBytesByRangeCompositeKey(stub shim.ChaincodeStubInterface, objectType string, startPrefixes, endPrefixes []string) ([][]byte, error) {
	endKey, startKey, err := getRangeCompositeKeys(stub, objectType, startPrefixes, endPrefixes)
	if err != nil {
		return nil, fmt.Errorf("could not create composite key for: %v", err)
	}

	states, err := getStatesBytesByRange(stub, startKey, endKey)
	return states, err
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

// TODO: TEST THIS
func GetStatesBytesByPartialCompositeKey(stub shim.ChaincodeStubInterface, objectType string, prefixes []string) ([][]byte, error) {
	startKey, err := stub.CreateCompositeKey(objectType, prefixes)
	endKey := startKey + string(maxUnicodeRuneValue)
	if err != nil {
		return nil, fmt.Errorf("could not create composite key for: %v", err)
	}

	states, err := getStatesBytesByRange(stub, startKey, endKey)
	return states, err
}

func GetStatesByPartialCompositeKey[T any](stub shim.ChaincodeStubInterface, objectType string, prefixes []string) ([]*T, error) {
	startKey, err := stub.CreateCompositeKey(objectType, prefixes)
	endKey := startKey + string(maxUnicodeRuneValue)
	if err != nil {
		return nil, fmt.Errorf("could not create composite key for: %v", err)
	}

	states, err := getStatesByRange[T](stub, startKey, endKey)
	return states, err
}

func getStatesByRange[T any](stub shim.ChaincodeStubInterface, startKey, endKey string) ([]*T, error) {
	states := []*T{}

	for {
		stateIterator, err := stub.GetStateByRange(startKey, endKey)
		if err != nil {
			return nil, fmt.Errorf("could not get state by range for start key %s and end key %s: %v", startKey, endKey, err)
		}
		defer stateIterator.Close()

		statesInRange, lastKey, err := readUnmarshalledIteratorStates[T](stateIterator)
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

func getStatesBytesByRange(stub shim.ChaincodeStubInterface, startKey, endKey string) ([][]byte, error) {
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

func readUnmarshalledIteratorStates[T any](stateIterator shim.StateQueryIteratorInterface) ([]*T, string, error) {
	statesInRange := []*T{}
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

		var unmarshalledState T
		err = json.Unmarshal(kv.GetValue(), &unmarshalledState)
		if err != nil {
			return nil, "", fmt.Errorf("could not unmarshal state: %v", err)
		}

		statesInRange = append(statesInRange, &unmarshalledState)
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

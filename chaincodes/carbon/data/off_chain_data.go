package data

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	OFF_CHAIN_DATA_PREFIX = "offChainData"
)

type OffChainData struct {
	Uri         string `json:"uri"`
	Method      string `json:"method"`
	ReflectType string `json:"relflectType"`
	Hash        []byte `json:"hash"`
	DataBytes   []byte `json:"-"` // this prevents the field from being serialized
}

var _ state.WorldStateManager = (*OffChainData)(nil)

// TODO:
func (ocd *OffChainData) loadDataFromUri() error {
	if ocd.Uri == "" {
		return fmt.Errorf("uri is empty")
	}
	if ocd.Method == "" {
		return fmt.Errorf("method is empty")
	}
	if ocd.ReflectType == "" {
		return fmt.Errorf("reflectType is empty")
	}
	if ocd.DataBytes != nil {
		return fmt.Errorf("dataBytes is not empty")
	}
	if ocd.Hash == nil {
		return fmt.Errorf("hash is nil")
	}

	// TODO: load from uri using http
	ocd.DataBytes = []byte{}

	calcHash := sha256.Sum256(ocd.DataBytes)
	result := bytes.Compare(calcHash[:], ocd.Hash[:])
	if result != 0 {
		return fmt.Errorf("hashes do not match")
	}

	return nil
}

func (ocd *OffChainData) FromJson() (any, error) {
	err := ocd.loadDataFromUri()
	if err != nil {
		return nil, fmt.Errorf("could not load data from uri: %v", err)
	}

	reflectType, ok := ReflectToTypes[ocd.ReflectType]
	if !ok {
		return nil, fmt.Errorf("could not find reflect type for %s", ocd.ReflectType)
	}
	reflectValue := reflect.New(reflectType).Interface()
	err = json.Unmarshal(ocd.DataBytes, &reflectValue)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal data bytes: %v", err)
	}

	return reflectValue, nil
}

// TODO:
func (ocd *OffChainData) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	panic("not implemented")
	return nil
}

func (ocd *OffChainData) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutStateWithCompositeKey(stub, OFF_CHAIN_DATA_PREFIX, ocd.GetID(), ocd)
}

func (ocd *OffChainData) GetID() *[][]string {
	return &[][]string{{ocd.Uri}}
}

package data

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	OFF_CHAIN_DATA_PREFIX = "offChainData"
)

// TODO: should this be private and deal with access tokens?
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

	var err error
	ocd.DataBytes, err = GetBytesFromUri(ocd.Uri, ocd.Method)
	if err != nil {
		return fmt.Errorf("could not get data from uri: %v", err)
	}

	calcHash := sha256.Sum256(ocd.DataBytes)
	result := bytes.Compare(calcHash[:], ocd.Hash[:])
	if result != 0 {
		return fmt.Errorf("hashes do not match")
	}

	return nil
}

func (ocd *OffChainData) FromJson() (any, error) {
	if ocd.DataBytes == nil {
		err := ocd.loadDataFromUri()
		if err != nil {
			return nil, fmt.Errorf("could not load data from uri: %v", err)
		}
	}

	reflectType, ok := ReflectToTypes[ocd.ReflectType]
	if !ok {
		return nil, fmt.Errorf("could not find reflect type for %s", ocd.ReflectType)
	}
	reflectValue := reflect.New(reflectType).Interface()
	err := json.Unmarshal(ocd.DataBytes, &reflectValue)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal data bytes: %v", err)
	}

	return reflectValue, nil
}

// TODO:
func (ocd *OffChainData) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetStateWithCompositeKey(stub, OFF_CHAIN_DATA_PREFIX, keyAttributes, ocd)
}

func (ocd *OffChainData) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutStateWithCompositeKey(stub, OFF_CHAIN_DATA_PREFIX, ocd.GetID(), ocd)
}

func (ocd *OffChainData) GetID() *[][]string {
	return &[][]string{{ocd.Uri}}
}

func PublishOffChainData(stub shim.ChaincodeStubInterface, uri string, method string, reflectType string, hash []byte) (*OffChainData, error) {
	if _, ok := ReflectToTypes[reflectType]; !ok {
		return nil, fmt.Errorf("could not find reflect type for %s", reflectType)
	}

	offChainData := &OffChainData{
		Uri:         uri,
		Method:      method,
		ReflectType: reflectType,
		Hash:        hash[:],
	}

	if _, err := offChainData.FromJson(); err != nil {
		return nil, fmt.Errorf("could not load data: %v", err)
	}

	err := offChainData.ToWorldState(stub)
	if err != nil {
		return nil, fmt.Errorf("could not put off chain data in world state: %v", err)
	}

	return offChainData, nil
}

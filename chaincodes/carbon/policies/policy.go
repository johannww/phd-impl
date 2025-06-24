package policies

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
)

const (
	ACTIVE_POL_PREFIX = "activePolicies"
)

type IndependentPolicy interface {
}

type CoupledPolicy interface {
}

type Name string

const (
	// Coupled policy names
	DISTANCE       Name = "distance"
	WIND_DIRECTION Name = "wind_direction"

	// Independent policy names
	VEGETATION   Name = "vegetation"
	AUDIT_METHOD Name = "audit_method"
	TEMPERATURE  Name = "temperature"
)

var DefinedPolicies = map[Name](func(*PolicyInput) float64){
	DISTANCE:       nil,
	WIND_DIRECTION: nil,
	VEGETATION:     nil,
	AUDIT_METHOD:   nil,
	TEMPERATURE:    nil,
}

// TODO: implement this
func MintIndependentMult(chunk *properties.PropertyChunk) float64 {
	return 1.0
}

func SetActivePolicies(stub shim.ChaincodeStubInterface, activePolicies []Name) error {
	if len(activePolicies) == 0 {
		return fmt.Errorf("active policies cannot be empty")
	}

	for _, policy := range activePolicies {
		if _, exists := DefinedPolicies[policy]; !exists {
			return fmt.Errorf("policy %s is not defined", policy)
		}
	}

	policyBytes, err := json.Marshal(activePolicies)
	if err != nil {
		return fmt.Errorf("could not marshal active policies: %v", err)
	}

	return stub.PutState(ACTIVE_POL_PREFIX, policyBytes)
}

func AppendActivePolicy(stub shim.ChaincodeStubInterface, policy Name) error {
	if _, exists := DefinedPolicies[policy]; !exists {
		return fmt.Errorf("policy %s is not coded in this chaincode", policy)
	}

	policyBytes, err := stub.GetState(ACTIVE_POL_PREFIX)
	if err != nil || policyBytes == nil {
		return fmt.Errorf("could not get active policies: %v", err)
	}

	var activePolicies []Name
	if err := json.Unmarshal(policyBytes, &activePolicies); err != nil {
		return fmt.Errorf("could not unmarshal active policies: %v", err)
	}

	// Check if policy already exists
	for _, p := range activePolicies {
		if p == policy {
			return nil
		}
	}

	activePolicies = append(activePolicies, policy)
	policyBytes, err = json.Marshal(activePolicies)
	if err != nil {
		return fmt.Errorf("could not marshal active policies: %v", err)
	}

	return stub.PutState(ACTIVE_POL_PREFIX, policyBytes)
}

func DeleteActivePolicy(stub shim.ChaincodeStubInterface, policy Name) error {
	policyBytes, err := stub.GetState(ACTIVE_POL_PREFIX)
	if err != nil || policyBytes == nil {
		return fmt.Errorf("could not get active policies: %v", err)
	}

	var activePolicies []Name
	if err := json.Unmarshal(policyBytes, &activePolicies); err != nil {
		return fmt.Errorf("could not unmarshal active policies: %v", err)
	}

	found := false
	newActivePolicies := slices.DeleteFunc(activePolicies, func(p Name) bool {
		mustDelete := p == policy
		if mustDelete {
			found = true
		}
		return mustDelete
	})

	if !found {
		return fmt.Errorf("policy %s is not active", policy)
	}

	policyBytes, err = json.Marshal(newActivePolicies)
	if err != nil {
		return fmt.Errorf("could not marshal active policies: %v", err)
	}

	return stub.PutState(ACTIVE_POL_PREFIX, policyBytes)
}

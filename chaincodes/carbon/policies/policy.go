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
	MULTIPLIER_SCALE  = 1000 // Scale for multipliers to avoid floating point precision issues
	MULTIPLIER_MIN    = 0    // Minimum multiplier value adjusted to the scale
	MULTIPLIER_MAX    = 2000 // Maximum multiplier value adjusted to the scale
)

type IndependentPolicy interface {
}

type CoupledPolicy interface {
}

type Name string

type PolicyFunc func(*PolicyInput) int64

const (
	// Coupled policy names
	DISTANCE       Name = "distance"
	WIND_DIRECTION Name = "wind_direction"

	// Independent policy names
	VEGETATION   Name = "vegetation"
	AUDIT_METHOD Name = "audit_method"
	TEMPERATURE  Name = "temperature"
)

// TODO: add the actual implementations of the policies
var DefinedPolicies = map[Name]PolicyFunc{
	DISTANCE:       DistancePolicy,
	WIND_DIRECTION: WindPolicy,
	VEGETATION:     nil,
	AUDIT_METHOD:   nil,
	TEMPERATURE:    nil,
}

// TODO: implement this
func MintIndependentMult(chunk *properties.PropertyChunk) int64 {
	return 1.0
}

// TODO: Review
func MintCoupledMult(input *PolicyInput, activePolicies []Name) (int64, error) {
	mult := int64(0)
	for _, policy := range activePolicies {
		if !isCoupledPolicy(policy) {
			continue // skip independent policies
		}

		if policyFunc, exists := DefinedPolicies[policy]; exists {
			if policyFunc == nil {
				return 0, fmt.Errorf("policy %s is not implemented", policy)
			}
			// Multiplier represents the extra quantity that can be acquired.
			// Since we have the integer scale, we do the following transformation:
			mult = ((mult + MULTIPLIER_SCALE) * (policyFunc(input) + MULTIPLIER_SCALE) / MULTIPLIER_SCALE) - MULTIPLIER_SCALE
			mult = boundMult(mult)
		} else {
			return 0, fmt.Errorf("policy %s is not defined", policy)
		}

	}

	return mult, nil
}

func isCoupledPolicy(policy Name) bool {
	return policy == DISTANCE || policy == WIND_DIRECTION
}

func boundMult(mult int64) int64 {
	return min(max(mult, MULTIPLIER_MIN), MULTIPLIER_MAX)
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

func GetActivePolicies(stub shim.ChaincodeStubInterface) ([]Name, error) {
	policyBytes, err := stub.GetState(ACTIVE_POL_PREFIX)
	if err != nil {
		return nil, fmt.Errorf("could not get active policies: %v", err)
	}

	if policyBytes == nil {
		return nil, fmt.Errorf("no active policies found")
	}

	var activePolicies []Name
	if err := json.Unmarshal(policyBytes, &activePolicies); err != nil {
		return nil, fmt.Errorf("could not unmarshal active policies: %v", err)
	}

	return activePolicies, nil
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

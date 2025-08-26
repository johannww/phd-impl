package policies

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoupledMultiplierOverlap(t *testing.T) {
	input := &PolicyInput{}

	mockMultValue := int64(300)
	DefinedPoliciesStatic[DISTANCE] = mockCoupledPolicy(mockMultValue)
	DefinedPoliciesStatic[WIND_DIRECTION] = mockCoupledPolicy(mockMultValue)
	pApplier := NewPolicyApplier()
	activePolicies := []Name{DISTANCE, WIND_DIRECTION}

	multiplier, err := pApplier.MintCoupledMult(input, activePolicies)
	require.NoError(t, err)

	floatCalculation := (1 + float64(mockMultValue)/MULTIPLIER_SCALE) * (1 + float64(mockMultValue)/MULTIPLIER_SCALE)

	require.Equal(t, int64((floatCalculation-1)*MULTIPLIER_SCALE), multiplier)

}

func TestMultiplierBound(t *testing.T) {
	input := &PolicyInput{}

	mockMultValue := int64(750)
	DefinedPoliciesStatic[DISTANCE] = mockCoupledPolicy(mockMultValue)
	DefinedPoliciesStatic[WIND_DIRECTION] = mockCoupledPolicy(mockMultValue)
	pApplier := NewPolicyApplier()
	activePolicies := []Name{DISTANCE, WIND_DIRECTION}

	multiplier, err := pApplier.MintCoupledMult(input, activePolicies)
	require.NoError(t, err)

	require.Equal(t, multiplier, int64(MULTIPLIER_MAX))

}

func mockCoupledPolicy(returnValue int64) PolicyFunc {
	return func(input *PolicyInput) int64 {
		return returnValue
	}
}

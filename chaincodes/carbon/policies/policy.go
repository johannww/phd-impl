package policies

import "github.com/johannww/phd-impl/chaincodes/carbon/properties"

type Name string

type PolicyFunc func(*PolicyInput) int64

type PolicyApplier interface {
	MintIndependentMult(chunk *properties.PropertyChunk) int64
	MintCoupledMult(input *PolicyInput, activePolicies []Name) (int64, error)
}

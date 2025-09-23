package policies

type Name string

type PolicyFunc func(*PolicyInput) int64

type PolicyApplier interface {
	MintIndependentMult(input *PolicyInput, activePolicies []Name) (int64, error)
	BurnIndependentMult(input *PolicyInput, activePolicies []Name) (int64, error)
	MintCoupledMult(input *PolicyInput, activePolicies []Name) (int64, error)
}

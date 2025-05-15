package policies

import "github.com/johannww/phd-impl/chaincodes/carbon/properties"

type IndependentPolicy interface {
}

type CoupledPolicy interface {
}

// TODO: implement this
func MintIndependentMult(chunk *properties.PropertyChunk) float64 {
	return 1.0
}

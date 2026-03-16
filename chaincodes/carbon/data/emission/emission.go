package emission

import "github.com/johannww/phd-impl/chaincodes/common/identities"

const (
	EMISSION_PREFIX = "emission"
)

type Emission struct {
	Emitter   identities.Identity
	Quantity  float64
	Timeframe [2]int
}

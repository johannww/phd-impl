package emission

import "github.com/johannww/phd-impl/chaincodes/carbon/identities"

const (
	EMISSION_PREFIX = "emission"
)

type Emission struct {
	emitter   identities.Identity
	quantity  float64
	timeframe [2]int
}

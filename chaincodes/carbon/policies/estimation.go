package policies

import (
	mathrand "math/rand"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
)

type Estimator struct{}

// TODO: implement
func (e *Estimator) Estimate(chunk *properties.PropertyChunk, intervalStart, intervalEnd time.Time) (int64, error) {
	quantity := int64(mathrand.Intn(1000) + 1) // Random quantity between 1 and 1000
	return quantity, nil
}

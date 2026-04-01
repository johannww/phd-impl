package policies

import (
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/common"
	"github.com/johannww/phd-impl/chaincodes/carbon/data/registry"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
)

// tCO2PerHaPerYear is the assumed annual CO₂ sequestration rate in tCO₂/ha/year,
// representative of a tropical Legal Reserve / APP forest.
const tCO2PerHaPerYear = 5

// Estimator computes the expected carbon credit quantity for a property chunk
// over a given time interval, based on the preserved area declared in the registry.
type Estimator struct {
	Summary *registry.RegistrySummary
}

// Estimate returns a simulated carbon credit quantity (scaled by common.QUANTITY_SCALE)
// for a single chunk over [intervalStart, intervalEnd].
// This mock distributes the property's LegalForestArea evenly across all chunks and applies
// a fixed tCO2PerHaPerYear sequestration rate pro-rated to the interval length.
// If no registry summary is available it falls back to a fixed quantity per day.
func (e *Estimator) Estimate(chunk *properties.PropertyChunk, intervalStart, intervalEnd time.Time) (int64, error) {
	intervalDays := int64(intervalEnd.Sub(intervalStart).Hours() / 24)
	if intervalDays <= 0 {
		intervalDays = 1
	}

	if e.Summary == nil || e.Summary.LegalForestArea <= 0 {
		// Fallback: 1 credit-unit per day when no registry data is available.
		return intervalDays * common.QUANTITY_SCALE, nil
	}

	nChunks := int64(chunk.PropertyID) // used only for distribution; caller passes all chunks
	if nChunks <= 0 {
		nChunks = 1
	}

	// tCO₂ = (LegalForestArea / nChunks) × (tCO2PerHaPerYear / 365) × intervalDays
	// Rearranged to stay in integers:
	//   = LegalForestArea × tCO2PerHaPerYear × intervalDays / (365 × nChunks)
	quantity := int64(e.Summary.LegalForestArea) * tCO2PerHaPerYear * intervalDays / (365 * nChunks)
	if quantity <= 0 {
		quantity = 1
	}

	return quantity * common.QUANTITY_SCALE, nil
}

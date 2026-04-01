package policies

import "time"

const (
	// freshnessMaxAgeDays is the age (in days) at which the multiplier reaches zero.
	freshnessMaxAgeDays = 5 * 365
	// freshnessFullBonusDays is the age (in days) below which the full bonus is awarded.
	freshnessFullBonusDays = 1 * 365
)

// RegistryFreshnessPolicy is an example policy and  returns a mint multiplier
// based on how recently the registry record was last updated.
func RegistryFreshnessPolicy(input *PolicyInput) int64 {
	if input.RegistrySummary == nil || input.RegistrySummary.LastUpdate.IsZero() {
		return 0
	}

	ageInDays := int64(time.Since(input.RegistrySummary.LastUpdate).Hours() / 24)

	if ageInDays <= freshnessFullBonusDays {
		return MULTIPLIER_SCALE
	}

	if ageInDays >= freshnessMaxAgeDays {
		return 0
	}

	// Linear decay between freshnessFullBonusDays and freshnessMaxAgeDays.
	decayRange := int64(freshnessMaxAgeDays - freshnessFullBonusDays)
	ageInDecayRange := ageInDays - freshnessFullBonusDays
	return MULTIPLIER_SCALE - (MULTIPLIER_SCALE * ageInDecayRange / decayRange)
}

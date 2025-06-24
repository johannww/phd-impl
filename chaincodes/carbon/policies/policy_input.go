package policies

import (
	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
)

// TODO: Add fields like property, companies, etc
type PolicyInput struct {
	Chunk   *properties.PropertyChunk
	Company *companies.Company
}

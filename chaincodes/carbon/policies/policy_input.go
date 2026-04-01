package policies

import (
	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/data/registry"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
)

type PolicyInput struct {
	Chunk           *properties.PropertyChunk `json:"chunk,omitempty"`
	Company         *companies.Company        `json:"company,omitempty"`
	RegistrySummary *registry.RegistrySummary `json:"registrySummary,omitempty"`
	DataFetcher     data.DataFetcher          `json:"-"`
}

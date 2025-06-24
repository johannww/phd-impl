package companies

import (
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
)

// TODO: continue implementing
type Company struct {
	Coordinate properties.Coordinate
	DataProps  []*data.ValidationProps
}

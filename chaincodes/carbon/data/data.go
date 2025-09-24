package data

import (
	"github.com/johannww/phd-impl/chaincodes/carbon/data/metheorological"
	"github.com/johannww/phd-impl/chaincodes/carbon/utils"
)

type DataFetcher interface {
	GetWindData(coord *utils.Coordinate) metheorological.Wind
	GetTemperatureData(coord *utils.Coordinate) metheorological.Temperature
}

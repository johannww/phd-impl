package data

import (
	"reflect"

	"github.com/johannww/phd-impl/chaincodes/carbon/data/emission"
)

var ReflectToTypes = map[string]reflect.Type{
	"emission": reflect.TypeOf(emission.Emission{}),
}

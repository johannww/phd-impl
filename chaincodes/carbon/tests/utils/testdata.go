package utils_test

import (
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
)

// TestData holds a list as an identity map
// The map key is a string and the value is generic interface{}
type TestData struct {
	Identities *setup.MockIdentities
	Properties []*properties.Property
}

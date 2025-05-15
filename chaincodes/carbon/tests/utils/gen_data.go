package utils_test

import (
	"fmt"
	"strings"

	mathrand "math/rand"

	"github.com/hyperledger/fabric-chaincode-go/pkg/attrmgr"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	"github.com/johannww/phd-impl/chaincodes/carbon/vegetation"
)

func GenData(
	nOwners int,
	nChunks int,
	nCompanies int,
) *TestData {
	data := &TestData{}

	mockIds := &setup.MockIdentities{}
	GenOwnerIDs(nOwners, mockIds)
	GenCompanyIDs(nCompanies, mockIds)
	props := GenProperties(nChunks, mockIds)

	data.Identities = mockIds
	data.Properties = props

	return data

}

func GenOwnerIDs(n int, mockIds *setup.MockIdentities) {
	for i := 0; i < n; i++ {
		ownerName := fmt.Sprintf("owner%d", i)

		mockId := setup.GenerateHFSerializedIdentity(
			setup.X509_TYPE,
			&attrmgr.Attributes{
				Attrs: map[string]string{},
			}, "GOV", ownerName,
		)
		(*mockIds)[ownerName] = mockId

	}
}

func GenCompanyIDs(n int, mockIds *setup.MockIdentities) {

	for i := 0; i < n; i++ {
		companyName := fmt.Sprintf("company%d", i)

		mockId := setup.GenerateHFSerializedIdentity(
			setup.IDEMIX_TYPE,
			&attrmgr.Attributes{
				Attrs: map[string]string{},
			}, "GOV", companyName,
		)
		(*mockIds)[companyName] = mockId

	}
}

func GenProperties(nChunks int, mockIds *setup.MockIdentities) []*properties.Property {

	props := []*properties.Property{}

	for key := range *mockIds {
		if !strings.Contains(key, "owner") {
			continue
		}

		// TODO: this should be tought later to avoid collisions
		id := mathrand.Uint64()

		prop := &properties.Property{
			OwnerID: key,
			ID:      id,
		}
		props = append(props, prop)

		for range nChunks {
			chunk := &properties.PropertyChunk{
				PropertyID: id,
				// NOTE: chunks will probably be spread
				Coordinates: []properties.Coordinate{
					{
						Latitude:  mathrand.Float64()*180 - 90,
						Longitude: mathrand.Float64()*360 - 180,
					},
				},
				VegetationsProps: []vegetation.VegetationProps{},
			}
			prop.Chunks = append(prop.Chunks, chunk)
		}
	}

	return props
}

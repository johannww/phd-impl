package utils_test

import (
	"fmt"
	"strings"
	"time"

	mathrand "math/rand"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/attrmgr"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	"github.com/johannww/phd-impl/chaincodes/carbon/vegetation"
)

func GenData(
	nOwners int,
	nChunks int,
	nCompanies int,
	startTimestamp string,
	endTimestamp string,
	issueInterval time.Duration,
) *TestData {
	data := &TestData{}

	mockIds := &setup.MockIdentities{}
	GenOwnerIDs(nOwners, mockIds)
	GenCompanyIDs(nCompanies, mockIds)
	props := GenProperties(nChunks, mockIds)

	startTs, err := time.Parse(time.RFC3339, startTimestamp)
	if err != nil {
		panic(err)
	}
	endTs, err := time.Parse(time.RFC3339, endTimestamp)
	if err != nil {
		panic(err)
	}

	mintCredits := GenMintCredits(props, startTs, endTs, issueInterval)

	data.Identities = mockIds
	data.Properties = props
	data.MintCredits = mintCredits

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
			chunk := chunkForProperty(prop)
			prop.Chunks = append(prop.Chunks, chunk)
		}
	}

	return props
}

func GenMintCredits(props []*properties.Property, startTs, endTs time.Time, issueInterval time.Duration) []*credits.MintCredit {
	nDurations := int64(endTs.Sub(startTs) / issueInterval)
	mintCredits := []*credits.MintCredit{}

	for _, prop := range props {
		for i := range prop.Chunks {
			chunk := prop.Chunks[i]
			for nDuration := int64(0); nDuration < nDurations; nDuration++ {
				lastIssue := startTs.Add(time.Duration(nDuration-1) * issueInterval)
				issueTs := startTs.Add(time.Duration(nDuration) * issueInterval)
				issueTsStr := issueTs.Format(time.RFC3339)

				estimator := policies.Estimator{}
				quantity, err := estimator.Estimate(chunk, lastIssue, issueTs)
				panicOnError(err)

				credit := creditForChunk(chunk, prop, quantity, issueTsStr)
				mintCredits = append(mintCredits, credit)
			}
		}
	}
	return mintCredits

}

func chunkForProperty(prop *properties.Property) *properties.PropertyChunk {
	chunk := &properties.PropertyChunk{
		PropertyID: prop.ID,
		// NOTE: chunks will be spread
		Coordinates: []properties.Coordinate{
			{
				Latitude:  mathrand.Float64()*180 - 90,
				Longitude: mathrand.Float64()*360 - 180,
			},
		},
		VegetationsProps: []vegetation.VegetationProps{},
		ValidationProps:  []data.ValidationProps{},
	}
	return chunk
}

func creditForChunk(chunk *properties.PropertyChunk,
	prop *properties.Property,
	quantity int64,
	issueTsStr string) *credits.MintCredit {
	credit := &credits.MintCredit{
		Credit: credits.Credit{
			OwnerID:  prop.OwnerID,
			ChunkID:  (*chunk.GetID())[0],
			Chunk:    chunk,
			Quantity: quantity,
		},
		MintMult:      policies.MintIndependentMult(chunk),
		MintTimeStamp: issueTsStr,
	}
	return credit
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

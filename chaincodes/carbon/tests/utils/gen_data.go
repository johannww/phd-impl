package utils_test

import (
	"fmt"
	"strings"
	"time"

	mathrand "math/rand"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/attrmgr"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/payment"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	"github.com/johannww/phd-impl/chaincodes/carbon/vegetation"
)

const (
	COMPANY_PREFIX = "company"
	OWNER_PREFIX   = "owner"
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

	defaultIdentities := (setup.SetupIdentities(nil))
	mockIds := &defaultIdentities
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

	creditWalletsMap := GenCreditWalletsMap(mockIds)

	mintCredits := GenMintCredits(props, creditWalletsMap, startTs, endTs, issueInterval)

	// TODO: Add virtual payment token
	tokenWallets := GenTokenWallets(mockIds)

	// Convert creditWalletsMap to slice
	creditWallets := []*credits.CreditWallet{}
	for _, cw := range creditWalletsMap {
		creditWallets = append(creditWallets, cw)
	}

	data.Identities = mockIds
	data.Properties = props
	data.MintCredits = mintCredits
	data.CreditWallets = creditWallets
	data.TokenWallets = tokenWallets

	return data

}

func GenOwnerIDs(n int, mockIds *setup.MockIdentities) {
	for i := 0; i < n; i++ {
		ownerName := fmt.Sprintf("%s%d", OWNER_PREFIX, i)

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
		companyName := fmt.Sprintf("%s%d", COMPANY_PREFIX, i)

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
		if !strings.Contains(key, OWNER_PREFIX) {
			continue
		}

		// TODO: this should be tought later to avoid collisions
		id := mathrand.Uint64()

		// TODOHP: I should get the OwnerID from the stub/cid creator
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

func GenCreditWalletsMap(mockIds *setup.MockIdentities) map[string]*credits.CreditWallet {
	creditWalletsMap := map[string]*credits.CreditWallet{}

	for ownerId := range *mockIds {
		creditWalletsMap[ownerId] = &credits.CreditWallet{
			OwnerID: ownerId,
		}
	}

	return creditWalletsMap
}

func GenMintCredits(
	props []*properties.Property,
	creditWalletsMap map[string]*credits.CreditWallet,
	startTs, endTs time.Time,
	issueInterval time.Duration,
) []*credits.MintCredit {
	nDurations := int64(endTs.Sub(startTs) / issueInterval)
	mintCredits := []*credits.MintCredit{}

	for _, prop := range props {
		if creditWalletsMap[prop.OwnerID] == nil {
			creditWalletsMap[prop.OwnerID] = &credits.CreditWallet{
				OwnerID: prop.OwnerID,
			}
		}

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
				creditWalletsMap[prop.OwnerID].Quantity += quantity
			}
		}
	}

	return mintCredits

}

func GenTokenWallets(mockIds *setup.MockIdentities) []*payment.VirtualTokenWallet {
	wallets := []*payment.VirtualTokenWallet{}

	for key := range *mockIds {
		quantity := int64(0)
		if strings.Contains(key, COMPANY_PREFIX) {
			// Companies start with a random quantity of tokens
			quantity = 20000000
		}

		wallet := &payment.VirtualTokenWallet{
			OwnerID:  key,
			Quantity: quantity,
		}
		wallets = append(wallets, wallet)
	}

	return wallets
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
		VegetationsProps: &vegetation.VegetationProps{},
		ValidationProps: &data.ValidationProps{
			// assuming satellite validation for simplicity
			Methods: []data.ValidationMethod{data.ValidationMethodSattelite},
		},
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

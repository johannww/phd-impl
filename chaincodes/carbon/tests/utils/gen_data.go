package utils_test

import (
	"fmt"
	"strings"
	"time"

	mathrand "math/rand"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/attrmgr"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/payment"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	"github.com/johannww/phd-impl/chaincodes/carbon/vegetation"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/state/mocks"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
)

const (
	COMPANY_PREFIX = "test_company"
	OWNER_PREFIX   = "test_owner"
)

// mockStub helps generating chaincode identity strings
var mockStub = mocks.NewMockStub("carbon", nil)

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
	data.Identities = &defaultIdentities
	GenOwnerIDs(nOwners, data.Identities)
	GenCompanyIDs(nCompanies, data.Identities)
	data.Properties = GenProperties(nChunks, data.Identities)

	data.Companies, data.PseudonymMap = GenCompaniesWithPseudonyms(nCompanies, data.Identities)

	startTs, err := time.Parse(time.RFC3339, startTimestamp)
	if err != nil {
		panic(err)
	}
	endTs, err := time.Parse(time.RFC3339, endTimestamp)
	if err != nil {
		panic(err)
	}

	creditWalletsMap := GenCreditWalletsMap(data.Identities, data.PseudonymMap)
	// Convert creditWalletsMap to slice
	for _, cw := range creditWalletsMap {
		data.CreditWallets = append(data.CreditWallets, cw)
	}

	data.MintCredits = GenMintCredits(data.Properties, creditWalletsMap, startTs, endTs, issueInterval)

	data.TokenWallets = GenTokenWallets(data.Identities, data.PseudonymMap)

	return data

}

func GenDataWithBids(
	nOwners int,
	nChunks int,
	nCompanies int,
	startTimestamp string,
	endTimestamp string,
	issueInterval time.Duration,
) *TestData {
	data := GenData(nOwners, nChunks, nCompanies, startTimestamp, endTimestamp, issueInterval)

	mintingEndTs, err := time.Parse(time.RFC3339, endTimestamp)
	panicOnError(err)

	bidsStart := mintingEndTs.Add(time.Minute)

	GenRandomBidsForMintCredits(bidsStart, data)

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

	for i := range n {
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

		// NOTE: this should be tought later to avoid collisions
		id := mathrand.Uint64()

		prop := &properties.Property{
			OwnerID: getCidFromMockIdentity((*mockIds)[key]),
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

func GenCompaniesWithPseudonyms(n int, mockIds *setup.MockIdentities,
) ([]*companies.Company, []*companies.PseudonymToCompanyID) {
	companyList := []*companies.Company{}
	pseudonymList := []*companies.PseudonymToCompanyID{}

	for i := range n {
		companyName := fmt.Sprintf("%s%d", COMPANY_PREFIX, i)

		company := &companies.Company{
			ID: companyName,
			Coordinate: &utils.Coordinate{
				Latitude:  mathrand.Float64()*180 - 90,
				Longitude: mathrand.Float64()*360 - 180,
			},
			DataProps: &data.ValidationProps{
				Methods: []data.ValidationMethod{data.ValidationMethodSattelite},
			},
		}

		pseudonym := identities.PseudonymStrFromID((*mockIds)[companyName])
		pseudonymToCompanyID := &companies.PseudonymToCompanyID{
			Pseudonym: pseudonym,
			CompanyID: companyName,
		}

		companyList = append(companyList, company)
		pseudonymList = append(pseudonymList, pseudonymToCompanyID)
	}

	return companyList, pseudonymList
}

func GenCreditWalletsMap(
	mockIds *setup.MockIdentities,
	pseudonymMap []*companies.PseudonymToCompanyID,
) map[string]*credits.CreditWallet {
	creditWalletsMap := map[string]*credits.CreditWallet{}

	for ownerId := range *mockIds {
		if !strings.Contains(ownerId, OWNER_PREFIX) {
			continue // skip non-owners
		}
		creditWalletsMap[ownerId] = &credits.CreditWallet{
			OwnerID: getCidFromMockIdentity((*mockIds)[ownerId]),
		}
	}

	// Create credit wallets for companies using their pseudonyms
	for companyId := range pseudonymMap {
		pseudonym := pseudonymMap[companyId].Pseudonym
		creditWalletsMap[pseudonym] = &credits.CreditWallet{
			OwnerID: pseudonym,
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
	pApplier := policies.NewPolicyApplier()

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

				credit := creditForChunk(pApplier, chunk, prop, quantity, issueTsStr)
				mintCredits = append(mintCredits, credit)
				creditWalletsMap[prop.OwnerID].Quantity += quantity
			}
		}
	}

	return mintCredits

}

func GenTokenWallets(
	mockIds *setup.MockIdentities,
	pseudonymMap []*companies.PseudonymToCompanyID,
) []*payment.VirtualTokenWallet {
	wallets := []*payment.VirtualTokenWallet{}

	for key := range *mockIds {
		if strings.Contains(key, COMPANY_PREFIX) {
			continue // skip companies
		}

		wallet := &payment.VirtualTokenWallet{
			OwnerID:  getCidFromMockIdentity((*mockIds)[key]),
			Quantity: 0,
		}
		wallets = append(wallets, wallet)
	}

	for _, pseudonymToCompanyID := range pseudonymMap {
		wallet := &payment.VirtualTokenWallet{
			OwnerID:  pseudonymToCompanyID.Pseudonym,
			Quantity: 200000000000000,
		}
		wallets = append(wallets, wallet)
	}

	return wallets
}

func GenRandomBidsForMintCredits(issueStart time.Time, testData *TestData) {
	sellMinPrice := int64(1000)
	buyMinPrice := int64(1000)

	buyerIds := testData.PseudonymMap

	var issueTsStr string
	for i, mintCredit := range testData.MintCredits {
		issueTs := issueStart.Add(time.Duration(time.Duration(i) * time.Second)).UTC()
		issueTsStr = issueTs.Format(time.RFC3339)
		sellPrice := sellMinPrice + int64(mathrand.Intn(1000)) // Randomize sell price
		sellBid := &bids.SellBid{
			SellerID:  mintCredit.OwnerID,
			CreditID:  (*mintCredit.GetID())[0],
			Timestamp: issueTsStr,
			PrivatePrice: &bids.PrivatePrice{
				Price: sellPrice,
			},
			Quantity: mintCredit.Quantity,
		}
		sellBid.PrivatePrice.BidID = (*sellBid.GetID())[0]

		testData.SellBids = append(testData.SellBids, sellBid)

		buyerIdIndex := mathrand.Intn(len(buyerIds))

		buyPrice := buyMinPrice + int64(mathrand.Intn(1000)) // Randomize buy price
		bidAskQuantity := mintCredit.Quantity
		buyBid := &bids.BuyBid{
			BuyerID:     buyerIds[buyerIdIndex].Pseudonym,
			AskQuantity: bidAskQuantity,
			Timestamp:   issueTsStr,
			PrivateQuantity: &bids.PrivateQuantity{
				AskQuantity: bidAskQuantity,
			},
			PrivatePrice: &bids.PrivatePrice{
				Price: buyPrice,
			},
		}
		buyBid.PrivatePrice.BidID = (*buyBid.GetID())[0]
		buyBid.PrivateQuantity.BidID = (*buyBid.GetID())[0]

		testData.BuyBids = append(testData.BuyBids, buyBid)
	}

	testData.BidIssueLastTs = issueTsStr
}

func chunkForProperty(prop *properties.Property) *properties.PropertyChunk {
	chunk := &properties.PropertyChunk{
		PropertyID: prop.ID,
		// NOTE: chunks will be spread
		Coordinates: []utils.Coordinate{
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

func creditForChunk(pApplier policies.PolicyApplier,
	chunk *properties.PropertyChunk,
	prop *properties.Property,
	quantity int64,
	issueTsStr string) *credits.MintCredit {
	mintMult, err := pApplier.MintIndependentMult(
		&policies.PolicyInput{Chunk: chunk},
		[]policies.Name{policies.VEGETATION})
	panicOnError(err)

	credit := &credits.MintCredit{
		Credit: credits.Credit{
			OwnerID:  prop.OwnerID,
			ChunkID:  (*chunk.GetID())[0],
			Chunk:    chunk,
			Quantity: quantity,
		},
		MintMult:      mintMult,
		MintTimeStamp: issueTsStr,
	}
	return credit
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func getCidFromMockIdentity(mockId []byte) string {
	mockStub.Creator = mockId
	return identities.GetID(mockStub)
}

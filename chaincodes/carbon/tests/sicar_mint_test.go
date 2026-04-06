package carbon_tests

import (
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/data/registry"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	mocks "github.com/johannww/phd-impl/chaincodes/common/state/mocks"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"github.com/stretchr/testify/require"
)

// expectedMintQuantity computes the effective minted quantity by running the real
// policy applier against a minimal PolicyInput, mirroring what mintCreditInternal does.
func expectedMintQuantity(baseQuantity int64, lastUpdateDate time.Time) int64 {
	pApplier := policies.NewPolicyApplier()
	pInput := &policies.PolicyInput{
		RegistrySummary: &registry.RegistrySummary{LastUpdate: lastUpdateDate},
	}
	mult, _ := pApplier.MintIndependentMult(pInput, []policies.Name{policies.VEGETATION, policies.REGISTRY_FRESHNESS})
	return baseQuantity + (baseQuantity * mult / policies.MULTIPLIER_SCALE)
}

func TestMintCreditWithSicarValidation(t *testing.T) {
	// 1. Setup Mock SICAR Server
	registryPropID := "BR-SP-12345"
	lastUpdateStr := "15/01/2024"
	lastUpdateDate, parseErr := time.Parse("02/01/2006", lastUpdateStr)
	require.NoError(t, parseErr)

	sicarData := registry.SicarData{
		CodigoImovel:                    registryPropID,
		SituacaoImovel:                  "AT",
		DataUltimaAtualizacaoCadastro:   lastUpdateStr,
		AreaTotalImovel:                 100.0,
		AreaPreservacaoPermanente:       20.0,
		AreaReservaLegalDeclarada:       20.0,
		AreaRemanescenteVegetacaoNativa: 50.0,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fmt.Sprintf("/sicar/demonstrativoDegustacao/1.0/%s", registryPropID) {
			resp := struct {
				Result []registry.SicarData `json:"result"`
			}{
				Result: []registry.SicarData{sicarData},
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	server := httptest.NewTLSServer(handler)
	defer server.Close()

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})

	// 2. Setup Chaincode Stub and Identities
	stub := mocks.NewMockStub("carbon", nil)
	mockIds := setup.SetupIdentities(stub)

	// Admin registers the trusted provider
	stub.MockTransactionStart("tx_reg")
	stub.Creator = mockIds[identities.TrustedDBRegistrator]
	err := registry.AddTrustedProvider(stub, "SICAR", server.URL, certPEM)
	require.NoError(t, err)
	stub.MockTransactionEnd("tx_reg")

	// 3. Create Property and Chunk linked to SICAR
	ownerName := setup.CREDIT_OWNER_ID
	stub.Creator = mockIds[ownerName]
	ownerID := identities.GetID(stub)

	propertyID := uint64(1)
	registryProvider := "SICAR"

	chunkID := []string{fmt.Sprintf("%d", propertyID), "0.000000", "0.000000"}
	chunk := &properties.PropertyChunk{
		PropertyID: propertyID,
		Coordinates: []utils.Coordinate{
			{Latitude: 0.0, Longitude: 0.0},
		},
	}

	prop := &properties.Property{
		OwnerID:          ownerID,
		ID:               propertyID,
		RegistryID:       registryPropID,
		RegistryProvider: registryProvider,
		Chunks:           []*properties.PropertyChunk{chunk},
	}
	stub.MockTransactionStart("tx_prop")
	err = prop.ToWorldState(stub)
	require.NoError(t, err)
	stub.MockTransactionEnd("tx_prop")

	// Verify property and chunk are stored correctly
	propFromState := &properties.Property{
		OwnerID: ownerID,
		ID:      propertyID,
	}
	stub.MockTransactionStart("tx_get_prop")
	err = propFromState.FromWorldState(stub, (*propFromState.GetID())[0])
	require.NoError(t, err)
	require.Equal(t, len(prop.Chunks), len(propFromState.Chunks))
	stub.MockTransactionEnd("tx_get_prop")

	// 4. Setup Active Policies
	stub.MockTransactionStart("tx_pol")
	stub.Creator = mockIds[identities.PolicySetter]
	pApplier := policies.NewPolicyApplier()
	err = pApplier.SetActivePolicies(stub, []policies.Name{policies.VEGETATION, policies.REGISTRY_FRESHNESS})
	require.NoError(t, err)
	stub.MockTransactionEnd("tx_pol")

	// 4.5 Refresh Registry Data (Pre-fetch before minting)
	stub.MockTransactionStart("tx_refresh")
	_, err = registry.RefreshRegistryData(stub, "SICAR", registryPropID)
	require.NoError(t, err)
	stub.MockTransactionEnd("tx_refresh")

	// 5. Attempt Minting
	stub.Creator = mockIds[identities.CreditMinter]

	intervalStart := time.Now().Format(utils.RFC3339WithMillis)
	intervalEnd := time.Now().Add(1 * time.Hour).Format(utils.RFC3339WithMillis)

	// Case A: Valid Minting with Estimated Quantity
	stub.MockTransactionStart("tx_mint_est_ok")
	propIDAttr := []string{ownerID, fmt.Sprintf("%d", propertyID)}
	mcs, err := credits.MintEstimatedCreditsForProperty(stub, propIDAttr, intervalStart, intervalEnd)
	require.NoError(t, err)
	require.Len(t, mcs, 1)
	require.Greater(t, mcs[0].Quantity, int64(0))
	stub.MockTransactionEnd("tx_mint_est_ok")

	// Case B: Valid Minting with Explicit Quantity
	stub.MockTransactionStart("tx_mint_qty_ok")
	mintQuantity := int64(100)
	mcQty, err := credits.MintQuantityCreditForChunk(stub, propIDAttr, chunkID, mintQuantity, intervalEnd)
	require.NoError(t, err)
	require.NotNil(t, mcQty)
	// Expected quantity applies VEGETATION (1x) and REGISTRY_FRESHNESS (time-dependent linear decay).
	require.Equal(t, expectedMintQuantity(mintQuantity, lastUpdateDate), mcQty.Quantity)
	stub.MockTransactionEnd("tx_mint_qty_ok")

	// Case B.2: Valid Minting for all chunks of Property
	stub.MockTransactionStart("tx_mint_prop_qty_ok")
	propMintQuantity := int64(150)
	mcsProp, err := credits.MintQuantityCreditsForProperty(stub, propIDAttr, propMintQuantity, intervalEnd)
	require.NoError(t, err)
	require.Len(t, mcsProp, 1) // Property has 1 chunk
	// Expected quantity applies VEGETATION (1x) and REGISTRY_FRESHNESS (time-dependent linear decay).
	require.Equal(t, expectedMintQuantity(propMintQuantity, lastUpdateDate), mcsProp[0].Quantity)
	stub.MockTransactionEnd("tx_mint_prop_qty_ok")

	// Case C: Deactivated Registry
	sicarData.SituacaoImovel = "Cancelado"
	stub.MockTransactionStart("tx_refresh_cancel")
	_, err = registry.RefreshRegistryData(stub, "SICAR", registryPropID)
	require.NoError(t, err)
	stub.MockTransactionEnd("tx_refresh_cancel")

	stub.MockTransactionStart("tx_mint_fail_status")
	_, err = credits.MintEstimatedCreditsForProperty(stub, propIDAttr, intervalStart, intervalEnd)
	require.Error(t, err)
	require.Contains(t, err.Error(), "property is not active")
	stub.MockTransactionEnd("tx_mint_fail_status")

	// 6. Test Burning with SICAR Validation and Multipliers
	// Get the ID of the minted credit from Case B
	mintCreditID := mcQty.GetID() // This returns &[][]string{{ownerID, "1", "0.000000", "0.000000", intervalEnd}}
	mintCreditIDAttr := (*mintCreditID)[0]

	// Setup company private data for multipliers
	company := &companies.Company{
		ID: ownerID,
		Coordinate: &utils.Coordinate{
			Latitude:  -23.550520,
			Longitude: -46.633308,
		},
		DataProps: &data.ValidationProps{
			Methods: []data.ValidationMethod{data.ValidationMethodSattelite},
		},
	}
	stub.MockTransactionStart("tx_company")
	err = company.ToWorldState(stub)
	require.NoError(t, err)
	stub.MockTransactionEnd("tx_company")

	// Case D: Valid Nominal Burning
	stub.MockTransactionStart("tx_burn_ok")
	stub.Creator = mockIds[ownerName] // Ensure caller is owner
	nominalBurnQuantity := int64(40)
	bc, err := credits.BurnNominalQuantity(stub, mintCreditIDAttr, nominalBurnQuantity)
	require.NoError(t, err)
	require.NotNil(t, bc)
	require.False(t, bc.Adjusted)
	stub.MockTransactionEnd("tx_burn_ok")

	// Case E: Apply Multipliers (Requires private data access)
	burnCreditIDAttr := (*bc.GetID())[0]
	stub.MockTransactionStart("tx_apply_mult")
	// In mock stub, we assume access to pvt data if the collection is registered (which it is by default in setup)
	err = credits.ApplyBurnMultipliers(stub, burnCreditIDAttr)
	require.NoError(t, err)
	stub.MockTransactionEnd("tx_apply_mult")

	// Verify adjusted status
	stub.MockTransactionStart("tx_verify_burn")
	adjustedBc := &credits.BurnCredit{}
	err = adjustedBc.FromWorldState(stub, burnCreditIDAttr)
	require.NoError(t, err)
	require.True(t, adjustedBc.Adjusted)
	require.NotEqual(t, int64(0), adjustedBc.BurnMult)
	stub.MockTransactionEnd("tx_verify_burn")

	// Case F: Burn fail if quantity too high
	stub.MockTransactionStart("tx_burn_fail_qty")
	_, err = credits.BurnNominalQuantity(stub, mintCreditIDAttr, 1000)
	require.Error(t, err)
	require.Contains(t, err.Error(), "burn quantity exceeds available quantity")
	stub.MockTransactionEnd("tx_burn_fail_qty")
}

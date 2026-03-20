package carbon_tests

import (
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/data/registry"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	mocks "github.com/johannww/phd-impl/chaincodes/common/state/mocks"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"github.com/stretchr/testify/require"
)

func TestMintCreditWithSicarValidation(t *testing.T) {
	// 1. Setup Mock SICAR Server
	registryPropID := "BR-SP-12345"
	sicarData := registry.SicarData{
		CodigoImovel:                    registryPropID,
		SituacaoImovel:                  "Ativo",
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
	ownerID := "test_owner"
	propertyID := uint64(1)
	registryProvider := "SICAR"

	prop := &properties.Property{
		OwnerID:          ownerID,
		ID:               propertyID,
		RegistryID:       registryPropID,
		RegistryProvider: registryProvider,
	}
	stub.MockTransactionStart("tx_prop")
	err = prop.ToWorldState(stub)
	require.NoError(t, err)
	stub.MockTransactionEnd("tx_prop")

	chunkID := []string{fmt.Sprintf("%d", propertyID), "0.000000", "0.000000"}
	chunk := &properties.PropertyChunk{
		PropertyID: propertyID,
		Coordinates: []utils.Coordinate{
			{Latitude: 0.0, Longitude: 0.0},
		},
	}
	stub.MockTransactionStart("tx_chunk")
	err = chunk.ToWorldState(stub)
	require.NoError(t, err)
	stub.MockTransactionEnd("tx_chunk")

	// 4. Setup Active Policies
	stub.MockTransactionStart("tx_pol")
	stub.Creator = mockIds[identities.PolicySetter]
	pApplier := policies.NewPolicyApplier()
	err = pApplier.SetActivePolicies(stub, []policies.Name{policies.VEGETATION})
	require.NoError(t, err)
	stub.MockTransactionEnd("tx_pol")

	// 4.5 Refresh Registry Data (Pre-fetch before minting)
	stub.MockTransactionStart("tx_refresh")
	_, err = registry.RefreshRegistryData(stub, "SICAR", registryPropID)
	require.NoError(t, err)
	stub.MockTransactionEnd("tx_refresh")

	// 5. Attempt Minting
	stub.Creator = mockIds[identities.CreditMinter]

	// Case A: Valid Minting (40 units <= 50 verified forest)
	stub.MockTransactionStart("tx_mint_ok")
	mc, err := credits.MintCreditForChunk(stub, ownerID, chunkID, 40, time.Now().Format(time.RFC3339))
	require.NoError(t, err)
	require.NotNil(t, mc)
	require.Equal(t, int64(40), mc.Quantity)
	stub.MockTransactionEnd("tx_mint_ok")

	// Case B: Invalid Minting (60 units > 50 verified forest)
	stub.MockTransactionStart("tx_mint_fail_qty")
	_, err = credits.MintCreditForChunk(stub, ownerID, chunkID, 60, time.Now().Format(time.RFC3339))
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds verified forest area")
	stub.MockTransactionEnd("tx_mint_fail_qty")

	// Case C: Deactivated Registry
	sicarData.SituacaoImovel = "Cancelado"
	stub.MockTransactionStart("tx_refresh_cancel")
	_, err = registry.RefreshRegistryData(stub, "SICAR", registryPropID)
	require.NoError(t, err)
	stub.MockTransactionEnd("tx_refresh_cancel")

	stub.MockTransactionStart("tx_mint_fail_status")
	_, err = credits.MintCreditForChunk(stub, ownerID, chunkID, 10, time.Now().Format(time.RFC3339))
	require.Error(t, err)
	require.Contains(t, err.Error(), "property is not active")
	stub.MockTransactionEnd("tx_mint_fail_status")
}

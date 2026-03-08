package sicar_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/johannww/phd-impl/data_api/internal/mock"
	"github.com/johannww/phd-impl/data_api/internal/sicar"
)

func newTestRouter(ids ...string) http.Handler {
	store := make(sicar.Store)
	for _, id := range ids {
		store[id] = mock.SicarImovel(id)
	}
	r := chi.NewRouter()
	r.Route("/sicar/pra/1.0", sicar.Routes(store))
	return r
}

func TestSicarReturns200(t *testing.T) {
	handler := newTestRouter("UF-9999999-ABC123")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sicar/pra/1.0/UF-9999999-ABC123", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestSicarReturns404ForUnknownID(t *testing.T) {
	handler := newTestRouter()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sicar/pra/1.0/UNKNOWN", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestSicarContentType(t *testing.T) {
	handler := newTestRouter("UF-0000001-XYZ")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sicar/pra/1.0/UF-0000001-XYZ", nil)
	handler.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Fatalf("expected application/json, got %q", ct)
	}
}

func TestSicarResponseShape(t *testing.T) {
	const id = "UF-1234567-TEST"
	handler := newTestRouter(id)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sicar/pra/1.0/"+id, nil)
	handler.ServeHTTP(rr, req)

	var resp sicar.SicarResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Result))
	}
	imovel := resp.Result[0]
	if imovel.CodigoImovel != id {
		t.Errorf("codigoimovel: expected %q, got %q", id, imovel.CodigoImovel)
	}
	if imovel.FracaoIdeal != nil {
		t.Errorf("fracaoideal: expected nil, got %v", imovel.FracaoIdeal)
	}
	if !strings.HasPrefix(imovel.AreaTotal, "MULTIPOLYGON") {
		t.Errorf("areatotal: expected MULTIPOLYGON prefix, got %q", imovel.AreaTotal)
	}
}

func TestSicarMockIsDeterministic(t *testing.T) {
	a := mock.SicarImovel("UF-STABLE-SEED")
	b := mock.SicarImovel("UF-STABLE-SEED")
	if a.Result[0].IdentificadorImovel != b.Result[0].IdentificadorImovel {
		t.Error("same input should produce same IdentificadorImovel")
	}
	if a.Result[0].AreaTotalImovel != b.Result[0].AreaTotalImovel {
		t.Error("same input should produce same AreaTotalImovel")
	}
}

func TestSicarDifferentIDsProduceDifferentData(t *testing.T) {
	a := mock.SicarImovel("UF-0000001-AAA")
	b := mock.SicarImovel("UF-0000002-BBB")
	if a.Result[0].AreaTotalImovel == b.Result[0].AreaTotalImovel {
		t.Error("different IDs should produce different AreaTotalImovel")
	}
}

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

// --- helpers ---

func newPraRouter(ids ...string) http.Handler {
	store := make(sicar.Store)
	for _, id := range ids {
		store[id] = sicar.ImovelToPra(mock.Imovel(id))
	}
	r := chi.NewRouter()
	r.Route("/sicar/pra/1.0", sicar.Routes(store))
	return r
}

func newDemonstrativoRouter(ids ...string) http.Handler {
	store := make(sicar.DemonstrativoStore)
	for _, id := range ids {
		store[id] = sicar.ImovelToDemonstrativo(mock.Imovel(id))
	}
	r := chi.NewRouter()
	r.Route("/sicar/demonstrativoDegustacao/1.0", sicar.DemonstrativoRoutes(store))
	return r
}

func newReciboRouter(ids ...string) http.Handler {
	store := make(sicar.ReciboStore)
	for _, id := range ids {
		store[id] = sicar.ImovelToRecibo(mock.Imovel(id))
	}
	r := chi.NewRouter()
	r.Route("/sicar/recibo/1.0", sicar.ReciboRoutes(store))
	return r
}

// --- Pra ---

func TestSicarReturns200(t *testing.T) {
	handler := newPraRouter("UF-9999999-ABC123")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sicar/pra/1.0/UF-9999999-ABC123", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestSicarReturns404ForUnknownID(t *testing.T) {
	handler := newPraRouter()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sicar/pra/1.0/UNKNOWN", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestSicarContentType(t *testing.T) {
	handler := newPraRouter("UF-0000001-XYZ")
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
	handler := newPraRouter(id)
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
	a := sicar.ImovelToPra(mock.Imovel("UF-STABLE-SEED"))
	b := sicar.ImovelToPra(mock.Imovel("UF-STABLE-SEED"))
	if a.Result[0].IdentificadorImovel != b.Result[0].IdentificadorImovel {
		t.Error("same input should produce same IdentificadorImovel")
	}
	if a.Result[0].AreaTotalImovel != b.Result[0].AreaTotalImovel {
		t.Error("same input should produce same AreaTotalImovel")
	}
}

func TestSicarDifferentIDsProduceDifferentData(t *testing.T) {
	a := sicar.ImovelToPra(mock.Imovel("UF-0000001-AAA"))
	b := sicar.ImovelToPra(mock.Imovel("UF-0000002-BBB"))
	if a.Result[0].AreaTotalImovel == b.Result[0].AreaTotalImovel {
		t.Error("different IDs should produce different AreaTotalImovel")
	}
}

// --- Demonstrativo ---

func TestDemonstrativoReturns200(t *testing.T) {
	const id = "UF-1000001-DEMO"
	handler := newDemonstrativoRouter(id)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sicar/demonstrativoDegustacao/1.0/"+id, nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestDemonstrativoReturns404ForUnknownID(t *testing.T) {
	handler := newDemonstrativoRouter()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sicar/demonstrativoDegustacao/1.0/UNKNOWN", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDemonstrativoResponseShape(t *testing.T) {
	const id = "UF-1234567-DEMO"
	handler := newDemonstrativoRouter(id)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sicar/demonstrativoDegustacao/1.0/"+id, nil)
	handler.ServeHTTP(rr, req)

	var resp sicar.DemonstrativoResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Result))
	}
	d := resp.Result[0]
	if d.CodigoImovel != id {
		t.Errorf("codigoImovel: expected %q, got %q", id, d.CodigoImovel)
	}
	if d.AreaTotalImovel <= 0 {
		t.Errorf("areaTotalImovel: expected > 0, got %f", d.AreaTotalImovel)
	}
	if !strings.HasPrefix(d.PoligonoAreaImovel, "MULTIPOLYGON") {
		t.Errorf("poligonoAreaImovel: expected MULTIPOLYGON prefix, got %q", d.PoligonoAreaImovel)
	}
	if d.DataUltimaAtualizacaoCadastro == "" {
		t.Error("dataUltimaAtualizacaoCadastro: expected non-empty")
	}
}

func TestDemonstrativoConsistentWithPra(t *testing.T) {
	const id = "UF-1234567-CROSS"
	im := mock.Imovel(id)
	pra := sicar.ImovelToPra(im)
	demo := sicar.ImovelToDemonstrativo(im)

	if demo.Result[0].CodigoImovel != pra.Result[0].CodigoImovel {
		t.Error("codigoImovel must be consistent between Pra and Demonstrativo")
	}
	if demo.Result[0].Municipio != pra.Result[0].Municipio {
		t.Error("municipio must be consistent between Pra and Demonstrativo")
	}
}

// --- Recibo ---

func TestReciboReturns200(t *testing.T) {
	const id = "UF-1000001-RECIBO"
	handler := newReciboRouter(id)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sicar/recibo/1.0/"+id, nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestReciboReturns404ForUnknownID(t *testing.T) {
	handler := newReciboRouter()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sicar/recibo/1.0/UNKNOWN", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestReciboResponseShape(t *testing.T) {
	const id = "UF-1234567-RECIBO"
	handler := newReciboRouter(id)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sicar/recibo/1.0/"+id, nil)
	handler.ServeHTTP(rr, req)

	var resp sicar.ReciboResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Result))
	}
	rec := resp.Result[0]
	if rec.CodigoImovel != id {
		t.Errorf("codigoImovel: expected %q, got %q", id, rec.CodigoImovel)
	}
	if rec.AreaTotalImovel <= 0 {
		t.Errorf("areaTotalImovel: expected > 0, got %f", rec.AreaTotalImovel)
	}
	if !strings.HasPrefix(rec.GeoImovel, "MULTIPOLYGON") {
		t.Errorf("geoImovel: expected MULTIPOLYGON prefix, got %q", rec.GeoImovel)
	}
	if rec.Proprietarios.TipoPessoa == "" {
		t.Error("proprietarios.tipoPessoa: expected non-empty")
	}
}

func TestReciboConsistentWithDemonstrativo(t *testing.T) {
	const id = "UF-1234567-CROSS2"
	im := mock.Imovel(id)
	demo := sicar.ImovelToDemonstrativo(im)
	recibo := sicar.ImovelToRecibo(im)

	if recibo.Result[0].CodigoImovel != demo.Result[0].CodigoImovel {
		t.Error("codigoImovel must be consistent between Demonstrativo and Recibo")
	}
	if recibo.Result[0].AreaTotalImovel != demo.Result[0].AreaTotalImovel {
		t.Error("areaTotalImovel must be consistent between Demonstrativo and Recibo")
	}
	if recibo.Result[0].Municipio != demo.Result[0].Municipio {
		t.Error("municipio must be consistent between Demonstrativo and Recibo")
	}
	if recibo.Result[0].AreaPreservacaoPermanente != demo.Result[0].AreaPreservacaoPermanente {
		t.Error("areaPreservacaoPermanente must be consistent between Demonstrativo and Recibo")
	}
}


package sicar_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/johannww/phd-impl/data_api/internal/mock"
	"github.com/johannww/phd-impl/data_api/internal/sicar"
)

// --- helpers ---

func loadTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}

func mustUnmarshal[T any](t *testing.T, data []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return v
}

// --- PraResponse ---

func TestPraUnmarshal(t *testing.T) {
	resp := mustUnmarshal[sicar.PraResponse](t, loadTestdata(t, "pra.json"))

	if len(resp.Result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Result))
	}
	im := resp.Result[0]

	check := []struct {
		field string
		got   any
		want  any
	}{
		{"identificadorimovel", im.IdentificadorImovel, 2345678},
		{"codigoimovel", im.CodigoImovel, "UF-9999999-A38E4A53E2494858B5FD8386DF178392"},
		{"codigoversao", im.CodigoVersao, "1.9.2"},
		{"tipoorigem", im.TipoOrigem, "EST"},
		{"protocolo", im.Protocolo, "UF-1234567-A38E4A53E2494858B5FD8386DF178392"},
		{"dataultimoprotocolo", im.DataUltimoProtocolo, "01/01/2024"},
		{"statusimovel", im.StatusImovel, "RE"},
		{"tipoimovel", im.TipoImovel, "IRU"},
		{"cpfcadastrante", im.CPFCadastrante, "01234567890"},
		{"nomecadastrante", im.NomeCadastrante, "JOSE CADASTRANTE"},
		{"nomeimovel", im.NomeImovel, "FAZENDA DO CADASTRANTE"},
		{"identificadormunicipio", im.IdentificadorMunicipio, 5102504},
		{"municipio", im.Municipio, "Cáceres"},
		{"areatotalimovel", im.AreaTotalImovel, "1234.5678"},
		{"numeromodulosfiscais", im.NumeroModulosFiscais, "12.3456"},
		{"datacriacaoregistro", im.DataCriacaoRegistro, "01/01/2024"},
		{"dataatualizacaoregistro", im.DataAtualizacaoRegistro, "01/06/2024"},
		{"declaracaoposterior", im.DeclaracaoPosterior, 1234567},
		{"aderiupra", im.AderiuPRA, "Não"},
	}
	for _, c := range check {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.field, c.got, c.want)
		}
	}

	if im.FracaoIdeal != nil {
		t.Errorf("fracaoideal: expected nil, got %v", im.FracaoIdeal)
	}
	if im.DeclaracaoAnterior != nil {
		t.Errorf("declaracaoanterior: expected nil, got %v", im.DeclaracaoAnterior)
	}
	if im.Situacao != false {
		t.Errorf("situacao: expected false")
	}
	if im.SituacaoMigracao != true {
		t.Errorf("situacaomigracao: expected true")
	}
	if !strings.HasPrefix(im.AreaTotal, "MULTIPOLYGON") {
		t.Errorf("areatotal: expected MULTIPOLYGON prefix, got %q", im.AreaTotal)
	}
}

func TestPraMarshal(t *testing.T) {
	resp := sicar.ImovelToPra(mock.Imovel("UF-1234567-TEST"))
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	var results []map[string]json.RawMessage
	if err := json.Unmarshal(raw["result"], &results); err != nil || len(results) == 0 {
		t.Fatalf("expected result array")
	}
	r := results[0]

	requiredKeys := []string{
		"identificadorimovel", "codigoimovel", "codigoversao", "tipoorigem",
		"protocolo", "dataultimoprotocolo", "statusimovel", "tipoimovel",
		"cpfcadastrante", "nomecadastrante", "nomeimovel", "fracaoideal",
		"identificadormunicipio", "municipio", "areatotalimovel",
		"numeromodulosfiscais", "datacriacaoregistro", "dataatualizacaoregistro",
		"areatotal", "declaracaoanterior", "declaracaoposterior",
		"situacao", "situacaomigracao", "aderiupra",
	}
	for _, k := range requiredKeys {
		if _, ok := r[k]; !ok {
			t.Errorf("marshal: missing key %q", k)
		}
	}
}

// --- DemonstrativoResponse ---

func TestDemonstrativoUnmarshal(t *testing.T) {
	resp := mustUnmarshal[sicar.DemonstrativoResponse](t, loadTestdata(t, "demonstrativo.json"))

	if len(resp.Result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Result))
	}
	im := resp.Result[0]

	// ImovelBase fields
	if im.CodigoImovel != "SC-4210100-EA1F3824FC3D4148B83F921E253C6DCD" {
		t.Errorf("codigoImovel: got %q", im.CodigoImovel)
	}
	if im.AreaTotalImovel != 95.5569 {
		t.Errorf("areaTotalImovel: got %v", im.AreaTotalImovel)
	}
	if im.CodigoMunicipio != 2902005 {
		t.Errorf("codigoMunicipio: got %v", im.CodigoMunicipio)
	}
	if im.Municipio != "Mafra" {
		t.Errorf("municipio: got %q", im.Municipio)
	}
	if im.UnidadeFederativa != "SC" {
		t.Errorf("unidadeFederativa: got %q", im.UnidadeFederativa)
	}
	if im.DataCadastro != "12/06/2015" {
		t.Errorf("dataCadastro: got %q", im.DataCadastro)
	}

	// AreaAmbiental fields
	if im.AreaPreservacaoPermanente != 16.5969602125317 {
		t.Errorf("areaPreservacaoPermanente: got %v", im.AreaPreservacaoPermanente)
	}
	if im.AreaConsolidada != 62.634481425266 {
		t.Errorf("areaConsolidada: got %v", im.AreaConsolidada)
	}
	if im.AreaRemanescenteVegetacaoNativa != 32.8631417737116 {
		t.Errorf("areaRemanescenteVegetacaoNativa: got %v", im.AreaRemanescenteVegetacaoNativa)
	}
	if im.AreaServidaoAdministrativa != 0 {
		t.Errorf("areaServidaoAdministrativa: got %v", im.AreaServidaoAdministrativa)
	}
	if im.AreaUsoRestrito != 0 {
		t.Errorf("areaUsoRestrito: got %v", im.AreaUsoRestrito)
	}

	// DemonstrativoImovel-specific fields
	check := []struct {
		field string
		got   any
		want  any
	}{
		{"situacaoImovel", im.SituacaoImovel, "PE"},
		{"descricaoEtapaCadastro", im.DescricaoEtapaCadastro, "Aguardando análise, não passível de revisão de dados"},
		{"quantidadeModulosFiscais", im.QuantidadeModulosFiscais, "5.9723"},
		{"dataUltimaAtualizacaoCadastro", im.DataUltimaAtualizacaoCadastro, "01/12/2022"},
		{"coordenadaImovelX", im.CoordenadaImovelX, -49.6610956240856},
		{"coordenadaImovelY", im.CoordenadaImovelY, -26.3179762554547},
		{"situacaoReservaLegal", im.SituacaoReservaLegal, "Não analisada"},
		{"areaReservaLegalAverbada", im.AreaReservaLegalAverbada, "0.0000"},
		{"areaReservaLegalAprovadaNaoAverbada", im.AreaReservaLegalAprovadaNaoAverbada, "0.0000"},
		{"areaReservaLegalProposta", im.AreaReservaLegalProposta, "11.9533"},
		{"areaReservaLegalDeclaradaProprietarioPossuidor", im.AreaReservaLegalDeclaradaProprietarioPossuidor, 12.8},
		{"areaPreservacaoPermanenteAreaRuralConsolida", im.AreaPreservacaoPermanenteAreaRuralConsolida, "0.1422"},
		{"areaPreservacaoPermanenteAreaRemanescenteVegetacaoNativa", im.AreaPreservacaoPermanenteAreaRemanescenteVegetacaoNativa, "16.4547"},
		{"areaUsoRestritoDeclividade", im.AreaUsoRestritoDeclividade, 0.0},
		{"areaReservaLegalExcedentePassivo", im.AreaReservaLegalExcedentePassivo, 7.15},
		{"areaReservaLegalRecompor", im.AreaReservaLegalRecompor, 0.0},
		{"areaPreservacaoPermanenteRecompor", im.AreaPreservacaoPermanenteRecompor, 0.07},
		{"areaUsoRestritoRecompor", im.AreaUsoRestritoRecompor, 0.0},
		{"sobreposicoesTerraIndigena", im.SobreposicoesTerraIndigena, 0.0},
		{"sobreposicoesUnidadeConservacao", im.SobreposicoesUnidadeConservacao, 0.0},
		{"sobreposicoesAreasEmbargadas", im.SobreposicoesAreasEmbargadas, 0.4086},
	}
	for _, c := range check {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.field, c.got, c.want)
		}
	}
	if !strings.HasPrefix(im.PoligonoAreaImovel, "MULTIPOLYGON") {
		t.Errorf("poligonoAreaImovel: expected MULTIPOLYGON prefix, got %q", im.PoligonoAreaImovel)
	}
}

func TestDemonstrativoMarshal(t *testing.T) {
	resp := sicar.ImovelToDemonstrativo(mock.Imovel("UF-1234567-TEST"))
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	var results []map[string]json.RawMessage
	if err := json.Unmarshal(raw["result"], &results); err != nil || len(results) == 0 {
		t.Fatalf("expected result array")
	}
	r := results[0]

	requiredKeys := []string{
		"codigoImovel", "areaTotalImovel", "codigoMunicipio", "municipio",
		"unidadeFederativa", "dataCadastro",
		"areaServidaoAdministrativa", "areaConsolidada", "areaRemanescenteVegetacaoNativa",
		"areaPreservacaoPermanente", "areaUsoRestrito",
		"situacaoImovel", "descricaoEtapaCadastro", "quantidadeModulosFiscais",
		"dataUltimaAtualizacaoCadastro", "coordenadaImovelX", "coordenadaImovelY",
		"situacaoReservaLegal", "areaReservaLegalAverbada", "areaReservaLegalAprovadaNaoAverbada",
		"areaReservaLegalProposta", "areaReservaLegalDeclaradaProprietarioPossuidor",
		"areaPreservacaoPermanenteAreaRuralConsolida",
		"areaPreservacaoPermanenteAreaRemanescenteVegetacaoNativa",
		"areaUsoRestritoDeclividade", "areaReservaLegalExcedentePassivo",
		"areaReservaLegalRecompor", "areaPreservacaoPermanenteRecompor",
		"areaUsoRestritoRecompor", "sobreposicoesTerraIndigena",
		"sobreposicoesUnidadeConservacao", "sobreposicoesAreasEmbargadas",
		"poligonoAreaImovel",
	}
	for _, k := range requiredKeys {
		if _, ok := r[k]; !ok {
			t.Errorf("marshal: missing key %q", k)
		}
	}
}

// --- ReciboResponse ---

func TestReciboUnmarshal(t *testing.T) {
	resp := mustUnmarshal[sicar.ReciboResponse](t, loadTestdata(t, "recibo.json"))

	if len(resp.Result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Result))
	}
	im := resp.Result[0]

	// ImovelBase fields
	if im.CodigoImovel != "TO-1714203-1C6DC93596CE4193A5E8519DDED552A4" {
		t.Errorf("codigoImovel: got %q", im.CodigoImovel)
	}
	if im.AreaTotalImovel != "6326.184" {
		t.Errorf("areaTotalImovel: got %q", im.AreaTotalImovel)
	}
	if im.CodigoMunicipio != 1714203 {
		t.Errorf("codigoMunicipio: got %v", im.CodigoMunicipio)
	}
	if im.Municipio != "Natividade" {
		t.Errorf("municipio: got %q", im.Municipio)
	}
	if im.UnidadeFederativa != "TO" {
		t.Errorf("unidadeFederativa: got %q", im.UnidadeFederativa)
	}

	// AreaAmbiental fields
	if im.AreaPreservacaoPermanente != 909.1799 {
		t.Errorf("areaPreservacaoPermanente: got %v", im.AreaPreservacaoPermanente)
	}
	if im.AreaConsolidada != 1675.4077 {
		t.Errorf("areaConsolidada: got %v", im.AreaConsolidada)
	}
	if im.AreaRemanescenteVegetacaoNativa != 4067.6421 {
		t.Errorf("areaRemanescenteVegetacaoNativa: got %v", im.AreaRemanescenteVegetacaoNativa)
	}

	// ReciboImovel-specific fields
	check := []struct {
		field string
		got   any
		want  any
	}{
		{"identificadorImovel", im.IdentificadorImovel, 12579165},
		{"situacaoImovel", im.SituacaoImovel, "RE"},
		{"tipoImovel", im.TipoImovel, "AST"},
		{"nomeImovel", im.NomeImovel, "JACUBINHA"},
		{"coordenadaImovelX", im.CoordenadaImovelX, -51.8516375657968},
		{"moduloFiscal", im.ModuloFiscal, "1.1297"},
		{"protocolo", im.Protocolo, "TO-1714203-037F2BE93ECB15CA6F2D256AAD97F2FB"},
		{"areaLiquidaImovel", im.AreaLiquidaImovel, 0.0},
		{"areaReservaLegal", im.AreaReservaLegal, 2182.3045},
		{"matricula", im.Matricula, "M-2.096"},
		{"livroMatricula", im.LivroMatricula, "2-J"},
		{"folhaMatricula", im.FolhaMatricula, "79"},
		{"municipioCartorio", im.MunicipioCartorio, "Natividade"},
		{"ufCartorio", im.UFCartorio, "TO"},
	}
	for _, c := range check {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.field, c.got, c.want)
		}
	}

	if im.Proprietarios.TipoPessoa != "PF" {
		t.Errorf("proprietarios.tipoPessoa: got %q", im.Proprietarios.TipoPessoa)
	}
	if im.Proprietarios.CPFCnpj != "03662243601" {
		t.Errorf("proprietarios.cpfCnpj: got %q", im.Proprietarios.CPFCnpj)
	}
	if im.Proprietarios.NomeProprietario != "Pessoa teste SFB" {
		t.Errorf("proprietarios.nomeProprietario: got %q", im.Proprietarios.NomeProprietario)
	}
	if im.Proprietarios.NomeFantasia != "Empresa teste SFB" {
		t.Errorf("proprietarios.nomeFantasia: got %q", im.Proprietarios.NomeFantasia)
	}
	if !strings.HasPrefix(im.GeoImovel, "MULTIPOLYGON") {
		t.Errorf("geoImovel: expected MULTIPOLYGON prefix, got %q", im.GeoImovel)
	}
	if im.InformacoesAdicionais == "" {
		t.Error("informacoesAdicionais: expected non-empty")
	}
}

func TestReciboMarshal(t *testing.T) {
	resp := sicar.ImovelToRecibo(mock.Imovel("UF-1234567-TEST"))
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	var results []map[string]json.RawMessage
	if err := json.Unmarshal(raw["result"], &results); err != nil || len(results) == 0 {
		t.Fatalf("expected result array")
	}
	r := results[0]

	requiredKeys := []string{
		"codigoImovel", "areaTotalImovel", "codigoMunicipio", "municipio",
		"unidadeFederativa", "dataCadastro",
		"areaServidaoAdministrativa", "areaConsolidada", "areaRemanescenteVegetacaoNativa",
		"areaPreservacaoPermanente", "areaUsoRestrito",
		"identificadorImovel", "situacaoImovel", "tipoImovel", "nomeImovel",
		"coordenadaImovelX", "coordenadaImovelY", "moduloFiscal", "protocolo",
		"informacoesAdicionais", "geoImovel", "proprietarios",
		"areaLiquidaImovel", "areaReservaLegal",
		"matricula", "dataMatricula", "livroMatricula", "folhaMatricula",
		"municipioCartorio", "ufCartorio",
	}
	for _, k := range requiredKeys {
		if _, ok := r[k]; !ok {
			t.Errorf("marshal: missing key %q", k)
		}
	}

	// verify proprietarios sub-keys
	var props map[string]json.RawMessage
	if err := json.Unmarshal(r["proprietarios"], &props); err != nil {
		t.Fatalf("unmarshal proprietarios: %v", err)
	}
	for _, k := range []string{"tipoPessoa", "cpfCnpj", "nomeProprietario", "nomeFantasia"} {
		if _, ok := props[k]; !ok {
			t.Errorf("marshal: missing proprietarios key %q", k)
		}
	}
}

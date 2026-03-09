package mock

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"

	"github.com/johannww/phd-impl/data_api/internal/imovel"
)

func seededRand(seed string) *rand.Rand {
	h := fnv.New64a()
	h.Write([]byte(seed))
	return rand.New(rand.NewSource(int64(h.Sum64())))
}

func date(r *rand.Rand) string {
	y := 2010 + r.Intn(14)
	m := 1 + r.Intn(12)
	d := 1 + r.Intn(28)
	return fmt.Sprintf("%02d/%02d/%d", d, m, y)
}

func multipolygon(r *rand.Rand) string {
	lat1 := -(r.Float64() * 30)
	lon1 := -(r.Float64() * 60)
	lat2 := lat1 - r.Float64()*2
	lon2 := lon1 - r.Float64()*2
	return fmt.Sprintf(
		"MULTIPOLYGON(((%.8f %.8f,%.8f %.8f,%.8f %.8f,%.8f %.8f,%.8f %.8f)))",
		lon1, lat1, lon1, lat2, lon2, lat2, lon2, lat1, lon1, lat1,
	)
}

var municipios = []string{"Cáceres", "Mafra", "Natividade", "Campinas", "Belém", "Cuiabá", "Porto Velho"}
var ufs = []string{"AC", "AL", "AM", "BA", "CE", "GO", "MG", "MT", "PA", "PR", "RJ", "RS", "SC", "SP"}

// Imovel generates a deterministic canonical Imovel from codigoImovel.
func Imovel(codigoImovel string) imovel.Imovel {
	r := seededRand(codigoImovel)

	total := 100.0 + r.Float64()*9900.0
	app := total * (0.05 + r.Float64()*0.20)
	consolidated := total * (0.30 + r.Float64()*0.30)
	remanescente := total - app - consolidated
	if remanescente < 0 {
		remanescente = 0
	}
	rl := total * (0.20 + r.Float64()*0.10)
	updDate := time.Date(2020+r.Intn(5), time.Month(1+r.Intn(12)), 1+r.Intn(28), 0, 0, 0, 0, time.UTC)
	uf := ufs[r.Intn(len(ufs))]
	municipio := municipios[r.Intn(len(municipios))]

	return imovel.Imovel{
		IdentificadorImovel:    r.Intn(9000000) + 1000000,
		CodigoImovel:           codigoImovel,
		CodigoVersao:           "1.9.2",
		Protocolo:              fmt.Sprintf("PROTO-%07d", r.Intn(9000000)+1000000),
		DataProtocolo:          date(r),
		StatusImovel:           "RE",
		TipoImovel:             "IRU",
		TipoOrigem:             "EST",
		DescricaoEtapaCadastro: "Aguardando análise, não passível de revisão de dados",

		CPFCadastrante:  fmt.Sprintf("%011d", r.Int63n(99999999999)),
		NomeCadastrante: "JOSE CADASTRANTE",
		NomeImovel:      "FAZENDA DO CADASTRANTE",
		FracaoIdeal:     nil,

		CodigoMunicipio:   r.Intn(9000000) + 1000000,
		Municipio:         municipio,
		UnidadeFederativa: uf,
		CoordenadaX:       -(r.Float64() * 60),
		CoordenadaY:       -(r.Float64() * 30),
		Poligono:          multipolygon(r),

		DataCadastro:          date(r),
		DataUltimaAtualizacao: updDate.Format("02/01/2006"),

		AreaTotalImovel:      total,
		NumeroModulosFiscais: total / 250.0,

		AreaPreservacaoPermanente:       app,
		AreaConsolidada:                 consolidated,
		AreaRemanescenteVegetacaoNativa: remanescente,
		AreaServidaoAdministrativa:      total * r.Float64() * 0.02,
		AreaUsoRestrito:                 total * r.Float64() * 0.05,
		AreaReservaLegal:                rl,
		AreaLiquidaImovel:               total * r.Float64() * 0.9,

		SituacaoReservaLegal:                           "Não analisada",
		AreaReservaLegalAverbada:                       0,
		AreaReservaLegalAprovadaNaoAverbada:            0,
		AreaReservaLegalProposta:                       rl * 0.5,
		AreaReservaLegalDeclaradaProprietarioPossuidor: rl,
		AreaReservaLegalExcedentePassivo:               r.Float64() * 10,
		AreaReservaLegalRecompor:                       0,

		AreaPPAreaRuralConsolida:          app * 0.01,
		AreaPPRemanescenteVegetacaoNativa: app * 0.99,
		AreaPreservacaoPermanenteRecompor: r.Float64() * 0.5,

		AreaUsoRestritoDeclividade:      0,
		AreaUsoRestritoRecompor:         0,
		SobreposicoesTerraIndigena:      0,
		SobreposicoesUnidadeConservacao: 0,
		SobreposicoesAreasEmbargadas:    r.Float64() * 2,

		Matricula:         fmt.Sprintf("M-%d", r.Intn(9000)+1000),
		DataMatricula:     "",
		LivroMatricula:    fmt.Sprintf("%d-J", r.Intn(9)+1),
		FolhaMatricula:    fmt.Sprintf("%d", r.Intn(200)+1),
		MunicipioCartorio: municipio,
		UFCartorio:        uf,

		TipoPessoa:       "PF",
		NomeProprietario: "Pessoa Teste",
		NomeFantasia:     "Empresa Teste",

		Situacao:            false,
		SituacaoMigracao:    true,
		AderiuPRA:           "Não",
		DeclaracaoPosterior: r.Intn(9000000) + 1000000,
	}
}

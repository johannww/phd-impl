package sicar

// ImovelBase holds fields shared between Demonstrativo and Recibo (camelCase APIs).
type ImovelBase struct {
	CodigoImovel      string `json:"codigoImovel"`
	CodigoMunicipio   int    `json:"codigoMunicipio"`
	Municipio         string `json:"municipio"`
	UnidadeFederativa string `json:"unidadeFederativa"`
	DataCadastro      string `json:"dataCadastro"`
}

// AreaAmbiental holds preservation and land-use breakdown fields
// shared between Demonstrativo and Recibo.
type AreaAmbiental struct {
	AreaServidaoAdministrativa      float64 `json:"areaServidaoAdministrativa"`
	AreaConsolidada                 float64 `json:"areaConsolidada"`
	AreaRemanescenteVegetacaoNativa float64 `json:"areaRemanescenteVegetacaoNativa"`
	AreaPreservacaoPermanente       float64 `json:"areaPreservacaoPermanente"`
	AreaUsoRestrito                 float64 `json:"areaUsoRestrito"`
}

// --- Pra ---

// PraImovel mirrors the /sicar/pra/1.0 response (all-lowercase keys per spec).
type PraImovel struct {
	IdentificadorImovel     int     `json:"identificadorimovel"`
	CodigoImovel            string  `json:"codigoimovel"`
	CodigoVersao            string  `json:"codigoversao"`
	TipoOrigem              string  `json:"tipoorigem"`
	Protocolo               string  `json:"protocolo"`
	DataUltimoProtocolo     string  `json:"dataultimoprotocolo"`
	StatusImovel            string  `json:"statusimovel"`
	TipoImovel              string  `json:"tipoimovel"`
	CPFCadastrante          string  `json:"cpfcadastrante"`
	NomeCadastrante         string  `json:"nomecadastrante"`
	NomeImovel              string  `json:"nomeimovel"`
	FracaoIdeal             *string `json:"fracaoideal"`
	IdentificadorMunicipio  int     `json:"identificadormunicipio"`
	Municipio               string  `json:"municipio"`
	AreaTotalImovel         string  `json:"areatotalimovel"`
	NumeroModulosFiscais    string  `json:"numeromodulosfiscais"`
	DataCriacaoRegistro     string  `json:"datacriacaoregistro"`
	DataAtualizacaoRegistro string  `json:"dataatualizacaoregistro"`
	AreaTotal               string  `json:"areatotal"`
	DeclaracaoAnterior      *int    `json:"declaracaoanterior"`
	DeclaracaoPosterior     int     `json:"declaracaoposterior"`
	Situacao                bool    `json:"situacao"`
	SituacaoMigracao        bool    `json:"situacaomigracao"`
	AderiuPRA               string  `json:"aderiupra"`
}

type PraResponse struct {
	Result []PraImovel `json:"result"`
}

// --- Demonstrativo ---

type DemonstrativoImovel struct {
	ImovelBase
	AreaAmbiental
	AreaTotalImovel                                          float64 `json:"areaTotalImovel"`
	SituacaoImovel                                           string  `json:"situacaoImovel"`
	DescricaoEtapaCadastro                                   string  `json:"descricaoEtapaCadastro"`
	QuantidadeModulosFiscais                                 string  `json:"quantidadeModulosFiscais"`
	DataUltimaAtualizacaoCadastro                            string  `json:"dataUltimaAtualizacaoCadastro"`
	CoordenadaImovelX                                        float64 `json:"coordenadaImovelX"`
	CoordenadaImovelY                                        float64 `json:"coordenadaImovelY"`
	SituacaoReservaLegal                                     string  `json:"situacaoReservaLegal"`
	AreaReservaLegalAverbada                                 string  `json:"areaReservaLegalAverbada"`
	AreaReservaLegalAprovadaNaoAverbada                      string  `json:"areaReservaLegalAprovadaNaoAverbada"`
	AreaReservaLegalProposta                                 string  `json:"areaReservaLegalProposta"`
	AreaReservaLegalDeclaradaProprietarioPossuidor           float64 `json:"areaReservaLegalDeclaradaProprietarioPossuidor"`
	AreaPreservacaoPermanenteAreaRuralConsolida              string  `json:"areaPreservacaoPermanenteAreaRuralConsolida"`
	AreaPreservacaoPermanenteAreaRemanescenteVegetacaoNativa string  `json:"areaPreservacaoPermanenteAreaRemanescenteVegetacaoNativa"`
	AreaUsoRestritoDeclividade                               float64 `json:"areaUsoRestritoDeclividade"`
	AreaReservaLegalExcedentePassivo                         float64 `json:"areaReservaLegalExcedentePassivo"`
	AreaReservaLegalRecompor                                 float64 `json:"areaReservaLegalRecompor"`
	AreaPreservacaoPermanenteRecompor                        float64 `json:"areaPreservacaoPermanenteRecompor"`
	AreaUsoRestritoRecompor                                  float64 `json:"areaUsoRestritoRecompor"`
	SobreposicoesTerraIndigena                               float64 `json:"sobreposicoesTerraIndigena"`
	SobreposicoesUnidadeConservacao                          float64 `json:"sobreposicoesUnidadeConservacao"`
	SobreposicoesAreasEmbargadas                             float64 `json:"sobreposicoesAreasEmbargadas"`
	PoligonoAreaImovel                                       string  `json:"poligonoAreaImovel"`
}

type DemonstrativoResponse struct {
	Result []DemonstrativoImovel `json:"result"`
}

// --- Recibo ---

type ReciboProprietario struct {
	TipoPessoa       string `json:"tipoPessoa"`
	CPFCnpj          string `json:"cpfCnpj"`
	NomeProprietario string `json:"nomeProprietario"`
	NomeFantasia     string `json:"nomeFantasia"`
}

type ReciboImovel struct {
	ImovelBase
	AreaAmbiental
	AreaTotalImovel       string             `json:"areaTotalImovel"`
	IdentificadorImovel   int                `json:"identificadorImovel"`
	SituacaoImovel        string             `json:"situacaoImovel"`
	TipoImovel            string             `json:"tipoImovel"`
	NomeImovel            string             `json:"nomeImovel"`
	CoordenadaImovelX     float64            `json:"coordenadaImovelX"`
	CoordenadaImovelY     float64            `json:"coordenadaImovelY"`
	ModuloFiscal          string             `json:"moduloFiscal"`
	Protocolo             string             `json:"protocolo"`
	InformacoesAdicionais string             `json:"informacoesAdicionais"`
	GeoImovel             string             `json:"geoImovel"`
	Proprietarios         ReciboProprietario `json:"proprietarios"`
	AreaLiquidaImovel     float64            `json:"areaLiquidaImovel"`
	AreaReservaLegal      float64            `json:"areaReservaLegal"`
	Matricula             string             `json:"matricula"`
	DataMatricula         string             `json:"dataMatricula"`
	LivroMatricula        string             `json:"livroMatricula"`
	FolhaMatricula        string             `json:"folhaMatricula"`
	MunicipioCartorio     string             `json:"municipioCartorio"`
	UFCartorio            string             `json:"ufCartorio"`
}

type ReciboResponse struct {
	Result []ReciboImovel `json:"result"`
}

// --- Legacy alias kept for backward compatibility with existing Store references ---

// SicarImovel is an alias for PraImovel.
type SicarImovel = PraImovel

// SicarResponse is an alias for PraResponse.
type SicarResponse = PraResponse

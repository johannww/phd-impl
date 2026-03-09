package imovel

// Imovel is the canonical data model for a rural property (imóvel rural).
// All API response projections are derived from this struct.
type Imovel struct {
	// Identity
	IdentificadorImovel int    `json:"identificadorImovel"`
	CodigoImovel        string `json:"codigoImovel"`
	CodigoVersao        string `json:"codigoVersao"`
	Protocolo           string `json:"protocolo"`
	DataProtocolo       string `json:"dataProtocolo"`

	// Status
	StatusImovel     string `json:"statusImovel"`
	TipoImovel       string `json:"tipoImovel"`
	TipoOrigem       string `json:"tipoOrigem"`
	DescricaoEtapaCadastro string `json:"descricaoEtapaCadastro"`

	// Owner
	CPFCadastrante   string  `json:"cpfCadastrante"`
	NomeCadastrante  string  `json:"nomeCadastrante"`
	NomeImovel       string  `json:"nomeImovel"`
	FracaoIdeal      *string `json:"fracaoIdeal"`

	// Location
	CodigoMunicipio   int     `json:"codigoMunicipio"`
	Municipio         string  `json:"municipio"`
	UnidadeFederativa string  `json:"unidadeFederativa"`
	CoordenadaX       float64 `json:"coordenadaX"`
	CoordenadaY       float64 `json:"coordenadaY"`
	Poligono          string  `json:"poligono"`

	// Dates
	DataCadastro          string `json:"dataCadastro"`
	DataUltimaAtualizacao string `json:"dataUltimaAtualizacao"`

	// Area totals
	AreaTotalImovel      float64 `json:"areaTotalImovel"`
	NumeroModulosFiscais float64 `json:"numeroModulosFiscais"`

	// Environmental areas
	AreaPreservacaoPermanente       float64 `json:"areaPreservacaoPermanente"`
	AreaConsolidada                 float64 `json:"areaConsolidada"`
	AreaRemanescenteVegetacaoNativa float64 `json:"areaRemanescenteVegetacaoNativa"`
	AreaServidaoAdministrativa      float64 `json:"areaServidaoAdministrativa"`
	AreaUsoRestrito                 float64 `json:"areaUsoRestrito"`
	AreaReservaLegal                float64 `json:"areaReservaLegal"`
	AreaLiquidaImovel               float64 `json:"areaLiquidaImovel"`

	// Reserva Legal breakdown
	SituacaoReservaLegal                           string  `json:"situacaoReservaLegal"`
	AreaReservaLegalAverbada                        float64 `json:"areaReservaLegalAverbada"`
	AreaReservaLegalAprovadaNaoAverbada             float64 `json:"areaReservaLegalAprovadaNaoAverbada"`
	AreaReservaLegalProposta                        float64 `json:"areaReservaLegalProposta"`
	AreaReservaLegalDeclaradaProprietarioPossuidor  float64 `json:"areaReservaLegalDeclaradaProprietarioPossuidor"`
	AreaReservaLegalExcedentePassivo                float64 `json:"areaReservaLegalExcedentePassivo"`
	AreaReservaLegalRecompor                        float64 `json:"areaReservaLegalRecompor"`

	// APP breakdown
	AreaPPAreaRuralConsolida             float64 `json:"areaPPAreaRuralConsolida"`
	AreaPPRemanescenteVegetacaoNativa    float64 `json:"areaPPRemanescenteVegetacaoNativa"`
	AreaPreservacaoPermanenteRecompor    float64 `json:"areaPreservacaoPermanenteRecompor"`

	// Other
	AreaUsoRestritoDeclividade  float64 `json:"areaUsoRestritoDeclividade"`
	AreaUsoRestritoRecompor     float64 `json:"areaUsoRestritoRecompor"`
	SobreposicoesTerraIndigena  float64 `json:"sobreposicoesTerraIndigena"`
	SobreposicoesUnidadeConservacao float64 `json:"sobreposicoesUnidadeConservacao"`
	SobreposicoesAreasEmbargadas    float64 `json:"sobreposicoesAreasEmbargadas"`

	// Registry
	Matricula         string `json:"matricula"`
	DataMatricula     string `json:"dataMatricula"`
	LivroMatricula    string `json:"livroMatricula"`
	FolhaMatricula    string `json:"folhaMatricula"`
	MunicipioCartorio string `json:"municipioCartorio"`
	UFCartorio        string `json:"ufCartorio"`

	// Proprietario
	TipoPessoa       string `json:"tipoPessoa"`
	NomeProprietario string `json:"nomeProprietario"`
	NomeFantasia     string `json:"nomeFantasia"`

	// Flags
	Situacao         bool `json:"situacao"`
	SituacaoMigracao bool `json:"situacaomigracao"`
	AderiuPRA        string `json:"aderiupra"`
	DeclaracaoPosterior int `json:"declaracaoPosterior"`
}

type Store map[string]Imovel

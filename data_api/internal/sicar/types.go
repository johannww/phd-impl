package sicar

// SicarImovel mirrors the SICAR API response schema.
type SicarImovel struct {
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

type SicarResponse struct {
	Result []SicarImovel `json:"result"`
}

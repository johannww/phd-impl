package sicar

import (
	"fmt"

	"github.com/johannww/phd-impl/data_api/internal/imovel"
)

// ImovelToPra projects a canonical Imovel into a Pra API response.
func ImovelToPra(im imovel.Imovel) PraResponse {
	return PraResponse{
		Result: []PraImovel{
			{
				IdentificadorImovel:     im.IdentificadorImovel,
				CodigoImovel:            im.CodigoImovel,
				CodigoVersao:            im.CodigoVersao,
				TipoOrigem:              im.TipoOrigem,
				Protocolo:               im.Protocolo,
				DataUltimoProtocolo:     im.DataProtocolo,
				StatusImovel:            im.StatusImovel,
				TipoImovel:              im.TipoImovel,
				CPFCadastrante:          im.CPFCadastrante,
				NomeCadastrante:         im.NomeCadastrante,
				NomeImovel:              im.NomeImovel,
				FracaoIdeal:             im.FracaoIdeal,
				IdentificadorMunicipio:  im.CodigoMunicipio,
				Municipio:               im.Municipio,
				AreaTotalImovel:         fmt.Sprintf("%.4f", im.AreaTotalImovel),
				NumeroModulosFiscais:    fmt.Sprintf("%.4f", im.NumeroModulosFiscais),
				DataCriacaoRegistro:     im.DataCadastro,
				DataAtualizacaoRegistro: im.DataUltimaAtualizacao,
				AreaTotal:               im.Poligono,
				DeclaracaoAnterior:      nil,
				DeclaracaoPosterior:     im.DeclaracaoPosterior,
				Situacao:                im.Situacao,
				SituacaoMigracao:        im.SituacaoMigracao,
				AderiuPRA:               im.AderiuPRA,
			},
		},
	}
}

// ImovelToDemonstrativo projects a canonical Imovel into a Demonstrativo API response.
func ImovelToDemonstrativo(im imovel.Imovel) DemonstrativoResponse {
	return DemonstrativoResponse{
		Result: []DemonstrativoImovel{
			{
				ImovelBase: ImovelBase{
					CodigoImovel:      im.CodigoImovel,
					AreaTotalImovel:   im.AreaTotalImovel,
					CodigoMunicipio:   im.CodigoMunicipio,
					Municipio:         im.Municipio,
					UnidadeFederativa: im.UnidadeFederativa,
					DataCadastro:      im.DataCadastro,
				},
				AreaAmbiental: AreaAmbiental{
					AreaServidaoAdministrativa:      im.AreaServidaoAdministrativa,
					AreaConsolidada:                 im.AreaConsolidada,
					AreaRemanescenteVegetacaoNativa: im.AreaRemanescenteVegetacaoNativa,
					AreaPreservacaoPermanente:       im.AreaPreservacaoPermanente,
					AreaUsoRestrito:                 im.AreaUsoRestrito,
				},
				SituacaoImovel:                im.StatusImovel,
				DescricaoEtapaCadastro:        im.DescricaoEtapaCadastro,
				QuantidadeModulosFiscais:       fmt.Sprintf("%.4f", im.NumeroModulosFiscais),
				DataUltimaAtualizacaoCadastro:  im.DataUltimaAtualizacao,
				CoordenadaImovelX:              im.CoordenadaX,
				CoordenadaImovelY:              im.CoordenadaY,
				SituacaoReservaLegal:           im.SituacaoReservaLegal,
				AreaReservaLegalAverbada:        fmt.Sprintf("%.4f", im.AreaReservaLegalAverbada),
				AreaReservaLegalAprovadaNaoAverbada: fmt.Sprintf("%.4f", im.AreaReservaLegalAprovadaNaoAverbada),
				AreaReservaLegalProposta:        fmt.Sprintf("%.4f", im.AreaReservaLegalProposta),
				AreaReservaLegalDeclaradaProprietarioPossuidor:           im.AreaReservaLegalDeclaradaProprietarioPossuidor,
				AreaPreservacaoPermanenteAreaRuralConsolida:              fmt.Sprintf("%.4f", im.AreaPPAreaRuralConsolida),
				AreaPreservacaoPermanenteAreaRemanescenteVegetacaoNativa: fmt.Sprintf("%.4f", im.AreaPPRemanescenteVegetacaoNativa),
				AreaUsoRestritoDeclividade:      im.AreaUsoRestritoDeclividade,
				AreaReservaLegalExcedentePassivo: im.AreaReservaLegalExcedentePassivo,
				AreaReservaLegalRecompor:        im.AreaReservaLegalRecompor,
				AreaPreservacaoPermanenteRecompor: im.AreaPreservacaoPermanenteRecompor,
				AreaUsoRestritoRecompor:         im.AreaUsoRestritoRecompor,
				SobreposicoesTerraIndigena:      im.SobreposicoesTerraIndigena,
				SobreposicoesUnidadeConservacao: im.SobreposicoesUnidadeConservacao,
				SobreposicoesAreasEmbargadas:    im.SobreposicoesAreasEmbargadas,
				PoligonoAreaImovel:              im.Poligono,
			},
		},
	}
}

// ImovelToRecibo projects a canonical Imovel into a Recibo API response.
func ImovelToRecibo(im imovel.Imovel) ReciboResponse {
	return ReciboResponse{
		Result: []ReciboImovel{
			{
				ImovelBase: ImovelBase{
					CodigoImovel:      im.CodigoImovel,
					AreaTotalImovel:   im.AreaTotalImovel,
					CodigoMunicipio:   im.CodigoMunicipio,
					Municipio:         im.Municipio,
					UnidadeFederativa: im.UnidadeFederativa,
					DataCadastro:      im.DataCadastro,
				},
				AreaAmbiental: AreaAmbiental{
					AreaServidaoAdministrativa:      im.AreaServidaoAdministrativa,
					AreaConsolidada:                 im.AreaConsolidada,
					AreaRemanescenteVegetacaoNativa: im.AreaRemanescenteVegetacaoNativa,
					AreaPreservacaoPermanente:       im.AreaPreservacaoPermanente,
					AreaUsoRestrito:                 im.AreaUsoRestrito,
				},
				IdentificadorImovel:   im.IdentificadorImovel,
				SituacaoImovel:        im.StatusImovel,
				TipoImovel:            im.TipoImovel,
				NomeImovel:            im.NomeImovel,
				CoordenadaImovelX:     im.CoordenadaX,
				CoordenadaImovelY:     im.CoordenadaY,
				ModuloFiscal:          fmt.Sprintf("%.4f", im.NumeroModulosFiscais),
				Protocolo:             im.Protocolo,
				InformacoesAdicionais: "",
				GeoImovel:             im.Poligono,
				Proprietarios: ReciboProprietario{
					TipoPessoa:       im.TipoPessoa,
					CPFCnpj:          im.CPFCadastrante,
					NomeProprietario: im.NomeProprietario,
					NomeFantasia:     im.NomeFantasia,
				},
				AreaLiquidaImovel: im.AreaLiquidaImovel,
				AreaReservaLegal:  im.AreaReservaLegal,
				Matricula:         im.Matricula,
				DataMatricula:     im.DataMatricula,
				LivroMatricula:    im.LivroMatricula,
				FolhaMatricula:    im.FolhaMatricula,
				MunicipioCartorio: im.MunicipioCartorio,
				UFCartorio:        im.UFCartorio,
			},
		},
	}
}

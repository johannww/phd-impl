package mock

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"

	"github.com/johannww/phd-impl/data_api/internal/sicar"
)

func seededRand(seed string) *rand.Rand {
	h := fnv.New64a()
	h.Write([]byte(seed))
	return rand.New(rand.NewSource(int64(h.Sum64())))
}

func SicarImovel(codigoImovel string) sicar.SicarResponse {
	r := seededRand(codigoImovel)

	date := func() string {
		y := 2010 + r.Intn(14)
		m := 1 + r.Intn(12)
		d := 1 + r.Intn(28)
		return fmt.Sprintf("%02d/%02d/%d", d, m, y)
	}

	updDate := time.Date(2020+r.Intn(5), time.Month(1+r.Intn(12)), 1+r.Intn(28), 0, 0, 0, 0, time.UTC)

	area := 100.0 + r.Float64()*9900.0
	modules := area / 250.0

	lat1 := -(r.Float64() * 30)
	lon1 := -(r.Float64() * 60)
	lat2 := lat1 - r.Float64()*2
	lon2 := lon1 - r.Float64()*2

	return sicar.SicarResponse{
		Result: []sicar.SicarImovel{
			{
				IdentificadorImovel:     r.Intn(9000000) + 1000000,
				CodigoImovel:            codigoImovel,
				CodigoVersao:            "1.9.2",
				TipoOrigem:              "EST",
				Protocolo:               fmt.Sprintf("PROTO-%07d", r.Intn(9000000)+1000000),
				DataUltimoProtocolo:     date(),
				StatusImovel:            "RE",
				TipoImovel:              "IRU",
				CPFCadastrante:          fmt.Sprintf("%011d", r.Int63n(99999999999)),
				NomeCadastrante:         "JOSE CADASTRANTE",
				NomeImovel:              "FAZENDA DO CADASTRANTE",
				FracaoIdeal:             nil,
				IdentificadorMunicipio:  r.Intn(9000000) + 1000000,
				Municipio:               "Cáceres",
				AreaTotalImovel:         fmt.Sprintf("%.4f", area),
				NumeroModulosFiscais:    fmt.Sprintf("%.4f", modules),
				DataCriacaoRegistro:     date(),
				DataAtualizacaoRegistro: updDate.Format("02/01/2006"),
				AreaTotal: fmt.Sprintf(
					"MULTIPOLYGON(((%.8f %.8f,%.8f %.8f,%.8f %.8f,%.8f %.8f,%.8f %.8f)))",
					lon1, lat1, lon1, lat2, lon2, lat2, lon2, lat1, lon1, lat1,
				),
				DeclaracaoAnterior:  nil,
				DeclaracaoPosterior: r.Intn(9000000) + 1000000,
				Situacao:            false,
				SituacaoMigracao:    true,
				AderiuPRA:           "Não",
			},
		},
	}
}

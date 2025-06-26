package vegetation

type ForestType int

const (
	AmazonRainforest ForestType = iota
	AtlanticForest
	Caatinga
	Cerrado
	Pantanal
	Pampa
	Mangroves
	AraucariaForest
	Restinga
)

const (
	AmazonRainforestStr = "Amazon Rainforest"
	AtlanticForestStr   = "Atlantic Forest"
	CaatingaStr         = "Caatinga"
	CerradoStr          = "Cerrado"
	PantanalStr         = "Pantanal"
	PampaStr            = "Pampa"
	MangrovesStr        = "Mangroves"
	AraucariaForestStr  = "Araucaria Forest"
	RestingaStr         = "Restinga"
)

var ForestTypeMap = map[ForestType]string{
	AmazonRainforest: AmazonRainforestStr,
	AtlanticForest:   AtlanticForestStr,
	Caatinga:         CaatingaStr,
	Cerrado:          CerradoStr,
	Pantanal:         PantanalStr,
	Pampa:            PampaStr,
	Mangroves:        MangrovesStr,
	AraucariaForest:  AraucariaForestStr,
	Restinga:         RestingaStr,
}

var ForestTypeReverseMap = map[string]ForestType{
	AmazonRainforestStr: AmazonRainforest,
	AtlanticForestStr:   AtlanticForest,
	CaatingaStr:         Caatinga,
	CerradoStr:          Cerrado,
	PantanalStr:         Pantanal,
	PampaStr:            Pampa,
	MangrovesStr:        Mangroves,
	AraucariaForestStr:  AraucariaForest,
	RestingaStr:         Restinga,
}

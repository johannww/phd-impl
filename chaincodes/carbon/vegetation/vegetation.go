package vegetation

type VegetationProps struct {
	ForestType    ForestType `json:"forestType,omitempty"`
	ForestDensity float64    `json:"forestDensity,omitempty"`
	CropType      CropType   `json:"cropType,omitempty"`
}

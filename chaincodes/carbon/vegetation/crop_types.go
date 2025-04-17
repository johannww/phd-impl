package vegetation

const (
	Soybean int = iota
	Corn
	Coffee
	Sugarcane
	Cotton
	Rice
	Wheat
	Beans
	Orange
	Banana
)

const (
	SoybeanStr   = "Soybean"
	CornStr      = "Corn"
	CoffeeStr    = "Coffee"
	SugarcaneStr = "Sugarcane"
	CottonStr    = "Cotton"
	RiceStr      = "Rice"
	WheatStr     = "Wheat"
	BeansStr     = "Beans"
	OrangeStr    = "Orange"
	BananaStr    = "Banana"
)

var CropTypeMap = map[int]string{
	Soybean:   SoybeanStr,
	Corn:      CornStr,
	Coffee:    CoffeeStr,
	Sugarcane: SugarcaneStr,
	Cotton:    CottonStr,
	Rice:      RiceStr,
	Wheat:     WheatStr,
	Beans:     BeansStr,
	Orange:    OrangeStr,
	Banana:    BananaStr,
}

var CropTypeReverseMap = map[string]int{
	SoybeanStr:   Soybean,
	CornStr:      Corn,
	CoffeeStr:    Coffee,
	SugarcaneStr: Sugarcane,
	CottonStr:    Cotton,
	RiceStr:      Rice,
	WheatStr:     Wheat,
	BeansStr:     Beans,
	OrangeStr:    Orange,
	BananaStr:    Banana,
}

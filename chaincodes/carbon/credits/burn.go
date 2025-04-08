package credits

// BurnCredit represents a minted carbon credit to be burned.
// it is associated to burn multiplier and burn timestamp.
type BurnCredit struct {
	MintCredit    `json:"mintCredit"`
	BurnMult      float64 `json:"burnMult"`
	BurnTimeStamp uint64  `json:"burnTimestamp"`
}

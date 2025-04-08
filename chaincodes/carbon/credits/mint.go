package carbon

// MintCredit represents a carbon credit that has been minted and
// it is associated to mint multiplier and mint timestamp.
type MintCredit struct {
	Credit        `json:"credit"`
	MintMult      float64 `json:"mintMult"`
	MintTimeStamp uint64  `json:"mintTimestamp"`
}

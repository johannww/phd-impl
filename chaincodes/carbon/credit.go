package carbon

// Credit represents a carbon unit minted for a property chunk
// at a specific time.
// TODO: enhance this struct
type Credit struct {
	ID       string        `json:"id"`
	Owner    string        `json:"owner"`
	Property Property      `json:"property"`
	Chunk    PropertyChunk `json:"chunk"`
}

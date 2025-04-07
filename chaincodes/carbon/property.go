package carbon

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// PropertyChunk represents a chunk of a property.
// It exists because properties might have heterogeneous chunks.
// It points to the property because otherwise---
type PropertyChunk struct {
	PropertyID  uint64
	Coordinates []Coordinates
}

type Property struct {
	ID uint64
}

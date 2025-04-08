package carbon

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// PropertyChunk represents a chunk of a property.
// It exists because properties might have heterogeneous chunks.
// It points to the property because otherwise---if in a slice in the
// property struct---it could generate MVCC_READ_CONFLICT errors.
// See: https://github.com/hyperledger/fabric/issues/3748
type PropertyChunk struct {
	PropertyID  uint64
	Coordinates []Coordinates
}

type Property struct {
	ID uint64
}

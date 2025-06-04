package credits

import (
	prop "github.com/johannww/phd-impl/chaincodes/carbon/properties"
)

// Credit represents a carbon unit minted for a property chunk
// at a specific time.
// TODO: enhance this struct
type Credit struct {
	OwnerID  string              `json:"owner"`
	ChunkID  []string            `json:"chunkId"`
	Chunk    *prop.PropertyChunk `json:"chunk"`
	Quantity int64               `json:"quantity"`
}

func (c *Credit) GetID() *[][]string {
	creditId := []string{c.OwnerID}
	creditId = append(creditId, (*c.Chunk.GetID())[0]...)
	return &[][]string{creditId}
}

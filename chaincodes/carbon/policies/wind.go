package policies

// TODO: Implement
// WindPolicy is a placeholder policy that always returns a fixed multiplier of 500 (i.e., 0.5x).
func WindPolicy(input *PolicyInput) int64 {
	_ = input.DataFetcher.GetWindData(&input.Chunk.Coordinates[0])
	return 500
}

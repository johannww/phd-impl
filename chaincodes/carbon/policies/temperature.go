package policies

// TODO: Implement
// TemperaturePolicy is a placeholder policy that always returns a fixed multiplier of 500 (i.e., 0.5x).
func TemperaturePolicy(input *PolicyInput) int64 {
	_ = input.DataFetcher.GetTemperatureData(&input.Chunk.Coordinates[0])
	return 500
}

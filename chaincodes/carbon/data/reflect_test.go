package data

import (
	"crypto/sha256"
	"encoding/json"
	"testing"

	"github.com/johannww/phd-impl/chaincodes/carbon/data/emission"
)

func TestReflectWithEmissionData(t *testing.T) {
	// Create an instance of the struct you want to reflect on
	emissionData := &emission.Emission{
		Emitter:   nil,
		Quantity:  15,
		Timeframe: [2]int{2023, 2024},
	}

	bytes, err := json.Marshal(emissionData)
	if err != nil {
		t.Fatalf("Failed to marshal emission data: %v", err)
	}
	// Create an instance of OffChainData
	hash := sha256.Sum256(bytes)
	offChainData := &OffChainData{
		Uri:         "http://example.com/data",
		Method:      "GET",
		ReflectType: "emission",
		Hash:        hash[:],
	}

	offChainData.DataBytes = bytes // simulate data get
	reflectValue, err := offChainData.FromJson()
	if err != nil {
		t.Fatalf("Failed to convert from JSON: %v", err)
	}

	reflectedEmissionData, ok := reflectValue.(*emission.Emission)
	if !ok {
		t.Fatalf("Failed to convert reflect value to emission data")
	}
	if reflectedEmissionData.Quantity != emissionData.Quantity {
		t.Fatalf("Expected quantity %v, got %v", emissionData.Quantity, reflectedEmissionData.Quantity)
	}
}

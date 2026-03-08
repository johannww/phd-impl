package main

import (
	"testing"
)

func TestGenerateProducesCorrectCount(t *testing.T) {
	store := generate("test-seed", 5)
	if len(store) != 5 {
		t.Errorf("expected 5 records, got %d", len(store))
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	a := generate("test-seed", 10)
	b := generate("test-seed", 10)

	for id, respA := range a {
		respB, ok := b[id]
		if !ok {
			t.Errorf("ID %q present in first run but not second", id)
			continue
		}
		if respA.Result[0].AreaTotalImovel != respB.Result[0].AreaTotalImovel {
			t.Errorf("ID %q: AreaTotalImovel differs between runs", id)
		}
		if respA.Result[0].IdentificadorImovel != respB.Result[0].IdentificadorImovel {
			t.Errorf("ID %q: IdentificadorImovel differs between runs", id)
		}
	}
}

func TestGenerateDifferentSeedsDifferentIDs(t *testing.T) {
	a := generate("seed-one", 5)
	b := generate("seed-two", 5)

	for id := range a {
		if _, conflict := b[id]; conflict {
			t.Errorf("ID %q appeared in both seed-one and seed-two outputs", id)
		}
	}
}

func TestGenerateIDsDifferentPerIndex(t *testing.T) {
	const seed = "test-seed"
	seen := make(map[string]bool)
	for i := range 20 {
		id := generateID(seed, i)
		if seen[id] {
			t.Errorf("duplicate ID %q at index %d", id, i)
		}
		seen[id] = true
	}
}

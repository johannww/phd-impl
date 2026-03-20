package main

import (
	"fmt"
	"testing"

	"github.com/johannww/phd-impl/data_api/internal/sicar"
)

func TestGenerateProducesCorrectCount(t *testing.T) {
	data := generate("test-seed", 5)
	if len(data) != 5 {
		t.Errorf("expected 5 records, got %d", len(data))
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	a := generate("test-seed", 10)
	b := generate("test-seed", 10)

	for id, imA := range a {
		imB, ok := b[id]
		if !ok {
			t.Errorf("ID %q present in first run but not second", id)
			continue
		}
		if imA.AreaTotalImovel != imB.AreaTotalImovel {
			t.Errorf("ID %q: AreaTotalImovel differs between runs", id)
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

func TestProjectionsAreConsistent(t *testing.T) {
	data := generate("test-seed", 5)
	for id, im := range data {
		pra := sicar.ImovelToPra(im)
		demo := sicar.ImovelToDemonstrativo(im)
		recibo := sicar.ImovelToRecibo(im)

		praArea := pra.Result[0].AreaTotalImovel
		demoArea := demo.Result[0].AreaTotalImovel
		reciboArea := recibo.Result[0].AreaTotalImovel

		if praArea != fmt.Sprintf("%.4f", im.AreaTotalImovel) {
			t.Errorf("ID %q: Pra area string %q doesn't match canonical float formatted as %%.4f", id, praArea)
		}
		if reciboArea != fmt.Sprintf("%g", im.AreaTotalImovel) {
			t.Errorf("ID %q: Recibo area string %q doesn't match canonical float formatted as %%g", id, reciboArea)
		}
		if demoArea != im.AreaTotalImovel {
			t.Errorf("ID %q: Demonstrativo area (%f) doesn't match canonical (%f)", id, demoArea, im.AreaTotalImovel)
		}

		if demo.Result[0].Municipio != recibo.Result[0].Municipio {
			t.Errorf("ID %q: Municipio inconsistent across endpoints", id)
		}
	}
}

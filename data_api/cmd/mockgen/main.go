package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/johannww/phd-impl/data_api/internal/imovel"
	"github.com/johannww/phd-impl/data_api/internal/mock"
)

func main() {
	out := flag.String("out", "data/sicar.json", "output JSON file path")
	n := flag.Int("n", 10, "number of imoveis to generate")
	seed := flag.String("seed", "data-api-mock", "seed for ID generation")
	flag.Parse()

	data := generate(*seed, *n)

	if err := os.MkdirAll(filepath.Dir(*out), 0755); err != nil {
		log.Fatalf("mkdir: %v", err)
	}

	f, err := os.Create(*out)
	if err != nil {
		log.Fatalf("create: %v", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		log.Fatalf("encode: %v", err)
	}

	log.Printf("wrote %d records to %s", *n, *out)
}

func generate(seed string, n int) imovel.Store {
	store := make(imovel.Store, n)
	for i := range n {
		id := generateID(seed, i)
		store[id] = mock.Imovel(id)
	}
	return store
}

// generateID produces a deterministic codigoImovel from a seed and index.
func generateID(seed string, index int) string {
	h := fnv.New64a()
	fmt.Fprintf(h, "%s-%d", seed, index)
	r := rand.New(rand.NewSource(int64(h.Sum64())))

	ufs := []string{"AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO", "MA",
		"MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI", "RJ", "RN",
		"RS", "RO", "RR", "SC", "SP", "SE", "TO"}
	uf := ufs[r.Intn(len(ufs))]
	num := r.Intn(9000000) + 1000000
	hex := fmt.Sprintf("%032X", r.Uint64())[:32]
	return fmt.Sprintf("%s-%07d-%s", uf, num, hex)
}

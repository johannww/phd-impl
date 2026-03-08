package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/johannww/phd-impl/data_api/internal/cert"
	"github.com/johannww/phd-impl/data_api/internal/sicar"
)

func main() {
	addr := getenv("ADDR", ":8443")
	certFile := getenv("CERT_FILE", "server.crt")
	keyFile := getenv("KEY_FILE", "server.key")
	dataFile := getenv("DATA_FILE", "data/sicar.json")

	store, err := loadStore(dataFile)
	if err != nil {
		log.Fatalf("load store: %v", err)
	}
	log.Printf("loaded %d SICAR records from %s", len(store), dataFile)

	if err := cert.EnsureCert(certFile, keyFile); err != nil {
		log.Fatalf("cert: %v", err)
	}
	tlsCfg, err := cert.LoadTLSConfig(certFile, keyFile)
	if err != nil {
		log.Fatalf("tls: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/sicar/pra/1.0", sicar.Routes(store))

	srv := &http.Server{
		Addr:      addr,
		Handler:   r,
		TLSConfig: tlsCfg,
	}

	log.Printf("listening on https://%s", addr)
	if err := srv.ListenAndServeTLS("", ""); err != nil {
		log.Fatal(err)
	}
}

func loadStore(path string) (sicar.Store, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var store sicar.Store
	return store, json.NewDecoder(f).Decode(&store)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

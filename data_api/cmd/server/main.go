package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/johannww/phd-impl/data_api/internal/cert"
	"github.com/johannww/phd-impl/data_api/internal/imovel"
	"github.com/johannww/phd-impl/data_api/internal/sicar"
)

func main() {
	addr := getenv("ADDR", ":8443")
	certFile := getenv("CERT_FILE", "server.crt")
	keyFile := getenv("KEY_FILE", "server.key")
	dataFilePath := getenv("DATA_FILE", "data/sicar.json")

	store, err := loadData(dataFilePath)
	if err != nil {
		log.Fatalf("load data: %v", err)
	}
	log.Printf("loaded %d canonical imoveis from %s", len(store), dataFilePath)

	praStore := make(sicar.Store, len(store))
	demoStore := make(sicar.DemonstrativoStore, len(store))
	reciboStore := make(sicar.ReciboStore, len(store))
	for id, im := range store {
		praStore[id] = sicar.ImovelToPra(im)
		demoStore[id] = sicar.ImovelToDemonstrativo(im)
		reciboStore[id] = sicar.ImovelToRecibo(im)
	}

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
	r.Route("/sicar/pra/1.0", sicar.Routes(praStore))
	r.Route("/sicar/demonstrativoDegustacao/1.0", sicar.DemonstrativoRoutes(demoStore))
	r.Route("/sicar/recibo/1.0", sicar.ReciboRoutes(reciboStore))

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

func loadData(path string) (imovel.Store, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var store imovel.Store
	return store, json.NewDecoder(f).Decode(&store)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

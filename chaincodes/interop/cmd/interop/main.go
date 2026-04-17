package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/johannww/phd-impl/chaincodes/interop/contract"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	metricsAddr := getenv("METRICS_ADDR", ":9443")
	metricsLn, err := net.Listen("tcp", metricsAddr)
	if err != nil {
		panic(err)
	}
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Printf("metrics endpoint listening on %s/metrics", metricsAddr)
		if serveErr := http.Serve(metricsLn, mux); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			panic(serveErr)
		}
	}()

	interopCC, err := contractapi.NewChaincode(contract.NewInteropContract())
	if err != nil {
		panic(err)
	}

	if err := interopCC.Start(); err != nil {
		panic(err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

package sicar

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Store maps codigoImovel to its pre-loaded response.
type Store map[string]SicarResponse

func Routes(store Store) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/{codigoImovel}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "codigoImovel")
			resp, ok := store[id]
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		})
	}
}

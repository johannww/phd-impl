package sicar

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Store = map[string]PraResponse
type DemonstrativoStore = map[string]DemonstrativoResponse
type ReciboStore = map[string]ReciboResponse

func routesByID[T any](store map[string]T) func(chi.Router) {
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

func Routes(store Store) func(chi.Router)              { return routesByID(store) }
func DemonstrativoRoutes(store DemonstrativoStore) func(chi.Router) { return routesByID(store) }
func ReciboRoutes(store ReciboStore) func(chi.Router)  { return routesByID(store) }

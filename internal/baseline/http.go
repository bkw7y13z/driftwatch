package baseline

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler returns an http.Handler that exposes baseline management over HTTP.
//
//	GET  /baseline/{service}        – retrieve pinned baseline
//	POST /baseline/{service}        – pin a new baseline (JSON body: Entry)
//	DELETE /baseline/{service}      – remove pinned baseline
func Handler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		service := strings.TrimPrefix(r.URL.Path, "/baseline/")
		service = strings.Trim(service, "/")
		if service == "" {
			http.Error(w, "service name required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			e, ok := s.Get(service)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(e)

		case http.MethodPost:
			var e Entry
			if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			e.Service = service
			if err := s.Pin(e); err != nil {
				http.Error(w, "store error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)

		case http.MethodDelete:
			if err := s.Remove(service); err != nil {
				http.Error(w, "store error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

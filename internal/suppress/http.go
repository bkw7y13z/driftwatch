package suppress

import (
	"encoding/json"
	"net/http"
	"time"
)

type addRequest struct {
	Service string        `json:"service"`
	Path    string        `json:"path"`
	Reason  string        `json:"reason"`
	TTL     time.Duration `json:"ttl_ns"` // nanoseconds for JSON simplicity
}

// Handler returns an http.Handler that exposes the suppression store over
// a simple JSON API.
//
//	POST /suppress  — add a new entry (body: addRequest JSON)
//	GET  /suppress  — list active entries
func Handler(s *Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/suppress", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req addRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if req.Service == "" || req.Path == "" {
				http.Error(w, "service and path are required", http.StatusBadRequest)
				return
			}
			if req.TTL <= 0 {
				req.TTL = time.Hour
			}
			s.Add(req.Service, req.Path, req.Reason, req.TTL)
			w.WriteHeader(http.StatusCreated)
		case http.MethodGet:
			active := s.Active()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(active)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}

package changelog

import (
	"encoding/json"
	"net/http"
	"time"
)

type entryJSON struct {
	RecordedAt  time.Time `json:"recorded_at"`
	ServiceName string    `json:"service_name"`
	Status      string    `json:"status"`
	FilePath    string    `json:"file_path,omitempty"`
}

// Handler returns an http.Handler that serves the changelog as JSON.
// DELETE /changelog clears the log.
func Handler(l *Log) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			l.Clear()
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			entries := l.Entries()
			out := make([]entryJSON, 0, len(entries))
			for _, e := range entries {
				out = append(out, entryJSON{
					RecordedAt:  e.RecordedAt,
					ServiceName: e.ServiceName,
					Status:      string(e.Event.Status),
					FilePath:    e.Event.FilePath,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(out)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

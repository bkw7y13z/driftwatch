// Package healthz exposes a simple HTTP health-check endpoint that reports
// the daemon's liveness and the timestamp of the last completed drift check.
package healthz

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"log/slog"
)

// Handler is an HTTP handler that serves health status.
type Handler struct {
	mu          sync.RWMutex
	lastCheck   time.Time
	lastDrifted bool
	log         *slog.Logger
}

// New returns a new Handler using the provided logger.
// If logger is nil a default logger is used.
func New(log *slog.Logger) *Handler {
	if log == nil {
		log = slog.Default()
	}
	return &Handler{log: log}
}

// RecordCheck updates the internal state after each drift-check cycle.
func (h *Handler) RecordCheck(drifted bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastCheck = time.Now().UTC()
	h.lastDrifted = drifted
}

type response struct {
	Status      string `json:"status"`
	LastCheck   string `json:"last_check,omitempty"`
	DriftedSeen bool   `json:"drifted_seen"`
}

// ServeHTTP implements http.Handler.
// Returns 200 OK when healthy, 503 if no check has been recorded yet.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	lc := h.lastCheck
	drifted := h.lastDrifted
	h.mu.RUnlock()

	resp := response{DriftedSeen: drifted}

	if lc.IsZero() {
		resp.Status = "starting"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		resp.Status = "ok"
		resp.LastCheck = lc.Format(time.RFC3339)
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("healthz: failed to encode response", "err", err)
	}
}

// Register mounts the health handler at /healthz on the given mux.
func Register(mux *http.ServeMux, h *Handler) {
	mux.Handle("/healthz", h)
}

// Package statuspage provides an HTTP handler that renders a human-readable
// summary of the most recent drift-check results for all monitored services.
package statuspage

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/driftwatch/internal/drift"
)

// ServiceStatus holds the last known status for a single service.
type ServiceStatus struct {
	Service   string    `json:"service"`
	HasDrift  bool      `json:"has_drift"`
	DriftedAt time.Time `json:"drifted_at,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
	Summary   string    `json:"summary"`
}

// Page aggregates per-service statuses and exposes an HTTP handler.
type Page struct {
	mu       sync.RWMutex
	statuses map[string]ServiceStatus
}

// New returns a ready-to-use Page.
func New() *Page {
	return &Page{statuses: make(map[string]ServiceStatus)}
}

// Record updates the stored status for the service referenced by the report.
func (p *Page) Record(report *drift.Report) {
	if report == nil {
		return
	}

	now := time.Now().UTC()
	st := ServiceStatus{
		Service:   report.Service,
		HasDrift:  report.HasDrift(),
		CheckedAt: now,
		Summary:   report.Summary(),
	}
	if st.HasDrift {
		st.DriftedAt = now
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Preserve original DriftedAt if service was already drifted.
	if prev, ok := p.statuses[report.Service]; ok && prev.HasDrift && st.HasDrift {
		st.DriftedAt = prev.DriftedAt
	}
	p.statuses[report.Service] = st
}

// Statuses returns a snapshot of all service statuses.
func (p *Page) Statuses() []ServiceStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]ServiceStatus, 0, len(p.statuses))
	for _, s := range p.statuses {
		out = append(out, s)
	}
	return out
}

// ServeHTTP renders the current status page as JSON.
func (p *Page) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	statuses := p.Statuses()

	code := http.StatusOK
	for _, s := range statuses {
		if s.HasDrift {
			code = http.StatusConflict
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"services": statuses,
		"generated_at": time.Now().UTC(),
	})
}

// Register mounts the handler at /statuspage on mux.
func Register(mux *http.ServeMux, p *Page) {
	mux.Handle("/statuspage", p)
}

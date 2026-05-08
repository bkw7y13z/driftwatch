package policycheck

import (
	"encoding/json"
	"net/http"

	"github.com/driftwatch/internal/drift"
)

type reportStore interface {
	Latest() *drift.Report
}

// Handler returns an http.Handler that evaluates the latest drift report
// against the checker's policies and returns any violations as JSON.
// Responds 204 when there are no violations, 409 when violations exist.
func Handler(c *Checker, store reportStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		report := store.Latest()
		violations := c.Evaluate(report)

		w.Header().Set("Content-Type", "application/json")

		if len(violations) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(violations)
	})
}

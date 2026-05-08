package diffrender

import (
	"net/http"
	"strings"

	"github.com/example/driftwatch/internal/drift"
)

// Handler returns an http.HandlerFunc that renders the most recent drift report
// as a plain-text unified diff. The report is retrieved via the provided
// getter function each time the endpoint is called.
func Handler(getReport func() *drift.Report) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := getReport()
		if report == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Detect whether the client prefers colour (e.g. curl with --color).
		accept := r.Header.Get("Accept")
		colours := strings.Contains(accept, "text/x-ansi")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if report.HasDrift() {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		renderer := New(w, colours)
		_ = renderer.Render(report) // best-effort; headers already sent
	}
}

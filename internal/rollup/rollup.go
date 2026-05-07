// Package rollup aggregates drift reports over a rolling window,
// providing summary statistics for dashboards and alerting.
package rollup

import (
	"sync"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Window holds aggregated drift statistics over a sliding time window.
type Window struct {
	mu       sync.Mutex
	entries  []entry
	maxAge   time.Duration
	now      func() time.Time
}

type entry struct {
	at        time.Time
	drifted   int
	missing   int
	matched   int
}

// New creates a Window that retains entries younger than maxAge.
func New(maxAge time.Duration) *Window {
	return &Window{
		maxAge: maxAge,
		now:    time.Now,
	}
}

// Record adds the counts from report to the rolling window.
func (w *Window) Record(report *drift.Report) {
	if report == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	var drifted, missing, matched int
	for _, ev := range report.Events {
		switch ev.Status {
		case drift.StatusDrifted:
			drifted++
		case drift.StatusMissing:
			missing++
		case drift.StatusMatch:
			matched++
		}
	}

	w.entries = append(w.entries, entry{
		at:      w.now(),
		drifted: drifted,
		missing: missing,
		matched: matched,
	})
	w.evict()
}

// Stats returns aggregated totals across all entries still in the window.
func (w *Window) Stats() (drifted, missing, matched int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict()
	for _, e := range w.entries {
		drifted += e.drifted
		missing += e.missing
		matched += e.matched
	}
	return
}

// Len returns the number of recorded entries currently in the window.
func (w *Window) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict()
	return len(w.entries)
}

// evict removes entries older than maxAge. Must be called with w.mu held.
func (w *Window) evict() {
	cutoff := w.now().Add(-w.maxAge)
	i := 0
	for i < len(w.entries) && w.entries[i].at.Before(cutoff) {
		i++
	}
	w.entries = w.entries[i:]
}

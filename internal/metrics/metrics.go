// Package metrics provides lightweight in-process counters and gauges
// for tracking drift detection activity over time.
package metrics

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Collector holds runtime metrics for the driftwatch daemon.
type Collector struct {
	mu sync.RWMutex

	ChecksTotal   int64
	DriftedTotal  int64
	MissingTotal  int64
	MatchedTotal  int64
	LastCheckTime time.Time
	LastDriftTime time.Time
}

// New returns an initialised Collector.
func New() *Collector {
	return &Collector{}
}

// RecordCheck records the outcome of a single drift-check cycle.
func (c *Collector) RecordCheck(drifted, missing, matched int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ChecksTotal++
	c.DriftedTotal += int64(drifted)
	c.MissingTotal += int64(missing)
	c.MatchedTotal += int64(matched)
	c.LastCheckTime = time.Now()

	if drifted > 0 || missing > 0 {
		c.LastDriftTime = time.Now()
	}
}

// Snapshot returns a point-in-time copy of the current metrics.
func (c *Collector) Snapshot() Collector {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return Collector{
		ChecksTotal:   c.ChecksTotal,
		DriftedTotal:  c.DriftedTotal,
		MissingTotal:  c.MissingTotal,
		MatchedTotal:  c.MatchedTotal,
		LastCheckTime: c.LastCheckTime,
		LastDriftTime: c.LastDriftTime,
	}
}

// WriteTo writes a human-readable summary of the metrics to w.
func (c *Collector) WriteTo(w io.Writer) {
	snap := c.Snapshot()
	fmt.Fprintf(w, "checks_total:   %d\n", snap.ChecksTotal)
	fmt.Fprintf(w, "drifted_total:  %d\n", snap.DriftedTotal)
	fmt.Fprintf(w, "missing_total:  %d\n", snap.MissingTotal)
	fmt.Fprintf(w, "matched_total:  %d\n", snap.MatchedTotal)
	if !snap.LastCheckTime.IsZero() {
		fmt.Fprintf(w, "last_check:     %s\n", snap.LastCheckTime.Format(time.RFC3339))
	}
	if !snap.LastDriftTime.IsZero() {
		fmt.Fprintf(w, "last_drift:     %s\n", snap.LastDriftTime.Format(time.RFC3339))
	}
}

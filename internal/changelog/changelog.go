// Package changelog records a bounded history of drift events so that
// operators can review what changed and when without consulting external
// storage.
package changelog

import (
	"sync"
	"time"

	"github.com/yourusername/driftwatch/internal/drift"
)

// Entry is a single record in the changelog.
type Entry struct {
	RecordedAt time.Time
	ServiceName string
	Event       drift.Event
}

// Log holds a capped, ordered list of drift entries.
type Log struct {
	mu      sync.RWMutex
	entries []Entry
	cap     int
	clock   func() time.Time
}

// New returns a Log that retains at most maxEntries records.
// Older entries are evicted when the cap is exceeded.
func New(maxEntries int) *Log {
	if maxEntries <= 0 {
		maxEntries = 100
	}
	return &Log{
		cap:   maxEntries,
		clock: time.Now,
	}
}

// Record appends every event in report to the log.
// Nil reports are silently ignored.
func (l *Log) Record(report *drift.Report) {
	if report == nil {
		return
	}
	now := l.clock()
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, ev := range report.Events {
		l.entries = append(l.entries, Entry{
			RecordedAt:  now,
			ServiceName: ev.ServiceName,
			Event:       ev,
		})
	}
	if len(l.entries) > l.cap {
		l.entries = l.entries[len(l.entries)-l.cap:]
	}
}

// Entries returns a snapshot of the current log contents, newest first.
func (l *Log) Entries() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Entry, len(l.entries))
	for i, e := range l.entries {
		out[len(l.entries)-1-i] = e
	}
	return out
}

// Clear removes all entries from the log.
func (l *Log) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = l.entries[:0]
}

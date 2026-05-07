// Package debounce provides a simple debouncer that suppresses repeated
// notifications for the same key within a configurable quiet period.
// This is useful for avoiding alert storms when a service oscillates
// between drifted and matched states in rapid succession.
package debounce

import (
	"sync"
	"time"
)

// Debouncer tracks the last time a key was allowed through and suppresses
// subsequent calls until the quiet period has elapsed.
type Debouncer struct {
	mu      sync.Mutex
	quiet   time.Duration
	lastSeen map[string]time.Time
	now     func() time.Time
}

// New creates a Debouncer with the given quiet period.
// Events for a given key are suppressed until quiet has elapsed since
// the last allowed event for that key.
func New(quiet time.Duration) *Debouncer {
	return &Debouncer{
		quiet:    quiet,
		lastSeen: make(map[string]time.Time),
		now:      time.Now,
	}
}

// Allow reports whether the event for key should be allowed through.
// The first call for a key is always allowed. Subsequent calls are
// suppressed until the quiet period has elapsed.
func (d *Debouncer) Allow(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	last, seen := d.lastSeen[key]
	if !seen || now.Sub(last) >= d.quiet {
		d.lastSeen[key] = now
		return true
	}
	return false
}

// Reset removes the record for key, allowing the next event through
// regardless of when the last event occurred.
func (d *Debouncer) Reset(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.lastSeen, key)
}

// Purge removes all keys whose last-seen time is older than the quiet
// period. Call periodically to prevent unbounded memory growth.
func (d *Debouncer) Purge() {
	d.mu.Lock()
	defer d.mu.Unlock()

	cutoff := d.now().Add(-d.quiet)
	for k, t := range d.lastSeen {
		if t.Before(cutoff) {
			delete(d.lastSeen, k)
		}
	}
}

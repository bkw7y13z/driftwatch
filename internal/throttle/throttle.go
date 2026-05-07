// Package throttle provides a token-bucket style check-rate limiter that
// prevents a single service from consuming disproportionate check cycles.
package throttle

import (
	"sync"
	"time"
)

// Throttle tracks per-service check budgets and enforces a maximum number of
// checks within a rolling window.
type Throttle struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	max     int
	window  time.Duration
	now     func() time.Time
}

type bucket struct {
	count     int
	windowEnd time.Time
}

// New returns a Throttle that allows at most maxChecks per service within
// the given window duration.
func New(maxChecks int, window time.Duration) *Throttle {
	return &Throttle{
		buckets: make(map[string]*bucket),
		max:     maxChecks,
		window:  window,
		now:     time.Now,
	}
}

// Allow reports whether the named service is permitted to run a check right
// now. It increments the service's counter if permitted.
func (t *Throttle) Allow(service string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	b, ok := t.buckets[service]
	if !ok || now.After(b.windowEnd) {
		t.buckets[service] = &bucket{count: 1, windowEnd: now.Add(t.window)}
		return true
	}
	if b.count >= t.max {
		return false
	}
	b.count++
	return true
}

// Remaining returns how many checks the named service may still perform in
// the current window. Returns max if no window is active.
func (t *Throttle) Remaining(service string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	b, ok := t.buckets[service]
	if !ok || now.After(b.windowEnd) {
		return t.max
	}
	rem := t.max - b.count
	if rem < 0 {
		return 0
	}
	return rem
}

// Reset clears the bucket for the named service, restoring its full budget.
func (t *Throttle) Reset(service string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.buckets, service)
}

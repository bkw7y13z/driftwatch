// Package circuitbreaker provides a simple circuit breaker that prevents
// repeated alert delivery to a failing downstream when errors exceed a
// configurable threshold within a rolling window.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned by Allow when the circuit is open.
var ErrOpen = errors.New("circuit breaker is open")

// State represents the current circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// Breaker is a thread-safe circuit breaker.
type Breaker struct {
	mu          sync.Mutex
	threshold   int
	window      time.Duration
	cooldown    time.Duration
	failures    []time.Time
	openUntil   time.Time
	now         func() time.Time
}

// New creates a Breaker that opens after threshold failures within window,
// and attempts recovery after cooldown.
func New(threshold int, window, cooldown time.Duration) *Breaker {
	return &Breaker{
		threshold: threshold,
		window:    window,
		cooldown:  cooldown,
		now:       time.Now,
	}
}

// Allow returns nil if the call should proceed, or ErrOpen if the circuit is
// open. A half-open probe is permitted once per cooldown period.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.now()

	if now.Before(b.openUntil) {
		return ErrOpen
	}

	// Prune failures outside the rolling window.
	cutoff := now.Add(-b.window)
	valid := b.failures[:0]
	for _, t := range b.failures {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	b.failures = valid

	return nil
}

// RecordSuccess resets the failure counter and closes the circuit.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = nil
	b.openUntil = time.Time{}
}

// RecordFailure records a failure. If the threshold is exceeded the circuit
// opens for the configured cooldown duration.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.now()
	b.failures = append(b.failures, now)
	if len(b.failures) >= b.threshold {
		b.openUntil = now.Add(b.cooldown)
	}
}

// State returns the current state of the circuit breaker.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.now()
	switch {
	case now.Before(b.openUntil):
		return StateOpen
	case !b.openUntil.IsZero():
		return StateHalfOpen
	default:
		return StateClosed
	}
}

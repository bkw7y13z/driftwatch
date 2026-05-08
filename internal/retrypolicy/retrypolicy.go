// Package retrypolicy provides a configurable retry policy for drift check
// operations, with exponential back-off and a maximum attempt ceiling.
package retrypolicy

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

// Policy holds the parameters that govern retry behaviour.
type Policy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	log         *slog.Logger
}

// New returns a Policy with sensible defaults.
// MaxAttempts=3, BaseDelay=200ms, MaxDelay=5s.
func New(log *slog.Logger) *Policy {
	if log == nil {
		log = slog.Default()
	}
	return &Policy{
		MaxAttempts: 3,
		BaseDelay:   200 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		log:         log,
	}
}

// Do executes fn, retrying on non-nil errors up to MaxAttempts times.
// Each retry waits BaseDelay * 2^(attempt-1), capped at MaxDelay.
// The context is checked before every attempt; cancellation stops retries
// immediately and returns the context error.
func (p *Policy) Do(ctx context.Context, fn func() error) error {
	var last error
	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		last = fn()
		if last == nil {
			return nil
		}
		if attempt == p.MaxAttempts {
			break
		}
		delay := p.backoff(attempt)
		p.log.Warn("retrypolicy: attempt failed, will retry",
			"attempt", attempt,
			"delay", delay,
			"err", last,
		)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return last
}

// backoff returns the delay for the given attempt number (1-based).
func (p *Policy) backoff(attempt int) time.Duration {
	delay := p.BaseDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay > p.MaxDelay {
			return p.MaxDelay
		}
	}
	return delay
}

// Permanent wraps an error to signal that retries should not be attempted.
type Permanent struct{ Err error }

func (p Permanent) Error() string { return p.Err.Error() }
func (p Permanent) Unwrap() error { return p.Err }

// IsPermanent reports whether err (or any error in its chain) is a Permanent
// sentinel, meaning the caller should not retry.
func IsPermanent(err error) bool {
	var p Permanent
	return errors.As(err, &p)
}

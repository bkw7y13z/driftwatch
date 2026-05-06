// Package scheduler provides a simple interval-based scheduler that
// triggers a callback on a fixed cadence, respecting context cancellation.
package scheduler

import (
	"context"
	"log/slog"
	"time"
)

// TickFunc is the function called on each scheduled tick.
type TickFunc func(ctx context.Context) error

// Scheduler runs a TickFunc at a fixed interval until the context is cancelled.
type Scheduler struct {
	interval time.Duration
	fn       TickFunc
	log      *slog.Logger
}

// New creates a new Scheduler with the given interval and tick function.
// If logger is nil, a default logger is used.
func New(interval time.Duration, fn TickFunc, logger *slog.Logger) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Scheduler{
		interval: interval,
		fn:       fn,
		log:      logger,
	}
}

// Run starts the scheduler loop. It fires immediately on the first tick, then
// repeats at the configured interval. Blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	s.log.Info("scheduler started", "interval", s.interval)

	// Fire immediately before waiting for the first interval.
	s.tick(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("scheduler stopped")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	start := time.Now()
	if err := s.fn(ctx); err != nil {
		s.log.Error("tick error", "err", err, "duration", time.Since(start))
		return
	}
	s.log.Debug("tick completed", "duration", time.Since(start))
}

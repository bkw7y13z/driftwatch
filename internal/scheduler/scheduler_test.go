package scheduler_test

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/scheduler"
)

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError + 10}))
}

func TestNew_DefaultLogger(t *testing.T) {
	s := scheduler.New(time.Second, func(_ context.Context) error { return nil }, nil)
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestRun_FiresImmediately(t *testing.T) {
	var count atomic.Int32

	ctx, cancel := context.WithCancel(context.Background())

	s := scheduler.New(10*time.Second, func(_ context.Context) error {
		count.Add(1)
		cancel() // cancel after first tick
		return nil
	}, silentLogger())

	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not stop within timeout")
	}

	if count.Load() != 1 {
		t.Fatalf("expected 1 tick, got %d", count.Load())
	}
}

func TestRun_RepeatsOnInterval(t *testing.T) {
	var count atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	s := scheduler.New(100*time.Millisecond, func(_ context.Context) error {
		count.Add(1)
		return nil
	}, silentLogger())

	s.Run(ctx)

	// Expect: immediate + ~3 interval ticks within 350ms
	if count.Load() < 3 {
		t.Fatalf("expected at least 3 ticks, got %d", count.Load())
	}
}

func TestRun_ContinuesOnError(t *testing.T) {
	var count atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	s := scheduler.New(80*time.Millisecond, func(_ context.Context) error {
		count.Add(1)
		return errors.New("simulated error")
	}, silentLogger())

	s.Run(ctx)

	if count.Load() < 2 {
		t.Fatalf("expected scheduler to continue after error, got %d ticks", count.Load())
	}
}

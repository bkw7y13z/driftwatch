package alerting_test

import (
	"context"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/alerting"
	"github.com/yourorg/driftwatch/internal/drift"
)

// TestAlerter_ConcurrentEvaluate verifies that concurrent Evaluate calls do
// not race or panic under the -race detector.
func TestAlerter_ConcurrentEvaluate(t *testing.T) {
	var fired atomic.Int64
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := alerting.New(5, func(_ context.Context, _ string, _ int) {
		fired.Add(1)
	}, log)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	const goroutines = 10
	const iterations = 20
	done := make(chan struct{}, goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for i := 0; i < iterations; i++ {
				a.Evaluate(ctx, makeReport(
					drift.Event{Service: "concurrent-svc", Status: drift.StatusDrifted},
				))
			}
		}()
	}

	for g := 0; g < goroutines; g++ {
		select {
		case <-done:
		case <-ctx.Done():
			t.Fatal("timed out waiting for goroutines")
		}
	}

	if fired.Load() == 0 {
		t.Error("expected at least one alert to fire")
	}
}

// TestAlerter_ResetMidStream verifies that resetting a counter mid-stream
// restarts the threshold window correctly.
func TestAlerter_ResetMidStream(t *testing.T) {
	fired := 0
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := alerting.New(3, func(_ context.Context, _ string, _ int) { fired++ }, log)
	ctx := context.Background()

	a.Evaluate(ctx, makeReport(drift.Event{Service: "svc", Status: drift.StatusDrifted}))
	a.Evaluate(ctx, makeReport(drift.Event{Service: "svc", Status: drift.StatusDrifted}))
	a.Reset("svc")
	// Counter is back to 0; need 3 more to fire.
	a.Evaluate(ctx, makeReport(drift.Event{Service: "svc", Status: drift.StatusDrifted}))
	a.Evaluate(ctx, makeReport(drift.Event{Service: "svc", Status: drift.StatusDrifted}))

	if fired != 0 {
		t.Errorf("expected no alerts after reset, got %d", fired)
	}

	a.Evaluate(ctx, makeReport(drift.Event{Service: "svc", Status: drift.StatusDrifted}))
	if fired != 1 {
		t.Errorf("expected exactly 1 alert after reaching threshold again, got %d", fired)
	}
}

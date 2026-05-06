package alerting_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/yourorg/driftwatch/internal/alerting"
	"github.com/yourorg/driftwatch/internal/drift"
)

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func makeReport(events ...drift.Event) *drift.Report {
	return &drift.Report{Events: events}
}

func TestEvaluate_NilReport(t *testing.T) {
	a := alerting.New(2, nil, slog.Default())
	// must not panic
	a.Evaluate(context.Background(), nil)
}

func TestEvaluate_MatchResetsCount(t *testing.T) {
	a := alerting.New(3, nil, slog.Default())
	ctx := context.Background()

	a.Evaluate(ctx, makeReport(drift.Event{Service: "svc", Status: drift.StatusDrifted}))
	if a.Count("svc") != 1 {
		t.Fatalf("expected count 1, got %d", a.Count("svc"))
	}
	a.Evaluate(ctx, makeReport(drift.Event{Service: "svc", Status: drift.StatusMatch}))
	if a.Count("svc") != 0 {
		t.Fatalf("expected count reset to 0, got %d", a.Count("svc"))
	}
}

func TestEvaluate_AlertFiredAtThreshold(t *testing.T) {
	fired := 0
	var firedSvc string
	var firedCount int

	fn := func(_ context.Context, svc string, count int) {
		fired++
		firedSvc = svc
		firedCount = count
	}

	a := alerting.New(3, fn, slog.Default())
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		a.Evaluate(ctx, makeReport(drift.Event{Service: "api", Status: drift.StatusDrifted}))
	}

	if fired != 1 {
		t.Fatalf("expected alert fired once, got %d", fired)
	}
	if firedSvc != "api" {
		t.Errorf("expected service 'api', got %q", firedSvc)
	}
	if firedCount != 3 {
		t.Errorf("expected count 3, got %d", firedCount)
	}
}

func TestEvaluate_AlertFiredOnEachSubsequentDrift(t *testing.T) {
	fired := 0
	a := alerting.New(2, func(_ context.Context, _ string, _ int) { fired++ }, slog.Default())
	ctx := context.Background()

	for i := 0; i < 4; i++ {
		a.Evaluate(ctx, makeReport(drift.Event{Service: "db", Status: drift.StatusMissing}))
	}
	// threshold=2: fires at 2, 3, 4 → 3 times
	if fired != 3 {
		t.Fatalf("expected 3 alerts, got %d", fired)
	}
}

func TestReset_ClearsCount(t *testing.T) {
	a := alerting.New(5, nil, slog.Default())
	ctx := context.Background()
	a.Evaluate(ctx, makeReport(drift.Event{Service: "svc", Status: drift.StatusDrifted}))
	a.Reset("svc")
	if a.Count("svc") != 0 {
		t.Errorf("expected count 0 after reset, got %d", a.Count("svc"))
	}
}

func TestEvaluate_IndependentServices(t *testing.T) {
	a := alerting.New(10, nil, slog.Default())
	ctx := context.Background()
	a.Evaluate(ctx, makeReport(
		drift.Event{Service: "alpha", Status: drift.StatusDrifted},
		drift.Event{Service: "beta", Status: drift.StatusMatch},
	))
	if a.Count("alpha") != 1 {
		t.Errorf("alpha: expected 1, got %d", a.Count("alpha"))
	}
	if a.Count("beta") != 0 {
		t.Errorf("beta: expected 0, got %d", a.Count("beta"))
	}
}

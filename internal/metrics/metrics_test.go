package metrics_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourorg/driftwatch/internal/metrics"
)

func TestNew(t *testing.T) {
	c := metrics.New()
	if c == nil {
		t.Fatal("expected non-nil Collector")
	}
	snap := c.Snapshot()
	if snap.ChecksTotal != 0 || snap.DriftedTotal != 0 {
		t.Errorf("expected zero counters, got %+v", snap)
	}
}

func TestRecordCheck_AllMatch(t *testing.T) {
	c := metrics.New()
	c.RecordCheck(0, 0, 5)

	snap := c.Snapshot()
	if snap.ChecksTotal != 1 {
		t.Errorf("ChecksTotal want 1, got %d", snap.ChecksTotal)
	}
	if snap.MatchedTotal != 5 {
		t.Errorf("MatchedTotal want 5, got %d", snap.MatchedTotal)
	}
	if !snap.LastDriftTime.IsZero() {
		t.Error("LastDriftTime should be zero when no drift")
	}
}

func TestRecordCheck_WithDrift(t *testing.T) {
	c := metrics.New()
	c.RecordCheck(2, 1, 3)

	snap := c.Snapshot()
	if snap.DriftedTotal != 2 {
		t.Errorf("DriftedTotal want 2, got %d", snap.DriftedTotal)
	}
	if snap.MissingTotal != 1 {
		t.Errorf("MissingTotal want 1, got %d", snap.MissingTotal)
	}
	if snap.LastDriftTime.IsZero() {
		t.Error("LastDriftTime should be set when drift detected")
	}
}

func TestRecordCheck_Accumulates(t *testing.T) {
	c := metrics.New()
	c.RecordCheck(1, 0, 2)
	c.RecordCheck(0, 1, 3)

	snap := c.Snapshot()
	if snap.ChecksTotal != 2 {
		t.Errorf("ChecksTotal want 2, got %d", snap.ChecksTotal)
	}
	if snap.DriftedTotal != 1 {
		t.Errorf("DriftedTotal want 1, got %d", snap.DriftedTotal)
	}
	if snap.MissingTotal != 1 {
		t.Errorf("MissingTotal want 1, got %d", snap.MissingTotal)
	}
	if snap.MatchedTotal != 5 {
		t.Errorf("MatchedTotal want 5, got %d", snap.MatchedTotal)
	}
}

func TestWriteTo(t *testing.T) {
	c := metrics.New()
	c.RecordCheck(1, 0, 4)

	var buf bytes.Buffer
	c.WriteTo(&buf)
	out := buf.String()

	for _, want := range []string{"checks_total", "drifted_total", "matched_total", "last_check", "last_drift"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

func TestWriteTo_NoDriftTime(t *testing.T) {
	c := metrics.New()
	c.RecordCheck(0, 0, 1)

	var buf bytes.Buffer
	c.WriteTo(&buf)
	out := buf.String()

	if strings.Contains(out, "last_drift") {
		t.Errorf("last_drift should not appear when no drift recorded\ngot:\n%s", out)
	}
}

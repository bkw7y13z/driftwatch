package rollup

import (
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeReport(drifted, missing, matched int) *drift.Report {
	events := make([]drift.Event, 0, drifted+missing+matched)
	for i := 0; i < drifted; i++ {
		events = append(events, drift.Event{Status: drift.StatusDrifted})
	}
	for i := 0; i < missing; i++ {
		events = append(events, drift.Event{Status: drift.StatusMissing})
	}
	for i := 0; i < matched; i++ {
		events = append(events, drift.Event{Status: drift.StatusMatch})
	}
	return &drift.Report{Events: events}
}

func TestRecord_NilReport(t *testing.T) {
	w := New(time.Minute)
	w.Record(nil)
	if w.Len() != 0 {
		t.Fatalf("expected 0 entries, got %d", w.Len())
	}
}

func TestStats_SingleRecord(t *testing.T) {
	w := New(time.Minute)
	w.Record(makeReport(2, 1, 3))
	d, m, ok := w.Stats()
	if d != 2 || m != 1 || ok != 3 {
		t.Fatalf("expected 2/1/3, got %d/%d/%d", d, m, ok)
	}
}

func TestStats_MultipleRecords_Accumulate(t *testing.T) {
	w := New(time.Minute)
	w.Record(makeReport(1, 0, 2))
	w.Record(makeReport(0, 1, 1))
	d, m, ok := w.Stats()
	if d != 1 || m != 1 || ok != 3 {
		t.Fatalf("expected 1/1/3, got %d/%d/%d", d, m, ok)
	}
}

func TestEviction_RemovesOldEntries(t *testing.T) {
	w := New(time.Second)

	past := time.Now().Add(-2 * time.Second)
	w.mu.Lock()
	w.entries = append(w.entries, entry{at: past, drifted: 5})
	w.mu.Unlock()

	w.Record(makeReport(1, 0, 0))

	d, _, _ := w.Stats()
	if d != 1 {
		t.Fatalf("expected old entry evicted, drifted should be 1, got %d", d)
	}
	if w.Len() != 1 {
		t.Fatalf("expected 1 entry after eviction, got %d", w.Len())
	}
}

func TestLen_ReflectsWindowSize(t *testing.T) {
	w := New(time.Minute)
	if w.Len() != 0 {
		t.Fatal("expected empty window")
	}
	w.Record(makeReport(0, 0, 1))
	w.Record(makeReport(0, 0, 1))
	if w.Len() != 2 {
		t.Fatalf("expected 2, got %d", w.Len())
	}
}

func TestStats_EmptyWindow(t *testing.T) {
	w := New(time.Minute)
	d, m, ok := w.Stats()
	if d != 0 || m != 0 || ok != 0 {
		t.Fatal("expected all zeros for empty window")
	}
}

package changelog

import (
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/drift"
)

func makeReport(names ...string) *drift.Report {
	r := &drift.Report{}
	for _, n := range names {
		r.Events = append(r.Events, drift.Event{
			ServiceName: n,
			Status:      drift.StatusDrifted,
		})
	}
	return r
}

func TestRecord_NilReport(t *testing.T) {
	l := New(10)
	l.Record(nil)
	if got := len(l.Entries()); got != 0 {
		t.Fatalf("expected 0 entries, got %d", got)
	}
}

func TestRecord_AppendsEvents(t *testing.T) {
	l := New(10)
	l.Record(makeReport("svc-a", "svc-b"))
	if got := len(l.Entries()); got != 2 {
		t.Fatalf("expected 2 entries, got %d", got)
	}
}

func TestEntries_NewestFirst(t *testing.T) {
	l := New(10)
	t0 := time.Unix(1000, 0)
	t1 := time.Unix(2000, 0)
	l.clock = func() time.Time { return t0 }
	l.Record(makeReport("first"))
	l.clock = func() time.Time { return t1 }
	l.Record(makeReport("second"))
	entries := l.Entries()
	if entries[0].ServiceName != "second" {
		t.Errorf("expected newest first, got %s", entries[0].ServiceName)
	}
}

func TestCap_EvictsOldest(t *testing.T) {
	l := New(3)
	for i := 0; i < 5; i++ {
		l.Record(makeReport("svc"))
	}
	if got := len(l.Entries()); got != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", got)
	}
}

func TestClear_RemovesAll(t *testing.T) {
	l := New(10)
	l.Record(makeReport("svc-a"))
	l.Clear()
	if got := len(l.Entries()); got != 0 {
		t.Fatalf("expected 0 after clear, got %d", got)
	}
}

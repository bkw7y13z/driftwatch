package rollup_test

import (
	"sync"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
	"github.com/yourorg/driftwatch/internal/rollup"
)

func TestRollup_ConcurrentRecord(t *testing.T) {
	w := rollup.New(time.Minute)

	var wg sync.WaitGroup
	const goroutines = 20
	const reportsEach = 50

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < reportsEach; j++ {
				r := &drift.Report{
					Events: []drift.Event{
						{Status: drift.StatusDrifted},
						{Status: drift.StatusMatch},
					},
				}
				w.Record(r)
			}
		}()
	}
	wg.Wait()

	expected := goroutines * reportsEach
	if w.Len() != expected {
		t.Fatalf("expected %d entries, got %d", expected, w.Len())
	}

	drifted, _, matched := w.Stats()
	if drifted != expected {
		t.Fatalf("expected %d drifted, got %d", expected, drifted)
	}
	if matched != expected {
		t.Fatalf("expected %d matched, got %d", expected, matched)
	}
}

func TestRollup_WindowExpiry_Integration(t *testing.T) {
	w := rollup.New(50 * time.Millisecond)

	w.Record(&drift.Report{
		Events: []drift.Event{{Status: drift.StatusDrifted}},
	})

	if w.Len() != 1 {
		t.Fatal("expected 1 entry before expiry")
	}

	time.Sleep(80 * time.Millisecond)

	w.Record(&drift.Report{
		Events: []drift.Event{{Status: drift.StatusMatch}},
	})

	if w.Len() != 1 {
		t.Fatalf("expected old entry evicted, got %d entries", w.Len())
	}

	d, _, _ := w.Stats()
	if d != 0 {
		t.Fatalf("expected 0 drifted after eviction, got %d", d)
	}
}

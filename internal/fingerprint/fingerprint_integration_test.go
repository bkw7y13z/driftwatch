package fingerprint_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/driftwatch/driftwatch/internal/fingerprint"
)

// TestSum_Concurrent verifies that Sum is safe to call from multiple
// goroutines simultaneously (no data races).
func TestSum_Concurrent(t *testing.T) {
	const workers = 50
	var wg sync.WaitGroup
	results := make([]fingerprint.Result, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = fingerprint.Sum(fmt.Sprintf("content-%d", idx))
		}(i)
	}
	wg.Wait()

	// Each result must be unique (different inputs).
	seen := make(map[string]bool, workers)
	for _, r := range results {
		if seen[r.Hex] {
			t.Fatalf("duplicate fingerprint hex %s", r.Hex)
		}
		seen[r.Hex] = true
	}
}

// TestSum_LargeContent ensures fingerprinting does not panic or produce
// an empty result for large inputs.
func TestSum_LargeContent(t *testing.T) {
	large := make([]byte, 1<<20) // 1 MiB
	for i := range large {
		large[i] = byte(i % 256)
	}
	r := fingerprint.Sum(string(large))
	if r.Hex == "" {
		t.Fatal("expected non-empty hex for large content")
	}
	if r.Algorithm != fingerprint.SHA256 {
		t.Fatalf("unexpected algorithm %q", r.Algorithm)
	}
}

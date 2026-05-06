// Package metrics provides a thread-safe in-process Collector for tracking
// driftwatch daemon activity.
//
// Usage:
//
//	c := metrics.New()
//
//	// after each detection cycle:
//	c.RecordCheck(driftedCount, missingCount, matchedCount)
//
//	// inspect at any time:
//	snap := c.Snapshot()
//	fmt.Println(snap.ChecksTotal)
//
//	// or write a human-readable summary:
//	c.WriteTo(os.Stdout)
//
// The Collector is safe for concurrent use from multiple goroutines.
package metrics

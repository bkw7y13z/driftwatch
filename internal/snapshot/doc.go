// Package snapshot provides persistence for drift detection baselines.
//
// A Snapshot captures the known-good state of a service's configuration
// files at a specific git ref. It is saved to disk as a JSON file so that
// driftwatch can compare the current live state against the last recorded
// baseline across daemon restarts.
//
// Typical usage:
//
//	snap := snapshot.New("my-service")
//	snap.Set("configs/app.yaml", hashOf(content), "HEAD")
//	if err := snap.Save("/var/lib/driftwatch"); err != nil {
//		log.Fatal(err)
//	}
//
//	// On next run:
//	prev, err := snapshot.Load("/var/lib/driftwatch", "my-service")
//	if prev == nil {
//		// No baseline yet; treat everything as new.
//	}
package snapshot

// Package rollup provides a rolling-window aggregator for drift reports.
//
// It is intended for use by dashboards and alerting pipelines that need
// trend data (e.g. "how many drift events occurred in the last 5 minutes")
// rather than point-in-time snapshots.
//
// Usage:
//
//	w := rollup.New(5 * time.Minute)
//	w.Record(report)
//	drifted, missing, matched := w.Stats()
package rollup

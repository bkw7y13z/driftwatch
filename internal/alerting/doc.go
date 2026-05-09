// Package alerting implements consecutive-drift threshold alerting for driftwatch.
//
// An Alerter tracks how many back-to-back check cycles each service has been
// in a drifted or missing state. Once the count reaches the configured
// threshold, the registered AlertFunc is invoked so callers can page on-call
// engineers, post to Slack, or trigger any other escalation mechanism.
//
// Counters are reset to zero as soon as a service reports a clean (match)
// result, preventing stale alerts after a drift is resolved.
//
// # Usage
//
//	alerter := alerting.New(alerting.Config{
//		Threshold: 3,
//		OnAlert:   myAlertFunc,
//	})
//
//	// Call Record after each check cycle:
//	alerter.Record("my-service", alerting.StateDrifted)
//
// The AlertFunc signature is:
//
//	type AlertFunc func(service string, consecutiveCount int)
//
// See [Alerter] and [Config] for full configuration options.
package alerting

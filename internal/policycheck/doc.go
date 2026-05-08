// Package policycheck provides named policy rules that are evaluated against
// drift reports to produce structured violations.
//
// A Policy matches services by a regular expression applied to the service
// name, and declares which drift statuses (Drifted, Missing, etc.) are
// prohibited. Each policy carries a Severity (warn or error) so callers can
// decide how to act on the results.
//
// Usage:
//
//	checker, err := policycheck.New([]policycheck.Policy{
//		{
//			Name:           "no-production-drift",
//			ServicePattern: "^prod-",
//			DenyStatuses:   []drift.Status{drift.StatusDrifted, drift.StatusMissing},
//			Severity:       policycheck.SeverityError,
//		},
//	})
//	violations := checker.Evaluate(report)
//
// The HTTP handler exposed by this package serves violations as JSON and
// returns 204 No Content when all policies pass.
package policycheck

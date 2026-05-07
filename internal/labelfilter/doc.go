// Package labelfilter implements label-selector filtering for driftwatch
// services.
//
// Services may be annotated with arbitrary key=value labels in the
// configuration file. A Filter is constructed from one or more "key=value"
// selector expressions and can be used to restrict drift checks to only those
// services whose labels satisfy all selectors (AND semantics).
//
// Example usage:
//
//	f, err := labelfilter.New([]string{"env=prod", "region=us-east-1"})
//	if err != nil { ... }
//	matched := f.MatchAll(serviceLabels)
package labelfilter

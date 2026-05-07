// Package labelfilter provides tag/label-based filtering so that drift
// checks can be scoped to a subset of services by arbitrary key=value labels.
package labelfilter

import (
	"fmt"
	"strings"
)

// Labels is a map of key/value metadata attached to a service.
type Labels map[string]string

// Selector represents a single key=value match expression.
type Selector struct {
	Key   string
	Value string
}

// Filter holds a set of selectors and matches services whose labels satisfy
// all of them (AND semantics).
type Filter struct {
	selectors []Selector
}

// New parses a slice of "key=value" expressions and returns a Filter.
// Returns an error if any expression is malformed.
func New(exprs []string) (*Filter, error) {
	selectors := make([]Selector, 0, len(exprs))
	for _, expr := range exprs {
		parts := strings.SplitN(expr, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("labelfilter: invalid selector %q: must be key=value", expr)
		}
		selectors = append(selectors, Selector{Key: parts[0], Value: parts[1]})
	}
	return &Filter{selectors: selectors}, nil
}

// Matches reports whether the given labels satisfy every selector in the
// filter. A filter with no selectors matches everything.
func (f *Filter) Matches(labels Labels) bool {
	for _, sel := range f.selectors {
		v, ok := labels[sel.Key]
		if !ok || v != sel.Value {
			return false
		}
	}
	return true
}

// MatchAll returns the subset of the provided label maps (keyed by service
// name) whose labels satisfy the filter.
func (f *Filter) MatchAll(services map[string]Labels) []string {
	matched := make([]string, 0)
	for name, labels := range services {
		if f.Matches(labels) {
			matched = append(matched, name)
		}
	}
	return matched
}

// String returns a human-readable representation of the filter.
func (f *Filter) String() string {
	if len(f.selectors) == 0 {
		return "<match-all>"
	}
	parts := make([]string, len(f.selectors))
	for i, s := range f.selectors {
		parts[i] = s.Key + "=" + s.Value
	}
	return strings.Join(parts, ",")
}

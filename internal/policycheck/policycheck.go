// Package policycheck evaluates drift events against a set of named policies
// and reports which policies are violated by the current report.
package policycheck

import (
	"fmt"
	"regexp"

	"github.com/driftwatch/internal/drift"
)

// Severity represents the importance of a policy violation.
type Severity string

const (
	SeverityWarn  Severity = "warn"
	SeverityError Severity = "error"
)

// Policy defines a rule that matches services by name pattern and declares
// which drift statuses are considered violations.
type Policy struct {
	Name            string
	ServicePattern  string
	DenyStatuses    []drift.Status
	Severity        Severity
	compiledPattern *regexp.Regexp
}

// Violation is produced when a report event breaches a policy.
type Violation struct {
	Policy  string
	Service string
	Status  drift.Status
	Severity Severity
}

// Checker holds a set of compiled policies.
type Checker struct {
	policies []Policy
}

// New compiles the supplied policies and returns a Checker.
// Returns an error if any service pattern is not a valid regular expression.
func New(policies []Policy) (*Checker, error) {
	compiled := make([]Policy, len(policies))
	for i, p := range policies {
		re, err := regexp.Compile(p.ServicePattern)
		if err != nil {
			return nil, fmt.Errorf("policycheck: invalid pattern %q in policy %q: %w", p.ServicePattern, p.Name, err)
		}
		compiled[i] = p
		compiled[i].compiledPattern = re
	}
	return &Checker{policies: compiled}, nil
}

// Evaluate inspects every event in the report and returns all violations.
// A nil report returns an empty slice.
func (c *Checker) Evaluate(r *drift.Report) []Violation {
	if r == nil {
		return nil
	}
	var violations []Violation
	for _, event := range r.Events {
		for _, pol := range c.policies {
			if !pol.compiledPattern.MatchString(event.Service) {
				continue
			}
			for _, denied := range pol.DenyStatuses {
				if event.Status == denied {
					violations = append(violations, Violation{
						Policy:   pol.Name,
						Service:  event.Service,
						Status:   event.Status,
						Severity: pol.Severity,
					})
					break
				}
			}
		}
	}
	return violations
}

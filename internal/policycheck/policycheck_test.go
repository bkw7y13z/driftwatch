package policycheck_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/policycheck"
)

func makeReport(events ...drift.Event) *drift.Report {
	return &drift.Report{Events: events}
}

func TestNew_InvalidPattern(t *testing.T) {
	_, err := policycheck.New([]policycheck.Policy{
		{Name: "bad", ServicePattern: "[", DenyStatuses: []drift.Status{drift.StatusDrifted}},
	})
	if err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
}

func TestEvaluate_NilReport(t *testing.T) {
	c, _ := policycheck.New(nil)
	if v := c.Evaluate(nil); v != nil {
		t.Fatalf("expected nil violations, got %v", v)
	}
}

func TestEvaluate_NoViolations(t *testing.T) {
	c, err := policycheck.New([]policycheck.Policy{
		{Name: "no-drift", ServicePattern: ".*", DenyStatuses: []drift.Status{drift.StatusDrifted}, Severity: policycheck.SeverityError},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := makeReport(drift.Event{Service: "api", Status: drift.StatusMatch})
	if v := c.Evaluate(r); len(v) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(v))
	}
}

func TestEvaluate_ViolationOnDrifted(t *testing.T) {
	c, err := policycheck.New([]policycheck.Policy{
		{Name: "no-drift", ServicePattern: ".*", DenyStatuses: []drift.Status{drift.StatusDrifted}, Severity: policycheck.SeverityError},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := makeReport(drift.Event{Service: "api", Status: drift.StatusDrifted})
	v := c.Evaluate(r)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Policy != "no-drift" || v[0].Service != "api" || v[0].Severity != policycheck.SeverityError {
		t.Errorf("unexpected violation: %+v", v[0])
	}
}

func TestEvaluate_PatternFiltersServices(t *testing.T) {
	c, err := policycheck.New([]policycheck.Policy{
		{Name: "prod-only", ServicePattern: "^prod-", DenyStatuses: []drift.Status{drift.StatusDrifted}, Severity: policycheck.SeverityWarn},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := makeReport(
		drift.Event{Service: "prod-api", Status: drift.StatusDrifted},
		drift.Event{Service: "staging-api", Status: drift.StatusDrifted},
	)
	v := c.Evaluate(r)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Service != "prod-api" {
		t.Errorf("expected prod-api violation, got %s", v[0].Service)
	}
}

func TestEvaluate_MultiplePoliciesMultipleViolations(t *testing.T) {
	c, err := policycheck.New([]policycheck.Policy{
		{Name: "no-drift", ServicePattern: ".*", DenyStatuses: []drift.Status{drift.StatusDrifted}, Severity: policycheck.SeverityWarn},
		{Name: "no-missing", ServicePattern: ".*", DenyStatuses: []drift.Status{drift.StatusMissing}, Severity: policycheck.SeverityError},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := makeReport(
		drift.Event{Service: "svc-a", Status: drift.StatusDrifted},
		drift.Event{Service: "svc-b", Status: drift.StatusMissing},
	)
	v := c.Evaluate(r)
	if len(v) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(v))
	}
}

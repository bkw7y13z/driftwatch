package throttle

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_FirstCallPermitted(t *testing.T) {
	th := New(3, time.Minute)
	if !th.Allow("svc-a") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_BlockedAtLimit(t *testing.T) {
	th := New(2, time.Minute)
	th.Allow("svc")
	th.Allow("svc")
	if th.Allow("svc") {
		t.Fatal("expected third call to be blocked")
	}
}

func TestAllow_DifferentServicesIndependent(t *testing.T) {
	th := New(1, time.Minute)
	th.Allow("svc-a")
	if !th.Allow("svc-b") {
		t.Fatal("svc-b should have its own budget")
	}
}

func TestAllow_PermittedAfterWindowExpires(t *testing.T) {
	base := time.Now()
	th := New(1, time.Second)
	th.now = fixedNow(base)
	th.Allow("svc")

	// advance past the window
	th.now = fixedNow(base.Add(2 * time.Second))
	if !th.Allow("svc") {
		t.Fatal("expected call to be allowed after window expired")
	}
}

func TestRemaining_FullBudgetWhenNoWindow(t *testing.T) {
	th := New(5, time.Minute)
	if got := th.Remaining("svc"); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestRemaining_DecreasesAfterAllow(t *testing.T) {
	th := New(3, time.Minute)
	th.Allow("svc")
	th.Allow("svc")
	if got := th.Remaining("svc"); got != 1 {
		t.Fatalf("expected 1 remaining, got %d", got)
	}
}

func TestRemaining_ZeroWhenExhausted(t *testing.T) {
	th := New(2, time.Minute)
	th.Allow("svc")
	th.Allow("svc")
	th.Allow("svc") // blocked but should not go negative
	if got := th.Remaining("svc"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestReset_RestoresBudget(t *testing.T) {
	th := New(1, time.Minute)
	th.Allow("svc")
	if th.Allow("svc") {
		t.Fatal("should be blocked before reset")
	}
	th.Reset("svc")
	if !th.Allow("svc") {
		t.Fatal("should be allowed after reset")
	}
}

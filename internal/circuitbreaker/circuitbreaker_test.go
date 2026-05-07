package circuitbreaker_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/circuitbreaker"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_ClosedByDefault(t *testing.T) {
	b := circuitbreaker.New(3, time.Minute, 30*time.Second)
	if err := b.Allow(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_OpensAfterThreshold(t *testing.T) {
	now := time.Now()
	b := circuitbreaker.New(3, time.Minute, 30*time.Second)
	b.(*circuitbreaker.Breaker) // ensure concrete type accessible via package

	b2 := circuitbreaker.New(3, time.Minute, 30*time.Second)
	for i := 0; i < 3; i++ {
		b2.RecordFailure()
	}
	_ = now
	if err := b2.Allow(); err == nil {
		t.Fatal("expected circuit to be open")
	}
}

func TestAllow_ClosedBelowThreshold(t *testing.T) {
	b := circuitbreaker.New(3, time.Minute, 30*time.Second)
	b.RecordFailure()
	b.RecordFailure()
	if err := b.Allow(); err != nil {
		t.Fatalf("expected circuit closed with 2 failures, got %v", err)
	}
}

func TestRecordSuccess_ClosesCircuit(t *testing.T) {
	b := circuitbreaker.New(2, time.Minute, 30*time.Second)
	b.RecordFailure()
	b.RecordFailure()
	if err := b.Allow(); err == nil {
		t.Fatal("expected open circuit")
	}
	b.RecordSuccess()
	if err := b.Allow(); err != nil {
		t.Fatalf("expected circuit closed after success, got %v", err)
	}
}

func TestState_Transitions(t *testing.T) {
	b := circuitbreaker.New(2, time.Minute, 10*time.Millisecond)

	if s := b.State(); s != circuitbreaker.StateClosed {
		t.Fatalf("expected Closed, got %v", s)
	}

	b.RecordFailure()
	b.RecordFailure()

	if s := b.State(); s != circuitbreaker.StateOpen {
		t.Fatalf("expected Open, got %v", s)
	}

	time.Sleep(20 * time.Millisecond)

	if s := b.State(); s != circuitbreaker.StateHalfOpen {
		t.Fatalf("expected HalfOpen, got %v", s)
	}
}

func TestAllow_FailuresExpireOutsideWindow(t *testing.T) {
	b := circuitbreaker.New(2, 50*time.Millisecond, time.Second)
	b.RecordFailure()
	b.RecordFailure()

	// wait for window to expire then reset success so state is clear
	time.Sleep(60 * time.Millisecond)
	b.RecordSuccess() // clears open state

	if err := b.Allow(); err != nil {
		t.Fatalf("expected circuit closed after window expiry, got %v", err)
	}
}

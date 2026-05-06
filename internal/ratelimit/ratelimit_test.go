package ratelimit_test

import (
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/ratelimit"
)

func TestAllow_FirstCallPermitted(t *testing.T) {
	l := ratelimit.New(time.Hour)
	if !l.Allow("svc-a") {
		t.Fatal("expected first Allow to return true")
	}
}

func TestAllow_SecondCallBlocked(t *testing.T) {
	l := ratelimit.New(time.Hour)
	l.Allow("svc-a")
	if l.Allow("svc-a") {
		t.Fatal("expected second Allow within cooldown to return false")
	}
}

func TestAllow_DifferentKeysIndependent(t *testing.T) {
	l := ratelimit.New(time.Hour)
	l.Allow("svc-a")
	if !l.Allow("svc-b") {
		t.Fatal("expected different key to be allowed independently")
	}
}

func TestAllow_PermittedAfterCooldown(t *testing.T) {
	l := ratelimit.New(10 * time.Millisecond)
	l.Allow("svc-a")
	time.Sleep(20 * time.Millisecond)
	if !l.Allow("svc-a") {
		t.Fatal("expected Allow to succeed after cooldown elapsed")
	}
}

func TestReset_ClearsKey(t *testing.T) {
	l := ratelimit.New(time.Hour)
	l.Allow("svc-a")
	l.Reset("svc-a")
	if !l.Allow("svc-a") {
		t.Fatal("expected Allow to succeed after Reset")
	}
}

func TestResetAll_ClearsAllKeys(t *testing.T) {
	l := ratelimit.New(time.Hour)
	l.Allow("svc-a")
	l.Allow("svc-b")
	l.ResetAll()
	if !l.Allow("svc-a") || !l.Allow("svc-b") {
		t.Fatal("expected all keys to be cleared after ResetAll")
	}
}

func TestNew_ZeroCooldownDefaultsToMinute(t *testing.T) {
	// Provide a zero duration; the limiter should not panic and should
	// enforce a positive cooldown (verified by blocking a second call).
	l := ratelimit.New(0)
	l.Allow("svc-a")
	if l.Allow("svc-a") {
		t.Fatal("expected second call to be blocked even with zero-duration input")
	}
}

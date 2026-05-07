package debounce

import (
	"testing"
	"time"
)

func TestAllow_FirstCallPermitted(t *testing.T) {
	d := New(5 * time.Second)
	if !d.Allow("svc-a") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_SecondCallSuppressed(t *testing.T) {
	d := New(5 * time.Second)
	d.Allow("svc-a")
	if d.Allow("svc-a") {
		t.Fatal("expected second call within quiet period to be suppressed")
	}
}

func TestAllow_PermittedAfterQuiet(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	d := New(5 * time.Second)
	d.now = func() time.Time { return now }

	d.Allow("svc-a")

	d.now = func() time.Time { return now.Add(6 * time.Second) }
	if !d.Allow("svc-a") {
		t.Fatal("expected call after quiet period to be allowed")
	}
}

func TestAllow_DifferentKeysIndependent(t *testing.T) {
	d := New(5 * time.Second)
	d.Allow("svc-a")
	if !d.Allow("svc-b") {
		t.Fatal("expected different key to be allowed independently")
	}
}

func TestReset_ClearsKey(t *testing.T) {
	d := New(5 * time.Second)
	d.Allow("svc-a")
	d.Reset("svc-a")
	if !d.Allow("svc-a") {
		t.Fatal("expected allow after reset")
	}
}

func TestPurge_RemovesStaleKeys(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	d := New(5 * time.Second)
	d.now = func() time.Time { return now }

	d.Allow("svc-old")
	d.Allow("svc-new")

	// Advance time so svc-old is stale but svc-new was just refreshed.
	d.now = func() time.Time { return now.Add(6 * time.Second) }
	d.Allow("svc-new") // refresh svc-new

	d.now = func() time.Time { return now.Add(7 * time.Second) }
	d.Purge()

	d.mu.Lock()
	_, oldPresent := d.lastSeen["svc-old"]
	_, newPresent := d.lastSeen["svc-new"]
	d.mu.Unlock()

	if oldPresent {
		t.Error("expected stale key to be purged")
	}
	if !newPresent {
		t.Error("expected fresh key to remain after purge")
	}
}

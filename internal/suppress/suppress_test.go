package suppress

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestIsSuppressed_NotPresent(t *testing.T) {
	s := New()
	if s.IsSuppressed("svc", "/etc/app.conf") {
		t.Fatal("expected not suppressed")
	}
}

func TestIsSuppressed_ActiveEntry(t *testing.T) {
	now := time.Now()
	s := &Store{now: fixedClock(now)}
	s.Add("svc", "/etc/app.conf", "maintenance", time.Hour)
	if !s.IsSuppressed("svc", "/etc/app.conf") {
		t.Fatal("expected suppressed")
	}
}

func TestIsSuppressed_ExpiredEntry(t *testing.T) {
	now := time.Now()
	s := &Store{now: fixedClock(now)}
	s.Add("svc", "/etc/app.conf", "old", -time.Minute)
	if s.IsSuppressed("svc", "/etc/app.conf") {
		t.Fatal("expired entry should not suppress")
	}
}

func TestIsSuppressed_DifferentService(t *testing.T) {
	now := time.Now()
	s := &Store{now: fixedClock(now)}
	s.Add("svc-a", "/etc/app.conf", "r", time.Hour)
	if s.IsSuppressed("svc-b", "/etc/app.conf") {
		t.Fatal("different service should not match")
	}
}

func TestPurge_RemovesExpired(t *testing.T) {
	now := time.Now()
	s := &Store{now: fixedClock(now)}
	s.Add("svc", "/a", "r", -time.Minute)
	s.Add("svc", "/b", "r", time.Hour)
	s.Purge()
	if len(s.entries) != 1 {
		t.Fatalf("expected 1 entry after purge, got %d", len(s.entries))
	}
	if s.entries[0].Path != "/b" {
		t.Fatalf("wrong entry retained: %s", s.entries[0].Path)
	}
}

func TestActive_ReturnsOnlyLive(t *testing.T) {
	now := time.Now()
	s := &Store{now: fixedClock(now)}
	s.Add("svc", "/a", "r", -time.Second)
	s.Add("svc", "/b", "r", time.Hour)
	s.Add("svc", "/c", "r", time.Hour)
	active := s.Active()
	if len(active) != 2 {
		t.Fatalf("expected 2 active entries, got %d", len(active))
	}
}

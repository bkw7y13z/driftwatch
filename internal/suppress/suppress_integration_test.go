package suppress_test

import (
	"sync"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/suppress"
)

func TestStore_ConcurrentAddAndCheck(t *testing.T) {
	s := suppress.New()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.Add("svc", "/path", "concurrent", time.Hour)
			_ = s.IsSuppressed("svc", "/path")
		}(i)
	}
	wg.Wait()
	active := s.Active()
	if len(active) != 50 {
		t.Fatalf("expected 50 active entries, got %d", len(active))
	}
}

func TestStore_PurgeUnderConcurrentLoad(t *testing.T) {
	s := suppress.New()
	s.Add("svc", "/a", "r", -time.Millisecond)
	s.Add("svc", "/b", "r", time.Hour)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Purge()
			_ = s.Active()
		}()
	}
	wg.Wait()

	active := s.Active()
	for _, e := range active {
		if e.Path == "/a" {
			t.Fatal("expired entry /a survived concurrent purge")
		}
	}
}

package retrypolicy_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/driftwatch/driftwatch/internal/retrypolicy"
)

func silentLogger(t *testing.T) interface{ any } { return nil } // unused; logger accepts nil

func TestDo_SucceedsOnFirstAttempt(t *testing.T) {
	p := retrypolicy.New(nil)
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesAndEventuallySucceeds(t *testing.T) {
	p := retrypolicy.New(nil)
	p.BaseDelay = time.Millisecond
	calls := 0
	sentinel := errors.New("transient")
	err := p.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return sentinel
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsAttempts(t *testing.T) {
	p := retrypolicy.New(nil)
	p.BaseDelay = time.Millisecond
	p.MaxAttempts = 3
	sentinel := errors.New("persistent")
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	p := retrypolicy.New(nil)
	p.BaseDelay = 10 * time.Second // long delay so cancel is the trigger
	p.MaxAttempts = 5
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := p.Do(ctx, func() error {
		calls++
		if calls == 1 {
			cancel()
		}
		return errors.New("fail")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestIsPermanent_True(t *testing.T) {
	err := retrypolicy.Permanent{Err: errors.New("fatal")}
	if !retrypolicy.IsPermanent(err) {
		t.Fatal("expected IsPermanent to return true")
	}
}

func TestIsPermanent_False(t *testing.T) {
	if retrypolicy.IsPermanent(errors.New("ordinary")) {
		t.Fatal("expected IsPermanent to return false")
	}
}

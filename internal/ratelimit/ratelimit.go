// Package ratelimit provides a simple token-bucket rate limiter for
// controlling how frequently drift notifications are emitted per service.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter tracks per-service notification rate limits.
type Limiter struct {
	mu       sync.Mutex
	buckets  map[string]time.Time
	cooldown time.Duration
}

// New creates a Limiter that enforces the given cooldown between
// successive notifications for the same service key.
func New(cooldown time.Duration) *Limiter {
	if cooldown <= 0 {
		cooldown = time.Minute
	}
	return &Limiter{
		buckets:  make(map[string]time.Time),
		cooldown: cooldown,
	}
}

// Allow reports whether a notification for key is permitted right now.
// If allowed, the internal timestamp for key is updated to now.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if last, ok := l.buckets[key]; ok && now.Sub(last) < l.cooldown {
		return false
	}
	l.buckets[key] = now
	return true
}

// Reset clears the rate-limit state for key, allowing the next call to
// Allow to succeed immediately.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// ResetAll clears all tracked state.
func (l *Limiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buckets = make(map[string]time.Time)
}

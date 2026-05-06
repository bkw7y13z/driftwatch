// Package ratelimit implements a lightweight per-service token-bucket
// rate limiter used by the notification pipeline to suppress repeated
// alerts for the same drifted service within a configurable cooldown
// window.
//
// Usage:
//
//	limiter := ratelimit.New(5 * time.Minute)
//
//	if limiter.Allow(serviceName) {
//		// send notification
//	}
//
// The zero value is not safe to use; always construct via New.
package ratelimit

// Package statuspage provides a lightweight HTTP status page that aggregates
// the most recent drift-check results for every monitored service.
//
// Usage:
//
//	page := statuspage.New()
//
//	// After each drift check, record the result:
//	page.Record(report)
//
//	// Mount on an existing mux:
//	statuspage.Register(mux, page)
//
// The handler responds with JSON containing a "services" array and a
// "generated_at" timestamp.  The HTTP status code is 200 when all services
// are clean and 409 (Conflict) when at least one service has drifted.
package statuspage

// Package healthz provides a lightweight HTTP liveness endpoint for the
// driftwatch daemon.
//
// Usage:
//
//	h := healthz.New(logger)
//	healthz.Register(http.DefaultServeMux, h)
//	go http.ListenAndServe(":8080", nil)
//
// After each drift-check cycle the runner should call h.RecordCheck(drifted)
// so the endpoint can reflect the current state.
//
// GET /healthz returns 503 until the first check completes, then 200.
// The JSON body includes the RFC3339 timestamp of the last check and whether
// drift was observed in that cycle.
package healthz

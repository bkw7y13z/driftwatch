// Package diffrender formats drift.Report values as human-readable unified
// diffs.
//
// The Renderer type writes output to any io.Writer, making it suitable for
// CLI output, structured log fields, or HTTP response bodies. Colour support
// is opt-in and uses ANSI escape codes.
//
// The Handler function exposes a drift diff over HTTP. It returns HTTP 200
// when no drift is detected, HTTP 409 (Conflict) when drift exists, and HTTP
// 204 when no report is available yet. Clients that send Accept: text/x-ansi
// receive ANSI-coloured output.
//
// # HTTP status codes
//
//   - 200 OK           – report present, no drift detected
//   - 204 No Content   – no report available yet (e.g. first sync pending)
//   - 409 Conflict     – report present, drift detected
//
// # Colour output
//
// Colour is never emitted unless explicitly requested. HTTP clients opt in by
// including "text/x-ansi" in their Accept header. CLI callers construct a
// Renderer with ColourEnabled set to true.
package diffrender

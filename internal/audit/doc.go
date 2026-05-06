// Package audit provides an append-only, newline-delimited JSON audit log
// for driftwatch check results.
//
// Each time a drift check completes, the caller records the resulting
// drift.Report via Logger.Record. One JSON line is emitted per event,
// capturing the service name, git ref, file path, drift status, and a
// UTC timestamp. The log file can be tailed, shipped to a SIEM, or
// parsed offline for compliance purposes.
//
// Usage:
//
//	l, err := audit.New("/var/log/driftwatch/audit.log")
//	if err != nil { ... }
//	l.Record(cfg.ServiceName, cfg.GitRef, report)
package audit

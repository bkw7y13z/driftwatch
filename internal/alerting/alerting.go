// Package alerting provides threshold-based alerting for repeated drift events.
// When a service has drifted more than a configured number of consecutive checks,
// an alert is emitted via the registered AlertFunc.
package alerting

import (
	"context"
	"log/slog"
	"sync"

	"github.com/yourorg/driftwatch/internal/drift"
)

// AlertFunc is called when a service exceeds the consecutive-drift threshold.
type AlertFunc func(ctx context.Context, serviceName string, consecutiveDrifts int)

// Alerter tracks consecutive drift counts per service and fires alerts.
type Alerter struct {
	mu        sync.Mutex
	counts    map[string]int
	threshold int
	alertFn   AlertFunc
	log       *slog.Logger
}

// New creates an Alerter that fires alertFn once the number of consecutive
// drifted checks for a service reaches threshold.
func New(threshold int, alertFn AlertFunc, log *slog.Logger) *Alerter {
	if log == nil {
		log = slog.Default()
	}
	return &Alerter{
		counts:    make(map[string]int),
		threshold: threshold,
		alertFn:   alertFn,
		log:       log,
	}
}

// Evaluate inspects a drift report and updates consecutive counters.
// If a service's counter reaches the threshold, alertFn is invoked.
func (a *Alerter) Evaluate(ctx context.Context, report *drift.Report) {
	if report == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	drifted := make(map[string]bool)
	for _, ev := range report.Events {
		if ev.Status != drift.StatusMatch {
			drifted[ev.Service] = true
		}
	}

	for svc, hasDrift := range drifted {
		if hasDrift {
			a.counts[svc]++
			count := a.counts[svc]
			a.log.Warn("consecutive drift detected", "service", svc, "count", count)
			if count >= a.threshold && a.alertFn != nil {
				a.alertFn(ctx, svc, count)
			}
		} else {
			a.counts[svc] = 0
		}
	}
}

// Reset clears the consecutive counter for a specific service.
func (a *Alerter) Reset(service string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.counts, service)
}

// Count returns the current consecutive drift count for a service.
func (a *Alerter) Count(service string) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.counts[service]
}

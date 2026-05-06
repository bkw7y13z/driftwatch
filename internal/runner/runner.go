// Package runner wires together the scheduler, git fetcher, drift detector,
// watcher, snapshot store, metrics recorder, and notifier into a single
// long-running daemon loop.
package runner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/yourorg/driftwatch/internal/config"
	"github.com/yourorg/driftwatch/internal/drift"
	"github.com/yourorg/driftwatch/internal/git"
	"github.com/yourorg/driftwatch/internal/metrics"
	"github.com/yourorg/driftwatch/internal/notify"
	"github.com/yourorg/driftwatch/internal/scheduler"
	"github.com/yourorg/driftwatch/internal/snapshot"
	"github.com/yourorg/driftwatch/internal/watcher"
)

// Runner orchestrates a single drift-check cycle on every scheduler tick.
type Runner struct {
	cfg     *config.Config
	fetcher *git.Fetcher
	watch   *watcher.Watcher
	snap    *snapshot.Store
	met     *metrics.Metrics
	log     *slog.Logger
}

// New validates the configuration and constructs a Runner ready to start.
func New(cfg *config.Config, log *slog.Logger) (*Runner, error) {
	if log == nil {
		log = slog.Default()
	}
	f, err := git.NewFetcher(cfg.RepoPath, log)
	if err != nil {
		return nil, fmt.Errorf("runner: init fetcher: %w", err)
	}
	w := watcher.New(cfg.WatchPaths, log)
	s := snapshot.New(cfg.SnapshotPath)
	m := metrics.New()
	return &Runner{cfg: cfg, fetcher: f, watch: w, snap: s, met: m, log: log}, nil
}

// Start blocks until ctx is cancelled, running a drift check on every tick
// produced by the scheduler.
func (r *Runner) Start(ctx context.Context) error {
	sched := scheduler.New(r.cfg.Interval, r.log)
	return sched.Run(ctx, func(ctx context.Context) error {
		return r.check(ctx)
	})
}

// check performs one full drift-detection cycle.
func (r *Runner) check(ctx context.Context) error {
	det := drift.NewDetector(r.log)

	states, err := r.watch.ReadAll(ctx, r.cfg.WatchFiles)
	if err != nil {
		return fmt.Errorf("runner: read live files: %w", err)
	}
	live := watcher.ToLiveContent(states)

	declared := make(map[string][]byte, len(r.cfg.WatchFiles))
	for _, relPath := range r.cfg.WatchFiles {
		data, err := r.fetcher.ReadFileAtRef(ctx, relPath, r.cfg.GitRef)
		if err != nil {
			r.log.Warn("runner: could not read declared file", "path", relPath, "err", err)
			continue
		}
		declared[relPath] = data
	}

	report := det.CompareAll(declared, live)
	r.met.RecordCheck(report)

	if err := r.snap.Save(ctx, report); err != nil {
		r.log.Warn("runner: snapshot save failed", "err", err)
	}

	n := notify.New(r.cfg, r.log)
	n.Notify(ctx, report)
	return nil
}

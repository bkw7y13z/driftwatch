package runner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/yourorg/driftwatch/internal/config"
	"github.com/yourorg/driftwatch/internal/drift"
	"github.com/yourorg/driftwatch/internal/git"
	"github.com/yourorg/driftwatch/internal/notify"
)

// Runner orchestrates periodic drift detection for all configured services.
type Runner struct {
	cfg      *config.Config
	fetcher  *git.Fetcher
	detector *drift.Detector
	notifier *notify.Notifier
	logger   *slog.Logger
}

// New creates a Runner from the provided config.
func New(cfg *config.Config, logger *slog.Logger) (*Runner, error) {
	fetcher, err := git.NewFetcher(cfg.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("runner: init fetcher: %w", err)
	}

	detector := drift.NewDetector(fetcher, cfg.GitRef)
	notifier := notify.New(cfg, logger)

	return &Runner{
		cfg:      cfg,
		fetcher:  fetcher,
		detector: detector,
		notifier: notifier,
		logger:   logger,
	}, nil
}

// RunOnce performs a single drift-detection pass across all services.
func (r *Runner) RunOnce(ctx context.Context) error {
	r.logger.Info("starting drift check", "ref", r.cfg.GitRef, "services", len(r.cfg.Services))

	report, err := r.detector.CompareAll(ctx, r.cfg.Services)
	if err != nil {
		return fmt.Errorf("runner: compare: %w", err)
	}

	if err := r.notifier.Notify(ctx, report); err != nil {
		return fmt.Errorf("runner: notify: %w", err)
	}

	r.logger.Info("drift check complete", "summary", report.Summary())
	return nil
}

// Start runs drift detection on the configured interval until ctx is cancelled.
func (r *Runner) Start(ctx context.Context) error {
	interval := time.Duration(r.cfg.IntervalSeconds) * time.Second
	r.logger.Info("runner started", "interval", interval)

	if err := r.RunOnce(ctx); err != nil {
		r.logger.Error("drift check failed", "err", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("runner stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := r.RunOnce(ctx); err != nil {
				r.logger.Error("drift check failed", "err", err)
			}
		}
	}
}

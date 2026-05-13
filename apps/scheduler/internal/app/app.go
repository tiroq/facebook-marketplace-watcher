package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/tiroq/fb-market-watcher/scheduler/internal/config"
	"github.com/tiroq/fb-market-watcher/scheduler/internal/policy"
	"github.com/tiroq/fb-market-watcher/scheduler/internal/publisher"
)

// App is the main scheduler application.
type App struct {
	nc  *nats.Conn
	js  jetstream.JetStream
	cfg config.Config
	pub *publisher.Publisher
}

// New creates a new App.
func New(nc *nats.Conn, js jetstream.JetStream, cfg config.Config) *App {
	pub := publisher.New(js, cfg.DryRun)
	return &App{nc: nc, js: js, cfg: cfg, pub: pub}
}

// Run starts the scheduler loop. It blocks until the context is cancelled.
func (a *App) Run(ctx context.Context) error {
	if !a.cfg.Enabled {
		slog.Info("scheduler is disabled, waiting for context cancellation")
		<-ctx.Done()
		return nil
	}

	slog.Info("scheduler started",
		"query", a.cfg.SearchQuery,
		"interval_minutes", a.cfg.IntervalMinutes,
		"jitter_minutes", a.cfg.JitterMinutes,
		"skip_probability", a.cfg.SkipProbability,
		"dry_run", a.cfg.DryRun,
	)

	p := policy.Policy{
		IntervalMinutes: a.cfg.IntervalMinutes,
		JitterMinutes:   a.cfg.JitterMinutes,
		SkipProbability: a.cfg.SkipProbability,
		Timezone:        a.cfg.Timezone,
	}

	// Run once immediately on startup
	a.maybeRunSearch(ctx, p)

	for {
		delay := policy.NextDelay(p.IntervalMinutes, p.JitterMinutes)
		slog.Info("next search scheduled", "delay", delay.Round(time.Second).String())

		select {
		case <-ctx.Done():
			slog.Info("scheduler stopping")
			return nil
		case <-time.After(delay):
			a.maybeRunSearch(ctx, p)
		}
	}
}

func (a *App) maybeRunSearch(ctx context.Context, p policy.Policy) {
	now := time.Now()

	if !policy.ShouldRun(now, p) {
		slog.Info("skipping search run", "reason", "policy skip")
		return
	}

	slog.Info("emitting search task", "query", a.cfg.SearchQuery)

	if err := a.pub.PublishSearchRequested(ctx, a.cfg.SearchQuery, a.cfg.LocationHint, a.cfg.MaxCards, a.cfg.MaxScrolls); err != nil {
		slog.Error("failed to publish search.requested", "error", err)
	}
}

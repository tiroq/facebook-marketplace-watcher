package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tiroq/fb-market-watcher/internal/logx"
	"github.com/tiroq/fb-market-watcher/internal/natsx"
	"github.com/tiroq/fb-market-watcher/scheduler/internal/app"
	"github.com/tiroq/fb-market-watcher/scheduler/internal/config"
)

func main() {
	logx.InitFromEnv()

	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nc, err := natsx.Connect(cfg.NATSUrl)
	if err != nil {
		slog.Error("failed to connect to nats", "error", err)
		os.Exit(1)
	}
	defer nc.Close()
	slog.Info("connected to nats")

	js, err := natsx.EnsureStream(ctx, nc)
	if err != nil {
		slog.Error("failed to ensure nats stream", "error", err)
		os.Exit(1)
	}

	a := app.New(nc, js, cfg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case <-quit:
			slog.Info("shutdown signal received")
			cancel()
		case <-ctx.Done():
		}
	}()

	if err := a.Run(ctx); err != nil {
		slog.Error("scheduler error", "error", err)
		os.Exit(1)
	}

	slog.Info("scheduler stopped")
}

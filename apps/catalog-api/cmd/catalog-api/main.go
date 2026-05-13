package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tiroq/fb-market-watcher/catalog-api/internal/app"
	"github.com/tiroq/fb-market-watcher/catalog-api/internal/config"
	"github.com/tiroq/fb-market-watcher/internal/logx"
	"github.com/tiroq/fb-market-watcher/internal/natsx"
)

func main() {
	logx.InitFromEnv()

	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping postgres", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to postgres")

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

	a := app.New(pool, nc, js, cfg)

	go func() {
		if err := a.StartConsumers(ctx); err != nil {
			slog.Error("consumer error", "error", err)
			cancel()
		}
	}()

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: a.HTTPHandler(),
	}

	go func() {
		slog.Info("starting HTTP server", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server error", "error", err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-quit:
		slog.Info("shutdown signal received")
	case <-ctx.Done():
		slog.Info("context cancelled")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("http server shutdown error", "error", err)
	}

	slog.Info("catalog-api stopped")
}

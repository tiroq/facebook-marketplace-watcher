package app

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/tiroq/fb-market-watcher/catalog-api/internal/config"
	"github.com/tiroq/fb-market-watcher/catalog-api/internal/httpapi"
	"github.com/tiroq/fb-market-watcher/catalog-api/internal/natsconsumers"
	"github.com/tiroq/fb-market-watcher/catalog-api/internal/repository"
	"github.com/tiroq/fb-market-watcher/catalog-api/internal/service"
)

// App is the main application container.
type App struct {
	pool     *pgxpool.Pool
	nc       *nats.Conn
	js       jetstream.JetStream
	cfg      config.Config
	repo     *repository.Repository
	svc      *service.Service
	handler  *httpapi.Handler
	consumer *natsconsumers.Consumer
}

// New creates a new App with all dependencies wired up.
func New(pool *pgxpool.Pool, nc *nats.Conn, js jetstream.JetStream, cfg config.Config) *App {
	repo := repository.New(pool)
	svc := service.New(repo, js)
	handler := httpapi.New(svc, repo)
	consumer := natsconsumers.New(js, svc)

	return &App{
		pool:     pool,
		nc:       nc,
		js:       js,
		cfg:      cfg,
		repo:     repo,
		svc:      svc,
		handler:  handler,
		consumer: consumer,
	}
}

// HTTPHandler returns the HTTP handler for the application.
func (a *App) HTTPHandler() http.Handler {
	return a.handler.Routes()
}

// StartConsumers starts all NATS consumers.
func (a *App) StartConsumers(ctx context.Context) error {
	return a.consumer.Start(ctx)
}

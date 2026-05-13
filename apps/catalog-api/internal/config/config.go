package config

import (
	"github.com/tiroq/fb-market-watcher/internal/config"
)

type Config struct {
	DatabaseURL string
	NATSUrl     string
	HTTPAddr    string
}

func Load() Config {
	return Config{
		DatabaseURL: config.GetString("DATABASE_URL", "postgres://fbwatcher:fbwatcher_secret@localhost:5432/fbwatcher?sslmode=disable"),
		NATSUrl:     config.GetString("NATS_URL", "nats://localhost:4222"),
		HTTPAddr:    config.GetString("HTTP_ADDR", ":8080"),
	}
}

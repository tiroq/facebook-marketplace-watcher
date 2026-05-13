package logx

import (
	"log/slog"
	"os"
)

// Init sets up the default slog logger with JSON output.
func Init(level slog.Level) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}

// InitFromEnv reads LOG_LEVEL from environment and initializes the logger.
func InitFromEnv() {
	level := slog.LevelInfo
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		var l slog.Level
		if err := l.UnmarshalText([]byte(v)); err == nil {
			level = l
		}
	}
	Init(level)
}

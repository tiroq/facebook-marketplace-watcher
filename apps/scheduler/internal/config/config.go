package config

import (
	baseconfig "github.com/tiroq/fb-market-watcher/internal/config"
)

// Config holds scheduler configuration.
type Config struct {
	NATSUrl         string
	Enabled         bool
	DryRun          bool
	SearchQuery     string
	LocationHint    string
	IntervalMinutes int
	JitterMinutes   int
	SkipProbability float64
	MaxCards        int
	MaxScrolls      int
	Timezone        string
}

// Load reads configuration from environment variables.
func Load() Config {
	return Config{
		NATSUrl:         baseconfig.GetString("NATS_URL", "nats://localhost:4222"),
		Enabled:         baseconfig.GetBool("SCHEDULER_ENABLED", true),
		DryRun:          baseconfig.GetBool("SCHEDULER_DRY_RUN", false),
		SearchQuery:     baseconfig.GetString("DEFAULT_SEARCH_QUERY", "laptop"),
		LocationHint:    baseconfig.GetString("DEFAULT_LOCATION_HINT", ""),
		IntervalMinutes: baseconfig.GetInt("DEFAULT_INTERVAL_MINUTES", 40),
		JitterMinutes:   baseconfig.GetInt("DEFAULT_JITTER_MINUTES", 10),
		SkipProbability: baseconfig.GetFloat64("DEFAULT_SKIP_PROBABILITY", 0.5),
		MaxCards:        baseconfig.GetInt("DEFAULT_MAX_CARDS", 50),
		MaxScrolls:      baseconfig.GetInt("DEFAULT_MAX_SCROLLS", 5),
		Timezone:        baseconfig.GetString("DEFAULT_TIMEZONE", "Asia/Bangkok"),
	}
}

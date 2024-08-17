package config

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/ivov/n8n-shortlink/internal/env"
)

// Config holds all configuration for the API.
type Config struct {
	Env  string
	Host string
	Port int
	DB   struct {
		FilePath string
	}
	Log struct {
		FilePath string
	}
	RateLimiter struct {
		Enabled    bool
		RPS        int
		Burst      int
		Inactivity time.Duration
	}
	Sentry struct {
		DSN string
	}
	MetadataMode *bool // whether to display binary metadata and exit
	Build        struct {
		CommitSha string
	}
}

// NewConfig instantiates a new Config from environment variables.
func NewConfig(commitSha string) Config {
	config := Config{}
	config.Build.CommitSha = commitSha

	flag.StringVar(
		&config.Env,
		"environment",
		env.GetStr("N8N_SHORTLINK_ENVIRONMENT", "development"),
		"Environment to run in (development, production, testing)",
	)

	flag.StringVar(
		&config.Host,
		"host",
		env.GetStr("N8N_SHORTLINK_HOST", "localhost"),
		"Host to listen on",
	)

	flag.IntVar(
		&config.Port,
		"port",
		env.GetInt("N8N_SHORTLINK_PORT", 3001),
		"Port to listen on",
	)

	flag.BoolVar(
		&config.RateLimiter.Enabled,
		"rate-limiter-enabled",
		env.GetBool("N8N_SHORTLINK_RATE_LIMITER_ENABLED", true),
		"Whether to enable rate limiter",
	)

	flag.IntVar(
		&config.RateLimiter.RPS,
		"rate-limiter-rps",
		env.GetInt("N8N_SHORTLINK_RATE_LIMITER_RPS", 2),
		"Max requests per second per client allowed by rate limiter",
	)

	flag.IntVar(
		&config.RateLimiter.Burst,
		"rate-limiter-burst",
		env.GetInt("N8N_SHORTLINK_RATE_LIMITER_BURST", 4),
		"Max burst per client allowed by rate limiter",
	)

	flag.DurationVar(
		&config.RateLimiter.Inactivity,
		"rate-limiter-inactivity",
		env.GetDuration("N8N_SHORTLINK_RATE_LIMITER_INACTIVITY", "3m"),
		"Duration after which inactive rate limiter clients are cleared",
	)

	const defaultSentryDSN = "https://f53e747195fcd00533f1f118ce69b44f@o4504685792460800.ingest.us.sentry.io/4507658952638464"

	flag.StringVar(
		&config.Sentry.DSN,
		"sentry-dsn",
		env.GetStr("N8N_SHORTLINK_SENTRY_DSN", defaultSentryDSN),
		"Sentry DSN",
	)

	config.MetadataMode = flag.Bool("metadata-mode", false, "Display binary metadata and exit")

	flag.Parse()

	return config
}

// SetupDotDir creates the .n8n-shortlink dir in the user's home dir.
func SetupDotDir() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	dotDirPath := filepath.Join(home, ".n8n-shortlink")

	if _, err := os.Stat(dotDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dotDirPath, 0755); err != nil {
			panic(err)
		}
	}
}

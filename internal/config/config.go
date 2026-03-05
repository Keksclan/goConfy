package config

import (
	"github.com/keksclan/goConfy/types"
)

// Config represents the canonical reference configuration for goConfy.
// It follows the Config-Policy defined in docs/CONFIG_POLICY.md.
type Config struct {
	// App settings.
	App AppConfig `yaml:"app" desc:"General application settings"`

	// Database settings.
	DB DBConfig `yaml:"db" desc:"Database connection and pool settings"`

	// Redis settings.
	Redis RedisConfig `yaml:"redis" desc:"Redis cache settings"`

	// Features toggle.
	Features FeaturesConfig `yaml:"features" desc:"Feature toggles"`
}

type AppConfig struct {
	// Name is the service name.
	Name string `yaml:"name" default:"goConfyService" desc:"Service name used in logs and traces"`

	// Env is the deployment environment.
	Env string `yaml:"env" default:"development" desc:"Deployment environment" options:"development, staging, production" example:"production"`

	// Port is the HTTP listener port.
	Port int `yaml:"port" default:"8080" desc:"HTTP server port" env:"APP_PORT"`

	// LogLevel sets the verbosity of logs.
	LogLevel string `yaml:"log_level" default:"info" desc:"Log verbosity level" options:"debug, info, warn, error"`
}

type DBConfig struct {
	// Host is the database server address.
	Host string `yaml:"host" default:"localhost" desc:"Database host" env:"DB_HOST"`

	// Port is the database server port.
	Port int `yaml:"port" default:"5432" desc:"Database port" env:"DB_PORT"`

	// User is the database username.
	User string `yaml:"user" default:"postgres" desc:"Database username" env:"DB_USER"`

	// Password is the database password.
	Password string `yaml:"password" secret:"true" desc:"Database password (secret)" env:"DB_PASSWORD"`

	// Name is the database name.
	Name string `yaml:"name" default:"goconfy_db" desc:"Database name"`

	// MaxConns is the maximum number of connections in the pool.
	MaxConns int `yaml:"max_conns" default:"20" desc:"Maximum connection pool size"`

	// Timeout is the connection timeout.
	Timeout types.Duration `yaml:"timeout" default:"5s" desc:"Database connection timeout"`
}

type RedisConfig struct {
	// URL is the Redis connection URL.
	URL string `yaml:"url" default:"redis://localhost:6379" desc:"Redis connection URL" env:"REDIS_URL" example:"redis://:password@host:6379/0"`

	// TTL is the default cache expiration.
	TTL types.Duration `yaml:"ttl" default:"1h" desc:"Default cache time-to-live"`
}

type FeaturesConfig struct {
	// EnableTracing enables distributed tracing.
	EnableTracing bool `yaml:"enable_tracing" default:"false" desc:"Enable distributed tracing (OpenTelemetry)"`

	// EnableMetrics enables Prometheus metrics.
	EnableMetrics bool `yaml:"enable_metrics" default:"true" desc:"Enable Prometheus metrics endpoint"`
}

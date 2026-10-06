// Package observability holds the telemetry setup shared by all
// trust-box services: slog logging, OTel tracing/metrics, and the
// localhost-only ops mux (/metrics, /healthz).
package observability

import "os"

// Configures a service's telemetry. Safe local-dev defaults;
// overridden via environment variables.
type Config struct {
	// Env: SERVICE_NAME (default "trust-box").
	ServiceName string
	// debug, info, warn, error. Env: LOG_LEVEL (default "info").
	Level string
	// "json" or "text". Env: LOG_FORMAT (default "json").
	Format string
	// Binds the ops mux. Localhost-only unless set explicitly.
	// Env: METRICS_ADDR (default "127.0.0.1:9090").
	MetricsAddr string
	// "stdout" or "none". Env: OTEL_TRACES_EXPORTER (default "stdout").
	TracesExporter string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// LoadFromEnv builds a Config from the environment.
func LoadFromEnv() Config {
	return Config{
		ServiceName:    getenv("SERVICE_NAME", "trust-box"),
		Level:          getenv("LOG_LEVEL", "info"),
		Format:         getenv("LOG_FORMAT", "json"),
		MetricsAddr:    getenv("METRICS_ADDR", "127.0.0.1:9090"),
		TracesExporter: getenv("OTEL_TRACES_EXPORTER", "stdout"),
	}
}

package observability

import (
	"context"
	"log/slog"
	"net/http"
	"os"
)

// Setup wires up the standard telemetry stack for a service: logging,
// tracing, and metrics. It returns the configured logger, the ops mux
// (serving /metrics and /healthz), and a cleanup function that shuts
// down tracing and metrics. The caller is responsible for defining any
// service-specific meters on top.
func Setup(ctx context.Context, defaultServiceName string) (*slog.Logger, Config, *http.ServeMux, func(context.Context) error, error) {
	cfg := LoadFromEnv()
	if os.Getenv("SERVICE_NAME") == "" {
		cfg.ServiceName = defaultServiceName
	}

	logger, err := NewLogger(cfg.ServiceName, cfg.Level, cfg.Format)
	if err != nil {
		return nil, cfg, nil, nil, err
	}

	traceShutdown, err := SetupTracing(ctx, cfg.TracesExporter)
	if err != nil {
		return logger, cfg, nil, nil, err
	}

	meterShutdown, err := SetupMetrics()
	if err != nil {
		_ = traceShutdown(context.Background())
		return logger, cfg, nil, nil, err
	}

	cleanup := func(c context.Context) error {
		var firstErr error
		if err := traceShutdown(c); err != nil {
			firstErr = err
		}
		if err := meterShutdown(c); err != nil && firstErr == nil {
			firstErr = err
		}
		return firstErr
	}

	mux := Mux()
	return logger, cfg, mux, cleanup, nil
}

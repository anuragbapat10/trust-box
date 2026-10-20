// Command node-daemon is the trust-box worker: node-local Firecracker
// lifecycle and exec. v1 skeleton: observability wiring plus the ops
// mux only. gRPC handlers land separately.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/anuragbapat10/trust-box.git/internal/observability"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	os.Exit(run(ctx))
}

func run(ctx context.Context) int {
	cfg := observability.LoadFromEnv()
	if os.Getenv("SERVICE_NAME") == "" {
		cfg.ServiceName = "node-daemon"
	}

	logger, err := observability.NewLogger(cfg.ServiceName, cfg.Level, cfg.Format)
	if err != nil {
		slog.Error("invalid logging config", slog.String("error", err.Error()))
		return 1
	}

	traceShutdown, err := observability.SetupTracing(ctx, cfg.TracesExporter)
	if err != nil {
		logger.Error("tracing setup failed", slog.String("error", err.Error()))
		return 1
	}
	defer func() {
		if err := traceShutdown(context.Background()); err != nil {
			logger.Error("tracer shutdown failed", slog.String("error", err.Error()))
		}
	}()

	meterShutdown, err := observability.SetupMetrics()
	if err != nil {
		logger.Error("metrics setup failed", slog.String("error", err.Error()))
		return 1
	}
	defer func() {
		if err := meterShutdown(context.Background()); err != nil {
			logger.Error("meter shutdown failed", slog.String("error", err.Error()))
		}
	}()

	triggerTotal, err := observability.MeterFor(cfg.ServiceName).Int64Counter(
		"trigger_total",
		metric.WithDescription("Manual debug triggers served."),
	)
	if err != nil {
		logger.Error("counter setup failed", slog.String("error", err.Error()))
		return 1
	}

	// TEMPORARY scaffolding for the observability smoke test. Deleted
	// with the first real gRPC handler.
	mux := observability.Mux()
	mux.HandleFunc("POST /triggerdebug", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.TracerFor(cfg.ServiceName).Start(r.Context(), "debug.trigger")
		defer span.End()

		triggerTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("route", "triggerdebug")))
		logger.InfoContext(ctx, "debug trigger served", observability.SpanAttrs(ctx)...)
		w.WriteHeader(http.StatusOK)
	})

	logger.Info("node-daemon starting", slog.String("ops_addr", cfg.MetricsAddr))
	if err := observability.Serve(ctx, cfg.MetricsAddr, mux); err != nil {
		logger.Error("ops server failed", slog.String("error", err.Error()))
		return 1
	}
	return 0
}

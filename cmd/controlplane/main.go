// Command controlplane is the trust-box control plane: instance
// lifecycle, state metadata, and capacity enforcement. Observability
// wiring plus the ops mux; gRPC handlers land separately.
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
	logger, mux, cleanup, err := observability.Setup(ctx, "control-plane")
	if err != nil {
		slog.Error("observability setup failed", slog.String("error", err.Error()))
		return 1
	}
	defer func() {
		if err := cleanup(context.Background()); err != nil {
			logger.Error("observability shutdown failed", slog.String("error", err.Error()))
		}
	}()

	triggerTotal, err := observability.MeterFor("control-plane").Int64Counter(
		"trigger_total",
		metric.WithDescription("Manual debug triggers served."),
	)
	if err != nil {
		logger.Error("counter setup failed", slog.String("error", err.Error()))
		return 1
	}

	// TEMPORARY scaffolding for the observability smoke test. Deleted
	// with the first real gRPC handler.
	mux.HandleFunc("POST /triggerdebug", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.TracerFor("control-plane").Start(r.Context(), "debug.trigger")
		defer span.End()

		triggerTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("route", "triggerdebug")))
		logger.InfoContext(ctx, "debug trigger served", observability.SpanAttrs(ctx)...)
		w.WriteHeader(http.StatusOK)
	})

	cfg := observability.LoadFromEnv()
	logger.Info("control-plane starting", slog.String("ops_addr", cfg.MetricsAddr))
	if err := observability.Serve(ctx, cfg.MetricsAddr, mux); err != nil {
		logger.Error("ops server failed", slog.String("error", err.Error()))
		return 1
	}
	return 0
}
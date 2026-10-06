package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// SetupMetrics registers the global meter provider backed by the
// Prometheus exporter (scraped via /metrics on the ops mux).
func SetupMetrics() (func(context.Context) error, error) {
	exp, err := prometheus.New()
	if err != nil {
		return func(context.Context) error { return nil }, err
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exp))
	otel.SetMeterProvider(mp)
	return mp.Shutdown, nil
}

// MeterFor returns a meter scoped to the named service.
func MeterFor(service string) metric.Meter {
	return otel.Meter(service)
}

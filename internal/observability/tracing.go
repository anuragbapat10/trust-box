package observability

import (
	"context"
	"fmt"
	"log/slog"

	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"

	"go.opentelemetry.io/otel"
)

// SetupTracing registers the global tracer provider and W3C
// propagator. "stdout" prints spans; "none" disables them. Nothing
// leaves the machine either way.
func SetupTracing(ctx context.Context, exporter string) (func(context.Context) error, error) {
	switch exporter {
	case "none", "off", "":
		otel.SetTracerProvider(nooptrace.NewTracerProvider())
		return func(context.Context) error { return nil }, nil
	case "stdout":
		exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
		tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.TraceContext{})
		return tp.Shutdown, nil
	default:
		return nil, fmt.Errorf("invalid OTEL_TRACES_EXPORTER %q: want stdout or none", exporter)
	}
}

// TracerFor returns a tracer scoped to the named service.
func TracerFor(service string) trace.Tracer {
	return otel.Tracer(service)
}

// SpanAttrs extracts trace/span IDs from ctx for log correlation.
func SpanAttrs(ctx context.Context) []any {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return nil
	}
	return []any{
		slog.String("trace_id", sc.TraceID().String()),
		slog.String("span_id", sc.SpanID().String()),
	}
}

package observability

import (
	"fmt"
	"log/slog"
	"os"
)

// A typo in LOG_LEVEL fails fast instead of silently logging
// at the wrong level.
func parseLevel(name string) (slog.Level, error) {
	var l slog.Level
	if err := l.UnmarshalText([]byte(name)); err != nil {
		return slog.LevelInfo, fmt.Errorf("invalid LOG_LEVEL %q: want debug, info, warn or error", name)
	}
	return l, nil
}

// NewLogger builds the service logger on stderr. Every line carries
// the service name.
func NewLogger(service, level, format string) (*slog.Logger, error) {
	lvl, err := parseLevel(level)
	if err != nil {
		return nil, err
	}
	opts := &slog.HandlerOptions{Level: lvl}
	var h slog.Handler
	switch format {
	case "text":
		h = slog.NewTextHandler(os.Stderr, opts)
	case "json", "":
		h = slog.NewJSONHandler(os.Stderr, opts)
	default:
		return nil, fmt.Errorf("invalid LOG_FORMAT %q: want json or text", format)
	}
	return slog.New(h.WithAttrs([]slog.Attr{slog.String("service", service)})), nil
}

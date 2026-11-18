package observability

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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

// NewLogger builds the service logger. By default it writes to
// /tmp/trust-box/<service>.log so logs persist across runs; override
// with LOG_FILE. Every line carries the service name.
func NewLogger(service, level, format string) (*slog.Logger, error) {
	lvl, err := parseLevel(level)
	if err != nil {
		return nil, err
	}
	opts := &slog.HandlerOptions{Level: lvl}

	logPath := os.Getenv("LOG_FILE")
	if logPath == "" {
		logPath = fmt.Sprintf("/tmp/trust-box/%s.log", service)
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file %s: %w", logPath, err)
	}

	var h slog.Handler
	switch format {
	case "text":
		h = slog.NewTextHandler(f, opts)
	case "json", "":
		h = slog.NewJSONHandler(f, opts)
	default:
		return nil, fmt.Errorf("invalid LOG_FORMAT %q: want json or text", format)
	}
	return slog.New(h.WithAttrs([]slog.Attr{slog.String("service", service)})), nil
}

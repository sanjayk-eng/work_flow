package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey struct{}

// New creates a slog.Logger writing JSON to stdout.
// level is read from the LOG_LEVEL env ("debug", "info", "warn", "error").
func New(level string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(level),
	}))
}

// WithContext stores a logger in the context for request-scoped logging.
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext retrieves the logger from context.
// Falls back to the default slog logger if none is set.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(contextKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

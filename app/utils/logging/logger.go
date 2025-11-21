package logging

import (
	"context"
	"log/slog"
	"os"

	"smart-cash/utils"
)

// InitLogger initializes and returns a configured JSON logger
// Always uses JSON format for structured logging
func InitLogger(config *Config, serviceName string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: false, // Set to true if you want file:line in logs
	})

	logger := slog.New(handler).With(
		slog.String("service", serviceName),
	)

	slog.SetDefault(logger)

	return logger
}

// LoggerWithTrace creates a logger that includes trace context from the given context
// Use this when you have a context available to automatically add trace_id and span_id to logs
func LoggerWithTrace(ctx context.Context, logger *slog.Logger) *slog.Logger {
	traceAttrs := utils.GetTraceContext(ctx)
	if traceAttrs == nil {
		return logger
	}

	// Create a new logger with trace context added
	attrs := make([]any, len(traceAttrs))
	for i, attr := range traceAttrs {
		attrs[i] = attr
	}

	return logger.With(attrs...)
}

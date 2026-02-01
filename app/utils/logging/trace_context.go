package logging

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TraceContextFromContext extracts trace ID and span ID from OpenTelemetry context
// Returns trace ID and span ID as strings, or empty strings if not found
func TraceContextFromContext(ctx context.Context) (traceID, spanID string) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return "", ""
	}

	spanContext := span.SpanContext()
	if !spanContext.IsValid() {
		return "", ""
	}

	return spanContext.TraceID().String(), spanContext.SpanID().String()
}

// WithTraceContext adds trace ID and span ID to log attributes if available in context
// Usage: logger.With(WithTraceContext(ctx)...).Info("message")
func WithTraceContext(ctx context.Context) []slog.Attr {
	traceID, spanID := TraceContextFromContext(ctx)

	attrs := []slog.Attr{}
	if traceID != "" {
		attrs = append(attrs, slog.String("trace_id", traceID))
	}
	if spanID != "" {
		attrs = append(attrs, slog.String("span_id", spanID))
	}

	return attrs
}

// LoggerWithTraceContext creates a new logger with trace context from the given context
// This is a convenience function for when you want to create a logger with trace context
func LoggerWithTraceContext(ctx context.Context, baseLogger *slog.Logger) *slog.Logger {
	traceAttrs := WithTraceContext(ctx)
	if len(traceAttrs) == 0 {
		return baseLogger
	}

	args := make([]any, len(traceAttrs))
	for i, attr := range traceAttrs {
		args[i] = attr
	}

	return baseLogger.With(args...)
}

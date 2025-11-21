package utils

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func StartSpan(ctx context.Context, serviceName, spanName string, attrs ...attribute.KeyValue) (context.Context, func()) {
	tr := otel.Tracer(serviceName)
	trContext, span := tr.Start(ctx, spanName, trace.WithAttributes(attrs...))

	return trContext, func() {
		span.End()
	}
}

func StartSpanWithComponent(ctx context.Context, serviceName, spanName, component string, attrs ...attribute.KeyValue) (context.Context, func()) {
	allAttrs := append([]attribute.KeyValue{
		attribute.String("component", component),
	}, attrs...)
	return StartSpan(ctx, serviceName, spanName, allAttrs...)
}

func RecordSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func SetSpanStatus(ctx context.Context, code codes.Code, description string) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetStatus(code, description)
	}
}

func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

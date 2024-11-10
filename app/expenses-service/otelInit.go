package main

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Set up openTelemetry
// 1. define to where the data will be sent (for this case stdout), exporter are defined per signal
// 2. Create the providers of the signals to use

func initOpenTelemetry() *sdktrace.TracerProvider {
	// define exporter for traces
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(os.Getenv("JAEGER_COLLECTOR")+":4318"),
		otlptracehttp.WithInsecure(),
	)

	if err != nil {
		logger.Error("error creating exporter",
			"error", err)
		panic(err)
	}

	// TracerProvider passing exporter created
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultScheduleDelay*time.Millisecond),
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
		),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("expenses-service"),
			),
		),
	)

	return tp
}

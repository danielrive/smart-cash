package utils

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Set up openTelemetry
// 1. define to where the data will be sent (for this case stdout), exporter are defined per signal
// 2. Create the providers of the signals to use

func InitOpenTelemetry(otelUrl string, serviceName string, logger *slog.Logger) *trace.TracerProvider {
	// Creating Resource

	res, err := resource.New(
		context.Background(),
		resource.WithFromEnv(),      // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithTelemetrySDK(), // Discover and provide information about the OpenTelemetry SDK used.
		resource.WithContainer(),    // Discover and provide container information.
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)), // Add custom resource attributes.
	)

	if errors.Is(err, resource.ErrPartialResource) || errors.Is(err, resource.ErrSchemaURLConflict) {
		log.Println(err) // Log non-fatal issues.
	} else if err != nil {
		log.Fatalln(err) // The error may be fatal.
	}

	// define exporter for traces
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(otelUrl+":4318"),
		otlptracehttp.WithInsecure(),
	)

	if err != nil {
		logger.Error("error creating exporter",
			"component", "otel",
			"error", err)
		panic(err)
	}

	// TracerProvider passing exporter created
	tp := trace.NewTracerProvider(
		trace.WithBatcher(
			exporter,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultScheduleDelay*time.Millisecond),
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
		),
		trace.WithResource(res),
	)

	return tp
}

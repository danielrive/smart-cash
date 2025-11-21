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

func InitOpenTelemetry(otelUrl string, serviceName string, logger *slog.Logger) *trace.TracerProvider {
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
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize), // Default: 512 spans
			trace.WithBatchTimeout(trace.DefaultScheduleDelay),            // Default: 5 seconds
		),
		trace.WithResource(res),
	)

	return tp
}

// ShutdownTracerProvider gracefully shuts down the TracerProvider

func ShutdownTracerProvider(ctx context.Context, tp *trace.TracerProvider, logger *slog.Logger) error {

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	logger.Info("shutting down tracer provider",
		"component", "otel")

	if err := tp.Shutdown(shutdownCtx); err != nil {
		logger.Error("error shutting down tracer provider",
			"component", "otel",
			"error", err)
		return err
	}

	logger.Info("tracer provider shut down successfully",
		"component", "otel")
	return nil
}

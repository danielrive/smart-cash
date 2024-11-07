package main

import (
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Set up openTelemetry
// 1. define to where the data will be sent (for this case stdout), exporter are defined per signal
// 2. Create the providers of the signals to use

func initOpenTelemetry() *sdktrace.TracerProvider {
	// define exporter for traces
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint())

	if err != nil {
		logger.Error("error creating exporter",
			"error", err)
		panic(err)
	}

	// TracerProvider passing exporter created
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)

	return tp
}

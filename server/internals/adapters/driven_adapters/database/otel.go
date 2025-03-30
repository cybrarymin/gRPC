package adapters

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up Jaeger exporter
	traceExporter, err := newJaegerTraceExporter(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	// Set up trace provider.
	tracerProvider, err := newTraceProvider(traceExporter)
	if err != nil {
		handleErr(err)
		return
	}

	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	return
}

// Propagator will be used in case you want to send a span from your application to another process or application. we will have it here later we will see samples of it
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// Create an exporter over HTTP for Jaeger endpoint. In latest version, Jaeger supports otlp endpoint
func newJaegerTraceExporter(ctx context.Context) (trace.SpanExporter, error) {
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("localhost:4318"), // Jaeger endpoint
		otlptracehttp.WithInsecure(),                 // us http instead of https
		otlptracehttp.WithTimeout(5*time.Second))

	if err != nil {
		return nil, err
	}

	return traceExporter, nil
}

// To be able to create span
// you need to define a exporter ( stdout , jaeger, prometheus or ....)
// Then with that exporter create a traceProvider
// Using traceProvider to setup the global tracer
// use the tracer to create span
func newTraceProvider(traceExporter trace.SpanExporter) (*trace.TracerProvider, error) {
	// define resource attributes. resource attributes are attrs such as pod name, service name, os, arch and...
	rattr, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName("bankservice")))
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(rattr),
	)
	return traceProvider, nil
}

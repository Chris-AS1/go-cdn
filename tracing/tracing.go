package tracing

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TODO test env. variable OTEL_EXPORTER_OTLP_ENDPOINT
// https://opentelemetry.io/docs/instrumentation/go/manual/
// https://github.com/open-telemetry/opentelemetry-go/blob/main/exporters/otlp/otlptrace/otlptracehttp/example_test.go
// https://github.com/open-telemetry/opentelemetry-go/tree/main/example/otel-collector

const (
	instrumentationName    = "go-cdn"
	instrumentationVersion = "0.1.0"
)

// Retrieves the global Tracer Provider
var Tracer = otel.GetTracerProvider().Tracer(
	instrumentationName,
	trace.WithInstrumentationVersion(instrumentationVersion),
	trace.WithSchemaURL(semconv.SchemaURL),
)

func add(ctx context.Context, x, y int64) int64 {
	var span trace.Span
	_, span = Tracer.Start(ctx, "Addition")
	defer span.End()

	/* span.AddEvent("custom logged event")
	span.SetAttributes(attribute.String("custom attribute", "val")) */

	return x + y
}

func multiply(ctx context.Context, x, y int64) int64 {
	var span trace.Span
	_, span = Tracer.Start(ctx, "Multiplication")
	defer span.End()

	return x * y
}

func newResource() *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("otlptrace-example"),
		semconv.ServiceVersion("0.0.1"),
	)
}

func InstallExportPipeline(ctx context.Context) (func(context.Context) error, error) {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint("localhost:4318")) // TODO test if this gets overwritten

	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(newResource()),
	)

	// Registers a tracer Provider globally.
	otel.SetTracerProvider(tracerProvider)
	return tracerProvider.Shutdown, nil
}

func Example() {
	ctx := context.Background()
	shutdown, err := InstallExportPipeline(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("the answer is", add(ctx, multiply(ctx, multiply(ctx, 2, 2), 10), 2))
}

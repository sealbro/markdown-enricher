package trace

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"markdown-enricher/domain/consts"
	"markdown-enricher/pkg/env"
)

const (
	service = consts.ApplicationName
	id      = 1
)

var (
	environment = env.EnvOrDefault("ENV", "default")
)

type JaegerConfig struct {
	Url string
}

type TracerProvider interface {
	Tracer(name string) trace.Tracer
	Shutdown(ctx context.Context) error
}

type JaegerTracerProvider struct {
	*tracesdk.TracerProvider
}

// MakeTraceProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func MakeTraceProvider(config *JaegerConfig) (TracerProvider, error) {
	if config.Url == "" {
		return &JaegerTracerProvider{}, nil
	}

	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.Url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
			attribute.Int64("ID", id),
		)),
	)

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	return &JaegerTracerProvider{tp}, nil
}

func (p *JaegerTracerProvider) Shutdown(ctx context.Context) error {
	if p.TracerProvider == nil {
		return nil
	}

	return p.TracerProvider.Shutdown(ctx)
}

func (p *JaegerTracerProvider) Tracer(name string) trace.Tracer {
	if p.TracerProvider == nil {
		return stubEmptyTracer
	}

	return p.TracerProvider.Tracer(name)
}

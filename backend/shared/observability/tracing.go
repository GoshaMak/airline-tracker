package observability

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type correlationKey struct{}

const correlationHeader = "x-correlation-id"

func Init(ctx context.Context, service string) (func(context.Context) error, error) {
	ratio := 1.0
	if value := os.Getenv("OTEL_TRACES_SAMPLER_ARG"); value != "" {
		var err error
		ratio, err = strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(ratio) || ratio < 0 || ratio > 1 {
			return nil, fmt.Errorf("OTEL_TRACES_SAMPLER_ARG must be between 0 and 1")
		}
	}
	options := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
		sdktrace.WithResource(resource.NewSchemaless(
			attribute.String("service.name", service),
			attribute.String("service.instance.id", InstanceID()),
		)),
	}
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		exporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithTimeout(2*time.Second),
			otlptracehttp.WithRetry(otlptracehttp.RetryConfig{Enabled: false}),
		)
		if err != nil {
			return nil, fmt.Errorf("configure tracing: %w", err)
		}
		options = append(options, sdktrace.WithBatcher(exporter,
			sdktrace.WithMaxQueueSize(1024),
			sdktrace.WithMaxExportBatchSize(128),
			sdktrace.WithBatchTimeout(time.Second),
			sdktrace.WithExportTimeout(2*time.Second),
		))
	}
	provider := sdktrace.NewTracerProvider(options...)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return func(ctx context.Context) error {
		shutdown, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		return provider.Shutdown(shutdown)
	}, nil
}

func InstanceID() string {
	if id := os.Getenv("INSTANCE_ID"); id != "" {
		return id
	}
	host, _ := os.Hostname()
	return host
}

func Start(ctx context.Context, name string, options ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("airline-tracker").Start(ctx, name, options...)
}

func CorrelationID(ctx context.Context) string {
	id, _ := ctx.Value(correlationKey{}).(string)
	return id
}

func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationKey{}, id)
}

func Inject(ctx context.Context) map[string]string {
	carrier := propagation.MapCarrier{}
	propagation.TraceContext{}.Inject(ctx, carrier)
	if id := CorrelationID(ctx); id != "" {
		carrier[correlationHeader] = id
	}
	return map[string]string(carrier)
}

func Extract(ctx context.Context, carrier map[string]string) context.Context {
	ctx = propagation.TraceContext{}.Extract(ctx, propagation.MapCarrier(carrier))
	if id := carrier[correlationHeader]; id != "" {
		ctx = WithCorrelationID(ctx, id)
	}
	return ctx
}

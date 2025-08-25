package tracing

import (
	"context"
	"fmt"
	"net/http"

	"github.com/VariableSan/go-factory-microservice/pkg/common/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	JaegerURL      string
	Logger         *logger.Logger
}

type TracingManager struct {
	tracer   trace.Tracer
	provider *tracesdk.TracerProvider
	logger   *logger.Logger
}

// NewTracingManager initializes OpenTelemetry with OTLP/HTTP exporter (compatible with Jaeger)
func NewTracingManager(config Config) (*TracingManager, error) {
	// Create OTLP HTTP exporter (works with Jaeger v1.35+)
	exp, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(config.JaegerURL),
		otlptracehttp.WithInsecure(), // Use HTTP instead of HTTPS
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(res),
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global text map propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer := tp.Tracer(config.ServiceName)

	return &TracingManager{
		tracer:   tracer,
		provider: tp,
		logger:   config.Logger.WithComponent("tracing"),
	}, nil
}

// StartSpan starts a new span with the given name
func (tm *TracingManager) StartSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return tm.tracer.Start(ctx, spanName)
}

// GetTracer returns the tracer instance
func (tm *TracingManager) GetTracer() trace.Tracer {
	return tm.tracer
}

// Shutdown gracefully shuts down the tracer provider
func (tm *TracingManager) Shutdown(ctx context.Context) error {
	if tm.provider != nil {
		return tm.provider.Shutdown(ctx)
	}
	return nil
}

// NoOpTracingManager creates a no-op tracing manager for cases where Jaeger is not available
func NoOpTracingManager(logger *logger.Logger) *TracingManager {
	return &TracingManager{
		tracer: otel.Tracer("noop"),
		logger: logger.WithComponent("tracing-noop"),
	}
}

// HTTPMiddleware creates HTTP middleware for tracing
func (tm *TracingManager) HTTPMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tm.StartSpan(r.Context(), fmt.Sprintf("%s %s", r.Method, r.URL.Path))
			defer span.End()

			// Add trace context to request
			r = r.WithContext(ctx)

			// Set span attributes
			span.SetAttributes(
				semconv.HTTPMethodKey.String(r.Method),
				semconv.HTTPURLKey.String(r.URL.String()),
				semconv.HTTPSchemeKey.String(r.URL.Scheme),
				semconv.NetHostNameKey.String(r.Host),
				semconv.HTTPTargetKey.String(r.URL.Path),
			)

			next.ServeHTTP(w, r)
		})
	}
}

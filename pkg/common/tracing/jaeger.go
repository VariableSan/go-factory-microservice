package tracing

import (
	"context"
	"io"
	"time"

	"github.com/opentracing/opentracing-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	jaegerLog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-client-go/zipkin"
)

// InitJaeger initializes Jaeger tracer
func InitJaeger(serviceName, jaegerURL string) (opentracing.Tracer, io.Closer, error) {
	cfg := jaegerConfig.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  "const",
			Param: 1, // sample all traces
		},
		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  "localhost:6831",
		},
	}

	if jaegerURL != "" {
		cfg.Reporter.CollectorEndpoint = jaegerURL + "/api/traces"
	}

	// Initialize tracer with Zipkin B3 propagation
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	
	tracer, closer, err := cfg.NewTracer(
		jaegerConfig.Logger(jaegerLog.StdLogger),
		jaegerConfig.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaegerConfig.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
	)
	
	if err != nil {
		return nil, nil, err
	}

	opentracing.SetGlobalTracer(tracer)
	return tracer, closer, nil
}

// StartSpan starts a new tracing span
func StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return opentracing.StartSpan(operationName, opts...)
}

// SpanFromContext returns span from context
func SpanFromContext(ctx context.Context) opentracing.Span {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		return span
	}
	return opentracing.NoopTracer{}.StartSpan("noop")
}

package trace

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	Service             string `envconfig:"TRACE_SERVICE"`
	Provider            string `envconfig:"TRACE_PROVIDER" default:"cloud_trace"`
	Environment         string `envconfig:"TRACE_ENVIRONMENT" default:"local"`
	JaegerEndpoint      string `envconfig:"TRACE_JAEGER_ENDPOINT" default:"http://localhost:14268/api/traces"`
	CloudTraceProjectID string `envconfig:"GCP_PROJECT_ID" default:"isu13-406204"`
}

const (
	ProviderJaeger     = "jaeger"
	ProviderCloudTrace = "cloud_trace"
)

// TraceIDFromHeader traceparentからtrace idを取り出す
func TraceIDFromHeader(header http.Header) string {
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.HeaderCarrier(header))
	return trace.SpanFromContext(ctx).SpanContext().TraceID().String()
}

// SpanFromRemote traceparentヘッダからtrace id, span idを取り出して分散trace/spanを作成する.
func SpanFromRemote(ctx context.Context, header http.Header) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
}

// StartSpan 新しいspanを作成する.
func StartSpan(ctx context.Context, name string, attributes ...attribute.KeyValue) context.Context {
	tr := otel.GetTracerProvider().Tracer(name)
	cctx, span := tr.Start(ctx, fmt.Sprintf("span-%s", name))
	if len(attributes) > 0 {
		span.SetAttributes(attributes...)
	}
	return cctx
}

// EndSpan spanの更新を終了する.
func EndSpan(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// default span attributes
var (
	environment = "local"
	version     = "devel"
)

func InitProvider(conf Config) (shutdownFn, error) {
	build, ok := debug.ReadBuildInfo()
	if ok {
		version = build.Main.Version
	}
	if conf.Environment != "" {
		environment = conf.Environment
	}
	service := ""
	if conf.Service != "" {
		service = conf.Service
	} else {
		service = getDefaultServicename()
	}

	// globalで使用するpropagation specを設定する
	otel.SetTextMapPropagator(propagation.TraceContext{})

	switch conf.Provider {
	case ProviderJaeger:
		return initJaeger(conf.JaegerEndpoint, service)
	case ProviderCloudTrace:
		return initCloudTrace(conf.CloudTraceProjectID, service)
	}
	return nopShutdown, nil
}

func getDefaultServicename() string {
	rev := os.Getenv("K_REVISION")
	if rev == "" {
		return "localhost"
	}
	return rev
}

type shutdownFn func(context.Context) error

var _ shutdownFn = nopShutdown

func nopShutdown(context.Context) error { return nil }

func initJaeger(uri, service string) (shutdownFn, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(uri)))
	if err != nil {
		return nopShutdown, nil
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
			attribute.String("version", version),
		)),
	)
	otel.SetTracerProvider(provider)
	return provider.ForceFlush, nil
}

func initCloudTrace(projectID, service string) (shutdownFn, error) {
	exp, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		return nopShutdown, err
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
			attribute.String("version", version),
		)),
	)
	otel.SetTracerProvider(provider)
	return provider.ForceFlush, nil
}

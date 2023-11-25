package trace

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/xerrors"
)

// tp 任意のexporterごとにTraceProviderを上書きする.
var tp = trace.NewNoopTracerProvider()

const gcpTraceHeader = "X-Cloud-Trace-Context"

// SpanFromRemote X-Cloud-Trace-Context からtrace id, span idを取り出して新規SpanContextに伝播させる
func SpanFromRemote(ctx context.Context, header http.Header) context.Context {
	h := header.Get(gcpTraceHeader)
	if h == "" {
		return ctx
	}
	sc, err := extract(h)
	if err != nil {
		return ctx
	}
	if sc.IsValid() {
		return trace.ContextWithRemoteSpanContext(ctx, sc)
	}
	return ctx
}

func extract(h string) (trace.SpanContext, error) {
	sc := trace.SpanContext{}

	// parse trace id
	trIdx := strings.Index(h, "/")
	trHex := h[:trIdx]
	traceID, err := trace.TraceIDFromHex(trHex)
	if err != nil {
		return sc, xerrors.Errorf("trace.TraceIDFromHex %s: %w", trHex, err)
	}
	sc = sc.WithTraceID(traceID)

	// parse span id
	spIdx := strings.Index(h, ";")
	spanRaw := h[trIdx+1 : spIdx]
	sid, err := strconv.ParseUint(spanRaw, 10, 64)
	if err != nil {
		return sc, fmt.Errorf("failed to parse value: %w", err)
	}
	spanID := sc.SpanID()
	binary.BigEndian.PutUint64(spanID[:], sid)
	sc = sc.WithSpanID(spanID)

	sc.WithTraceFlags(trace.FlagsSampled)
	return sc, nil
}

// Fix: attributesをlib側で定義したinterfaceを使うようにする
func StartSpan(ctx context.Context, name string, attributes ...attribute.KeyValue) context.Context {
	tr := tp.Tracer(name)
	cctx, span := tr.Start(ctx, name)
	if len(attributes) > 0 {
		span.SetAttributes(attributes...)
	}
	return cctx
}

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
	service     = "isu13"
	environment = "local"
	version     = "devel"
)

const (
	ProviderJaeger     = "jaeger"
	ProviderCloudTrace = "cloud_trace"
)

const project = "isu13-406204"

func InitProvider() (shutdown func(), _ error) {
	return initCloudTrace(project)
}

func initJaeger(uri string) (shutdown func(), _ error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(uri)))
	if err != nil {
		return nopShutdown, nil
	}

	tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
			attribute.String("version", version),
		)),
	)
	return nopShutdown, nil
}

func nopShutdown() {}

func initCloudTrace(projectID string) (shutdown func(), _ error) {
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

	tp = provider
	return shutdown, nil
}

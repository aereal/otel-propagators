package datadog

import (
	"context"
	"fmt"
	"strconv"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	datadogIDBase = 16

	keyTraceID  = "x-datadog-trace-id"
	keyParentID = "x-datadog-parent-id"
	keyPriority = "x-datadog-sampling-priority"

	samplingNo  = "0"
	samplingYes = "1"

	nilContext = trace.SpanContext{}
)

// Propagator serializes Span Context to/from Datadog APM traces headers.
type Propagator struct{}

var _ propagation.TextMapPropagator = &Propagator{}

// Inject injects a context to the carreir following Datadog APM traces format.
func (Propagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanFromContext(ctx).SpanContext()
	traceID := sc.TraceID()
	spanID := sc.SpanID()
	if !(traceID.IsValid() && spanID.IsValid()) {
		return
	}
	isSampled := samplingNo
	if sc.IsSampled() {
		isSampled = samplingYes
	}
	carrier.Set(keyPriority, isSampled)
	carrier.Set(keyTraceID, ToDatadogID(traceID.String()))
	carrier.Set(keyParentID, ToDatadogID(spanID.String()))
}

// Extract gets a context from the carrier if it contains Datadog APM traces headers.
func (Propagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	sc, err := extractSpanContext(
		carrier.Get(keyTraceID),
		carrier.Get(keyParentID),
		carrier.Get(keyPriority),
	)
	if err != nil || !sc.IsValid() {
		return ctx
	}
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

// Fields returns a list of fields used by HTTPTextFormat.
func (Propagator) Fields() []string {
	return []string{keyTraceID, keyParentID, keyPriority}
}

// ToDatadogID converts the ID that comes from OpenTelemetry SDK to Datadog ID.
func ToDatadogID(otelID string) string {
	if len(otelID) < datadogIDBase {
		return ""
	}
	if len(otelID) > datadogIDBase {
		otelID = otelID[datadogIDBase:]
	}
	iv, err := strconv.ParseUint(otelID, datadogIDBase, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatUint(iv, 10)
}

// ToOpenTelemetryID converts the ID that comes from Datadog APM traces to OpenTelemetry ID.
//
// It may fail if the ID has invalid format.
func ToOpenTelemetryID(ddID string) (string, error) {
	uv, err := strconv.ParseUint(ddID, 10, 64)
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(uv, datadogIDBase), nil
}

func extractSpanContext(traceID, spanID, sampled string) (trace.SpanContext, error) {
	cfg := trace.SpanContextConfig{}

	parsedTraceID, err := ToOpenTelemetryID(traceID)
	if err != nil {
		return nilContext, fmt.Errorf("invalid trace ID: %w", err)
	}
	if len(parsedTraceID) < 32 {
		parsedTraceID = fmt.Sprintf("%032s", parsedTraceID)
	}
	cfg.TraceID, err = trace.TraceIDFromHex(parsedTraceID)
	if err != nil {
		return nilContext, fmt.Errorf("invalid trace ID: %w", err)
	}

	parsedSpanID, err := ToOpenTelemetryID(spanID)
	if err != nil {
		return nilContext, fmt.Errorf("invalid span ID: %w", err)
	}
	if len(parsedSpanID) < 16 {
		parsedSpanID = fmt.Sprintf("%016s", parsedSpanID)
	}
	cfg.SpanID, err = trace.SpanIDFromHex(parsedSpanID)
	if err != nil {
		return nilContext, fmt.Errorf("invalid span ID: %w", err)
	}

	cfg.TraceFlags = cfg.TraceFlags.WithSampled(sampled == samplingYes)

	return trace.NewSpanContext(cfg), nil
}

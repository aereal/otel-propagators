package datadog_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/aereal/otel-propagators/datadog"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	p = datadog.Propagator{}

	emptySpanContextJSON []byte

	ddTraceID   = "9530669991610245"
	otelTraceID = trace.TraceID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x21, 0xdc, 0x18, 0x7, 0x52, 0x47, 0x85}

	ddSpanID   = "9455715668862222"
	otelSpanID = trace.SpanID{0x0, 0x21, 0x97, 0xec, 0x5d, 0x8a, 0x25, 0xe}

	headerTraceID  = http.CanonicalHeaderKey("x-datadog-trace-id")
	headerParentID = http.CanonicalHeaderKey("x-datadog-parent-id")
	headerPriority = http.CanonicalHeaderKey("x-datadog-sampling-priority")

	validTestCases = []struct {
		name        string
		header      http.Header
		spanContext trace.SpanContext
	}{
		{
			"not sampled",
			http.Header{
				headerTraceID:  []string{ddTraceID},
				headerParentID: []string{ddSpanID},
				headerPriority: []string{"0"},
			},
			trace.NewSpanContext(trace.SpanContextConfig{
				TraceID: otelTraceID,
				SpanID:  otelSpanID,
				Remote:  true,
			}),
		},
		{
			"sampled",
			http.Header{
				headerTraceID:  []string{ddTraceID},
				headerParentID: []string{ddSpanID},
				headerPriority: []string{"1"},
			},
			trace.NewSpanContext(trace.SpanContextConfig{
				TraceID:    otelTraceID,
				SpanID:     otelSpanID,
				Remote:     true,
				TraceFlags: trace.FlagsSampled,
			}),
		},
	}
)

func init() {
	var err error
	emptySpanContextJSON, err = json.MarshalIndent(trace.SpanContext{}, "", "  ")
	if err != nil {
		panic(err)
	}
}

func TestPropagator_Inject_valid(t *testing.T) {
	for _, tc := range validTestCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := trace.ContextWithRemoteSpanContext(context.Background(), tc.spanContext)
			h := http.Header{}
			p.Inject(ctx, propagation.HeaderCarrier(h))
			if !reflect.DeepEqual(h, tc.header) {
				t.Errorf("mismatch:\n\tgot: %#v\n\twant: %#v", h, tc.header)
			}
		})
	}
}

func TestPropagator_Extract_valid(t *testing.T) {
	for _, tc := range validTestCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := p.Extract(context.Background(), propagation.HeaderCarrier(tc.header))
			got := trace.SpanContextFromContext(ctx)
			gotJSON, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Fatalf("failed to marshal got span context: %+v", err)
			}
			wantJSON, err := json.MarshalIndent(tc.spanContext, "", "  ")
			if err != nil {
				t.Fatalf("failed to marshal want span context: %+v", err)
			}
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("mismatch:\n\tgot: %s\n\twant: %s", string(gotJSON), string(wantJSON))
			}
		})
	}
}

func TestPropagator_Extract_invalid(t *testing.T) {
	testCases := []struct {
		name   string
		header http.Header
	}{
		{
			"malformed trace ID",
			http.Header{
				headerTraceID:  []string{"a"},
				headerParentID: []string{ddSpanID},
				headerPriority: []string{"0"},
			},
		},
		{
			"malformed span ID",
			http.Header{
				headerTraceID:  []string{ddTraceID},
				headerParentID: []string{"a"},
				headerPriority: []string{"0"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := p.Extract(context.Background(), propagation.HeaderCarrier(tc.header))
			got := trace.SpanContextFromContext(ctx)
			gotJSON, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Fatalf("failed to marshal got span context: %+v", err)
			}
			if string(gotJSON) != string(emptySpanContextJSON) {
				t.Errorf("mismatch:\n\tgot: %s\n\twant: %s", string(gotJSON), string(emptySpanContextJSON))
			}
		})
	}
}

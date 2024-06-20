package carrier_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/aereal/otel-propagators/carrier"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	traceID trace.TraceID
	spanID  trace.SpanID
)

func init() {
	var err error
	traceID, err = trace.TraceIDFromHex(traceIDStr)
	if err != nil {
		panic(err)
	}
	spanID, err = trace.SpanIDFromHex(spanIDStr)
	if err != nil {
		panic(err)
	}
}

const (
	traceIDStr = "4bf92f3577b34da6a3ce929d0e0e4736"
	spanIDStr  = "00f067aa0ba902b7"
)

func TestURLCarrier(t *testing.T) {
	var prop propagation.TraceContext
	t.Run("initial state", func(t *testing.T) {
		testURL, err := url.Parse("http://test.example/")
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		sc := trace.SpanContextFromContext(prop.Extract(ctx, carrier.NewURLCarrier(testURL)))
		want := trace.NewSpanContext(trace.SpanContextConfig{})
		if !sc.Equal(want) {
			t.Errorf("extracted span context mismatch:\n\twant: %#v\n\tgot: %#v", want, sc)
		}
	})
	t.Run("injected/not sampled", func(t *testing.T) {
		testURL, err := url.Parse("http://test.example/?traceparent=00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00")
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		prop.Inject(ctx, carrier.NewURLCarrier(testURL))
		sc := trace.SpanContextFromContext(prop.Extract(ctx, carrier.NewURLCarrier(testURL)))
		want := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
			Remote:  true,
		})
		if !sc.Equal(want) {
			t.Errorf("extracted span context mismatch:\n\twant: %#v\n\tgot: %#v", want, sc)
		}
	})
	t.Run("injected/sampled", func(t *testing.T) {
		testURL, err := url.Parse("http://test.example/?traceparent=00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		prop.Inject(ctx, carrier.NewURLCarrier(testURL))
		sc := trace.SpanContextFromContext(prop.Extract(ctx, carrier.NewURLCarrier(testURL)))
		want := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    traceID,
			SpanID:     spanID,
			Remote:     true,
			TraceFlags: trace.FlagsSampled,
		})
		if !sc.Equal(want) {
			t.Errorf("extracted span context mismatch:\n\twant: %#v\n\tgot: %#v", want, sc)
		}
	})
}

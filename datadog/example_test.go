package datadog_test

import (
	"github.com/aereal/otel-propagators/datadog"
	"go.opentelemetry.io/otel"
)

func ExamplePropagator() {
	otel.SetTextMapPropagator(datadog.Propagator{})
}

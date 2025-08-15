package metrics

import (
	"context"
	"go.opentelemetry.io/otel/metric"
)

type ExtendMeter struct {
	ctx           context.Context
	embeddedMeter metric.Meter
}

func NewExtendMeter(ctx context.Context, serviceName string, provider metric.MeterProvider) *ExtendMeter {
	return &ExtendMeter{
		ctx:           ctx,
		embeddedMeter: provider.Meter(serviceName),
	}
}

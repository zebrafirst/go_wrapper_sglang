package metrics

import (
	"context"
	"errors"
	"git.iflytek.com/AIaaS/otlp-self/v3/zaplog"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
)

const (
	MILLISECOND = "millisecond"
	SECOND      = "second"
)

// 单位
const (
	DEFAULT_COUNTER_UNIT string = "1"
)

type ExtendLables struct {
	Key string
	Val string
}

// 自定义  counter
type ExtendInt64Counter struct {
	ctx          context.Context
	int64Counter api.Int64Counter
}

func (m *ExtendMeter) NewExtendInt64Counter(metricName, metricDesc, metricUnit string) (*ExtendInt64Counter, error) {
	if m == nil || m.embeddedMeter == nil {
		return nil, errors.New("extendMeter is nil or embeddedMeter is nil")
	}
	int64Counter, err := m.embeddedMeter.Int64Counter(metricName,
		api.WithDescription(metricDesc),
		api.WithUnit(metricUnit))
	if err != nil {
		zaplog.SDKLogger.Errorf("ExtendInt64Counter with metricName: %v err %v", metricName, err)
		return nil, err
	}

	return &ExtendInt64Counter{
		ctx:          m.ctx,
		int64Counter: int64Counter,
	}, nil
}

func (e *ExtendInt64Counter) Record(labels []*ExtendLables, value int64) error {
	if e == nil || e.int64Counter == nil {
		return errors.New("metrics can not be empty")
	}

	if labels == nil {
		return errors.New("metrics can not with empty labels")
	}

	attr := make([]attribute.KeyValue, 0, len(labels))
	for _, v := range labels {
		attr = append(attr, attribute.KeyValue{
			Key:   attribute.Key(v.Key),
			Value: attribute.StringValue(v.Val),
		})
	}

	opt := api.WithAttributeSet(
		attribute.NewSet(attr...),
	)

	e.int64Counter.Add(e.ctx, value, opt)
	return nil
}

type ExtendFloat64Counter struct {
	ctx            context.Context
	float64Counter api.Float64Counter
}

func (m *ExtendMeter) NewExtendFloat64Counter(metricName, metricDesc, metricUnit string) (*ExtendFloat64Counter, error) {
	if m == nil || m.embeddedMeter == nil {
		return nil, errors.New("extendMeter is nil or embeddedMeter is nil")
	}
	float64Counter, err := m.embeddedMeter.Float64Counter(metricName,
		api.WithDescription(metricDesc),
		api.WithUnit(metricUnit))

	if err != nil {
		zaplog.SDKLogger.Errorf("ExtendInt64Counter with metricName: %v err %v", metricName, err)
		return nil, err
	}
	return &ExtendFloat64Counter{
		ctx:            m.ctx,
		float64Counter: float64Counter,
	}, nil
}

func (e *ExtendFloat64Counter) Record(labels []*ExtendLables, value float64) error {
	if e == nil || e.float64Counter == nil {
		return errors.New("metrics can not be empty")
	}

	if labels == nil {
		return errors.New("metrics can not with empty labels")
	}
	attr := make([]attribute.KeyValue, 0, len(labels))
	for _, v := range labels {
		attr = append(attr, attribute.KeyValue{
			Key:   attribute.Key(v.Key),
			Value: attribute.StringValue(v.Val),
		})
	}

	opt := api.WithAttributeSet(
		attribute.NewSet(attr...),
	)

	e.float64Counter.Add(e.ctx, value, opt)
	return nil
}

type ExtendInt64UpDownCounter struct {
	ctx                context.Context
	int64UpDownCounter api.Int64UpDownCounter
}

func (m *ExtendMeter) NewExtendInt64UpDownCounter(metricName, metricDesc, metricUnit string) (*ExtendInt64UpDownCounter, error) {
	if m == nil || m.embeddedMeter == nil {
		return nil, errors.New("extendMeter is nil or embeddedMeter is nil")
	}
	int64UpDownCounter, err := m.embeddedMeter.Int64UpDownCounter(metricName,
		api.WithDescription(metricDesc),
		api.WithUnit(metricUnit))

	if err != nil {
		zaplog.SDKLogger.Errorf("ExtendInt64UpDownCounter with metricName: %v err %v", metricName, err)
		return nil, err
	}
	return &ExtendInt64UpDownCounter{
		ctx:                m.ctx,
		int64UpDownCounter: int64UpDownCounter,
	}, nil
}

func (e *ExtendInt64UpDownCounter) Record(labels []*ExtendLables, value int64) error {
	if e == nil || e.int64UpDownCounter == nil {
		return errors.New("metrics can not be empty")
	}

	if labels == nil {
		return errors.New("metrics can not with empty labels")
	}

	attr := make([]attribute.KeyValue, 0, len(labels))
	for _, v := range labels {
		attr = append(attr, attribute.KeyValue{
			Key:   attribute.Key(v.Key),
			Value: attribute.StringValue(v.Val),
		})
	}
	opt := api.WithAttributeSet(
		attribute.NewSet(attr...),
	)

	e.int64UpDownCounter.Add(e.ctx, value, opt)
	return nil
}

type ExtendFloat64UpDownCounter struct {
	ctx                  context.Context
	float64UpDownCounter api.Float64UpDownCounter
}

func (m *ExtendMeter) NewExtendFloat64UpDownCounter(metricName, metricDesc, metricUnit string) (*ExtendFloat64UpDownCounter, error) {
	if m == nil || m.embeddedMeter == nil {
		return nil, errors.New("extendMeter is nil or embeddedMeter is nil")
	}
	float64UpDownCounter, err := m.embeddedMeter.Float64UpDownCounter(metricName,
		api.WithDescription(metricDesc),
		api.WithUnit(metricUnit))

	if err != nil {
		zaplog.SDKLogger.Errorf("ExtendFloat64UpDownCounter with metricName: %v err %v", metricName, err)
		return nil, err
	}
	return &ExtendFloat64UpDownCounter{
		ctx:                  m.ctx,
		float64UpDownCounter: float64UpDownCounter,
	}, nil
}

func (e *ExtendFloat64UpDownCounter) Record(labels []*ExtendLables, value float64) error {
	if e == nil || e.float64UpDownCounter == nil {
		return errors.New("metrics can not be empty")
	}
	if labels == nil {
		return errors.New("metrics can not with empty labels")
	}
	attr := make([]attribute.KeyValue, 0, len(labels))
	for _, v := range labels {
		attr = append(attr, attribute.KeyValue{
			Key:   attribute.Key(v.Key),
			Value: attribute.StringValue(v.Val),
		})
	}
	opt := api.WithAttributeSet(
		attribute.NewSet(attr...),
	)

	e.float64UpDownCounter.Add(e.ctx, value, opt)
	return nil
}

type ExtendInt64Histogram struct {
	ctx            context.Context
	int64Histogram api.Int64Histogram
}

func (m *ExtendMeter) NewExtendInt64Histogram(metricName, metricDesc, metricUnit string, bucket []float64) (*ExtendInt64Histogram, error) {
	if m == nil || m.embeddedMeter == nil {
		return nil, errors.New("extendMeter is nil or embeddedMeter is nil")
	}
	if bucket == nil {
		bucket = []float64{0.0, 5.0, 10.0, 25.0, 50.0, 75.0, 100.0, 250.0, 500.0, 750.0, 1000.0, 2500.0, 5000.0, 7500.0, 10000.0, 200000.0}
	}
	histogram, err := m.embeddedMeter.Int64Histogram(metricName,
		api.WithDescription(metricDesc),
		api.WithUnit(metricUnit),
		api.WithExplicitBucketBoundaries(bucket...))

	if err != nil {
		zaplog.SDKLogger.Errorf("ExtendInt64Histogram with metricName: %v, err %v", metricName, err)
		return nil, err
	}

	return &ExtendInt64Histogram{
		ctx:            m.ctx,
		int64Histogram: histogram,
	}, nil
}

func (e *ExtendInt64Histogram) Record(labels []*ExtendLables, value int64) error {
	if e == nil || e.int64Histogram == nil {
		return errors.New("metrics can not be empty")
	}
	if labels == nil {
		return errors.New("metrics can not with empty labels")
	}

	attr := make([]attribute.KeyValue, 0, len(labels))
	for _, v := range labels {
		attr = append(attr, attribute.KeyValue{
			Key:   attribute.Key(v.Key),
			Value: attribute.StringValue(v.Val),
		})
	}
	opt := api.WithAttributeSet(
		attribute.NewSet(attr...),
	)
	e.int64Histogram.Record(e.ctx, value, opt)
	return nil
}

type ExtendFloat64Histogram struct {
	ctx              context.Context
	float64Histogram api.Float64Histogram
}

func (m *ExtendMeter) NewExtendFloat64Histogram(metricName, metricDesc, metricUnit string, bucket []float64) (*ExtendFloat64Histogram, error) {
	if m == nil || m.embeddedMeter == nil {
		return nil, errors.New("extendMeter is nil or embeddedMeter is nil")
	}
	if bucket == nil {
		bucket = []float64{0.0, 5.0, 10.0, 25.0, 50.0, 75.0, 100.0, 250.0, 500.0, 750.0, 1000.0, 2500.0, 5000.0, 7500.0, 10000.0, 200000.0}
	}
	histogram, err := m.embeddedMeter.Float64Histogram(metricName,
		api.WithDescription(metricDesc),
		api.WithUnit(metricUnit),
		api.WithExplicitBucketBoundaries(bucket...))

	if err != nil {
		zaplog.SDKLogger.Errorf("ExtendFloat64Histogram with metricName: %v, err %v", metricName, err)
		return nil, err
	}

	return &ExtendFloat64Histogram{
		ctx:              m.ctx,
		float64Histogram: histogram,
	}, nil
}

func (e *ExtendFloat64Histogram) Record(labels []*ExtendLables, value float64) error {
	if e == nil || e.float64Histogram == nil {
		return errors.New("metrics can not be empty")
	}
	if labels == nil {
		return errors.New("metrics can not with empty labels")
	}

	attr := make([]attribute.KeyValue, 0, len(labels))

	for _, v := range labels {
		attr = append(attr, attribute.KeyValue{
			Key:   attribute.Key(v.Key),
			Value: attribute.StringValue(v.Val),
		})
	}
	opt := api.WithAttributeSet(
		attribute.NewSet(attr...),
	)
	e.float64Histogram.Record(e.ctx, value, opt)
	return nil
}

type ExtendInt64Gauge struct {
	ctx        context.Context
	int64Gauge api.Int64Gauge
}

func (m *ExtendMeter) NewExtendInt64Gauge(metricName, metricDesc, metricUnit string) (*ExtendInt64Gauge, error) {
	if m == nil || m.embeddedMeter == nil {
		return nil, errors.New("extendMeter is nil or embeddedMeter is nil")
	}
	int64Gauge, err := m.embeddedMeter.Int64Gauge(metricName,
		api.WithDescription(metricDesc),
		api.WithUnit(metricUnit))

	if err != nil {
		zaplog.SDKLogger.Errorf("ExtendInt64Gauge with metricName: %v err %v", metricName, err)
		return nil, err
	}
	return &ExtendInt64Gauge{
		ctx:        m.ctx,
		int64Gauge: int64Gauge,
	}, nil
}

func (e *ExtendInt64Gauge) Record(labels []*ExtendLables, value int64) error {
	if e == nil || e.int64Gauge == nil {
		return errors.New("metrics can not be empty")
	}
	if labels == nil {
		return errors.New("metrics can not with empty labels")
	}

	attr := make([]attribute.KeyValue, 0, len(labels))
	for _, v := range labels {
		attr = append(attr, attribute.KeyValue{
			Key:   attribute.Key(v.Key),
			Value: attribute.StringValue(v.Val),
		})
	}
	opt := api.WithAttributeSet(
		attribute.NewSet(attr...),
	)

	e.int64Gauge.Record(e.ctx, value, opt)
	return nil
}

type ExtendFloat64Gauge struct {
	ctx          context.Context
	float64Gauge api.Float64Gauge
}

func (m *ExtendMeter) NewExtendFloat64Gauge(metricName, metricDesc, metricUnit string) (*ExtendFloat64Gauge, error) {
	if m == nil || m.embeddedMeter == nil {
		return nil, errors.New("extendMeter is nil or embeddedMeter is nil")
	}
	float64Gauge, err := m.embeddedMeter.Float64Gauge(metricName,
		api.WithDescription(metricDesc),
		api.WithUnit(metricUnit))

	if err != nil {
		zaplog.SDKLogger.Errorf("ExtendInt64Gauge with metricName: %v err %v", metricName, err)
		return nil, err
	}
	return &ExtendFloat64Gauge{
		ctx:          m.ctx,
		float64Gauge: float64Gauge,
	}, nil
}

func (e *ExtendFloat64Gauge) Record(labels []*ExtendLables, value float64) error {
	if e == nil || e.float64Gauge == nil {
		return errors.New("metrics can not be empty")
	}
	if labels == nil {
		return errors.New("metrics can not with empty labels")
	}
	attr := make([]attribute.KeyValue, 0, len(labels))
	for _, v := range labels {
		attr = append(attr, attribute.KeyValue{
			Key:   attribute.Key(v.Key),
			Value: attribute.StringValue(v.Val),
		})
	}

	opt := api.WithAttributeSet(
		attribute.NewSet(attr...),
	)
	e.float64Gauge.Record(e.ctx, value, opt)
	return nil
}

package monitor

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	api "go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"sync/atomic"
	"time"
)

const (
	default_service_name = "otlp_sdk"
	exportTimeOut        = time.Second * 30
	exportInterval       = time.Second * 1
)

// global mertrics
var (
	sdkMeterProvider = &metricsdk.MeterProvider{}
	sdkMeter         = &SdkMeter{}
	OtlpSdkMetrics   = &SdkMetrics{}
	SDKMetricEnable  uint32
)

type SdkMetrics struct {
	ctx         context.Context
	serviceName string
	counter     api.Int64Counter
	gauge       api.Int64Gauge
	histogram   api.Float64Histogram
}

func setSDKMetricsEnable(value bool) {
	if value {
		atomic.StoreUint32(&SDKMetricEnable, 1)
	} else {
		atomic.StoreUint32(&SDKMetricEnable, 0)
	}
}

func SDKMetricsEnable() bool {
	return atomic.LoadUint32(&SDKMetricEnable) != 0
}

type SdkMeter struct {
	ctx   context.Context
	meter api.Meter
}

func InitSDKMetrics(serviceName, reportAddr string) (err error) {
	ctx := context.Background()
	conn, err := grpc.NewClient(reportAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("InitSDKMetrics grpc new connection err: %v", err)
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(),
	)
	// 创建metrics exporter
	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("InitSDKMetrics new exporter err: %v", err)
	}
	reader := metricsdk.NewPeriodicReader(exporter,
		metricsdk.WithTimeout(exportTimeOut),   // export timeout 超过则取消上报 默认30s
		metricsdk.WithInterval(exportInterval)) // export interval export间隔 默认60s

	provider := metricsdk.NewMeterProvider(
		metricsdk.WithReader(reader),
		metricsdk.WithResource(res),
	)
	otel.SetMeterProvider(provider)
	sdkMeterProvider = provider
	sdkMeter = newMeter(default_service_name)
	// init metrics
	OtlpSdkMetrics, err = initMetrics(serviceName)
	if err == nil {
		setSDKMetricsEnable(true)
	}
	return
}

func newMeter(serviceName string) *SdkMeter {
	return &SdkMeter{
		ctx:   context.Background(),
		meter: sdkMeterProvider.Meter(serviceName),
	}
}

func initMetrics(serviceName string) (sdkMetrics *SdkMetrics, err error) {
	if serviceName == "" {
		serviceName = default_service_name
	}
	counter, err := sdkMeter.meter.Int64Counter("otlp_sdk_counter")
	if err != nil {
		return nil, err
	}
	gauge, err := sdkMeter.meter.Int64Gauge("otlp_sdk_gauge")
	if err != nil {
		return nil, err
	}
	bucket := []float64{0.0, 5.0, 10.0, 25.0, 50.0, 75.0, 100.0, 250.0, 500.0, 750.0, 1000.0, 2500.0, 5000.0, 7500.0, 10000.0, 200000.0, 300000.0}
	histogram, err := sdkMeter.meter.Float64Histogram("otlp_sdk_histogram", api.WithExplicitBucketBoundaries(bucket...))
	if err != nil {
		return nil, err
	}
	sdkMetrics = &SdkMetrics{
		serviceName: serviceName,
		ctx:         sdkMeter.ctx,
		counter:     counter,
		gauge:       gauge,
		histogram:   histogram,
	}
	return sdkMetrics, nil
}

func (sm *SdkMetrics) RecordCounter(ns string, value int64) {
	opt := api.WithAttributeSet(
		attribute.NewSet(
			attribute.KeyValue{Key: "ns", Value: attribute.StringValue(ns)},
			attribute.KeyValue{Key: "server_name", Value: attribute.StringValue(sm.serviceName)},
		),
	)
	sm.counter.Add(sm.ctx, value, opt)
}

func (sm *SdkMetrics) RecordHistogram(ns string, value float64) {
	opt := api.WithAttributeSet(
		attribute.NewSet(
			attribute.KeyValue{Key: "ns", Value: attribute.StringValue(ns)},
			attribute.KeyValue{Key: "server_name", Value: attribute.StringValue(sm.serviceName)},
		),
	)
	sm.histogram.Record(sm.ctx, value, opt)
}

func (sm *SdkMetrics) RecordGauge(ns string, addr string, value int64) {
	opt := api.WithAttributeSet(
		attribute.NewSet(
			attribute.KeyValue{Key: "ns", Value: attribute.StringValue(ns)},
			attribute.KeyValue{Key: "addr", Value: attribute.StringValue(addr)},
			attribute.KeyValue{Key: "server_name", Value: attribute.StringValue(sm.serviceName)},
		),
	)
	sm.gauge.Record(sm.ctx, value, opt)
}

// 逆初始化
func Fini() (err error) {
	if err := sdkMeterProvider.Shutdown(context.Background()); err != nil {
		log.Printf("shutdown MeterProvider err :%v \n", err)
		return err
	}
	log.Println("metrics fini ok")
	return
}

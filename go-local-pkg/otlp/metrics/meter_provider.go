package metrics

import (
	"context"
	"errors"
	"fmt"
	"git.iflytek.com/AIaaS/otlp-self/v3/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

// 获取MeterProvider
func NewMeterProvider(opts ...metricOptions) (*metricsdk.MeterProvider, error) {
	conf := newMetricConfig(opts...)
	log.Printf("NewMeterProvider with metricsConf: %#+v \n", conf)
	ctx := context.Background()
	conn, err := grpc.NewClient(conf.reportAddr, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(conf.maxGrpcSendSize*1024*1024)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("new grpc connection err: %v", err)
	}
	// 创建资源 表示关于非临时进程的底层元数据 如协议 等
	res, err := resource.New(ctx)
	// 创建metrics exporter
	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("create metrics exporter err: %v", err)
	}
	// export timeout 超过则取消上报 默认30s export interval export间隔 默认60s
	reader := metricsdk.NewPeriodicReader(exporter,
		metricsdk.WithTimeout(conf.exportTimeOut),
		metricsdk.WithInterval(conf.exportInterval))

	provider := metricsdk.NewMeterProvider(
		metricsdk.WithReader(reader),
		metricsdk.WithResource(res),
	)
	otel.SetMeterProvider(provider)
	return provider, nil
}

// 逆初始化
func Fini(ctx context.Context, m *metricsdk.MeterProvider) error {
	defer utils.Catch("metrics Fini")
	if m == nil {
		return errors.New("MeterProvider is nil")
	}
	if err := m.Shutdown(ctx); err != nil {
		log.Printf("shutdown MeterProvider err :%v \n", err)
		return err
	}
	log.Println("metrics fini ok")
	return nil
}

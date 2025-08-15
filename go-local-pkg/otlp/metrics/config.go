package metrics

import "time"

const (
	DEFAULT_METRICS_MAX_SIZE        int    = 4 //单位MB
	DEFAULT_METRICS_REPORT_ADDR     string = "127.0.0.1:4317"
	DEFAULT_METRICS_EXPORT_TIMEOUT         = time.Second * 30
	DEFAULT_METRICS_EXPORT_INTERVAL        = time.Second
)

type MetricConfig struct {
	reportAddr      string        //上报ip 生产环境配置127.0.0.1:4317
	maxGrpcSendSize int           // grpc客户端数据限制大小
	exportTimeOut   time.Duration // 上报超时时间，默认30s
	exportInterval  time.Duration // 上报间隔时间，默认1s
}

type metricOptions func(*MetricConfig)

func newMetricConfig(opts ...metricOptions) *MetricConfig {
	// api default
	defaultConfig := &MetricConfig{
		maxGrpcSendSize: DEFAULT_METRICS_MAX_SIZE,
		reportAddr:      DEFAULT_METRICS_REPORT_ADDR,
		exportTimeOut:   DEFAULT_METRICS_EXPORT_TIMEOUT,
		exportInterval:  DEFAULT_METRICS_EXPORT_INTERVAL,
	}
	for _, opt := range opts {
		opt(defaultConfig)
	}
	return defaultConfig
}

func (oc *MetricConfig) Option(opts ...metricOptions) {
	for _, opt := range opts {
		opt(oc)
	}
}

func WithReportAddr(reportAddr string) metricOptions {
	return func(oc *MetricConfig) {
		oc.reportAddr = reportAddr
	}
}

func WithMaxGrpcSendSize(maxGrpcSendSize int) metricOptions {
	return func(oc *MetricConfig) {
		oc.maxGrpcSendSize = maxGrpcSendSize
	}
}

func WithExportTimeOut(exportTimeOut time.Duration) metricOptions {
	return func(oc *MetricConfig) {
		oc.exportTimeOut = exportTimeOut
	}
}

func WithExportInterval(exportInterval time.Duration) metricOptions {
	return func(oc *MetricConfig) {
		oc.exportInterval = exportInterval
	}
}

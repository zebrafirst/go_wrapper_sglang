package logging

import "time"

const (
	// log兜底配置 log提供富媒体接口 不保障不丢失
	DEFAULT_LOG_REPORT_ADDR string = "127.0.0.1:4317"
)

type logConfig struct {
	serviceName        string
	exportInterval     time.Duration // 上报间隔  默认1s
	exportTimeOut      time.Duration // 上报超时时间 默认30s
	exportMaxBatchSize int           // 批量发送大小默认512
	maxQueueSize       int           // record
	maxGrpcSendSize    int           // grpc客户端数据限制大小
	flushQueueSize     int           // flush队列大小
	flushWorkerNum     int           // flush工作协程数
	flushBlock         bool          // flush是否阻塞 默认非阻塞
	flushTimeOut       time.Duration
	dumpEnable         bool // 是否开启dump 默认为false 研测调试，线上不要开启
}

type logOptions func(*logConfig)

func newOtlpLogConfig(opts ...logOptions) *logConfig {
	defaultlogConfig := &logConfig{
		// 通用兜底配置
		serviceName:        DEFAULT_SERVICE_NAME,
		exportInterval:     DEFAULT_EXPORT_INTERVAL,
		exportTimeOut:      DEFAULT_EXPORT_TIMEOUT,
		exportMaxBatchSize: DEFAULT_EXPORT_MAX_BATCH_SIZE,
		maxQueueSize:       DEFAULT_MAX_QUEUE_SIZE,
		maxGrpcSendSize:    DEFAULT_MAX_GRPC_SEND_SIZE,
		flushQueueSize:     DEFAULT_FLUSH_QUEUE_SIZE,
		flushWorkerNum:     DEFAULT_FLUSH_WORKER_NUM,
		flushBlock:         DEFAULT_FLUSH_BLOCK,
		flushTimeOut:       DEFAULT_FLUSH_TIMEOUT,
	}
	for _, opt := range opts {
		opt(defaultlogConfig)
	}
	return defaultlogConfig
}

func (oc *logConfig) Option(opts ...logOptions) {
	for _, opt := range opts {
		opt(oc)
	}
}

func WithLogServiceName(serviceName string) logOptions {
	return func(oc *logConfig) {
		oc.serviceName = serviceName
	}
}

// 默认 1s
func WithLogExportInterval(exportInterval time.Duration) logOptions {
	return func(oc *logConfig) {
		oc.exportInterval = exportInterval
	}
}

// 默认值30s
func WithLogExportTimeOut(exportTimeOut time.Duration) logOptions {
	return func(oc *logConfig) {
		oc.exportTimeOut = exportTimeOut
	}
}

// 默认 400
func WithLogExportMaxBatchSize(exportMaxBatchSize int) logOptions {
	return func(oc *logConfig) {
		oc.exportMaxBatchSize = exportMaxBatchSize
	}
}

// 默认 1000
func WithLogMaxQueueSize(maxQueueSize int) logOptions {
	return func(oc *logConfig) {
		oc.maxQueueSize = maxQueueSize
	}
}

// grpc传递消息大小 单位MB  默认4MB  客户端和服务端需保持一致
func WithLogMaxGrpcSendSize(maxGrpcSendSize int) logOptions {
	return func(oc *logConfig) {
		oc.maxGrpcSendSize = maxGrpcSendSize
	}
}

func WithLogFlushQueueSize(flushQueueSize int) logOptions {
	return func(oc *logConfig) {
		oc.flushQueueSize = flushQueueSize
	}
}

func WithLogFlushWorkerNum(flushWorkerNum int) logOptions {
	return func(oc *logConfig) {
		oc.flushWorkerNum = flushWorkerNum
	}
}

// 默认非阻塞 false
func WithLogFlushBlock(flushBlock bool) logOptions {
	return func(oc *logConfig) {
		oc.flushBlock = flushBlock
	}
}

// 默认 10s
func WithLogFlushTimeOut(flushTimeOut time.Duration) logOptions {
	return func(oc *logConfig) {
		oc.flushTimeOut = flushTimeOut
	}
}

func WithDumpEnable(dumpEnable bool) logOptions {
	return func(oc *logConfig) {
		oc.dumpEnable = dumpEnable
	}
}

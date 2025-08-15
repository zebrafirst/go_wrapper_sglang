package api

const (
	// 日志相关
	DEFAULT_LOG_LEVEL       = "warn"
	DEFAULT_MAX_SIZE        = 100
	DEFAULT_MAX_BACKUPS     = 10
	DEFAULT_MAX_AGE         = 30
	DEFAULT_BATCH_SIZE      = 100
	DEFAULT_CALLER          = true
	DEFAULT_ASYNC           = true
	DEFAULT_CACHE_MAX_COUNT = -1
	DEFAULT_WASH            = 60

	// 监控相关
	DEFAULT_SDK_METRICS_ENABLE       bool   = false
	DEFAULT_SDK_METRICS_ADDR         string = "127.0.0.1:4317"
	DEFAULT_SDK_METRICS_SERVICE_NAME string = "otlpService"
)

type OtlpConfig struct {
	fileName      string
	logLevel      string // 日志级别 默认 warn
	maxSize       int    // 日志大小 单位 MB 默认 10
	maxBackups    int    // 日志备份数量 默认 10
	maxAge        int    // 保存时间 单位天
	async         bool   // 是否异步 默认true
	cacheMaxCount int    // 缓存大小，单位日志条数 超过则丢弃 缺省-1 表示不丢数据，堆积在内存中
	batchSize     int    // 批量日志大小
	wash          int    // 写入磁盘的时间
	caller        bool   // 是否添加调用行信息

	sdkMetricsEnable      bool
	sdkMetricsAddr        string // 默认 127.0.0.1:4317
	sdkMetricsServiceName string // 默认 "otlpService"
}

type option func(*OtlpConfig)

func NewOtlpConfig(opts ...option) *OtlpConfig {
	// api default
	defaultOtlpConfig := &OtlpConfig{
		logLevel:              DEFAULT_LOG_LEVEL,
		maxSize:               DEFAULT_MAX_SIZE,
		maxBackups:            DEFAULT_MAX_BACKUPS,
		maxAge:                DEFAULT_MAX_AGE,
		batchSize:             DEFAULT_BATCH_SIZE,
		async:                 DEFAULT_ASYNC,
		caller:                DEFAULT_CALLER,
		cacheMaxCount:         DEFAULT_CACHE_MAX_COUNT,
		wash:                  DEFAULT_WASH,
		sdkMetricsEnable:      DEFAULT_SDK_METRICS_ENABLE,
		sdkMetricsAddr:        DEFAULT_SDK_METRICS_ADDR,
		sdkMetricsServiceName: DEFAULT_SDK_METRICS_SERVICE_NAME,
	}
	for _, opt := range opts {
		opt(defaultOtlpConfig)
	}
	return defaultOtlpConfig
}

func (oc *OtlpConfig) Option(opts ...option) {
	for _, opt := range opts {
		opt(oc)
	}
}

func WithFileName(fileName string) option {
	return func(oc *OtlpConfig) {
		oc.fileName = fileName
	}
}

func WithLogLevel(logLevel string) option {
	return func(oc *OtlpConfig) {
		oc.logLevel = logLevel
	}
}

func WithMaxSize(maxSize int) option {
	return func(oc *OtlpConfig) {
		oc.maxSize = maxSize
	}
}

func WithMaxBackups(maxBackups int) option {
	return func(oc *OtlpConfig) {
		oc.maxBackups = maxBackups
	}
}

func WithMaxAge(maxAge int) option {
	return func(oc *OtlpConfig) {
		oc.maxAge = maxAge
	}
}

func WithAsync(async bool) option {
	return func(oc *OtlpConfig) {
		oc.async = async
	}
}

func WithCacheMaxCount(cacheMaxCount int) option {
	return func(oc *OtlpConfig) {
		oc.cacheMaxCount = cacheMaxCount
	}
}

func WithBatchSize(batchSize int) option {
	return func(oc *OtlpConfig) {
		oc.batchSize = batchSize
	}
}

func WithWash(wash int) option {
	return func(oc *OtlpConfig) {
		oc.wash = wash
	}
}

func WithCaller(caller bool) option {
	return func(oc *OtlpConfig) {
		oc.caller = caller
	}
}

func WithSdkMetricsEnable(sdkMetricsEnable bool) option {
	return func(oc *OtlpConfig) {
		oc.sdkMetricsEnable = sdkMetricsEnable
	}
}

func WithSdkMetricsAddr(sdkMetricsAddr string) option {
	return func(oc *OtlpConfig) {
		oc.sdkMetricsAddr = sdkMetricsAddr
	}
}

func WithSdkMetricsServiceName(sdkMetricsServiceName string) option {
	return func(oc *OtlpConfig) {
		oc.sdkMetricsServiceName = sdkMetricsServiceName
	}
}

package logging

import "time"

// 完全兼容quiver sdk
const (
	// 富媒体文件
	DEFAULT_MEDIA_MAX_SIZE       = 30
	DEFAULT_MEDIA_UP_TIMEOUT     = time.Second * 5
	DEFAULT_MEDIA_QUEUE_SIZE     = 10000
	DEFAULT_MEDIA_WORKER_NUM     = 32
	DEFAULT_MEDIA_RETRY_NUM  int = 3
	DEFAULT_MEDIA_BLOCK          = false
	DEFAULT_SGW_LOG_LEVEL    int = 2
	DEFAULT_SGW_LINK_TTL         = 3600 * 24 // 默认外链生效时间为1天

	// 以下配置为日志防止丢失使用配置
	// eventlog兜底备份配置
	DEFAULT_EVENT_REPORT_ADDR    string = "127.0.0.1:4319"
	DEFAULT_EVENT_BACKUP_PATH    string = "/log/server/otlp/event/event_backup.log"
	DEFAULT_EVENT_PROBE_TIMEOUT         = time.Second * 5
	DEFAULT_EVENT_PROBE_INTERVAL        = time.Second * 30

	// 富媒体文件兜底备份配置
	DEFAULT_MEDIA_BACKUP_PATH    = "/log/server/otlp/sgw/sgw_backup.log"
	DEFAULT_MEDIA_PROBE_TIMEOUT  = 5 * time.Second
	DEFAULT_MEDIA_PROBE_INTERVAL = 30 * time.Second
)

type eventConfig struct {
	serviceName string
	// event 配置
	exportInterval     time.Duration // 上报间隔  默认1s
	exportTimeOut      time.Duration // 上报超时时间 默认30s
	exportMaxBatchSize int           // 批量发送大小默认512
	maxQueueSize       int           // record
	maxGrpcSendSize    int           // grpc客户端数据限制大小
	flushQueueSize     int           // flush队列大小 备份复用
	flushWorkerNum     int           // flush工作协程数
	flushBlock         bool          // flush是否阻塞 默认非阻塞
	flushTimeOut       time.Duration

	eventDumpEnable bool // 是否开启dump 默认为false 研测调试，线上不要开启
	// 富媒体配置
	mediaEnable    bool          // 是否使用富媒体上传
	mediaMaxSize   int           // 富媒体文件上传限制大小 默认30MB
	mediaUpTimeOut time.Duration // 上传富媒体文件超时时间
	mediaRetryNum  int           // 上传重试次数
	// 自定义func
	mediaFunc      MediaFunc
	mediaQueueSize int
	mediaWorkerNum int
	mediaBlock     bool // 上传富媒体文件是否阻塞

	sgwAccessKey string // 存储网关四元组
	sgwSecretKey string
	sgwAddress   string
	sgwNameSpace string
	sgwLinkTTL   int // 外链生效时间 重要 单位s
	sgwLogPath   string
	sgwLogLevel  int // 0 warn 1 info 2 error

	// event备份相关配置
	eventBackUpEnable  bool
	eventBackUpPath    string
	eventProbeAddr     string
	eventProbeTimeOut  time.Duration
	eventProbeInterVal time.Duration

	// 富媒体备份相关配置  支持使用存储网关的探测和
	mediaBackUpEnable  bool
	mediaBackUpPath    string
	mediaProbeAddr     string
	mediaProbeTimeOut  time.Duration
	mediaProbeInterVal time.Duration
}

type eventOptions func(*eventConfig)

func newEventConfig(opts ...eventOptions) *eventConfig {
	defaulteventConfig := &eventConfig{
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

		// eventlog兜底配置
		eventBackUpPath:    DEFAULT_EVENT_BACKUP_PATH,
		eventProbeAddr:     DEFAULT_EVENT_REPORT_ADDR,
		eventProbeTimeOut:  DEFAULT_EVENT_PROBE_TIMEOUT,
		eventProbeInterVal: DEFAULT_EVENT_PROBE_INTERVAL,

		// 富媒体兜底配置
		mediaMaxSize:   DEFAULT_MEDIA_MAX_SIZE,
		mediaUpTimeOut: DEFAULT_MEDIA_UP_TIMEOUT,
		mediaQueueSize: DEFAULT_MEDIA_QUEUE_SIZE,
		mediaWorkerNum: DEFAULT_MEDIA_WORKER_NUM,
		sgwLinkTTL:     DEFAULT_SGW_LINK_TTL,
		mediaBlock:     DEFAULT_MEDIA_BLOCK,
		sgwLogLevel:    DEFAULT_SGW_LOG_LEVEL,
		mediaRetryNum:  DEFAULT_MEDIA_RETRY_NUM,

		mediaBackUpPath:    DEFAULT_MEDIA_BACKUP_PATH,
		mediaProbeTimeOut:  DEFAULT_MEDIA_PROBE_TIMEOUT,
		mediaProbeInterVal: DEFAULT_MEDIA_PROBE_INTERVAL,
	}
	for _, opt := range opts {
		opt(defaulteventConfig)
	}
	return defaulteventConfig
}

func (oc *eventConfig) Option(opts ...eventOptions) {
	for _, opt := range opts {
		opt(oc)
	}
}

func WithServiceName(serviceName string) eventOptions {
	return func(oc *eventConfig) {
		oc.serviceName = serviceName
	}
}

func WithExportInterval(exportInterval time.Duration) eventOptions {
	return func(oc *eventConfig) {
		oc.exportInterval = exportInterval
	}
}

func WithExportTimeOut(exportTimeOut time.Duration) eventOptions {
	return func(oc *eventConfig) {
		oc.exportTimeOut = exportTimeOut
	}
}

func WithExportMaxBatchSize(exportMaxBatchSize int) eventOptions {
	return func(oc *eventConfig) {
		oc.exportMaxBatchSize = exportMaxBatchSize
	}
}

func WithMaxQueueSize(maxQueueSize int) eventOptions {
	return func(oc *eventConfig) {
		oc.maxQueueSize = maxQueueSize
	}
}

// grpc传递消息大小 单位4MB  默认4MB  客户端和服务端需保持一致
func WithMaxGrpcSendSize(maxGrpcSendSize int) eventOptions {
	return func(oc *eventConfig) {
		oc.maxGrpcSendSize = maxGrpcSendSize
	}
}

func WithFlushQueueSize(flushQueueSize int) eventOptions {
	return func(oc *eventConfig) {
		oc.flushQueueSize = flushQueueSize
	}
}

func WithFlushWorkerNum(flushWorkerNum int) eventOptions {
	return func(oc *eventConfig) {
		oc.flushWorkerNum = flushWorkerNum
	}
}

func WithFlushBlock(flushBlock bool) eventOptions {
	return func(oc *eventConfig) {
		oc.flushBlock = flushBlock
	}
}

func WithFlushTimeOut(flushTimeOut time.Duration) eventOptions {
	return func(oc *eventConfig) {
		oc.flushTimeOut = flushTimeOut
	}
}

func WithMediaEnable(mediaEnable bool) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaEnable = mediaEnable
	}
}

func WithEventDumpEnable(eventDumpEnable bool) eventOptions {
	return func(oc *eventConfig) {
		oc.eventDumpEnable = eventDumpEnable
	}
}

func WithEventBackUpEnable(eventBackUpEnable bool) eventOptions {
	return func(oc *eventConfig) {
		oc.eventBackUpEnable = eventBackUpEnable
	}
}

func WithEventBackUpPath(eventBackUpPath string) eventOptions {
	return func(oc *eventConfig) {
		oc.eventBackUpPath = eventBackUpPath
	}
}

func WithEventProbeAddr(eventProbeAddr string) eventOptions {
	return func(oc *eventConfig) {
		oc.eventProbeAddr = eventProbeAddr
	}
}

func WithEventProbeTimeOut(eventProbeTimeOut time.Duration) eventOptions {
	return func(oc *eventConfig) {
		oc.eventProbeTimeOut = eventProbeTimeOut
	}
}

func WithEventProbeInterVal(eventProbeInterVal time.Duration) eventOptions {
	return func(oc *eventConfig) {
		oc.eventProbeInterVal = eventProbeInterVal
	}
}

func WithMediaMaxSize(mediaMaxSize int) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaMaxSize = mediaMaxSize
	}
}

func WithSgwAccessKey(sgwAccessKey string) eventOptions {
	return func(oc *eventConfig) {
		oc.sgwAccessKey = sgwAccessKey
	}
}

func WithSgwSecretKey(sgwSecretKey string) eventOptions {
	return func(oc *eventConfig) {
		oc.sgwSecretKey = sgwSecretKey
	}
}

func WithSgwNameSpace(sgwNameSpace string) eventOptions {
	return func(oc *eventConfig) {
		oc.sgwNameSpace = sgwNameSpace
	}
}

func WithSgwAddress(sgwAddress string) eventOptions {
	return func(oc *eventConfig) {
		oc.sgwAddress = sgwAddress
	}
}

func WithMediaUpTimeOut(mediaUpTimeOut time.Duration) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaUpTimeOut = mediaUpTimeOut
	}
}

func WithSgwLinkTTL(sgwLinkTTL int) eventOptions {
	return func(oc *eventConfig) {
		oc.sgwLinkTTL = sgwLinkTTL
	}
}

func WithMediaQueueSize(mediaQueueSize int) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaQueueSize = mediaQueueSize
	}
}

func WithMediaWorkerNum(mediaWorkerNum int) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaWorkerNum = mediaWorkerNum
	}
}

func WithMediaRetryNum(mediaRetryNum int) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaRetryNum = mediaRetryNum
	}
}

func WithMediaBlock(mediaBlock bool) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaBlock = mediaBlock
	}
}

func WithSgwLogPath(sgwLogPath string) eventOptions {
	return func(oc *eventConfig) {
		oc.sgwLogPath = sgwLogPath
	}
}

func WithSgwLogLevel(sgwLogLevel int) eventOptions {
	return func(oc *eventConfig) {
		oc.sgwLogLevel = sgwLogLevel
	}
}

func WithMediaBackUpEnable(mediaBackUpEnable bool) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaBackUpEnable = mediaBackUpEnable
	}
}

func WithMediaBackUpPath(mediaBackUpPath string) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaBackUpPath = mediaBackUpPath
	}
}

func WithMediaProbeAddr(mediaProbeAddr string) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaProbeAddr = mediaProbeAddr
	}
}

func WithMediaProbeTimeOut(mediaProbeTimeOut time.Duration) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaProbeTimeOut = mediaProbeTimeOut
	}
}

func WithMediaProbeInterVal(mediaProbeInterVal time.Duration) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaProbeInterVal = mediaProbeInterVal
	}
}

func WithMediaFunc(mediaFunc MediaFunc) eventOptions {
	return func(oc *eventConfig) {
		oc.mediaFunc = mediaFunc
	}
}

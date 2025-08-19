package logging

import (
	"context"
	"errors"
	minilog "git.iflytek.com/AIaaS/mini-log"
	"git.iflytek.com/AIaaS/otlp-self/v3"
	"git.iflytek.com/AIaaS/otlp-self/v3/logging/kv"
	"git.iflytek.com/AIaaS/otlp-self/v3/probe"
	"git.iflytek.com/AIaaS/otlp-self/v3/utils"
	"git.iflytek.com/AIaaS/otlp-self/v3/zaplog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	logapi "go.opentelemetry.io/otel/log"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"time"
)

type MediaFunc func(md *MediaData) (string, error)

type EventLogProvider struct {
	serviceName string
	// 内置
	embedEventProvider *logsdk.LoggerProvider
	// dump
	eventDumpEnable bool
	eventDumpDir    string

	flushTimeOut         time.Duration
	eventFlushPool       *WorkerPool
	eventBackUpFlushPool *WorkerPool

	// 富媒体相关
	mediaEnable    bool
	mediaMaxSize   int
	mediaRetryNum  int
	mediaUpTimeOut time.Duration
	mediaWorkPool  *WorkerPool
	mediaFunc      MediaFunc // 自定义func
	sgwLinkTTL     int
	sgwNameSpace   string

	mediaBackUpWorkPool *WorkerPool

	// 备份相关
	eventBackUpEnable   bool
	eventBackupQuitChan chan struct{}
	mediaBackUpEnable   bool
	mediaBackupQuitChan chan struct{}

	sgwLogger   *minilog.Logger // 记录备份日志 存储网关日志
	eventLogger *minilog.Logger // 记录备份日志 eventlog
}

func NewEventProvider(reportAddr string, opts ...eventOptions) (*EventLogProvider, error) {
	conf := newEventConfig(opts...)
	log.Printf("NewEventProvider with conf: %#+v \n", conf)
	if reportAddr == "" {
		reportAddr = DEFAULT_EVENT_REPORT_ADDR
	}
	// 创建grpc 连接
	conn, err := grpc.NewClient(reportAddr,
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(conf.maxGrpcSendSize*1024*1024)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	// 创建资源 表示关于非临时进程的底层元数据 如协议等
	res := resource.NewWithAttributes(semconv.SchemaURL,
		attribute.String("environment", otlp.Environment),
		attribute.String("version", otlp.Version))
	// 创建 exporter
	exporter, err := otlploggrpc.New(context.Background(), otlploggrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}
	// 创建批量处理器
	batchProcessor := logsdk.NewBatchProcessor(exporter,
		logsdk.WithMaxQueueSize(conf.maxQueueSize),             //dfltMaxQSize        = 2048
		logsdk.WithExportInterval(conf.exportInterval),         //dfltExpInterval     = time.Second
		logsdk.WithExportTimeout(conf.exportTimeOut),           //dfltExpTimeout      = 30 * time.Second
		logsdk.WithExportMaxBatchSize(conf.exportMaxBatchSize)) //dfltExpMaxBatchSize = 512
	// 创建provider
	provider := logsdk.NewLoggerProvider(logsdk.WithResource(res),
		logsdk.WithProcessor(batchProcessor),
		logsdk.WithAttributeCountLimit(DEFAULT_ATTR_COUNT_LIMIT),         //defaultAttrCntLim    = 128   属性个数
		logsdk.WithAttributeValueLengthLimit(DEFAULT_VALUE_LENGTH_LIMIT)) //defaultAttrValLenLim = -1 不限制 单个属性长度

	eventProvider := &EventLogProvider{
		serviceName:        conf.serviceName,
		embedEventProvider: provider,
		flushTimeOut:       conf.flushTimeOut,
	}
	// 开启本地下载
	if conf.eventDumpEnable {
		dumpDir := "." + string(os.PathSeparator) + "eventlog"
		if err = os.MkdirAll(dumpDir, 0755); err != nil {
			zaplog.SDKLogger.Errorf("mkdir dumpdir err : %v", err)
			return nil, err
		}
		eventProvider.eventDumpEnable = conf.eventDumpEnable
		eventProvider.eventDumpDir = dumpDir
	}
	// 开启备份
	if conf.eventBackUpEnable {
		eventLogger, err := zaplog.InitBusinessLogger(&minilog.LogConf{Level: "error", File: conf.eventBackUpPath})
		if err != nil {
			return nil, err
		}
		eventProvider.eventLogger = eventLogger
		quitChan := make(chan struct{})
		probe.ProbeEventCollector(quitChan, conf.eventProbeAddr, conf.eventProbeTimeOut, conf.eventProbeInterVal)
		eventProvider.eventBackUpEnable = conf.eventBackUpEnable
		eventProvider.eventBackupQuitChan = quitChan
		eventProvider.eventBackUpFlushPool = newWorkerPool("event_flush_backup", conf.flushQueueSize, conf.flushWorkerNum, conf.flushBlock)
	}
	// 初始化富媒体相关
	if conf.mediaEnable {
		if err = eventProvider.initMedia(conf); err != nil {
			return nil, err
		}
	}
	// event pool
	eventProvider.eventFlushPool = newWorkerPool("event_flush", conf.flushQueueSize, conf.flushWorkerNum, conf.flushBlock)
	eventProvider.start()
	return eventProvider, nil
}

// 自定义富媒体文件
func (ep *EventLogProvider) initMedia(conf *eventConfig) (err error) {
	ep.mediaEnable = conf.mediaEnable

	ep.mediaMaxSize = conf.mediaMaxSize
	ep.mediaUpTimeOut = conf.mediaUpTimeOut
	ep.mediaRetryNum = conf.mediaRetryNum
	ep.mediaWorkPool = newWorkerPool("event_media", conf.mediaQueueSize, conf.mediaWorkerNum, conf.mediaBlock)
	// 如果自定义了就不需要存储网关了
	if conf.mediaFunc != nil {
		ep.mediaFunc = conf.mediaFunc
		return nil
	}
	// 初始化存储网关
	sgwConf := &SgwConfig{
		SgwAccessKey: conf.sgwAccessKey,
		SgwSecretKey: conf.sgwSecretKey,
		SgwAddress:   conf.sgwAddress,
		SgwUpTimeOut: conf.mediaUpTimeOut,
		SgwLogPath:   conf.sgwLogPath,
		SgwLogLevel:  conf.sgwLogLevel,
	}
	if err = initSgw(sgwConf); err != nil {
		zaplog.SDKLogger.Errorf("initSgw err: %v with sgwConfig: %v", err, conf)
		return err
	}
	ep.sgwLinkTTL = conf.sgwLinkTTL
	ep.sgwNameSpace = conf.sgwNameSpace

	if conf.mediaBackUpEnable {
		sgwLogger, err := zaplog.InitBusinessLogger(&minilog.LogConf{Level: "error", File: conf.mediaBackUpPath})
		if err != nil {
			return err
		}
		ep.sgwLogger = sgwLogger
		quitChan := make(chan struct{})
		probe.ProbeMedia(quitChan, conf.mediaProbeAddr, conf.mediaProbeTimeOut, conf.mediaProbeInterVal)
		ep.mediaBackUpEnable = conf.mediaBackUpEnable
		ep.mediaBackupQuitChan = quitChan
		ep.mediaBackUpWorkPool = newWorkerPool("event_sgw_backup", conf.mediaQueueSize, conf.mediaWorkerNum, conf.mediaBlock)
	}

	return
}

func (p *EventLogProvider) start() {
	if p.eventFlushPool != nil {
		p.eventFlushPool.start()
	}

	if p.mediaEnable && p.mediaWorkPool != nil {
		p.mediaWorkPool.start()
	}
	if p.eventBackUpEnable && p.eventBackUpFlushPool != nil {
		p.eventBackUpFlushPool.start()
	}

	if p.mediaBackUpEnable && p.mediaBackUpWorkPool != nil {
		p.mediaBackUpWorkPool.start()
	}
}

// 创建一个EventLog
func (ep *EventLogProvider) EventLog(serviceName, sid, sub, endpoint string) *EventLog {
	if ep == nil || ep.embedEventProvider == nil {
		return nil
	}
	event := newEvent(sid, serviceName, sub, endpoint)
	eventLog := &EventLog{
		//serviceName 一致 logger为同一个
		logger:    ep.embedEventProvider.Logger(serviceName, logapi.WithInstrumentationVersion(otlp.Version)),
		eventData: event,
		p:         ep,
	}
	// 初始化计数器、ctx 保证flush前一个eventlog中的图片全部上传完成 用于通知上传成功 同时防止flush阻塞
	if ep.mediaEnable {
		eventLog.mediaCtx, eventLog.mediaCancel = context.WithCancel(context.Background())
		eventLog.ops = 0
	}
	return eventLog
}

func newEvent(sid, serviceName, sub, endpoint string) *eventData {
	event := &eventData{
		Type:      0,
		Sid:       sid,
		Timestamp: utils.CurrentTimeMillis()}
	event.Tags = make(map[string]string)
	event.Outputs = make(map[string][]string)
	event.Name = serviceName
	return event.Tag(kv.KV{"serviceName", serviceName}).WithEndpoint(endpoint).WithSub(sub)
}

func (p *EventLogProvider) Fini() error {
	defer utils.Catch("EventLogProvider Fini")
	if p == nil || p.embedEventProvider == nil {
		return errors.New("EventLogProvider is nil")
	}
	if err := p.embedEventProvider.Shutdown(context.Background()); err != nil {
		return err
	}
	if p.eventFlushPool != nil {
		p.eventFlushPool.stop()
	}
	if p.mediaEnable && p.mediaWorkPool != nil {
		p.mediaWorkPool.stop()
	}
	// 停止探测collector
	if p.eventBackUpEnable {
		if p.eventLogger != nil {
			p.eventLogger.Close()
		}
		if p.eventBackupQuitChan != nil {
			p.eventBackupQuitChan <- struct{}{}
		}
	}
	// 停止探测存储网关
	if p.mediaBackUpEnable {
		if p.sgwLogger != nil {
			p.sgwLogger.Close()
		}
		if p.mediaBackupQuitChan != nil {
			p.mediaBackupQuitChan <- struct{}{}
		}
	}
	return nil
}

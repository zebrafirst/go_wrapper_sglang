package logging

import (
	"context"
	"errors"
	"git.iflytek.com/AIaaS/otlp-self/v3"
	"git.iflytek.com/AIaaS/otlp-self/v3/logging/kv"
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

type OtlpLogProvider struct {
	serviceName string
	// dump
	dumpEnable           bool
	dumpDir              string
	embedOtlpLogProvider *logsdk.LoggerProvider
	otlpLogFlushPool     *WorkerPool
	flushTimeOut         time.Duration
}

func NewOtlpLogProvider(reportAddr string, opts ...logOptions) (*OtlpLogProvider, error) {
	conf := newOtlpLogConfig(opts...)
	log.Printf("NewOtlpLogProvider with conf: %#+v \n", conf)
	if reportAddr == "" {
		reportAddr = DEFAULT_LOG_REPORT_ADDR
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

	otlpLogProvider := &OtlpLogProvider{
		serviceName:          conf.serviceName,
		embedOtlpLogProvider: provider,
		otlpLogFlushPool:     newWorkerPool("otlp_flush", conf.flushQueueSize, conf.flushWorkerNum, conf.flushBlock),
		flushTimeOut:         conf.flushTimeOut,
	}

	if conf.dumpEnable {
		dumpDir := "." + string(os.PathSeparator) + "otlplog"
		if err = os.MkdirAll(dumpDir, 0755); err != nil {
			zaplog.SDKLogger.Errorf("mkdir dumpdir err : %v", err)
			return nil, err
		}
		otlpLogProvider.dumpEnable = conf.dumpEnable
		otlpLogProvider.dumpDir = dumpDir
	}
	otlpLogProvider.start()
	return otlpLogProvider, nil
}

func (p *OtlpLogProvider) start() {
	for _, w := range p.otlpLogFlushPool.workers {
		w.run(p.otlpLogFlushPool.ctx, p.otlpLogFlushPool.queue)
	}
}

// 每次 一个log
func (op *OtlpLogProvider) OtlpLog(serviceName, sid, host string) (*OtlpLog, error) {
	if op == nil || op.embedOtlpLogProvider == nil {
		return nil, errors.New("logProvider or embedOtlpLogProvider can not be nil")
	}

	if serviceName == "" {
		serviceName = op.serviceName
	}

	otlpLog := &OtlpLog{
		logger:  op.embedOtlpLogProvider.Logger(serviceName, logapi.WithInstrumentationVersion(otlp.Version)),
		message: map[string]interface{}{SERVICE_NAME: serviceName, SID: sid, HOST: host, TIMESTAMP: utils.CurrentTimeMillis()},
		p:       op,
	}

	return otlpLog, nil
}

// 记录kv
func (op *OtlpLogProvider) Log(serviceName, sid, host string, kv ...kv.KV) {
	if op == nil || op.embedOtlpLogProvider == nil {
		return
	}
	if serviceName == "" {
		serviceName = op.serviceName
	}
	otlpLog := &OtlpLog{
		logger:  op.embedOtlpLogProvider.Logger(serviceName, logapi.WithInstrumentationVersion(otlp.Version)),
		message: map[string]interface{}{SERVICE_NAME: serviceName, SID: sid, HOST: host, TIMESTAMP: utils.CurrentTimeMillis()},
		p:       op,
	}
	if kv != nil {
		for _, v := range kv {
			otlpLog.message[v.Key] = v.Value
		}
	}
	otlpLog.Flush()
}

// 调用此接口记录msg 需要手动刷新
func (op *OtlpLogProvider) NewLog(serviceName, sid, host string) *OtlpLog {
	if op == nil || op.embedOtlpLogProvider == nil {
		return nil
	}
	if serviceName == "" {
		serviceName = op.serviceName
	}
	otlpLog := &OtlpLog{
		logger:  op.embedOtlpLogProvider.Logger(op.serviceName, logapi.WithInstrumentationVersion(otlp.Version)),
		message: map[string]interface{}{SERVICE_NAME: serviceName, SID: sid, HOST: host, TIMESTAMP: utils.CurrentTimeMillis()},
		p:       op,
	}
	return otlpLog
}

func (o *OtlpLogProvider) Fini() (err error) {
	if o == nil || o.embedOtlpLogProvider == nil {
		return errors.New("OtlpLogProvider or embedOtlpLogProvider is nil")
	}
	if err = o.embedOtlpLogProvider.Shutdown(context.Background()); err != nil {
		return err
	}
	if o.otlpLogFlushPool != nil {
		o.otlpLogFlushPool.stop()
	}
	return
}

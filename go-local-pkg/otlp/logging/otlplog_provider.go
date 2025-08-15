package logging

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"git.iflytek.com/AIaaS/otlp-self/v3"
	"git.iflytek.com/AIaaS/otlp-self/v3/global/kv"
	"git.iflytek.com/AIaaS/otlp-self/v3/internal"
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
)

type OtlpLogProvider struct {
	serviceName string
	// dump
	dumpEnable           bool
	dumpDir              string
	embedOtlpLogProvider *logsdk.LoggerProvider
	otlpLogFlushPool     *internal.Pool
	flushTimeOut         time.Duration
}

// 创建并初始化 OtlpLogProvider。
// reportAddr 为日志上报地址（如为空则使用默认地址）。
// opts 为可选参数，支持自定义日志导出、批量处理等配置。
// 返回 OtlpLogProvider 实例指针和错误信息。
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
		otlpLogFlushPool:     internal.NewWorkPool(conf.flushWorkerNum, conf.finiEnableWait),
		flushTimeOut:         conf.flushTimeOut,
	}

	if conf.logDumpEnable {
		dumpDir := "." + string(os.PathSeparator) + "otlplog"
		if err = os.MkdirAll(dumpDir, 0755); err != nil {
			zaplog.SDKLogger.Errorf("mkdir dumpdir err : %v", err)
			return nil, err
		}
		otlpLogProvider.dumpEnable = conf.logDumpEnable
		otlpLogProvider.dumpDir = dumpDir
	}
	return otlpLogProvider, nil
}

// 创建一个 OtlpLog 实例。
// serviceName 为空时使用默认服务名。
// 返回 OtlpLog 实例指针和错误信息。
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

// 记录一条日志（带自定义 kv 参数），并立即刷新到后端。
// serviceName 为空时使用默认服务名。
// sid 为日志标识符，按照规范填写。
// host 为日志记录的主机地址或标识符。
// kv 为日志的键值对信息。
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
	for _, v := range kv {
		otlpLog.message[v.Key] = v.Value
	}
	otlpLog.Flush()
}

// 创建一个 OtlpLog 实例（单条日志），但不会自动刷新到后端，需要调用 Flush 方法手动提交。
// serviceName 为空时使用默认服务名。
// sid 为日志标识符，按照规范填写。
// host 为日志记录的主机地址或标识符。
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
		o.otlpLogFlushPool.Stop()
	}
	return
}

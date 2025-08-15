package api

import (
	"context"
	"log"
	"time"

	minilog "git.iflytek.com/AIaaS/mini-log"
	"git.iflytek.com/AIaaS/otlp-self/v3"
	"git.iflytek.com/AIaaS/otlp-self/v3/metrics"
	"git.iflytek.com/AIaaS/otlp-self/v3/zaplog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
)

// 全局变量，用于在eventlog中访问metrics
var globalOtlpHandler *OtlpHandler

type OtlpHandler struct {
	sdkLogger     *minilog.Logger
	meterProvider *metric.MeterProvider
	metricsCtx    context.Context
	internalMeter *metrics.ExtendMeter
	otlpCounter   *metrics.ExtendInt64Counter
	apmCounter    *metrics.ExtendInt64Counter
	serviceName   string
}

// OtlpQuiverLogger 实现 quiver.CustomLogInterface 接口
type OtlpQuiverLogger struct {
	logger *minilog.Logger
}

func (o *OtlpHandler) setOtelLog() {
	otel.SetErrorHandler(o)
}

func (o *OtlpHandler) Handle(err error) {
	o.sdkLogger.Errorf("otlp sdk err:%v", err)
}

// InitOtlp 初始化 OTLP SDK。
// 参数 opts 为可选配置项，详见 option 定义。
// 返回值为初始化后的 OTLP 句柄（*OtlpHandler），可用于后续资源释放等操作，以及初始化失败时的错误信息。
func InitOtlp(opts ...option) (otlpHandler *OtlpHandler, err error) {
	conf := NewOtlpConfig(opts...)
	log.Printf("about to init otlp | version: %v \n", otlp.Version)

	// 初始化logger
	logConf := &minilog.LogConf{
		Level:         conf.logLevel,
		File:          conf.fileName,
		MaxSize:       conf.maxSize,
		MaxBackups:    conf.maxBackups,
		MaxAge:        conf.maxAge,
		Async:         conf.async,
		CacheMaxCount: conf.cacheMaxCount,
		BatchSize:     conf.batchSize,
		Wash:          conf.wash,
		Caller:        conf.caller,
	}
	logger, err := zaplog.InitSDKLogger(logConf)
	if err != nil {
		log.Printf("new sdk logger error: %v \n", err)
		return nil, err
	} // 处理日志打屏问题
	otlpHandler = &OtlpHandler{
		sdkLogger: logger,
	}
	// otel打屏日志问题
	otlpHandler.setOtelLog()

	// 初始化sdk metrics
	err = otlpHandler.initMetrics(conf)
	if err != nil {
		log.Printf("init metrics failed: %v \n", err)
		// metrics初始化失败不影响整体初始化，继续执行
	}

	// 设置全局变量
	globalOtlpHandler = otlpHandler

	log.Println("init otlp ok")
	log.Printf("%#v\n", conf)
	return
}

// grateful offlinetrace、metrcis
func (l *OtlpHandler) Fini() (err error) {
	log.Println("about to fini otlp")

	// 清理metrics
	if l.meterProvider != nil {
		ctx := l.metricsCtx
		if ctx == nil {
			ctx = context.Background() // 兜底使用Background context
		}
		err = metrics.Fini(ctx, l.meterProvider)
		if err != nil {
			log.Printf("fail to fini otlp sdkMetrics: %v", err)
		} else {
			log.Println("fini otlp sdkMetrics ok")
		}
	}

	l.sdkLogger.Close()
	log.Println("fini otlp sdkLogger ok")
	return
}

// NewOtlpQuiverLogger 创建新的 OTLP Quiver Logger
func NewOtlpQuiverLogger() *OtlpQuiverLogger {
	return &OtlpQuiverLogger{
		logger: zaplog.SDKLogger,
	}
}

func (l *OtlpQuiverLogger) Infof(format string, params ...interface{}) {
	if l.logger != nil {
		l.logger.Infof("quiver - "+format, params...)
	}
}

func (l *OtlpQuiverLogger) Debugf(format string, params ...interface{}) {
	if l.logger != nil {
		l.logger.Debugf("quiver - "+format, params...)
	}
}

func (l *OtlpQuiverLogger) Errorf(format string, params ...interface{}) {
	if l.logger != nil {
		l.logger.Errorf("quiver - "+format, params...)
	}
}

// initMetrics 初始化SDK内部metrics
func (otlpHandler *OtlpHandler) initMetrics(conf *OtlpConfig) error {
	if !conf.sdkMetricsEnable {
		return nil
	}

	ctx := context.Background()
	meterProvider, err := metrics.NewMeterProvider(
		metrics.WithReportAddr(conf.sdkMetricsAddr),
		metrics.WithExportInterval(time.Second*10),
		metrics.WithExportTimeOut(time.Second*30),
	)
	if err != nil {
		log.Printf("init sdk metrics error: %v \n", err)
		return err
	}

	otlpHandler.meterProvider = meterProvider
	otlpHandler.metricsCtx = ctx                         // 保存context用于后续销毁
	otlpHandler.serviceName = conf.sdkMetricsServiceName // 保存服务名

	// 创建内部meter，用于SDK内部监控
	otlpHandler.internalMeter = metrics.NewExtendMeter(ctx, conf.sdkMetricsServiceName, meterProvider)

	// 创建OTLP eventlog counter
	otlpHandler.otlpCounter, err = otlpHandler.internalMeter.NewExtendInt64Counter(
		"otlp_eventlog_flush_total",
		"OTLP eventlog flush次数",
		"1",
	)
	if err != nil {
		log.Printf("create otlp counter error: %v \n", err)
		return err
	}

	// 创建APM eventlog counter
	otlpHandler.apmCounter, err = otlpHandler.internalMeter.NewExtendInt64Counter(
		"apm_eventlog_flush_total",
		"APM eventlog flush次数",
		"1",
	)
	if err != nil {
		log.Printf("create apm counter error: %v \n", err)
		return err
	}

	log.Println("init sdk metrics ok")
	return nil
}

// GetOtlpCounter 获取OTLP EventLog计数器（SDK内部使用）
func GetOtlpCounter() *metrics.ExtendInt64Counter {
	if globalOtlpHandler != nil {
		return globalOtlpHandler.otlpCounter
	}
	return nil
}

// GetApmCounter 获取APM EventLog计数器（SDK内部使用）
func GetApmCounter() *metrics.ExtendInt64Counter {
	if globalOtlpHandler != nil {
		return globalOtlpHandler.apmCounter
	}
	return nil
}

// GetServiceName 获取服务名（SDK内部使用）
func GetServiceName() string {
	if globalOtlpHandler != nil {
		return globalOtlpHandler.serviceName
	}
	return ""
}

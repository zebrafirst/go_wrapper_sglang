package th3apiutils

import (
	"os"
	"th3api/common"
	"th3api/config"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   config.GetBaseConf().LogDir,        // 日志文件路径
		MaxSize:    config.GetBaseConf().LogMaxSize,    // 单个日志文件最大大小(MB)
		MaxBackups: config.GetBaseConf().LogMaxBackups, // 保留旧日志文件的最大数量
		MaxAge:     config.GetBaseConf().LogMaxAge,     // 保留旧日志文件的最大天数
		Compress:   config.GetBaseConf().LogCompress,   // 是否压缩/归档旧日志文件
	}
	return zapcore.AddSync(lumberJackLogger)
}

var WLogger *zap.Logger

// 获取日志级别
func getZapLevel(level string) zapcore.Level {
	switch level {
	case common.LOG_LEVEL_DEBUG:
		return zapcore.DebugLevel
	case common.LOG_LEVEL_INFO:
		return zapcore.InfoLevel
	case common.LOG_LEVEL_WARN:
		return zapcore.WarnLevel
	case common.LOG_LEVEL_ERROR:
		return zapcore.ErrorLevel
	case common.LOG_LEVEL_DPANIC:
		return zapcore.DPanicLevel
	case common.LOG_LEVEL_PANIC:
		return zapcore.PanicLevel
	case common.LOG_LEVEL_FATAL:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func NewLocalLog() error {

	// 获取日志级别
	logLevel := getZapLevel(config.GetBaseConf().LogLevel)

	// 文件输出
	fileWriteSyncer := getLogWriter()

	// 自定义时间格式
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006/01/02 15:04:05"))
	}

	// 创建自定义的Encoder配置
	customEncoderConfig := zap.NewProductionEncoderConfig()
	customEncoderConfig.EncodeTime = customTimeEncoder // 设置自定义时间格式

	// 控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(customEncoderConfig)
	consoleWriteSyncer := zapcore.AddSync(os.Stdout)

	// 创建多输出核心
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleWriteSyncer, logLevel),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(customEncoderConfig),
			fileWriteSyncer,
			logLevel,
		),
	)

	logger := zap.New(core, zap.AddCaller())
	defer logger.Sync()
	WLogger = logger
	logger.Info("This message will go to both console and file")
	return nil
}

package mini_log

import (
	"errors"
	"os"
	"strings"
	"time"

	"git.iflytek.com/rdg_ai_services/lumberjack-ccr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	lumberjack    *lumberjack.Logger
	sugaredLogger *zap.SugaredLogger
}

type LogConf struct {
	Level         string // 日志等级 [debug, info, warn, error]，默认为 info
	File          string // 日志地址
	MaxSize       int    // 日志最大大小，单位 MB(megabytes)，默认为100MB
	MaxBackups    int    // 归档日志保存的最大数量，默认保存所有归档日志(0)
	MaxAge        int    // 归档日志最大保存时间，单位 day，默认保存所有归档日志(0)
	LocalTime     bool   // 是否使用本地时间戳，默认使用UTC时间
	Compress      bool   // 归档日志是否进行压缩，默认不压缩
	Async         bool   // 是否允许异步写入日志，当为true时先将日志保存到缓冲区
	CacheMaxCount int    // 缓存区最大大小 单位 B(byte)。当为 0 时不限制缓存区大小
	BatchSize     int    // 触发批量写入日志事件的阈值，单位 B(byte)。当 BatchSize <= 0 时设置为 16KB
	Wash          int    // 将缓存区日志写入文件的协程时间，单位 S(second)
	Caller        bool   // 是否记录调用方的文件名、行号和函数名称注释每条消息，默认为true
	CallerSkip    int    // 值为 0 时直接记录日志输出函数的位置，值为 1 时向上跳 1 层记录调用日志输出函数的位置
}

var (
	pid = int64(os.Getpid())
	c   = &LogConf{
		Level:         "info",
		File:          time.Now().Format("2006-01-02T15-04-05") + ".log",
		MaxSize:       100,
		MaxBackups:    0,
		MaxAge:        0,
		LocalTime:     true,
		Compress:      false,
		Async:         false,
		CacheMaxCount: 0,
		BatchSize:     16,
		Wash:          0,
		Caller:        true,
		CallerSkip:    1,
	}
)

func NewLogger(conf *LogConf) (*Logger, error) {
	var logger *Logger

	if conf == nil {
		conf = c
	}

	conf.Level = strings.ToLower(conf.Level)
	if err := func() error {
		if conf.Level != "info" && conf.Level != "debug" && conf.Level != "warn" && conf.Level != "error" {
			return errors.New("params is illegal")
		} else {
			return nil
		}
	}(); err != nil {
		return nil, err
	}

	if conf.CallerSkip == 0 {
		conf.CallerSkip = 1
	}

	levelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= func() zapcore.Level {
			switch conf.Level {
			case "debug":
				{
					return zapcore.DebugLevel
				}
			case "info":
				{
					return zapcore.InfoLevel
				}
			case "warn":
				{
					return zapcore.WarnLevel
				}
			case "error":
				{
					return zapcore.ErrorLevel
				}
			default:
				{
					return zapcore.ErrorLevel
				}
			}
		}()
	})

	lumberjackLogger := &lumberjack.Logger{
		Filename:      conf.File,
		MaxSize:       conf.MaxSize,
		MaxBackups:    conf.MaxBackups,
		LocalTime:     conf.LocalTime,
		Compress:      conf.Compress,
		MaxAge:        conf.MaxAge,
		Async:         conf.Async,
		CacheMaxCount: conf.CacheMaxCount,
		BatchSize:     conf.BatchSize,
		Wash:          conf.Wash,
	}
	lumberjackLogger.Start()

	writeSyncer := zapcore.AddSync(lumberjackLogger)
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time_key",
		LevelKey:       "level",
		CallerKey:      "caller",
		NameKey:        "logger",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, writeSyncer, levelEnabler),
	)

	if conf.Caller {
		logger = &Logger{
			lumberjack: lumberjackLogger,
			sugaredLogger: zap.New(core,
				zap.AddCaller(),
				zap.AddCallerSkip(conf.CallerSkip),
				zap.Fields(zapcore.Field{
					Key:     "pid",
					Type:    zapcore.Int64Type,
					Integer: pid,
				})).Sugar(),
		}
	} else {
		logger = &Logger{
			lumberjack: lumberjackLogger,
			sugaredLogger: zap.New(core,
				zap.AddCallerSkip(1),
				zap.Fields(zapcore.Field{
					Key:     "pid",
					Type:    zapcore.Int64Type,
					Integer: pid,
				})).Sugar(),
		}
	}
	return logger, nil
}

func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.sugaredLogger.Errorw(msg, keysAndValues...)
}
func (l *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	l.sugaredLogger.Warnw(msg, keysAndValues...)
}
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.sugaredLogger.Infow(msg, keysAndValues...)
}

func (l *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	l.sugaredLogger.Debugw(msg, keysAndValues...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.sugaredLogger.Errorf(template, args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.sugaredLogger.Warnf(template, args...)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.sugaredLogger.Debugf(template, args...)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.sugaredLogger.Infof(template, args...)
}
func (l *Logger) Close() {
	l.lumberjack.Stop()
}

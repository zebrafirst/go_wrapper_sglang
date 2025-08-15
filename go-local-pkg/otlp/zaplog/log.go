package zaplog

import minilog "git.iflytek.com/AIaaS/mini-log"

var (
	SDKLogger *minilog.Logger // sdk自身日志
)

func InitSDKLogger(conf *minilog.LogConf) (logger *minilog.Logger, err error) {
	logger, err = minilog.NewLogger(conf)
	if err != nil {
		return nil, err
	}
	SDKLogger = logger
	return logger, nil
}

func InitBusinessLogger(conf *minilog.LogConf) (logger *minilog.Logger, err error) {
	return minilog.NewLogger(conf)
}

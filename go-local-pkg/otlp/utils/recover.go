package utils

import (
	"git.iflytek.com/AIaaS/otlp-self/v3/zaplog"
	"runtime/debug"
)

func Catch(site string) {
	if err := recover(); err != nil {
		zaplog.SDKLogger.Errorf("Error occur [%v] at [%s] with \n%s.", err, site, string(debug.Stack()))
	}
}

package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"sync"

	"git.iflytek.com/AIaaS/otlp-self/v3/global/kv"
	"git.iflytek.com/AIaaS/otlp-self/v3/utils"
	"git.iflytek.com/AIaaS/otlp-self/v3/zaplog"
	logapi "go.opentelemetry.io/otel/log"
)

const (
	SERVICE_NAME string = "serviceName"
	SID          string = "sid"
	HOST         string = "host"
	TIMESTAMP    string = "timestamp"
)

type OtlpLog struct {
	lock    sync.RWMutex
	logger  logapi.Logger
	message map[string]interface{}
	p       *OtlpLogProvider
}

// 追加 kv 参数到日志内容，并立即刷新到后端。
// kv 为日志的键值对内容。
func (ol *OtlpLog) LogMsgAndFlush(kv ...kv.KV) {
	if ol == nil || ol.message == nil {
		return
	}
	ol.lock.RLock()
	for _, v := range kv {
		ol.message[v.Key] = v.Value
	}
	ol.lock.RUnlock()
	ol.Flush()
}

// 追加 map 类型的日志内容，并立即刷新到后端。
// msgs 为map类型的键值对内容。
func (ol *OtlpLog) LogMsgsAndFlush(msgs map[string]interface{}) {
	if ol == nil || ol.message == nil || msgs == nil {
		return
	}
	ol.lock.RLock()
	maps.Copy(ol.message, msgs)
	ol.lock.RUnlock()
	ol.Flush()
}

// 追加 kv 参数到日志内容，但不会自动刷新到后端，需要手动调用 Flush。
// kv 为日志的键值对内容。
// 返回 OtlpLog 实例指针，便于链式调用。
func (ol *OtlpLog) LogMsg(kv ...kv.KV) *OtlpLog {
	if ol == nil || ol.message == nil || kv == nil {
		return ol
	}
	ol.lock.RLock()
	defer ol.lock.RUnlock()
	for _, v := range kv {
		ol.message[v.Key] = v.Value
	}
	return ol
}

func (ol *OtlpLog) dump(dir string) {
	filename := fmt.Sprintf("%s%s%s", dir, string(os.PathSeparator), "otlplog.log")
	if fp, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666); err == nil {
		// Dump
		msg, _ := json.Marshal(ol.message)
		fmt.Fprintln(fp, string(msg))
		fp.Close()
	} else {
		zaplog.SDKLogger.Errorf("open dump file %s failed with error %v", filename, err)
	}
}

// 刷新日志到后端。
func (ol *OtlpLog) Flush() {
	defer utils.Catch("OtlpLog Flush")
	if ol == nil || ol.p == nil {
		return
	}
	f := func() {
		ctx, cancel := context.WithTimeout(context.Background(), ol.p.flushTimeOut)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				zaplog.SDKLogger.Errorf("otlplog flush err: %v", ctx.Err())
				return
			default:
				logBytes, _ := json.Marshal(ol.message)
				record := logapi.Record{}
				record.SetBody(logapi.BytesValue(logBytes))
				ol.logger.Emit(context.Background(), record)
				// 本地下载验证
				if ol.p.dumpEnable {
					ol.dump(ol.p.dumpDir)
				}
				return
			}
		}
	}
	ol.p.otlpLogFlushPool.Submit(f)
}

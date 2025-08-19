package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"git.iflytek.com/AIaaS/otlp-self/v3/logging/kv"
	"git.iflytek.com/AIaaS/otlp-self/v3/utils"
	"git.iflytek.com/AIaaS/otlp-self/v3/zaplog"
	logapi "go.opentelemetry.io/otel/log"
	"os"
	"sync"
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

// 记录kv 标准使用
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

// 记录map
func (ol *OtlpLog) LogMsgsAndFlush(msgs map[string]interface{}) {
	if ol == nil || ol.message == nil || msgs == nil {
		return
	}
	ol.lock.RLock()
	for k, v := range msgs {
		ol.message[k] = v
	}
	ol.lock.RUnlock()
	ol.Flush()
}

// 此接口需要自己手动flush
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

func (ol *OtlpLog) Flush() {
	defer utils.Catch("otlplog flush")
	if ol == nil || ol.p == nil {
		return
	}
	ol.p.otlpLogFlushPool.AppendJob(&Task{
		f: func() {
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
		},
	})
}

package logging

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"git.iflytek.com/AIaaS/otlp-self/v3/logging/kv"
	"git.iflytek.com/AIaaS/otlp-self/v3/monitor"
	"git.iflytek.com/AIaaS/otlp-self/v3/probe"
	"git.iflytek.com/AIaaS/otlp-self/v3/utils"
	"git.iflytek.com/AIaaS/otlp-self/v3/zaplog"
	logapi "go.opentelemetry.io/otel/log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	IMAGE_MAX_LEN int    = 1024 * 1024
	MEDIA_DATA    string = "media_data"
)

type EventLog struct {
	eventData   *eventData
	media       []*MediaDataWithUrl
	logger      logapi.Logger
	ops         int64
	mediaCtx    context.Context
	mediaCancel context.CancelFunc
	lock        sync.RWMutex
	errs        []error
	p           *EventLogProvider
}

type eventData struct {
	// server type
	Type int64 `json:"type" bson:"type"`
	// sid
	Sid string `json:"sid" bson:"sid"`
	// uid
	Uid string `json:"uid" bson:"uid"`
	// syncid
	Syncid int64 `json:"syncid" bson:"syncid"`
	// sub
	Sub string `json:"sub" bson:"sub"`
	// timestamp
	Timestamp int64 `json:"timestamp" bson:"timestamp"`
	// name
	Name string `json:"name" bson:"name"`
	// endpoint
	Endpoint string `json:"endpoint" bson:"endpoint"`
	// tags
	Tags map[string]string `json:"tags" bson:"tags"`
	// outputs
	Outputs map[string][]string `json:"outputs" bson:"outputs"`
	// desc
	Descs []string `json:"descs" bson:"descs"`
	// media
	Medias []*MediaDataWithUrl `json:"media" bson:"media"`
}

type MediaDataWithUrl struct {
	DataFlag              string                 `json:"data_flag"`
	DataUrl               string                 `json:"data_url"`
	DataTime              string                 `json:"data_time"`
	DataType              string                 `json:"data_type"`
	DataIndex             int                    `json:"data_index"`
	DataClassficationType string                 `json:"data_classfication_type"`
	DataOcrType           string                 `json:"data_ocr_type"`
	ExtendInfos           map[string]interface{} `json:"extend_infos"`
}

type MediaData struct {
	DataFlag              string                 `json:"data_flag"`
	Data                  []byte                 `json:"data"`
	DataTime              string                 `json:"data_time"`
	DataType              string                 `json:"data_type"`
	DataIndex             int                    `json:"data_index"`
	DataClassficationType string                 `json:"data_classfication_type"`
	DataOcrType           string                 `json:"data_ocr_type"`
	ExtendInfos           map[string]interface{} `json:"extend_infos"`
}

func (el *EventLog) Tag(kv ...kv.KV) *EventLog {
	if el == nil || el.eventData == nil {
		return el
	}
	el.lock.Lock()
	defer el.lock.Unlock()

	for _, v := range kv {
		el.eventData.Tag(v)
	}
	return el
}

// add tag value
func (event *eventData) Tag(tag kv.KV) *eventData {
	var value = ""
	switch tag.Value.(type) {
	case bool:
		if tag.Value == true {
			value = "true"
		} else {
			value = "false"
		}
	case int, int64, int32, int16, int8, uint8, uint16, uint32, uint64, uint:
		value = fmt.Sprintf("%d", tag.Value)
	case float32, float64:
		value = fmt.Sprintf("%f", tag.Value)
	case string:
		value = fmt.Sprintf("%s", tag.Value)
	default:
		zaplog.SDKLogger.Errorf("unsupported tag type with key: %v, value: %v", tag.Key, tag.Value)
	}
	event.Tags[tag.Key] = value
	return event
}

func (el *EventLog) Output(k, v string) *EventLog {
	if el == nil || el.eventData == nil {
		return el
	}
	el.lock.Lock()
	defer el.lock.Unlock()

	el.eventData.Output(k, v)
	return el
}

// add outputs
func (event *eventData) Output(key string, value string) *eventData {
	event.Outputs[key] = append(event.Outputs[key], value)
	return event
}

func (el *EventLog) WithUid(uid string) *EventLog {
	if el == nil || el.eventData == nil {
		return el
	}
	el.lock.Lock()
	defer el.lock.Unlock()
	el.eventData.WithUid(uid)
	//el.eventData.WithUid(uid)
	return el
}

func (el *EventLog) Media(media *MediaData) (*EventLog, error) {
	defer utils.Catch("eventlog media")
	if el == nil || el.eventData == nil || el.p == nil {
		return el, errors.New("eventLog is nil or provider is nil")
	}
	if !el.p.mediaEnable {
		zaplog.SDKLogger.Errorf("media drop data with mediaDisable")
		return el, errors.New("media drop data with mediaDisable")
	}
	if media == nil {
		zaplog.SDKLogger.Errorf("media is nil")
		return el, errors.New("media drop data with mediaDisable")
	}
	if len(media.Data) >= (el.p.mediaMaxSize * IMAGE_MAX_LEN) {
		if monitor.SDKMetricsEnable() {
			monitor.OtlpSdkMetrics.RecordCounter("event_media_size_exceed", 1)
		}
		zaplog.SDKLogger.Errorf("media drop data length : %v", len(media.Data))
		return el, errors.New("media drop data length")
	}
	// 计数器加一
	atomic.AddInt64(&el.ops, 1)
	el.p.mediaWorkPool.AppendJob(&Task{
		f: func() {
			ctx, cancel := context.WithTimeout(context.Background(), el.p.mediaUpTimeOut)
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					zaplog.SDKLogger.Errorf("eventlog media upload timeout err: %v", ctx.Err())
					return
				default:
					var err error
					var url string
					if el.p.mediaFunc != nil {
						for i := 0; i < el.p.mediaRetryNum; i++ {
							url, err = el.p.mediaFunc(media)
							if err == nil {
								break
							}
						}
					} else {
						// 使用默认实现
						if probe.MediaOk() {
							now := time.Now()
							for i := 0; i < el.p.mediaRetryNum; i++ {
								url, err = uploadSgw(ctx, media.Data, el.p.sgwNameSpace, el.p.sgwLinkTTL)
								if err == nil {
									break
								}
							}
							// 记录metrics
							if monitor.SDKMetricsEnable() {
								monitor.OtlpSdkMetrics.RecordHistogram("event_media_ts", float64(time.Since(now).Milliseconds()))
							}
						} else {
							err = errors.New("media not ok")
						}
					}
					// 如果以上两种方式都没有成功
					if err != nil {
						// 记录metrics
						if monitor.SDKMetricsEnable() {
							monitor.OtlpSdkMetrics.RecordCounter("event_media_fail", 1)
						}
						el.lock.Lock()
						el.errs = append(el.errs, err)
						el.lock.Unlock()
						if zaplog.SDKLogger != nil {
							zaplog.SDKLogger.Errorf("mediaUpload err : %v", err)
						}
						// 存储网关不可用情况 备份
						if el.p.mediaBackUpEnable {
							el.p.mediaBackUpWorkPool.AppendJob(&Task{
								f: func() {
									el.backupEventMedia(media)
									if monitor.SDKMetricsEnable() {
										monitor.OtlpSdkMetrics.RecordCounter("event_sgw_backup", 1)
									}
									if atomic.AddInt64(&el.ops, -1) <= 0 {
										el.mediaCancel() // 备份完成也算上传完成
									}
									return
								},
							})
							return
						}
					}
					mdata := &MediaDataWithUrl{
						DataUrl:               url,
						DataFlag:              media.DataFlag,
						DataTime:              media.DataTime,
						DataType:              media.DataType,
						DataIndex:             media.DataIndex,
						DataClassficationType: media.DataClassficationType,
						DataOcrType:           media.DataOcrType,
						ExtendInfos:           media.ExtendInfos,
					}
					el.lock.Lock()
					el.media = append(el.media, mdata)
					el.lock.Unlock()
					//	处理完计数器减一 如果此时还有任务在处理 则不取消  flush会等待至超时
					if atomic.AddInt64(&el.ops, -1) <= 0 {
						el.mediaCancel() // 确定上传完成取消
					}
					return
				}
			}
		},
	})
	return el, nil
}

type eventMediaBackUpMeta struct {
	Sid                   string
	TimeStamp             int64
	DataFlag              string
	DataUrl               string
	DataTime              string
	DataType              string
	DataIndex             int
	DataClassficationType string
	DataOcrType           string
	ExtendInfos           map[string]interface{}
}

// el.eventData.Sid, el.eventData.Timestamp, media
func (el *EventLog) backupEventMedia(media *MediaData) {
	base64Str := base64.StdEncoding.EncodeToString(media.Data)
	meta := &eventMediaBackUpMeta{
		Sid:                   el.eventData.Sid,
		TimeStamp:             el.eventData.Timestamp,
		DataUrl:               base64Str, // 工具处理成url替换
		DataFlag:              media.DataFlag,
		DataTime:              media.DataTime,
		DataType:              media.DataType,
		DataIndex:             media.DataIndex,
		DataClassficationType: media.DataClassficationType,
		DataOcrType:           media.DataOcrType,
		ExtendInfos:           media.ExtendInfos,
	}
	el.p.sgwLogger.Errorf("%v", mediaInfoToString(meta))
	return
}

func mediaInfoToString(meta *eventMediaBackUpMeta) string {
	var builder strings.Builder
	builder.WriteString("Sid:")
	builder.WriteString(meta.Sid)
	builder.WriteString("#")
	builder.WriteString("TimeStamp:")
	builder.WriteString(strconv.FormatInt(meta.TimeStamp, 10))
	builder.WriteString("#")
	builder.WriteString("DataFlag:")
	builder.WriteString(meta.DataFlag)
	builder.WriteString("#")
	builder.WriteString("DataUrl:")
	builder.WriteString(meta.DataUrl)
	builder.WriteString("#")
	builder.WriteString("DataTime:")
	builder.WriteString(meta.DataTime)
	builder.WriteString("#")
	builder.WriteString("DataType:")
	builder.WriteString(meta.DataType)
	builder.WriteString("#")
	builder.WriteString("DataIndex:")
	builder.WriteString(strconv.Itoa(meta.DataIndex))
	builder.WriteString("#")
	builder.WriteString("DataClassficationType:")
	builder.WriteString(meta.DataClassficationType)
	builder.WriteString("#")
	builder.WriteString("DataOcrType:")
	builder.WriteString(meta.DataOcrType)
	builder.WriteString("#")

	builder.WriteString("ExtendInfos:")
	if meta.ExtendInfos != nil {
		for k, v := range meta.ExtendInfos {
			builder.WriteString("@")
			builder.WriteString(k)
			builder.WriteString("$")
			builder.WriteString(utils.AnyToString(v))
		}
	}

	return builder.String()
}

func (el *EventLog) Flush() (err error) {
	if el == nil || el.p == nil {
		return
	}
	defer utils.Catch("event flush")
	if len(el.errs) > 0 {
		if err = getErrMsg(el.errs); err != nil {
			zaplog.SDKLogger.Errorf("event flush err: %v", err)
		}
	}
	el.p.eventFlushPool.AppendJob(&Task{
		f: func() {
			// 富媒体文件信息
			ctx, cancel := context.WithTimeout(context.Background(), el.p.flushTimeOut)
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					zaplog.SDKLogger.Errorf("eventlog flush with err: %v", ctx.Err())
					return
				default:
					if el.p.mediaEnable && el.ops > 0 {
						<-el.mediaCtx.Done() //阻塞直到到达超时时间或者收到信号
						el.lock.RLock()
						mediaBytes, _ := json.Marshal(el.media)
						el.lock.RUnlock()
						el.eventData.Tag(kv.KV{MEDIA_DATA, utils.B2S(mediaBytes)})
					}
					// 本地下载验证
					if el.p.eventDumpEnable {
						el.dump(el.p.eventDumpDir)
					}
					el.lock.RLock()
					eventBytes, _ := json.Marshal(el.eventData)
					el.lock.RUnlock()
					if probe.EventOk() {
						record := logapi.Record{}
						record.SetBody(logapi.BytesValue(eventBytes))
						el.logger.Emit(ctx, record)
					} else {
						if el.p.eventBackUpEnable {
							el.p.eventBackUpFlushPool.AppendJob(&Task{
								f: func() {
									el.p.eventLogger.Errorf("%v", utils.B2S(eventBytes))
									if monitor.SDKMetricsEnable() {
										monitor.OtlpSdkMetrics.RecordCounter("event_backup", 1)
									}
								},
							})
						}
					}
					return
				}
			}
		},
	})

	return
}

func getErrMsg(errs []error) error {
	var errMsg string
	for _, e := range errs {
		if e != nil {
			errMsg += e.Error() + ";"
		}
	}
	if errMsg != "" {
		errMsg = errMsg[:len(errMsg)-2]
	}
	return fmt.Errorf(errMsg)
}

func (el *EventLog) dump(dir string) {
	filename := fmt.Sprintf("%s%s%s", dir, string(os.PathSeparator), el.eventData.Sid)
	if fp, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666); err == nil {
		// Dump
		fmt.Fprintln(fp, el.ToString())
		fp.Close()
	} else {
		zaplog.SDKLogger.Errorf("open dump file %s failed with error %v", filename, err)
	}
}

// convert to string in json
func (e *EventLog) ToString() string {
	if res, err := json.MarshalIndent(e.eventData, "", "  "); err == nil {
		return string(res)
	} else {
		zaplog.SDKLogger.Errorf("marshal eventlog data error with :%v", err)
		return ""
	}
}

// add endpoint
func (event *eventData) WithEndpoint(endpoint string) *eventData {
	event.Endpoint = endpoint
	return event
}

// add sub
func (event *eventData) WithSub(sub string) *eventData {
	event.Sub = sub
	return event
}

// add uid
func (event *eventData) WithUid(uid string) *eventData {
	event.Uid = uid
	return event
}

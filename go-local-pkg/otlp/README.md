## OTLP-SDK(V3)使用说明

otlp致力于建设一套可观测系统，提供统一的接入方式，帮助研发和运维提高工作效率，保障线上应用服务稳定运行：

- 提供统一的接入方式采集和上报trace、metrics和log（包含event log）
- 应用服务新增或者扩容时无需手动配置Promethus采集新服务的metrics，提高服务扩容效率

<img src="/Users/zhulin/Library/Application Support/typora-user-images/image-20240718171451566.png" alt="image-20240718171451566" style="zoom:50%;" />

其中，otlp主要提供三种日志功能：

* trace：链路跟踪日志，提供基于 RPC 级业务链路调用性能日志级及相关分析，建议链路组件记录sid(会话唯一标识)，appid，uid，关键入参，响应参数，错误码等信息便于排查链路问题；
* metric：监控日志，提供基于客户端的性能数据指标；
* log：服务日志，用于排障，同时提供 eventlog用于运营管理和数据回流。

### otlp 使用(dev环境)

#### trace

trace查看地址：http://172.30.209.27:16686/  ，使用文档参考仓库中doc中trace使用文档。

示例：

![image-20240612143534116](docs/img/image-20240612143534116.png)

![image-20240612143213475](docs/img/image-20240612143213475.png)

<img src="/Users/zhulin/Library/Application Support/typora-user-images/image-20240612143505063.png" alt="image-20240612143505063" style="zoom:50%;" />

#### metrics

metrics查看地址：http://172.30.209.27:7777/metrics

示例：

![image-20240612143356362](docs/img/image-20240612143356362.png)

可以在grafana对metrics采集展示。

#### log

日志链路使用elk，kibana地址为：http://10.1.87.67:5601/   用户名密码为：elastic/d7mhXaGqsBe65xp0bZdn

![image-20240904153017783](/Users/zhulin/Library/Application Support/typora-user-images/image-20240904153017783.png)

### 使用示例

#### trace

```go
package main

import (
	"git.iflytek.com/AIaaS/otlp/v3/api"
	"git.iflytek.com/AIaaS/otlp/v3/trace"
	"time"
)

func main() {
	//初始化sdk
	otlpHandler, err := api.InitOtlp(api.WithLogLevel("info"), api.WithFileName("./otlp/otlpsdk.log"),
		api.WithSdkMetricsEnable(true), api.WithSdkMetricsAddr("172.30.209.28:4317"))
	if err != nil {
		panic(err)
	}
	defer otlpHandler.Fini()

	// 初始化trace
	// 需要注意参数WithExportBatchSize 批量大小*数据大小不要超过默认的WithMaxGrpcSendSize(4MB)
	// 设业务trace数据大小为10KB,则batchSize最大不应超过4*1024/10=409.6
	traceProvider, err := trace.NewTraceProvider("test-serviceName", "172.30.209.27:4317",
		trace.WithRate(1.0), trace.WithQueueSize(2048), trace.WithMaxGrpcSendSize(4),
		trace.WithExportBatchSize(100), trace.WithExportTimeout(time.Second*30),
		trace.WithBatchTimeout(time.Second*5), trace.WithBlock(false))
	if err != nil {
		panic(err)
	}
	// 逆初始化
	defer traceProvider.Fini()
	// 创建root span
	rootSpan := traceProvider.NewAndStartSpan("root-span")
	rootSpan.WithTag("sid", "root-sid").WithTag("uid", "root-uid").WithStatus(trace.Ok, "is ok")
	time.Sleep(time.Millisecond * 100)
	// 创建子span 一般在同一进程内
	nextSpan := rootSpan.Next("next-span").WithTag("sid", "next-sid").WithTag("uid", "next-uid").WithStatus(trace.Error, "is error")
	time.Sleep(time.Millisecond * 100)
	// 根据父span信息创建子span 一般为跨进程
	fromMetaSpan := traceProvider.NewSpanFromParentMeta(rootSpan.Meta(), "NewSpanFromParentMeta-span").WithTag("sid", "fromMeta-sid").WithTag("uid", "fromMeta-uid").WithStatus(trace.Error, "is error")
	time.Sleep(time.Millisecond * 100)
	// 结束span 上报信息
	fromMetaSpan.End()
	nextSpan.End()
	rootSpan.End()
}
```

#### metrics

```go
package main

import (
	"context"
	"fmt"
	"git.iflytek.com/AIaaS/otlp/v3/api"
	"git.iflytek.com/AIaaS/otlp/v3/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"os"
	"os/signal"
	"time"
)

func main() {
	//初始化sdk
	otlpHandler, err := api.InitOtlp(api.WithLogLevel("info"), api.WithFileName("./otlp/otlpsdk.log"),
		api.WithSdkMetricsEnable(true), api.WithSdkMetricsAddr("172.30.209.28:4317"))
	if err != nil {
		panic(err)
	}
	defer otlpHandler.Fini()

	// 初始化metrics
	// 注意WithExportInterval 尽量不要设置过大，间隔过大会导致metrics堆积超过GrpcSendSize(4MB)，导致上报失败
	ctx := context.Background()
	meterProvider, err := metrics.NewMeterProvider(metrics.WithReportAddr("172.30.209.27:4317"),
		metrics.WithExportInterval(time.Second*1), metrics.WithExportTimeOut(time.Second*30),
		metrics.WithMaxGrpcSendSize(4))
	if err != nil {
		fmt.Printf("NewMeterProvider err: %v \n", err)
	}
	// 逆初始化
	defer metrics.Fini(ctx, meterProvider)

	// 使用方式1 sdk简单封装自定义扩展接口 meter建议保持全局唯一，用于创建不同的metrics
	meter := metrics.NewExtendMeter(ctx, "test-serviceName", meterProvider)
	//使用counter
	int64Counter, err := meter.NewExtendInt64Counter("extend_int64_counter", "extend_int64_counter_desc", "1")
	if err != nil {
		fmt.Println(err)
	}
	// 对应metrics中的lable，如serviceName/serviceId sub dc func appid（不建议埋点，grafana展示压力大）等
	labels := []*metrics.ExtendLables{
		{Key: "test_key1", Val: "text_val1"},
		{Key: "test_key2", Val: "text_val2"},
		{Key: "test_key3", Val: "text_val3"},
	}
	int64Counter.Record(labels, 1)

	// 使用 histogram  记录耗时或者大小分布
	bucket := []float64{0.0, 5.0, 10.0, 25.0, 50.0, 75.0, 100.0, 250.0, 500.0, 750.0, 1000.0, 2500.0, 5000.0, 7500.0, 10000.0, 200000.0}
	extendInt64Histogram, err := meter.NewExtendInt64Histogram("extend_int64_counter", "extend_int64_counter_desc", "1", bucket)
	if err != nil {
		fmt.Println(err)
	}
	labels1 := []*metrics.ExtendLables{
		{Key: "test_key1", Val: "text_val1"},
		{Key: "test_key2", Val: "text_val2"},
		{Key: "test_key3", Val: "text_val3"},
	}
	extendInt64Histogram.Record(labels1, 1)

	// 使用方式二 使用官方原生接口
	m := meterProvider.Meter("serviceName")
	counter, err := m.Int64Counter("counter")
	if err != nil {
		fmt.Println(err)
		return
	}
	opt := metric.WithAttributes(
		attribute.Key("dc").String("hf"),
		attribute.Key("ent").String("ist-upload"),
		attribute.Key("sub").String("ist1"),
	)
	if err == nil {
		counter.Add(ctx, 1, opt)
	}

	ctx, _ = signal.NotifyContext(context.Background(), os.Interrupt)
	<-ctx.Done()
}
```

#### log

```go
package main

import (
	"context"
	"fmt"
	"git.iflytek.com/AIaaS/otlp/v3/api"
	"git.iflytek.com/AIaaS/otlp/v3/logging"
	"git.iflytek.com/AIaaS/otlp/v3/logging/kv"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func main() {
	//初始化sdk 日志文件 和 sdk 监控
	otlpHandler, err := api.InitOtlp(api.WithLogLevel("error"), api.WithFileName("./otlp/otlpsdk.log"),
		api.WithSdkMetricsEnable(true), api.WithSdkMetricsAddr("172.30.209.28:4317"))
	if err != nil {
		panic(err)
	}
	defer otlpHandler.Fini()

	// 服务日志使用 类似于本地日志  上报一行日志
	otlpLogProvider, err := logging.NewOtlpLogProvider("172.30.209.28:4317",
		logging.WithLogServiceName("serviceName"),
		logging.WithLogExportTimeOut(time.Second*30), // 日志上报超时时间
		logging.WithLogExportInterval(time.Second*1), // 日志上报间隔
		logging.WithLogExportMaxBatchSize(100),       //重要重要 日志上报批量大小 根据日志大小设置 参考公式 4（WithLogMaxGrpcSendSize）*1024/日志大小（10KB）= 400   重要
		logging.WithLogMaxQueueSize(2048),            // 日志队列大小 sdk默认2048 超过会丢弃
		logging.WithLogMaxGrpcSendSize(4),            //  grpc默认大小 4MB 该值不建议修改 需要服务端同步修改
		logging.WithLogFlushQueueSize(1),             // flush队列大小 默认10000
		logging.WithLogFlushWorkerNum(1),             // flush 工作协程数 默认 32
		logging.WithLogFlushBlock(false),             // flush 是否阻塞 若设置阻塞，达到flush队列大小 会阻塞刷新
		logging.WithLogFlushTimeOut(time.Second*10),  // flush 超时时间 超时放弃flush
		logging.WithDumpEnable(false))                // 开启dump  开放调试 线上不要开启
	if err != nil {
		fmt.Printf("NewEventProvider: %v \n", err)
		return
	}
	defer otlpLogProvider.Fini()

	//方式一
	for i := 0; i < 1; i++ {
		// 线上需要对数据进行清洗 需要埋点服务名标识服务
		otlpLog, err := otlpLogProvider.OtlpLog("sparkAPI-testOtlpLog", fmt.Sprintf("eee44400001@hf%v", randomString(18)), "127.0.0.1:8080")
		if err != nil {
			fmt.Println(err)
		}
		// 服务日志每个log仅允许调用该接口一次  建议记录 error日志相关信息 其他记录eventlog
		otlpLog.LogMsgAndFlush(kv.KV{Key: "otlpLogFlag", Value: "otlpLogFlag"})
	}

	// 方式二 传map
	log, _ := otlpLogProvider.OtlpLog("sparkAPI", fmt.Sprintf("eee44400001@hf%v",
		randomString(18)), "127.0.0.1:8080")
	log.LogMsgsAndFlush(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": 3,
	})

	// 方式三
	for i := 0; i < 10; i++ {
		otlpLogProvider.Log("sparkAPI-testLog", fmt.Sprintf("eee44400001@hf%v",
			randomString(18)), "127.0.0.1:8080",
			kv.KV{Key: "key1", Value: "value1"},
			kv.KV{Key: "key2", Value: "value2"},
			kv.KV{Key: "key3", Value: 3},
		)
	}

	// 方式四 手动flush
	otlpLogProvider.NewLog("sparkAPI-testNewLog", "sid", "127.0.0.1").LogMsg(
		kv.KV{Key: "key1", Value: "value1"},
		kv.KV{Key: "key2", Value: "value2"},
		kv.KV{Key: "key3", Value: 3}).Flush()

	// 运营日志 EventLog使用
	eventProvider, err := logging.NewEventProvider("172.30.209.28:4319",
		logging.WithExportTimeOut(time.Second*30), // 日志上报超时时间
		logging.WithExportInterval(time.Second),   // 日志上报间隔
		logging.WithExportMaxBatchSize(100),       //重要重要 日志上报批量大小 根据日志大小设置 4（WithLogMaxGrpcSendSize）*1024/日志大小（10KB）= 400
		logging.WithMaxQueueSize(2048),            // 日志队列大小 默认2048 超过会丢弃
		logging.WithMaxGrpcSendSize(4),            // grpc默认大小 4MB 该值不建议修改 需要服务端同步修改
		logging.WithFlushQueueSize(1),             // flush队列大小
		logging.WithFlushWorkerNum(1),             // flush 工作协程数
		logging.WithFlushBlock(false),             // flush 是否阻塞 若设置阻塞，达到flush队列大小 会阻塞刷新
		logging.WithFlushTimeOut(time.Second*10),  // flush 超时时间 超时放弃flush

		logging.WithEventDumpEnable(false), // 是否开启dump 研测环境调试使用  线上不要开启
		// 富媒体上传相关配置
		logging.WithMediaEnable(false),            // 是否使用富媒体文件上传功能
		logging.WithMediaMaxSize(0),               // 富媒体文件最大大小
		logging.WithMediaUpTimeOut(time.Second*3), // 富媒体上传超时时间
		logging.WithMediaQueueSize(1),             // 富媒体文件上传队列大小
		logging.WithMediaWorkerNum(1),             // 富媒体文件上传工作协程数
		logging.WithMediaRetryNum(3),              // 富媒体文件上传最大重试次数
		logging.WithMediaBlock(false),             // 富媒体文件上传是否阻塞
		logging.WithSgwAddress("10.1.87.69:8211"), // 富媒体上传使用存储网关地址
		logging.WithSgwNameSpace("sjliu"),         // 富媒体上传使用存储网关ns
		logging.WithSgwAccessKey("sjliu777"),      // 富媒体上传使用存储网关AccessKey
		logging.WithSgwSecretKey("12345678"),      // 富媒体上传使用存储网关SecretKey
		logging.WithSgwLinkTTL(3600),              // 富媒体上传文件外链保存时间 单位s
		logging.WithSgwLogPath("./sgwsdk.log"),    // 富媒体文件上传日志目录
		logging.WithSgwLogLevel(2),                // // 富媒体文件上传日志级别(兼容sgw sdk) 0 warn 1 info  2 error

		// 备份配置
		// event配置
		logging.WithEventBackUpEnable(true),                   // 是否开启eventlog备份
		logging.WithEventBackUpPath("./otlp/event/event.log"), //备份日志地址
		logging.WithEventProbeAddr("172.30.209.28:4319"),      //备份使用探测地址
		logging.WithEventProbeTimeOut(time.Second*10),         //备份使用探测超时时间
		logging.WithEventProbeInterVal(time.Second*30),        //备份使用探测间隔
		// 富媒体配置
		logging.WithMediaBackUpEnable(true),               // 是否开启富媒体备份
		logging.WithMediaBackUpPath("./otlp/sgw/sgw.log"), //富媒体备份日志地址
		logging.WithMediaProbeAddr("10.1.87.69:8211"),     //备份使用探测地址 研测使用实际地址 线上端口为域名+80
		logging.WithMediaProbeTimeOut(time.Second*10),     //备份使用探测超时时间
		logging.WithMediaProbeInterVal(time.Second*30),    //备份使用探测间隔
	)

	if err != nil {
		fmt.Printf("NewEventProvider: %v \n", err)
		return
	}

	defer eventProvider.Fini()

	for i := 0; i < 10; i++ {
		eventLog := eventProvider.EventLog("sparkAPI", fmt.Sprintf("%de44400001@hf%v", i, randomString(18)), "cht", "127.0.0.1:8080")
		eventLog.WithUid("uid123")
		eventLog.Tag(kv.KV{Key: "eventFlag", Value: "eventFlag"})
		eventLog.Tag(kv.KV{Key: "answerToken", Value: 452})
		eventLog.Tag(kv.KV{Key: "appid", Value: "12a0a7e2"})
		eventLog.Tag(kv.KV{Key: "auditAnswerPerf", Value: 168})
		eventLog.Tag(kv.KV{Key: "auditQuestion", Value: "pass"})
		eventLog.Tag(kv.KV{Key: "classificationPerf", Value: 113})
		eventLog.Tag(kv.KV{Key: "clientFr", Value: 18665})

		eventLog.Output("output_key1", "output_value1").Output("output_key1", "output_value2").Output("output_key1", "output_value3")

		media1 := &logging.MediaData{
			DataFlag:              "dataFlagValue",
			Data:                  []byte("test_media" + strconv.Itoa(i)),
			DataTime:              time.Now().String(),
			DataType:              "answer",
			DataIndex:             i,
			DataClassficationType: "class",
			DataOcrType:           "ocr",
			ExtendInfos:           map[string]interface{}{"index": i, "key1": "value1"},
		}
		eventLog, err = eventLog.Media(media1)
		if err != nil {
			fmt.Printf("media err:%v \n", err)
			return
		}
		media2 := &logging.MediaData{
			Data:        []byte("test_media" + strconv.Itoa(i)),
			ExtendInfos: map[string]interface{}{"index": i + 1},
		}
		eventLog, err = eventLog.Media(media2)
		if err != nil {
			fmt.Printf("media err:%v \n", err)
			return
		}
		eventLog.Tag(kv.KV{Key: "domain", Value: "max-32k"})
		eventLog.Tag(kv.KV{Key: "domain1", Value: "max-32k"})
		eventLog.Tag(kv.KV{Key: "env", Value: "dx-v3-5pro"})
		eventLog.Tag(kv.KV{Key: "fafr", Value: 15177})
		eventLog.Tag(kv.KV{Key: "logType", Value: "session"})
		eventLog.Tag(kv.KV{Key: "perf", Value: 33730})

		if err = eventLog.Flush(); err != nil {
			fmt.Printf("flush err: %v \n", err)
		}
	}

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	<-ctx.Done()
}

const charset = "abcdefghijklmnopqrstuvwxyz"

func randomString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}
```





### 配置说明

```toml
# trace 相关配置
reportAddr      string //上报ip 生产环境配置127.0.0.1:4317
queueSize       int   // trace队列大小 默认值 2048 建议大于服务qps
rate            float64 // 采样比例，默认 1.0，全部采样
batchTimeout    int  // 批量上报时间 默认5s 
exportTimeout   int // 上报超时时间，默认30s
exportBatchSize int // 默认批量上报 400个span  8kb会小于4mb  注意 批量发送数据总大小超过maxGrpcSendSize（4MB）会失败，建议根据日志数据大小进行调整
maxGrpcSendSize int // grpc客户端数据限制大小
block           bool // 是否阻塞，设为false，超过queueSize会丢弃

# metrics 相关配置
reportAddr      string //上报ip 生产环境配置127.0.0.1:4317
maxGrpcSendSize int    // grpc客户端数据限制大小
exportTimeOut   int    // 上报超时时间，默认30s
exportInterval  int    // 上报间隔时间，默认1s

# log相关配置
exportInterval     time.Duration // 上报间隔  默认1s
exportTimeOut      time.Duration // 上报超时时间 默认30s
exportMaxBatchSize int           // 批量发送大小默认40  批量大小*日志大小需要小于maxGrpcSendSize（默认 4MB）
maxQueueSize       int           // 日志队列大小 默认2048
maxGrpcSendSize    int           // grpc客户端数据限制大小
flushQueueSize     int           // flush队列大小
flushWorkerNum     int           // flush工作协程数
flushBlock         bool          // flush是否阻塞 默认非阻塞
flushTimeOut       time.Duration

// eventlog 配置  运营日志相关
exportInterval     time.Duration // 上报间隔  默认1s
exportTimeOut      time.Duration // 上报超时时间 默认30s
exportMaxBatchSize int           // 批量发送大小默认40
maxQueueSize       int           // 日志队列大小 默认2048
maxGrpcSendSize    int           // grpc客户端数据限制大小
flushQueueSize     int           // flush队列大小 备份复用
flushWorkerNum     int           // flush工作协程数
flushBlock         bool          // flush是否阻塞 默认非阻塞
flushTimeOut       time.Duration

eventDumpEnable bool // 是否开启dump 默认为false 研测调试，线上不要开启
// 富媒体配置
mediaEnable    bool          // 是否使用富媒体上传
mediaMaxSize   int           // 富媒体文件上传限制大小 默认30MB
mediaUpTimeOut time.Duration // 上传富媒体文件超时时间
mediaRetryNum  int           // 上传重试次数
// 自定义func
mediaFunc      MediaFunc
mediaQueueSize int
mediaWorkerNum int
mediaBlock     bool // 上传富媒体文件是否阻塞

sgwAccessKey string // 存储网关四元组
sgwSecretKey string
sgwAddress   string
sgwNameSpace string
sgwLinkTTL   int // 外链生效时间 重要 单位s
sgwLogPath   string
sgwLogLevel  int // 0 warn 1 info 2 error

// event备份相关配置
eventBackUpEnable  bool
eventBackUpPath    string
eventProbeAddr     string
eventProbeTimeOut  time.Duration
eventProbeInterVal time.Duration

// 富媒体备份相关配置  支持使用存储网关的探测和
mediaBackUpEnable  bool
mediaBackUpPath    string
mediaProbeAddr     string
mediaProbeTimeOut  time.Duration
mediaProbeInterVal time.Duration

```

### 支持&维护

陈建（jianchen15/18721197851） 李远川（ycli15/18326686970） 金诗文（swjin/18305628595）  汪建圩 （jwwang48/18855038797） 朱琳（linzhu13/15651900298）


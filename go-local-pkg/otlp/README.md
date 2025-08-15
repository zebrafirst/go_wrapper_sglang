## EventLog-SDK(ASE)使用说明

基于OTLP-SDK的event_log和log修改，提供统一的接口兼容OTLP和旧有APM(quiver)的运营日志event_log，以及OTLP服务日志log的上报，同时APM的日志重定向到OtlpHandler中。

和OTLP-SDK相比没有日志和富媒体本地备份功能。

OTLP上报的日志可以在Kibana中查看到数据：

![OTLP上报日志](./docs/images/Kibana监控.png)

APM（quiver）上报的日志在APM监控台查看：

![APM监控](./docs/images/apm监控.png)

### SDK使用示例(dev环境)

OTLP链路的event_log使用elk，kibana地址为：http://10.104.57.188/zone-hfyc-1/24384/es-cVfE8KsVsN/kibana/app/discover/  用户名密码为：esadmin/d7mhXaGqsBe65xp0bZdn

Grafana监控地址：https://hawkeye.xfyun.cn/d/a45992d7-46b5-4ccc-81dc-d35780f3a249/otlp-collector?orgId=1

log和event的示例代码位置如下所示：

```md
OTLP
├─...
├─examples      // 调用示例
│  ├─eventlog   // event log调用示例
│  ├─log        // log调用示例
│  └─main.go    // SDK完整调用示例
└─...
```

### 配置说明

```toml
# BEGIN OTLP log 配置
[OTLP_log]
	serviceName        string
	exportInterval     time.Duration  # 上报间隔  默认1s
	exportTimeOut      time.Duration  # 上报超时时间 默认30s
	exportMaxBatchSize int            # 批量发送大小默认
	maxQueueSize       int            # 日志队列大小
	maxGrpcSendSize    int            # grpc客户端数据限制大小
	flushWorkerNum     int            # flush工作协程数
	finiEnableWait     bool           # flush是否阻塞 默认非阻塞
	flushTimeOut       time.Duration
	logDumpEnable      bool           # 是否开启dump 默认为false 研测调试，线上不要开启
# END OTLP log 配置

# BEGIN OTLP eventlog 配置
[OTLP_evetlog]
	exportInterval     time.Duration # 上报间隔  默认1s
	exportTimeOut      time.Duration # 上报超时时间 默认30s
	exportMaxBatchSize int           # 批量发送大小默认40
	maxQueueSize       int           # 日志队列大小 默认2048
	maxGrpcSendSize    int           # grpc客户端数据限制大小
	flushWorkerNum     int           # flush工作协程数
	finiEnableWait     bool          # flush是否阻塞 默认非阻塞
	flushTimeOut       time.Duration

	eventDumpEnable bool # 是否开启dump 默认为false 研测调试，线上不要开启
	# 富媒体配置
	mediaEnable    bool          # 是否使用富媒体上传
	mediaMaxSize   int           # 富媒体文件上传限制大小 默认30MB
	mediaUpTimeOut time.Duration # 上传富媒体文件超时时间
	mediaRetryNum  int           # 上传重试次数
	# 自定义func
	mediaFunc      MediaFunc
	mediaQueueSize int
	mediaWorkerNum int
	mediaBlock     bool # 上传富媒体文件是否阻塞

	sgwAccessKey string # 存储网关四元组
	sgwSecretKey string
	sgwAddress   string
	sgwNameSpace string
	sgwLinkTTL   int # 外链生效时间 重要 单位s
	sgwLogPath   string
	sgwLogLevel  int # 0 warn 1 info 2 error

	# event备份相关配置
	eventBackUpEnable  bool
	eventBackUpPath    string
	eventProbeAddr     string
	eventProbeTimeOut  time.Duration
	eventProbeInterVal time.Duration
	# 富媒体备份相关配置  支持使用存储网关的探测
	mediaBackUpEnable  bool
	mediaBackUpPath    string
	mediaProbeAddr     string
	mediaProbeTimeOut  time.Duration
	mediaProbeInterVal time.Duration
# END OTLP eventlog 配置

# BEGIN APM (quiver) 配置
[apm]
	eventHost			string
	eventPort			string
	eventConsumerNum	int		# 消费者数量
	s3AccessKey			string
	s3SecretKey			string
	s3Endpoint			string
	hbaseZKHosts		string	
	eventDumpEnable		bool	# 开启开启本地 dump
	eventDumpDir		string	# dump到本地文件目录，默认 ./dump
	eventSpillEnable	bool	# 是否溢出兜底到本地
	eventSpillDir		string	# 溢出兜底到本地目录
	buffSize			int32	# 缓冲区大小
	lingerSec			int		# 批次发送前的最大等待时间，单位s
	batchSize			int		# 批量大小
	maxSpillContentSize int64 	# 最大溢出大小 单位 GB
	flushRetryCount     int
# END APM (quiver) 配置

```

### 支持&维护

陈建（jianchen15/18721197851） 李远川（ycli15/18326686970） 金诗文（swjin/18305628595）  汪建圩 （jwwang48/18855038797） 朱琳（linzhu13/15651900298）


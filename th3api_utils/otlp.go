package th3apiutils

import (
	"th3api/config"
	"time"

	"git.iflytek.com/AIaaS/otlp-self/v3/logging"
)

var OtlpLogProvider *logging.OtlpLogProvider

func InitOtlp() error {
	// //初始化sdk 日志文件 和 sdk 监控
	// OtlpHandler, err := api.InitOtlp(api.WithLogLevel("error"), api.WithFileName("./otlp/otlpsdk.log"),
	// 	api.WithSdkMetricsEnable(true), api.WithSdkMetricsAddr("172.30.209.28:4317"))
	// if err != nil {
	// 	panic(err)
	// }
	// defer otlpHandler.Fini()

	// 服务日志使用 类似于本地日志  上报一行日志
	var err error
	OtlpLogProvider, err = logging.NewOtlpLogProvider(config.GetBaseConf().OtlpLogAddr,
		logging.WithLogServiceName(config.GetBaseConf().ServiceId),
		logging.WithLogExportTimeOut(time.Second*30), // 日志上报超时时间
		logging.WithLogExportInterval(time.Second*1), // 日志上报间隔
		logging.WithLogExportMaxBatchSize(100),       //重要重要 日志上报批量大小 根据日志大小设置 参考公式 4（WithLogMaxGrpcSendSize）*1024/日志大小（10KB）= 400   重要
		logging.WithLogMaxQueueSize(2048),            // 日志队列大小 sdk默认2048 超过会丢弃
		logging.WithLogMaxGrpcSendSize(4),            //  grpc默认大小 4MB 该值不建议修改 需要服务端同步修改
		logging.WithLogFlushWorkerNum(32),            // flush 工作协程数 默认 32
		logging.WithLogFlushTimeOut(time.Second*10),  // flush 超时时间 超时放弃flush
	)
	if err != nil {
		return err
	}
	return nil

	// // 方式二 传map
	// log, _ := OtlpLogProvider.OtlpLog("sparkAPI", fmt.Sprintf("eee44400001@hf%v",
	// 	RandomString(18)), "127.0.0.1:8080")
	// log.LogMsgsAndFlush(map[string]interface{}{
	// 	"key1": "value1",
	// 	"key2": "value2",
	// 	"key3": 3,
	// })

}

func GetOtlpLog(sid string) (*logging.OtlpLog, error) {
	return OtlpLogProvider.OtlpLog(config.GetBaseConf().ServiceId, sid, config.GetBaseConf().OtlpLogHost)
}

func FiniOtlp() {
	OtlpLogProvider.Fini()
}

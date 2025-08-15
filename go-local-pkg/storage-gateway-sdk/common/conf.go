package common

import (
	finderHttp "git.iflytek.com/AIaaS/finderhttp-self"
)

type CliConf struct {
	UseFinder bool                    // 是否使用配置中心
	Address   []string                // 存储网关地址，使用直连地址时，地址必须为 ip:port
	FinderURL string                  // 配置中心地址 如：10.1.87.79:6868/AIaaS/dx/webgate/1.0.9
	LB        finderHttp.Lb           // 负载均衡策略
	Ping      func(addr string) error // 健康检查探测器，默认使用 net.Dial
	LogName   string                  // 日志存储地址
	LogLevel  int                     // 日志等级。0: debug; 1: info; 2: error
	UseTLS    bool                    // 是否启动 TLS 加密协议
	TimeOut   int                     // http请求的超时时间，单位 s
}

type Auth struct {
	ApiKey    string // 网关校验信息
	ApiSecret string // 网关校验信息
}

func CheckConf(conf *CliConf) {
	if conf.LogLevel < 0 || conf.LogLevel > 2 {
		conf.LogLevel = 1
	}
	if conf.TimeOut < 0 {
		conf.TimeOut = 0
	}
}

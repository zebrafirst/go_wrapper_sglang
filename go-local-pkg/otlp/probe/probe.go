package probe

import (
	"fmt"
	"git.iflytek.com/AIaaS/otlp-self/v3/monitor"
	"git.iflytek.com/AIaaS/otlp-self/v3/zaplog"
	"log"
	"net"
	"runtime/debug"
	"sync/atomic"
	"time"
)

var (
	otlpOkFlag  uint32 = 1
	eventOkFlag uint32 = 1
	sgwOkFlag   uint32 = 1
)

var retryNum = 3

func OtlpOk() bool {
	return atomic.LoadUint32(&otlpOkFlag) != 0
}

func setOtlpOkFlag(value bool) {
	if value {
		atomic.StoreUint32(&otlpOkFlag, 1) // ok
	} else {
		atomic.StoreUint32(&otlpOkFlag, 0)
	}
}

func EventOk() bool {
	return atomic.LoadUint32(&eventOkFlag) != 0
}

func setEventOkFlag(value bool) {
	if value {
		atomic.StoreUint32(&eventOkFlag, 1) // ok
	} else {
		atomic.StoreUint32(&eventOkFlag, 0)
	}
}

func MediaOk() bool {
	return atomic.LoadUint32(&sgwOkFlag) != 0
}

func setSgwOkFlag(value bool) {
	if value {
		atomic.StoreUint32(&sgwOkFlag, 1) // ok
	} else {
		atomic.StoreUint32(&sgwOkFlag, 0)
	}
}

// addr 是ip和端口 组成 判断是否做解析 去掉http
func ProbeMedia(quit chan struct{}, addr string, dialTmo, probeInterVal time.Duration) {
	ticker := time.NewTicker(probeInterVal)
	go func() {
		defer catch("probe media")
		for {
			select {
			case <-ticker.C:
				var err error
				for i := 0; i < retryNum; i++ {
					err = probeTCP(addr, dialTmo)
					if err == nil {
						break
					}
				}
				if err != nil {
					zaplog.SDKLogger.Errorf("ProbeSgw err %v", err)
					setSgwOkFlag(false)
					if monitor.SDKMetricsEnable() {
						monitor.OtlpSdkMetrics.RecordGauge("probe_media", addr, int64(0))
					}
				} else {
					setSgwOkFlag(true)
					if monitor.SDKMetricsEnable() {
						monitor.OtlpSdkMetrics.RecordGauge("probe_media", addr, int64(1))
					}
				}
			case <-quit:
				log.Println("stop ProbeSgw ticker")
				ticker.Stop()
				return
			}
		}
	}()
}

func ProbeLogCollector(quit chan struct{}, addr string, dialTmo, probeInterVal time.Duration) {
	ticker := time.NewTicker(probeInterVal)
	go func() {
		defer catch("probe log")
		for {
			select {
			case <-ticker.C:
				var err error
				for i := 0; i < retryNum; i++ {
					err = probeTCP(addr, dialTmo)
					if err == nil {
						break
					}
				}
				if err != nil {
					setOtlpOkFlag(false)
					if monitor.SDKMetricsEnable() {
						monitor.OtlpSdkMetrics.RecordGauge("probe_log", addr, int64(0))
					}
					zaplog.SDKLogger.Errorf("ProbeCollector err %v", err)
				} else {
					setOtlpOkFlag(true)
					if monitor.SDKMetricsEnable() {
						monitor.OtlpSdkMetrics.RecordGauge("probe_log", addr, int64(1))
					}
				}
			case <-quit:
				fmt.Println("stop ProbeLogCollector ticker")
				ticker.Stop()
				return
			}
		}
	}()
}

func ProbeEventCollector(quit chan struct{}, addr string, dialTmo, probeInterVal time.Duration) {
	ticker := time.NewTicker(probeInterVal)
	go func() {
		defer catch("probe event")
		for {
			select {
			case <-ticker.C:
				var err error
				for i := 0; i < retryNum; i++ {
					err = probeTCP(addr, dialTmo)
					if err == nil {
						break
					}
				}
				if err != nil {
					setEventOkFlag(false)
					if monitor.SDKMetricsEnable() {
						monitor.OtlpSdkMetrics.RecordGauge("probe_event", addr, int64(0))
					}
					zaplog.SDKLogger.Errorf("ProbeEventCollector err %v", err)
				} else {
					setEventOkFlag(true)
					if monitor.SDKMetricsEnable() {
						monitor.OtlpSdkMetrics.RecordGauge("probe_event", addr, int64(1))
					}
				}
			case <-quit:
				log.Println("stop ProbeEventCollector ticker")
				ticker.Stop()
				return
			}
		}
	}()
}

func probe(network string, addr string, tmo time.Duration) error {
	if c, err := net.DialTimeout(network, addr, tmo); err != nil {
		return err
	} else {
		c.Close()
		return nil
	}
}

func probeTCP(addr string, tmo time.Duration) error {
	return probe("tcp", addr, tmo)
}

func probeUDP(addr string, tmo time.Duration) error {
	return probe("udp", addr, tmo)
}

func catch(site string) {
	if err := recover(); err != nil {
		zaplog.SDKLogger.Errorf("Error occur [%v] at [%s] with \n%s.", err, site, string(debug.Stack()))
	}
}

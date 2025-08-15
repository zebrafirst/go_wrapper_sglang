package finderhttp

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"strings"

	"git.iflytek.com/AIaaS/finder-go-self"
	"git.iflytek.com/AIaaS/finder-go-self/common"
)

type Request struct {
	Path    string
	Method  string
	Headers map[string]string
	Body    io.Reader

	// 设置请求的ContentLength
	ContentLength int
	// Handler ,会在请求前执行，可以在这里添加一些执行逻辑，例如鉴权
	Handlers []Handler
	// 指定本次使用的lb，LB 不用每次请求都创建新的实例，默认为 ClientOptions 中指定的LB
	Lb Lb
	// 指定远程的地址,如果不指定默认为从服务发现使用lb选择的地址
	PeerAddr string
}

type Handler func(r *http.Request, url string) error

type Lb interface {
	Get(ctx context.Context, addr []*Target) (string, error)
}

type Target struct {
	Address      string
	failCount    atomic.Int64
	AddressSum32 uint32
}

func (t *Target) String() string {
	return fmt.Sprintf("%s|%v", t.Address, t.AddressSum32)
}
func newTarget(addr string) *Target {
	return &Target{
		Address:      addr,
		AddressSum32: crc32.ChecksumIEEE(BytesOf(addr)),
	}
}

func targetsOf(addr []string) []*Target {
	res := make([]*Target, 0, len(addr))
	for _, v := range addr {
		res = append(res, &Target{
			Address: v,
		})
	}
	return res
}

type ClientOptions struct {
	Address   []string                // 远程地址，如果没有配置FinderUrl ，会使用这个地址访问
	FinderUrl string                  //配置中心地址：例如  10.1.87.79:6868/AIaaS/dx/webate/1.0.9
	LB        Lb                      // 负载均衡策略，支持RoundRobin 和 Hash
	Ping      func(addf string) error // 健康检查探测器
	Logger    Logger                  // 日志接口，默认INFO 级别

	UseTLS bool // 是否启用tls

	HttpClient *http.Client
}

type clientPool struct {
	// 不健康的地址集合
	unhealthy SyncMapList[string, *Target]

	// 当前使用的地址集合
	address SyncMapList[string, *Target]

	// lb选择算法
	lb Lb

	httpcli *http.Client

	opt *ClientOptions

	done   chan struct{}
	closed atomic.Bool

	logger Logger

	ping func(addf string) error
}

type Client interface {
	Request(ctx context.Context, req *Request) (resp *Response, err error)
	//根据负载均衡算法获取一个可用的地址
	GetAvailablePeerAddr(ctx context.Context) (string, error)

	GetAvailablePeerAddrWithLB(ctx context.Context, lb Lb) (string, error)

	NetDialContext(hosts []string) func(ctx context.Context, network string, addr string) (net.Conn, error)

	NetDial(hosts []string, timeout time.Duration) func(network string, addr string) (net.Conn, error)

	Destory()
}

// 创建一个可以通过服务发现调用http接口的client，全局创建一个实例即可
func NewClient(opt *ClientOptions) (Client, error) {
	c := &clientPool{
		opt:  opt,
		done: make(chan struct{}),
	}

	err := c.init()
	if err != nil {
		return nil, err
	}
	return c, nil

}

func (c *clientPool) init() error {
	err := c.checkInitOptions()
	if err != nil {
		return err
	}

	c.logger = c.opt.Logger
	if c.logger == nil {
		c.logger, err = NewLogger("./finderhttp.log", LogInfo)
		if err != nil {
			return err
		}
	}

	if c.opt.FinderUrl != "" {
		err = c.initWithCenter()
		if err != nil {
			return err
		}
	}
	for _, v := range c.opt.Address {
		c.address.Store(v, newTarget(v))
	}

	c.lb = c.opt.LB
	if c.lb == nil {
		c.lb = &roundRobin{}
	}

	c.httpcli = c.opt.HttpClient
	if c.httpcli == nil {
		c.httpcli = &http.Client{}
	}

	c.ping = c.opt.Ping
	if c.ping == nil {
		c.ping = telnet
	}

	go c.healthyCheck()
	go c.unhealthyCheck()
	return nil
}

func (c *clientPool) checkInitOptions() error {
	if c.opt == nil {
		return fmt.Errorf("options is nil")
	}
	o := c.opt
	if o.FinderUrl == "" {
		if len(o.Address) == 0 {
			return fmt.Errorf("storageGateway is empty when not use center")
		}
	}
	return nil
}

func (c *clientPool) initWithCenter() error {
	schema := ""
	url := c.opt.FinderUrl
	if strings.HasPrefix(url, "http://") {
		schema = "http://"
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		schema = "https://"
		url = strings.TrimPrefix(url, "https://")
	} else {
		schema = "http://"
	}

	ls := strings.Split(url, "/")
	if len(ls) != 5 {
		return errors.New("invalid finder url:" + c.opt.FinderUrl)
	}

	fd, err := finder.NewFinderWithLogger(common.BootConfig{
		CompanionUrl:  schema + ls[0],
		CachePath:     "./findercache",
		CacheConfig:   false,
		CacheService:  false,
		ExpireTimeout: 0,
		MeteData: &common.ServiceMeteData{
			Project: ls[1],
			Group:   ls[2],
			Service: ls[3],
			Version: "1.0.0",
			Address: "",
		},
	}, nil)
	if err != nil {
		return err
	}
	services, err := fd.ServiceFinder.UseAndSubscribeService([]common.ServiceSubscribeItem{
		{
			ServiceName: ls[3],
			ApiVersion:  ls[4],
		},
	}, c)
	if err != nil {
		return err
	}
	// sv := services[cf.Service]
	for _, sv := range services {
		for _, s := range sv.ProviderList {
			c.address.Store(s.Addr, newTarget(s.Addr))
			c.logger.Info("[subscribe] addr:", s.Addr)
		}
	}

	return nil
}

// 服务实例上的配置信息发生变化
func (c *clientPool) OnServiceInstanceConfigChanged(name string, apiVersion string, addr string, config *common.ServiceInstanceConfig) bool {
	c.logger.Info("[service instance config changed]", name, apiVersion, addr, config.UserConfig)
	return true
}

// 服务整体配置信息发生变化
func (c *clientPool) OnServiceConfigChanged(name string, apiVersion string, config *common.ServiceConfig) bool {
	c.logger.Info("[service config changed]", name, apiVersion, config.JsonConfig)
	return true
}

// 服务实例发生变化
func (c *clientPool) OnServiceInstanceChanged(name string, apiVersion string, eventList []*common.ServiceInstanceChangedEvent) bool {

	for _, e := range eventList {
		switch e.EventType {
		case common.INSTANCEREMOVE:
			for _, s := range e.ServerList {
				c.address.Delete(s.Addr)
				c.unhealthy.Delete(s.Addr)
				c.logger.Info("[instance remove]", s.Addr)
			}
		case common.INSTANCEADDED:
			for _, s := range e.ServerList {
				c.address.Store(s.Addr, newTarget(s.Addr))
				c.logger.Info("[instance add]", s.Addr)
			}
		}
	}
	return true
}

func (c *clientPool) NetDialContext(hosts []string) func(ctx context.Context, network string, addr string) (net.Conn, error) {
	d := &net.Dialer{}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		ip, _, _ := strings.Cut(addr, ":")
		// 当前是ipv4，直接dial
		if net.ParseIP(ip).To4() != nil {
			return d.DialContext(ctx, network, addr)
		}

		hasHost := false
		for _, v := range hosts {
			if v == ip || v == "*" {
				hasHost = true
				break
			}
		}

		if !hasHost {
			return d.DialContext(ctx, network, addr)
		}

		laddr, err := c.GetAvailablePeerAddr(ctx)
		if err != nil {
			return nil, err
		}
		return d.DialContext(ctx, network, laddr)

	}
}

func (c *clientPool) NewHttpClient(hosts []string) *http.Client {
	cli := &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, DialContext: c.NetDialContext(hosts)}}
	return cli
}

func (c *clientPool) NetDial(hosts []string, timeout time.Duration) func(network string, addr string) (net.Conn, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	d := c.NetDialContext(hosts)
	return func(network, addr string) (net.Conn, error) {
		ctx, cf := context.WithTimeout(context.Background(), timeout)
		defer cf()
		return d(ctx, network, addr)
	}
}

type Response struct {
	*http.Response
	RemoteAddr string
}

func (c *clientPool) Request(ctx context.Context, req *Request) (resp *Response, err error) {
	lb := req.Lb
	if lb == nil {
		lb = c.lb
	}
	peerAddr := req.PeerAddr
	if peerAddr == "" {
		peerAddr, err = c.GetAvailablePeerAddrWithLB(ctx, lb)
		if err != nil {
			return nil, err
		}
	}

	schema := "http://"
	if c.opt.UseTLS {
		schema = "https://"
	}
	httpUrl := schema + peerAddr + req.Path
	r, err := http.NewRequestWithContext(ctx, req.Method, httpUrl, req.Body)
	if err != nil {
		c.logger.Error("[create request error]", err)
		return nil, err
	}

	for k, v := range req.Headers {
		r.Header.Set(k, v)
	}

	for _, hd := range req.Handlers {
		err = hd(r, httpUrl)

		if err != nil {
			return nil, err
		}
	}

	start := time.Now()
	r.ContentLength = int64(req.ContentLength)
	resp = new(Response)
	resp.Response, err = c.httpcli.Do(r)
	if err != nil {
		c.logger.Error("[request error]", httpUrl, err)
		return nil, err
	}
	resp.RemoteAddr = peerAddr
	c.logger.Debug("[request]", resp.StatusCode, "url", httpUrl, "length", resp.ContentLength, "cost", time.Since(start))
	return
}

// 获取一个可以用的远程地址
func (c *clientPool) GetAvailablePeerAddr(ctx context.Context) (string, error) {
	return c.GetAvailablePeerAddrWithLB(ctx, c.lb)
}

func (c *clientPool) GetAvailablePeerAddrWithLB(ctx context.Context, lb Lb) (string, error) {
	return lb.Get(ctx, c.address.ValList())
}

func (c *clientPool) healthyCheck() {
	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
		case <-c.done:
			return
		}
		for _, addr := range c.address.ValList() {
			err := c.ping(addr.Address)
			var fc int64
			if err != nil {
				fc = addr.failCount.Add(1)
			} else {
				addr.failCount.Store(0)
			}
			if fc > 3 {
				c.address.Delete(addr.Address)
				c.unhealthy.Store(addr.Address, addr)
			}
		}
	}
}

func (c *clientPool) deleteAddr(addr string) {
	c.address.Delete(addr)
}

func (c *clientPool) unhealthyCheck() {
	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
		case <-c.done:
			return
		}
		for _, addr := range c.unhealthy.ValList() {
			err := c.ping(addr.Address)

			if err != nil {
				addr.failCount.Add(1)
				c.logger.Error("healthy check error:", addr.Address, err)
			} else {
				addr.failCount.Store(0)
				c.unhealthy.Delete(addr.Address)
				c.address.Store(addr.Address, addr)
				c.logger.Info("healthy check ok:", addr.Address)
			}

		}
	}
}

func (c *clientPool) Destory() {
	if !c.closed.CompareAndSwap(false, true) {
		return
	}
	close(c.done)
}

func telnet(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

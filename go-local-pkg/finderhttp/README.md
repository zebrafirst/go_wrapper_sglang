# finderhttp
支持通过服务发现调用http 服务
#### 使用示例

##### 示例1
````go
	log, err := NewLogger("./finderhttp.log", LogInfo)
	if err != nil {
		panic(err)
	}
	cli, err := NewClient(&ClientOptions{
		FinderUrl: "10.1.87.69:6868/guiderAllService/gas/storage-gateway/1.0.0",
		Logger:    log,
	})
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
	resp, err := cli.Request(ctx, &Request{
		Path:          "/",
		Method:        GET,
		Headers:       map[string]string{},
		Body:          nil,
		ContentLegnth: 0,
			Handlers: []Handler{
				func(r *http.Request, url string) error {
					
					for k, v := range NewAuthHeaders(url, GET, "apikey", "apisecret", nil) {
						r.Header.Set(k, v)
					}
					return nil
				},
			},
	})
	if err != nil {
		panic(err)
	}
    cancel()
	resp.Write(os.Stdout)
    resp.Body.Close()
	cli.Destory()

````
 ##### 示例2
````go
	log, err := NewLogger("./finderhttp.log", LogDebug)
	if err != nil {
		panic(err)
	}
	cli, err := NewClient(&ClientOptions{
		FinderUrl: "10.1.87.69:6868/guiderAllService/gas/storage-gateway/1.0.0",
		// Address: []string{"10.1.87.70:8210"},
		Logger: log,
		UseTLS: false,
		LB:     LBConsistencyHash(nil, 10),
	})

	hc := http.Client{
		Transport: &http.Transport{
			Proxy:       http.ProxyFromEnvironment,
			DialContext: cli.NetDialContext([]string{"abc.test.com"}),
		},
	}

	resp, err := hc.Get("http://abc.test.com")
	if err != nil {
		panic(err)
	}

	resp.Write(os.Stdout)

````


#### clientOptions 解释

```
type ClientOptions struct {
	Address   []string                // 远程地址，如果没有配置FinderUrl ，会使用这个地址访问
	FinderUrl string                  // 配置中心地址：例如  10.1.87.79:6868/AIaaS/dx/webate/1.0.9
	LB        Lb                      // 负载均衡策略，支持RoundRobin 和 Hash 和 ConsistencyHash 
	Ping      func(addf string) error // 健康检查探测器
	Logger    Logger                  // 日志接口，默认INFO 级别

	UseTLS bool // 是否启用tls

	HttpClient *http.Client
}

```
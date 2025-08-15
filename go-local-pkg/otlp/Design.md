````
### 1 大模型定制版本

将工具兜底重传逻辑移至sdk 侧，通常情况下，用户仅需提供一个兜底重传的开关即可，其他的配置项也开放出来，如果提供则用提供的，没有则用默认。

异常情况兜底：

* otlp collector不可用、collector端口不通

* 存储网关不可用

日志配置尽量轻量化一点，关键配置必须配置即可，比如，sdk自身日志配置，提供个默认合理的即可。

写入本地磁盘还是使用bager：

使用bager：
需要验证重启是否会影响

设计如何更简单

保证一小时内的重传即可：

qps=300,1000000 万左右。

配置项下划线保持统一。

备份模块设计：如何通用可扩展

备份模块：

写入+发送（是否需要该方法）

```go
type Backuper interface {
    WriteLogs(logs ...logging.EventLog) error
    Send() error
    Start() error
    ShutDown(ctx context.Context) error
    StatusOk() bool
}
```

clientProxy（localCache）同时实现写入和发送接口：



流程：

Backup:

包含多个对象：

* 缓存对象：
* 探测对象：
* 写入：向缓存对象写入
* 发送：向探测对象发送

创建一个Backup 实例，

* 初始化Backup 实例，启动异步发送模块，时间间隔触发；
* 探测对象每个一定时间探测；
* 若探测失败，写入，写入成功，删除备份



Metrics 关键字告警：

```bash
ResourceExhausted
```

Log 关键字告警：

```bash

```

验证断开后重连状态：





otlp 自身修复问题：

### 2 ASE 定制版本
精简模式：移除备份逻辑

EventLog 兼容：


Trace兼容：


Sonar二选一：
````

# mini-log
轻量级滚动日志。引用 `zap` 和 `lumberjack-ccr`，提供一个轻量级滚动日志的服务组件。
* 归档日志：不再允许写入的日志文件
* 当 `MaxAge` 和 `MaxBackup` 都为 0 时，不再删除归档日志

## 使用
```go
import "git.iflytek.com/AIaaS/mini-log"

logger, err := NewLocalLog(nil)
if err!=nil {
    panic("create log failed.")
}

```


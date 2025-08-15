package logging

import "time"

const (
	DEFAULT_SERVICE_NAME                        = "otlpService"
	DEFAULT_MAX_QUEUE_SIZE        int           = 2048
	DEFAULT_EXPORT_TIMEOUT        time.Duration = 30 * time.Second //单位ms
	DEFAULT_EXPORT_INTERVAL       time.Duration = time.Second      //单位ms
	DEFAULT_EXPORT_MAX_BATCH_SIZE int           = 40
	DEFAULT_ATTR_COUNT_LIMIT      int           = 128
	DEFAULT_VALUE_LENGTH_LIMIT    int           = -1
	DEFAULT_MAX_GRPC_SEND_SIZE    int           = 4
	DEFAULT_FLUSH_WORKER_NUM                    = 32
	DEFAULT_FINI_ENABLE_WAIT                    = false
	DEFAULT_FLUSH_TIMEOUT                       = time.Second * 10
)

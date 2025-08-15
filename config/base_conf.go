package config

import (
	"fmt"
	"strconv"
	"th3api/common"
)

const (
	CFG_KEY_SERVICE_ID = "serviceId"

	CFG_KEY_TH3API_AK               = "th3apiAK"
	CFG_KEY_TH3API_BASE_URL         = "th3apiBaseUrl"
	CFG_KEY_TH3API_TIMEOUT          = "th3apiTimeout"
	CFG_KEY_TH3API_TCP_DIAL_TIMEOUT = "th3apiTcpDialTimeout"
	CFG_KEY_TH3API_MODEL_NAME       = "th3apiModelName"

	CFG_KEY_LOG_DIR         = "logDir"
	CFG_KEY_LOG_MAX_SIZE    = "logMaxSize"
	CFG_KEY_LOG_MAX_BACKUPS = "logMaxBackups"
	CFG_KEY_LOG_MAX_AGE     = "logMaxAge"
	CFG_KEY_LOG_COMPRESS    = "logCompress"
	CFG_KEY_LOG_LEVEL       = "logLevel"

	CFG_KEY_OTLP_LOG_ADDR = "otlpLogAddr"
	CFG_KEY_OTLP_LOG_HOST = "otlpLogHost"

	CFG_KEY_PROMPT_SEARCH_TEMPLATE          = "prompt_search_template"
	CFG_KEY_PROMPT_SEARCH_TEMPLATE_NO_INDEX = "prompt_search_template_no_index"
)

type BaseConf struct {
	ServiceId string // 服务名称

	Th3ApiAK             string // 接入三方api口令
	Th3ApiBaseUrl        string // 接入三方api请求URL
	Th3ApiTimeOut        int    // 请求三方http总超时时间,  默认15*60s
	Th3ApiTCPDialTimeout int    // 请求三方tcp dial超时时间， 默认5s
	Th3ApiModelName      string // 接入三方模型名称

	LogDir        string // 日志输出路径， 默认在 ./log/{serviceId}.log
	LogMaxSize    int    // 单个日志文件最大大小(MB)，  默认50MB
	LogMaxBackups int    // 保留旧日志文件的最大数量， 默认保留5个备份
	LogMaxAge     int    // 保留旧日志文件的最大天数， 默认保留7天
	LogCompress   bool   // 是否压缩/归档旧日志文件, 默认不开启压缩
	LogLevel      string // 日志等级, 默认info级别

	OtlpLogAddr string // otlp日志上报ip
	OtlpLogHost string // otlp日志上报host

	PromptSearchTemplate        string // 联网搜索有抽槽模板
	PromptSearchTemplateNoIndex string // 联网搜索无抽槽模板

}

var baseConf BaseConf

func GetBaseConf() BaseConf {
	return baseConf
}

func NewBaseConf(cfg map[string]string) error {
	if v, ok := cfg[CFG_KEY_SERVICE_ID]; ok {
		baseConf.ServiceId = v
	} else {
		return fmt.Errorf(" Wrapper Conf[%s] is empty! ", CFG_KEY_SERVICE_ID)
	}

	// th3api 配置初始化
	if v, ok := cfg[CFG_KEY_TH3API_AK]; ok {
		baseConf.Th3ApiAK = v
	} else {
		return fmt.Errorf(" Wrapper Conf[%s] is empty! ", CFG_KEY_TH3API_AK)
	}

	if v, ok := cfg[CFG_KEY_TH3API_BASE_URL]; ok {
		baseConf.Th3ApiBaseUrl = v
	} else {
		return fmt.Errorf(" Wrapper Conf[%s] is empty! ", CFG_KEY_TH3API_BASE_URL)
	}

	if v, ok := cfg[CFG_KEY_TH3API_TIMEOUT]; ok {
		timeout, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return fmt.Errorf(" Wrapper Conf[%s=%s] is not num", CFG_KEY_TH3API_TIMEOUT, v)
		}
		if timeout <= 0 {
			return fmt.Errorf(" Wrapper Conf[%s=%s] is less than or equal to 0", CFG_KEY_TH3API_TIMEOUT, v)
		}
		baseConf.Th3ApiTimeOut = int(timeout)
	} else {
		baseConf.Th3ApiTimeOut = 15 * 60
	}

	if v, ok := cfg[CFG_KEY_TH3API_TCP_DIAL_TIMEOUT]; ok {
		timeout, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return fmt.Errorf(" Wrapper Conf[%s=%s] is not num", CFG_KEY_TH3API_TCP_DIAL_TIMEOUT, v)
		}
		if timeout <= 0 {
			return fmt.Errorf(" Wrapper Conf[%s=%s] is less than or equal to 0", CFG_KEY_TH3API_TCP_DIAL_TIMEOUT, v)
		}
		baseConf.Th3ApiTCPDialTimeout = int(timeout)
	} else {
		baseConf.Th3ApiTCPDialTimeout = 5
	}

	if v, ok := cfg[CFG_KEY_TH3API_MODEL_NAME]; ok {
		baseConf.Th3ApiModelName = v
	} else {
		return fmt.Errorf(" Wrapper Conf[%s] is empty! ", CFG_KEY_TH3API_MODEL_NAME)
	}

	// 日志配置初始化
	if v, ok := cfg[CFG_KEY_LOG_DIR]; ok {
		baseConf.LogDir = v
	} else {
		baseConf.LogDir = "./log/" + baseConf.ServiceId + ".log"
	}

	// 日志最大大小(MB)
	if v, ok := cfg[CFG_KEY_LOG_MAX_SIZE]; ok {
		maxSize, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return fmt.Errorf(" Wrapper Conf[%s=%s] is not num", CFG_KEY_LOG_MAX_SIZE, v)
		}
		if maxSize <= 0 {
			return fmt.Errorf(" Wrapper Conf[%s=%s] is less than or equal to 0", CFG_KEY_LOG_MAX_SIZE, v)
		}
		baseConf.LogMaxSize = int(maxSize)
	} else {
		baseConf.LogMaxSize = 50 // 默认50MB
	}

	// 日志最大备份数量
	if v, ok := cfg[CFG_KEY_LOG_MAX_BACKUPS]; ok {
		maxBackups, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return fmt.Errorf(" Wrapper Conf[%s=%s] is not num", CFG_KEY_LOG_MAX_BACKUPS, v)
		}
		if maxBackups < 0 {
			return fmt.Errorf(" Wrapper Conf[%s=%s] is less than 0", CFG_KEY_LOG_MAX_BACKUPS, v)
		}
		baseConf.LogMaxBackups = int(maxBackups)
	} else {
		baseConf.LogMaxBackups = 5 // 默认保留5个备份
	}

	// 日志最大保存天数
	if v, ok := cfg[CFG_KEY_LOG_MAX_AGE]; ok {
		maxAge, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return fmt.Errorf(" Wrapper Conf[%s=%s] is not num", CFG_KEY_LOG_MAX_AGE, v)
		}
		if maxAge < 0 {
			return fmt.Errorf(" Wrapper Conf[%s=%s] is less than 0", CFG_KEY_LOG_MAX_AGE, v)
		}
		baseConf.LogMaxAge = int(maxAge)
	} else {
		baseConf.LogMaxAge = 7 // 默认保留7天
	}

	// 日志是否压缩
	if v, ok := cfg[CFG_KEY_LOG_COMPRESS]; ok {
		compress, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf(" Wrapper Conf[%s=%s] is not bool", CFG_KEY_LOG_COMPRESS, v)
		}
		baseConf.LogCompress = compress
	} else {
		baseConf.LogCompress = false // 默认不开启压缩
	}

	// 日志级别
	if v, ok := cfg[CFG_KEY_LOG_LEVEL]; ok {
		baseConf.LogLevel = v
	} else {
		baseConf.LogLevel = common.LOG_LEVEL_INFO // 默认info级别
	}

	// otlp日志上报ip
	if v, ok := cfg[CFG_KEY_OTLP_LOG_ADDR]; ok {
		baseConf.OtlpLogAddr = v
	} else {
		baseConf.OtlpLogAddr = "172.30.209.28:4317" // 默认addr
	}

	// otlp日志上报host
	if v, ok := cfg[CFG_KEY_OTLP_LOG_HOST]; ok {
		baseConf.OtlpLogHost = v
	} else {
		baseConf.OtlpLogHost = "127.0.0.1:8080" // 默认addr
	}

	baseConf.PromptSearchTemplate = cfg[CFG_KEY_PROMPT_SEARCH_TEMPLATE]
	baseConf.PromptSearchTemplateNoIndex = cfg[CFG_KEY_PROMPT_SEARCH_TEMPLATE_NO_INDEX]

	return nil
}

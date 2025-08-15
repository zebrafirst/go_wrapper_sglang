package common

// openai 请求参数key
const (
	BASE_REQ_KEY_MODEL                 = "model"
	BASE_REQ_KEY_MESSAGES              = "messages"
	BASE_REQ_KEY_MAX_TOKENS            = "max_tokens"
	BASE_REQ_KEY_MAX_COMPLETION_TOKENS = "max_completion_tokens"
	BASE_REQ_KEY_TEMPERATURE           = "temperature"
	BASE_REQ_KEY_TOP_P                 = "top_p"
	BASE_REQ_KEY_N                     = "n"
	BASE_REQ_KEY_STREAM                = "stream"
	BASE_REQ_KEY_STOP                  = "stop"
	BASE_REQ_KEY_PRESENCE_PENALTY      = "presence_penalty"
	BASE_REQ_KEY_RESPONSE_FORMAT       = "response_format"
	BASE_REQ_KEY_SEED                  = "seed"
	BASE_REQ_KEY_FREQUENCY_PENALTY     = "frequency_penalty"
	BASE_REQ_KEY_LOGIT_BIAS            = "logit_bias"
	BASE_REQ_KEY_LOGPROBS              = "logprobs"
	BASE_REQ_KEY_TOP_LOGPROBS          = "top_logprobs"
	BASE_REQ_KEY_USER                  = "user"
	BASE_REQ_KEY_FUNCTIONS             = "functions"
	BASE_REQ_KEY_FUNCTION_CALL         = "function_call"
	BASE_REQ_KEY_TOOLS                 = "tools"
	BASE_REQ_KEY_TOOL_CHOICE           = "tool_choice"
	BASE_REQ_KEY_STREAM_OPTIONS        = "stream_options"
	BASE_REQ_KEY_PARALLEL_TOOL_CALLS   = "parallel_tool_calls"
	BASE_REQ_KEY_STORE                 = "store"
	BASE_REQ_KEY_REASONING_EFFORT      = "reasoning_effort"
	BASE_REQ_KEY_METADATA              = "metadata"
	BASE_REQ_KEY_PREDICTION            = "prediction"
	BASE_REQ_KEY_CHAT_TEMPLATE_KWARGS  = "chat_template_kwargs"
	BASE_REQ_KEY_SERVICE_TIER          = "service_tier"

	BASE_REQ_KEY_EXTRA_BODY = "extra_body"
)

// 日志等级
const (
	LOG_LEVEL_DEBUG  = "debug"
	LOG_LEVEL_INFO   = "info"
	LOG_LEVEL_WARN   = "warn"
	LOG_LEVEL_ERROR  = "error"
	LOG_LEVEL_DPANIC = "dpanic"
	LOG_LEVEL_PANIC  = "panic"
	LOG_LEVEL_FATAL  = "fatal"
)

// llm chat message role
const (
	ROLE_ASSISTANT = "assistant"
	ROLE_TOOL      = "tool"
	ROLE_USER      = "user"
)

// 流式会话状态
const (
	SESSION_STATUS_BEGIN    = iota // 0 - 会话开始
	SESSION_STATUS_CONTINUE        // 1 - 会话继续
	SESSION_STATUS_END             // 2 - 会话结束
)

// 工具id
const (
	TOOL_CALL_ID_IFLY_SEARCH = "ifly_search"
)

// msg key
const (
	MSG_KEY_CONTENT = "content"
	MSG_KEY_USAGE   = "usage"
)

var MeterFunc func(usrTag string, key string, count int) (code int)

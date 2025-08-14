package inst

import (
	"comwrapper"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"th3api/common"
	"th3api/common/model"
	"th3api/config"
	errconvert "th3api/inst/err_convert"
	th3apiutils "th3api/th3api_utils"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type OpenAITh3API struct {
	Inst    *InstAdaptor
	Ak      string
	BaseUrl string
	LogMap  map[string]interface{}
	Ret     []string
}

func NewOpenAITh3API(inst *InstAdaptor) *OpenAITh3API {
	return &OpenAITh3API{
		Inst:    inst,
		Ak:      config.GetBaseConf().Th3ApiAK,
		BaseUrl: config.GetBaseConf().Th3ApiBaseUrl,
		LogMap:  make(map[string]interface{}),
	}
}

func (th3api *OpenAITh3API) PushBack(cb comwrapper.CallBackPtr) error {
	var err error
	logMap := th3api.LogMap
	logMap["sid"] = th3api.Inst.Sid
	logMap["code"] = 0
	logMap["massage"] = "success"

	start := time.Now()
	isFirstRet := true
	var retTT []int64

	defer func() {
		logMap["ret"] = th3api.Ret
		logMap["retTT"] = retTT
		if err != nil {
			switch v := err.(type) {
			case *errconvert.Th3apiErr:
				logMap["code"] = v.Code
				logMap["massage"] = v.Message
			default:
				logMap["code"] = errconvert.UnknowErrCode.Code
				logMap["massage"] = v.Error()
			}
		}
		if log, err1 := th3apiutils.GetOtlpLog(th3api.Inst.Sid); err1 == nil {
			th3apiutils.WLogger.Info("", zap.Any("logMap", logMap))
			log.LogMsgsAndFlush(logMap)
		} else {
			th3apiutils.WLogger.Warn("Get otlp log failed", zap.Any("err", err1), zap.Any("logMap", logMap))
		}
	}()

	stream, err := th3api.doReqStream(context.Background())
	if err != nil {
		th3apiutils.WLogger.Error("请求三方api失败", zap.Any("err", err), zap.String("sid", th3api.Inst.Sid))
		if err1 := cb(th3api.Inst.UsrTag, nil, err); err1 != nil {
			th3apiutils.WLogger.Error("Callback loader failed", zap.Any("err", err1), zap.String("sid", th3api.Inst.Sid))
		}
		return errconvert.WrapperErr(err, errconvert.CallTh3ApiUnknowErrCode.Code)
	}
	defer stream.Close()

	for {
		var resp []comwrapper.WrapperData

		response, err := stream.Recv()
		elapsed := time.Since(start).Milliseconds()
		if isFirstRet {
			logMap["firstRetTT"] = elapsed
			isFirstRet = false
		}
		retTT = append(retTT, elapsed)

		if errors.Is(err, io.EOF) { // 处理最后一帧
			th3apiutils.WLogger.Info("Stream finished")
			return nil
		}

		if err != nil {
			th3apiutils.WLogger.Error("流式请求三方api失败", zap.Any("err", err), zap.String("sid", th3api.Inst.Sid))
			if err1 := cb(th3api.Inst.UsrTag, nil, err); err1 != nil {
				th3apiutils.WLogger.Error("Callback loader failed", zap.Any("err", err1), zap.String("sid", th3api.Inst.Sid))
			}
			return errconvert.WrapperErr(err, errconvert.CallTh3ApiUnknowErrCode.Code)
		}

		th3apiutils.WLogger.Debug("", zap.Any("resp", response))

		// 状态变更
		if th3api.Inst.IsFirstRet {
			th3api.Inst.Status = common.SESSION_STATUS_BEGIN
			th3api.Inst.IsFirstRet = false
		} else if response.Usage != nil {
			th3api.Inst.Status = common.SESSION_STATUS_END
		} else {
			th3api.Inst.Status = common.SESSION_STATUS_CONTINUE
		}

		if len(response.Choices) > 0 {
			for _, v := range response.Choices {
				// 记录th3api返回
				th3api.Ret = append(th3api.Ret, v.Delta.Content)
				content := model.Content{
					Choices: []model.Choice{
						{
							Content:          v.Delta.Content,
							Index:            th3api.Inst.ReqNo,
							Role:             common.ROLE_ASSISTANT,
							ReasoningContent: "",
						},
					},
					QuestionType: "",
				}
				th3api.Inst.ReqNo++
				contentJsonBytes, _ := json.Marshal(content)

				resp = append(resp, comwrapper.WrapperData{
					Key:      common.MSG_KEY_CONTENT,
					Data:     contentJsonBytes,
					Desc:     nil,
					Encoding: "utf-8",
					Type:     comwrapper.DataText,
					Status:   comwrapper.DataStatus(th3api.Inst.Status),
				})
			}
		}

		var usage model.Usage
		if response.Usage != nil {
			usage.TotalTokens = response.Usage.TotalTokens
			usage.PromptTokens = response.Usage.PromptTokens
			usage.CompletionTokens = response.Usage.CompletionTokens
			logMap["usage"] = usage

			// 上报自定义计量数据
			if code := common.MeterFunc(th3api.Inst.UsrTag, "total_tokens", usage.TotalTokens); code != 0 {
				th3apiutils.WLogger.Warn("自定义计量失败", zap.Int("total_tokens", usage.TotalTokens), zap.String("sid", th3api.Inst.Sid), zap.String("usrTag", th3api.Inst.UsrTag))
			}
			if code := common.MeterFunc(th3api.Inst.UsrTag, "completion_tokens", usage.CompletionTokens); code != 0 {
				th3apiutils.WLogger.Warn("自定义计量失败", zap.Int("completion_tokens", usage.CompletionTokens), zap.String("sid", th3api.Inst.Sid), zap.String("usrTag", th3api.Inst.UsrTag))
			}
			if code := common.MeterFunc(th3api.Inst.UsrTag, "prompt_tokens", usage.PromptTokens); code != 0 {
				th3apiutils.WLogger.Warn("自定义计量失败", zap.Int("prompt_tokens", usage.PromptTokens), zap.String("sid", th3api.Inst.Sid), zap.String("usrTag", th3api.Inst.UsrTag))
			}
		}
		usageJsonBytes, _ := json.Marshal(usage)
		resp = append(resp, comwrapper.WrapperData{
			Key:      common.MSG_KEY_USAGE,
			Data:     usageJsonBytes,
			Desc:     nil,
			Encoding: "utf-8",
			Type:     comwrapper.DataText,
			Status:   comwrapper.DataStatus(th3api.Inst.Status),
		},
		)

		if err := cb(th3api.Inst.UsrTag, resp, nil); err != nil {
			th3apiutils.WLogger.Error("Callback loader failed", zap.Any("err", err), zap.String("sid", th3api.Inst.Sid))
			return errconvert.WrapperErr(err, errconvert.UnknowErrCode.Code)
		}
	}
}

func (th3api *OpenAITh3API) doReqStream(ctx context.Context) (stream *openai.ChatCompletionStream, err error) {
	client := th3api.buildOpenAIClient()
	req, err := th3api.buildOpenAIChatCompletionRequest()
	if err != nil {
		th3apiutils.WLogger.Error("Th3api build chat req failed", zap.Any("err", err), zap.String("sid", th3api.Inst.Sid))
		return nil, err
	}
	th3api.LogMap["th3apiReq"] = req
	th3apiutils.WLogger.Info("th3api openAI req", zap.Any("th3apiReq", req), zap.String("sid", th3api.Inst.Sid))
	return client.CreateChatCompletionStream(ctx, req)
}

func (th3api *OpenAITh3API) buildOpenAIClient() *openai.Client {
	config := openai.DefaultConfig(th3api.Ak)
	config.BaseURL = th3api.BaseUrl

	// 设置超时
	config.HTTPClient = &http.Client{
		Timeout: th3api.Inst.TimeOut,
	}

	return openai.NewClientWithConfig(config)
}

func (th3api *OpenAITh3API) buildOpenAIChatCompletionRequest() (openai.ChatCompletionRequest, error) {
	req := openai.ChatCompletionRequest{
		Model:  th3api.Inst.Model,
		Stream: true,
		StreamOptions: &openai.StreamOptions{
			IncludeUsage: true,
		},
	}
	if messages, err := th3api.buildChatReqMessage(); err != nil {
		return req, err
	} else {
		req.Messages = messages
		th3apiutils.WLogger.Debug("", zap.Any("chatReqMessage", req))
	}

	if err := th3api.attachBaseParam(&req); err != nil {
		return req, err
	}
	return req, nil
}

func (th3api *OpenAITh3API) buildChatReqMessage() ([]openai.ChatCompletionMessage, error) {
	chatMessage := make([]openai.ChatCompletionMessage, 0, len(th3api.Inst.InDatas))

	for _, v := range th3api.Inst.InDatas {
		ldMessage := LoaderMessage{}
		if err := json.Unmarshal([]byte(v), &ldMessage); err != nil {
			th3apiutils.WLogger.Error("Unmarshal loader message failed", zap.String("message", v), zap.Any("err", err), zap.String("sid", th3api.Inst.Sid))
			return nil, err
		}

		for i, v := range ldMessage.Messages {
			if v.Role == common.ROLE_TOOL && v.ToolCallID == common.TOOL_CALL_ID_IFLY_SEARCH {
				var sources []Source
				if err := json.Unmarshal([]byte(v.Content), &sources); err != nil {
					th3apiutils.WLogger.Error("Unmarshal web search source failed", zap.String("content", v.Content), zap.Any("err", err), zap.String("sid", th3api.Inst.Sid))
					return nil, err
				}
				v.Sources = sources
				th3apiutils.WLogger.Debug("", zap.Any("webSearchSources", sources))
				ldMessage.Messages[i] = v
			}
		}

		th3apiutils.WLogger.Debug("", zap.Any("loadMessage", ldMessage))

		ws := NewWebSearch(&ldMessage)
		if ws.NeedWebSearch {
			th3apiutils.WLogger.Debug("Need web search", zap.Any("ws", ws))
			for _, v1 := range ldMessage.Messages {
				chatMessage = append(chatMessage, openai.ChatCompletionMessage{
					Content:    v1.Content,
					Role:       v1.Role,
					ToolCallID: v1.ToolCallID,
				})
			}
			th3apiutils.WLogger.Debug("", zap.Any("webSearchRawChatMsg", chatMessage))
			// 移除用户最后一次请求msg
			lastUserIndex := -1
			question := ""
			for i := len(chatMessage) - 1; i >= 0; i-- {
				if chatMessage[i].Role == common.ROLE_USER {
					question = chatMessage[i].Content
					lastUserIndex = i
					break
				}
			}
			if lastUserIndex != -1 {
				chatMessage = append(chatMessage[:lastUserIndex], chatMessage[lastUserIndex+1:]...)
			}

			// 移除联网搜索 msg
			filterChatMessage := make([]openai.ChatCompletionMessage, 0, len(chatMessage))
			for i := len(chatMessage) - 1; i >= 0; i-- {
				if chatMessage[i].Role == common.ROLE_TOOL {
					continue
				}
				filterChatMessage = append(filterChatMessage, chatMessage[i])
			}

			// 添加联网搜索user msg
			filterChatMessage = append(filterChatMessage, openai.ChatCompletionMessage{
				Role:    common.ROLE_USER,
				Content: ws.BuildWebSearchPrompt(question),
			})

			return filterChatMessage, nil
		} else {
			for _, v1 := range ldMessage.Messages {
				chatMessage = append(chatMessage, openai.ChatCompletionMessage{
					Content: v1.Content,
					Role:    v1.Role,
				})
			}
		}
	}

	return chatMessage, nil
}

func (th3api *OpenAITh3API) attachBaseParam(chatReq *openai.ChatCompletionRequest) error {
	// temperature top_k max_tokens chat_id? tools:暂不支持 enable_thinking：不支持

	if v, ok := th3api.Inst.Params[common.BASE_REQ_KEY_TEMPERATURE]; ok {
		if temperature, err := strconv.ParseFloat(v, 32); err == nil {
			chatReq.Temperature = float32(temperature)
		} else {
			th3apiutils.WLogger.Warn("Parse temperature failed", zap.String("temperature", v), zap.Any("err", err), zap.String("sid", th3api.Inst.Sid))
		}
	}

	// top_k openai 协议无top_k参数，先不管这个参数

	if v, ok := th3api.Inst.Params[common.BASE_REQ_KEY_MAX_TOKENS]; ok {
		if maxTokens, err := strconv.ParseInt(v, 10, 32); err == nil {
			chatReq.MaxTokens = int(maxTokens)
		} else {
			th3apiutils.WLogger.Warn("Parse maxTokens failed", zap.String("maxTokens", v), zap.Any("err", err), zap.String("sid", th3api.Inst.Sid))
		}
	}

	return nil
}

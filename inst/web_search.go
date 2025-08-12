package inst

import (
	"fmt"
	"strings"
	"th3api/common"
	"time"
)

type WebSearch struct {
	NeedWebSearch bool
	Msg           Message
}

func NewWebSearch(ldMsg *LoaderMessage) *WebSearch {
	webSearch := &WebSearch{}
	for _, v := range ldMsg.Messages {
		if v.Role == common.ROLE_TOOL && v.ToolCallID == common.TOOL_CALL_ID_IFLY_SEARCH {
			webSearch.NeedWebSearch = true
			webSearch.Msg = v
		}
	}
	return webSearch
}

// 将联网搜索结果转成拼接到prompt中的标准字符串
func (ws *WebSearch) fmtSource2StdStr() string {
	searchResultsStdPromptStr := ""
	for _, src := range ws.Msg.Sources {
		searchResultsStdPromptStr += fmt.Sprintf("[webpage %d begin]%s[webpage %d end]\n", src.Index, src.Document, src.Index)
	}
	return searchResultsStdPromptStr
}

func (ws *WebSearch) BuildWebSearchPrompt(question string) string {
	searchResultsStdPromptStr := ws.fmtSource2StdStr()
	curDate := time.Now().Format("2006年01月02日 15时04分05秒")

	prompt := `# 以下内容是基于用户发送的消息的搜索结果:\n{{SearchResults}}\n在我给你的搜索结果中，每个结果都是[webpage X begin]...[webpage X end]格式的，X代表每篇文章的数字索引。请在适当的情况下在句子末尾引用上下文。请按照引用编号[citation:X]的格式在答案中对应部分引用上下文。如果一句话源自多个上下文，请列出所有相关的引用编号，例如[citation:3][citation:5]，切记不要将引用集中在最后返回引用编号，而是在答案对应部分列出。\n在回答时，请注意以下几点：\n- 今天是{{CurDate}}。\n- 并非搜索结果的所有内容都与用户的问题密切相关，你需要结合问题，对搜索结果进行甄别、筛选。\n- 对于列举类的问题（如列举所有航班信息），尽量将答案控制在10个要点以内，并告诉用户可以查看搜索来源、获得完整信息。优先提供信息完整、最相关的列举项；如非必要，不要主动告诉用户搜索结果未提供的内容。\n- 对于创作类的问题（如写论文），请务必在正文的段落中引用对应的参考编号，例如[citation:3][citation:5]，不能只在文章末尾引用。你需要解读并概括用户的题目要求，选择合适的格式，充分利用搜索结果并抽取重要信息，生成符合用户要求、极具思想深度、富有创造力与专业性的答案。你的创作篇幅需要尽可能延长，对于每一个要点的论述要推测用户的意图，给出尽可能多角度的回答要点，且务必信息量大、论述详尽。\n- 如果回答很长，请尽量结构化、分段落总结。如果需要分点作答，尽量控制在5个点以内，并合并相关的内容。\n- 对于客观类的问答，如果问题的答案非常简短，可以适当补充一到两句相关信息，以丰富内容。\n- 你需要根据用户要求和回答内容选择合适、美观的回答格式，确保可读性强。\n- 你的回答应该综合多个相关网页来回答，不能重复引用一个网页。\n- 除非用户要求，否则你回答的语言需要和用户提问的语言保持一致。\n\n# 用户消息为：\n{{Question}}"
	"prompt_search_template_no_index" = "# 以下内容是基于用户发送的消息的搜索结果:\n{{SearchResults}}\n在我给你的搜索结果中，每个结果都是[webpage X begin]...[webpage X end]格式的，X代表每篇文章的数字索引。\n在回答时，请注意以下几点：\n- 今天是{{CurDate}}。\n- 并非搜索结果的所有内容都与用户的问题密切相关，你需要结合问题，对搜索结果进行甄别、筛选。\n- 对于列举类的问题（如列举所有航班信息），尽量将答案控制在10个要点以内，并告诉用户可以查看搜索来源、获得完整信息。优先提供信息完整、最相关的列举项；如非必要，不要主动告诉用户搜索结果未提供的内容。\n- 对于创作类的问题（如写论文）。你需要解读并概括用户的题目要求，选择合适的格式，充分利用搜索结果并抽取重要信息，生成符合用户要求、极具思想深度、富有创造力与专业性的答案。你的创作篇幅需要尽可能延长，对于每一个要点的论述要推测用户的意图，给出尽可能多角度的回答要点，且务必信息量大、论述详尽。\n- 如果回答很长，请尽量结构化、分段落总结。如果需要分点作答，尽量控制在5个点以内，并合并相关的内容。\n- 对于客观类的问答，如果问题的答案非常简短，可以适当补充一到两句相关信息，以丰富内容。\n- 你需要根据用户要求和回答内容选择合适、美观的回答格式，确保可读性强。\n- 你的回答应该综合多个相关网页来回答，不能重复引用一个网页。\n- 除非用户要求，否则你回答的语言需要和用户提问的语言保持一致。\n- 请在回答中不要包含搜索结果的引用标记\n\n# 用户消息为：\n{{Question}}"`
	prompt = strings.ReplaceAll(prompt, "{{SearchResults}}", searchResultsStdPromptStr)
	prompt = strings.ReplaceAll(prompt, "{{CurDate}}", curDate)
	prompt = strings.ReplaceAll(prompt, "{{Question}}", question)

	return prompt
}

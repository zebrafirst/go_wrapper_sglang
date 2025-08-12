package model

type Choice struct {
	Content          string `json:"content"`           // 内容字段（含Unicode转义）
	Index            int    `json:"index"`             // 索引
	Role             string `json:"role"`              // 角色
	ReasoningContent string `json:"reasoning_content"` // 思考分析
}
type Content struct {
	Choices      []Choice `json:"choices"`
	QuestionType string   `json:"question_type"`
}

type Usage struct {
	CompletionTokens int `json:"completion_tokens,omitempty"`
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	QuestionTokens   int `json:"question_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

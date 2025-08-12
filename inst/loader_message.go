package inst

// 定义主消息结构
type LoaderMessage struct {
	Messages  []Message   `json:"messages"`
	Functions interface{} `json:"functions"`
}

// 定义消息项结构
type Message struct {
	Content          string   `json:"content"`
	ReasoningContent string   `json:"reasoning_content"`
	Role             string   `json:"role"`
	Index            int      `json:"index"`
	ToolCallID       string   `json:"tool_call_id,omitempty"`
	ShowRefLabel     bool     `json:"show_ref_label,omitempty"`
	Sources          []Source `json:"-"`
}

// 定义来源结构
type Source struct {
	Index    int    `json:"index"`
	Source   string `json:"source"`
	DocID    string `json:"docid"`
	Document string `json:"document"`
}

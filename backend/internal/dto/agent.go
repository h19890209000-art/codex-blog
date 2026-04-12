package dto

// AgentDraftRequest 是“根据素材生成文章草稿”接口的请求结构。
// source_text 里放已经抽取好的正文内容。
// 这样前端可以先上传文件抽取，再把结果交给 AI 继续整理。
type AgentDraftRequest struct {
	SourceText   string `json:"source_text"`
	Goal         string `json:"goal"`
	Tone         string `json:"tone"`
	CategoryHint string `json:"category_hint"`
}

// AgentChatRequest 是后台 Agent 日常聊天接口的请求结构。
// context 是可选上下文，比如刚刚抽取出来的素材或生成的草稿。
type AgentChatRequest struct {
	Message string `json:"message"`
	Context string `json:"context"`
}

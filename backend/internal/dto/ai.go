package dto

// AnalyzeTitleRequest 是标题解析接口的请求结构。
type AnalyzeTitleRequest struct {
	Title  string `json:"title"`
	Stream bool   `json:"stream"`
}

// GenerateSummaryRequest 是摘要生成接口的请求结构。
type GenerateSummaryRequest struct {
	Content string `json:"content"`
	Stream  bool   `json:"stream"`
}

// SuggestTagsRequest 是标签推荐接口的请求结构。
type SuggestTagsRequest struct {
	Content string `json:"content"`
}

// BrainstormRequest 是灵感风暴接口的请求结构。
type BrainstormRequest struct {
	Keyword string `json:"keyword"`
}

// RewriteRequest 是润色改写接口的请求结构。
type RewriteRequest struct {
	Content string `json:"content"`
	Style   string `json:"style"`
}

// GenerateCoverRequest 是封面图生成接口的请求结构。
type GenerateCoverRequest struct {
	Title string `json:"title"`
}

// QARequest 是问答接口的通用请求结构。
type QARequest struct {
	Question string `json:"question"`
	Stream   bool   `json:"stream"`
}

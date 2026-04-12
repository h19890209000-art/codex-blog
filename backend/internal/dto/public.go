package dto

// CreateCommentRequest 是前台提交评论的请求结构。
type CreateCommentRequest struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

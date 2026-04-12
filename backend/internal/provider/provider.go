package provider

import "context"

// Capability 表示某个 Provider 是否支持某类能力。
type Capability string

const (
	CapabilityChat          Capability = "chat"
	CapabilityStreamChat    Capability = "stream_chat"
	CapabilityEmbedding     Capability = "embedding"
	CapabilityModerate      Capability = "moderate"
	CapabilityImageGenerate Capability = "image_generate"
	CapabilityTextToSpeech  Capability = "text_to_speech"
)

// ChatMessage 表示一条对话消息。
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 是对话请求。
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

// ChatResponse 是对话响应。
type ChatResponse struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Text     string `json:"text"`
}

// ModerateRequest 是内容审核请求。
type ModerateRequest struct {
	Input string `json:"input"`
}

// ModerateResponse 是内容审核结果。
type ModerateResponse struct {
	Provider    string `json:"provider"`
	Flagged     bool   `json:"flagged"`
	Reason      string `json:"reason"`
	RawDecision string `json:"raw_decision"`
}

// EmbeddingRequest 是向量请求。
type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// EmbeddingResponse 是向量结果。
type EmbeddingResponse struct {
	Provider string    `json:"provider"`
	Vector   []float64 `json:"vector"`
}

// ImageRequest 是图片生成请求。
type ImageRequest struct {
	Prompt string `json:"prompt"`
	Size   string `json:"size"`
}

// ImageResponse 是图片生成结果。
type ImageResponse struct {
	Provider string `json:"provider"`
	URL      string `json:"url"`
}

// TTSRequest 是文本转语音请求。
type TTSRequest struct {
	Text  string `json:"text"`
	Voice string `json:"voice"`
}

// TTSResponse 是文本转语音结果。
type TTSResponse struct {
	Provider string `json:"provider"`
	URL      string `json:"url"`
}

// Provider 是统一的大模型适配接口。
// 业务层只依赖这一个接口，不直接依赖任何具体厂商。
type Provider interface {
	Name() string
	Supports(capability Capability) bool
	Chat(ctx context.Context, request ChatRequest) (ChatResponse, error)
	StreamChat(ctx context.Context, request ChatRequest) (<-chan string, <-chan error)
	Embedding(ctx context.Context, request EmbeddingRequest) (EmbeddingResponse, error)
	Moderate(ctx context.Context, request ModerateRequest) (ModerateResponse, error)
	ImageGenerate(ctx context.Context, request ImageRequest) (ImageResponse, error)
	TextToSpeech(ctx context.Context, request TTSRequest) (TTSResponse, error)
}

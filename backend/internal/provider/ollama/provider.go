package ollama

import (
	"context"
	"errors"
	"strings"

	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/provider"
	"ai-blog/backend/internal/provider/base"
	"ai-blog/backend/internal/stream"
)

// ProviderInstance 是 Ollama Provider 的实现。
type ProviderInstance struct {
	config config.ProviderConfig
	client *base.JSONClient
}

// New 创建 Ollama Provider。
func New(cfg config.ProviderConfig) provider.Provider {
	return &ProviderInstance{
		config: cfg,
		client: base.NewJSONClient(),
	}
}

// Name 返回 Provider 名称。
func (providerInstance *ProviderInstance) Name() string {
	return "ollama"
}

// Supports 返回 Ollama 是否支持某类能力。
func (providerInstance *ProviderInstance) Supports(capability provider.Capability) bool {
	switch capability {
	case provider.CapabilityChat, provider.CapabilityStreamChat, provider.CapabilityEmbedding:
		return true
	default:
		return false
	}
}

// Chat 调用 Ollama 的 /api/chat 接口。
func (providerInstance *ProviderInstance) Chat(ctx context.Context, request provider.ChatRequest) (provider.ChatResponse, error) {
	if providerInstance.config.BaseURL == "" {
		return provider.ChatResponse{}, errors.New("ollama base_url 不能为空")
	}

	payload := map[string]any{
		"model":    chooseModel(request.Model, providerInstance.config.Model),
		"messages": request.Messages,
		"stream":   false,
	}

	responseBody := struct {
		Model   string `json:"model"`
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}{}

	err := providerInstance.client.PostJSON(
		ctx,
		strings.TrimRight(providerInstance.config.BaseURL, "/")+"/api/chat",
		nil,
		payload,
		&responseBody,
	)
	if err != nil {
		return provider.ChatResponse{}, err
	}

	return provider.ChatResponse{
		Provider: "ollama",
		Model:    responseBody.Model,
		Text:     responseBody.Message.Content,
	}, nil
}

// StreamChat 通过拆文本的方式模拟流式输出。
func (providerInstance *ProviderInstance) StreamChat(ctx context.Context, request provider.ChatRequest) (<-chan string, <-chan error) {
	chunkChannel := make(chan string)
	errorChannel := make(chan error, 1)

	go func() {
		defer close(chunkChannel)
		defer close(errorChannel)

		response, err := providerInstance.Chat(ctx, request)
		if err != nil {
			errorChannel <- err
			return
		}

		for _, chunk := range stream.SplitText(response.Text) {
			chunkChannel <- chunk
		}
	}()

	return chunkChannel, errorChannel
}

// Embedding 调用 Ollama 的 embeddings 接口。
func (providerInstance *ProviderInstance) Embedding(ctx context.Context, request provider.EmbeddingRequest) (provider.EmbeddingResponse, error) {
	if providerInstance.config.BaseURL == "" {
		return provider.EmbeddingResponse{}, errors.New("ollama base_url 不能为空")
	}

	payload := map[string]any{
		"model":  chooseModel(request.Model, providerInstance.config.Model),
		"prompt": request.Input,
	}

	responseBody := struct {
		Embedding []float64 `json:"embedding"`
	}{}

	err := providerInstance.client.PostJSON(
		ctx,
		strings.TrimRight(providerInstance.config.BaseURL, "/")+"/api/embeddings",
		nil,
		payload,
		&responseBody,
	)
	if err != nil {
		return provider.EmbeddingResponse{}, err
	}

	return provider.EmbeddingResponse{
		Provider: "ollama",
		Vector:   responseBody.Embedding,
	}, nil
}

// Moderate 表示当前 Provider 不支持审核能力。
func (providerInstance *ProviderInstance) Moderate(ctx context.Context, request provider.ModerateRequest) (provider.ModerateResponse, error) {
	_ = ctx
	_ = request
	return provider.ModerateResponse{}, errors.New("ollama 默认实现不支持 moderate")
}

// ImageGenerate 表示当前 Provider 不支持图片生成能力。
func (providerInstance *ProviderInstance) ImageGenerate(ctx context.Context, request provider.ImageRequest) (provider.ImageResponse, error) {
	_ = ctx
	_ = request
	return provider.ImageResponse{}, errors.New("ollama 默认实现不支持 image_generate")
}

// TextToSpeech 表示当前 Provider 不支持 TTS。
func (providerInstance *ProviderInstance) TextToSpeech(ctx context.Context, request provider.TTSRequest) (provider.TTSResponse, error) {
	_ = ctx
	_ = request
	return provider.TTSResponse{}, errors.New("ollama 默认实现不支持 text_to_speech")
}

func chooseModel(requestModel string, fallbackModel string) string {
	if strings.TrimSpace(requestModel) != "" {
		return requestModel
	}

	return fallbackModel
}

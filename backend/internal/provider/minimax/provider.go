package minimax

import (
	"context"
	"errors"
	"strings"

	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/provider"
	"ai-blog/backend/internal/provider/base"
	"ai-blog/backend/internal/stream"
)

// ProviderInstance 是 MiniMax Provider 的实现。
type ProviderInstance struct {
	config config.ProviderConfig
	client *base.JSONClient
}

// New 创建 MiniMax Provider。
func New(cfg config.ProviderConfig) provider.Provider {
	return &ProviderInstance{
		config: cfg,
		client: base.NewJSONClient(),
	}
}

// Name 返回 Provider 名称。
func (providerInstance *ProviderInstance) Name() string {
	return "minimax"
}

// Supports 返回当前 Provider 支持的能力。
func (providerInstance *ProviderInstance) Supports(capability provider.Capability) bool {
	switch capability {
	case provider.CapabilityChat, provider.CapabilityStreamChat, provider.CapabilityTextToSpeech:
		return true
	default:
		return false
	}
}

// Chat 调用 MiniMax 文本聊天接口。
func (providerInstance *ProviderInstance) Chat(ctx context.Context, request provider.ChatRequest) (provider.ChatResponse, error) {
	if providerInstance.config.BaseURL == "" {
		return provider.ChatResponse{}, errors.New("minimax base_url 不能为空")
	}

	if providerInstance.config.APIKey == "" {
		return provider.ChatResponse{}, errors.New("minimax api_key 不能为空")
	}

	systemPrompt := ""
	payloadMessages := make([]map[string]any, 0, len(request.Messages))

	for _, message := range request.Messages {
		if message.Role == "system" {
			if systemPrompt == "" {
				systemPrompt = message.Content
			} else {
				systemPrompt += "\n" + message.Content
			}

			continue
		}

		payloadMessages = append(payloadMessages, map[string]any{
			"role": message.Role,
			"content": []map[string]string{
				{
					"type": "text",
					"text": message.Content,
				},
			},
		})
	}

	payload := map[string]any{
		"model":       chooseModel(request.Model, providerInstance.config.Model),
		"messages":    payloadMessages,
		"max_tokens":  chooseMaxTokens(request.MaxTokens),
		"temperature": chooseTemperature(request.Temperature),
		"system":      systemPrompt,
	}

	responseBody := struct {
		Model   string `json:"model"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}{}

	err := providerInstance.client.PostJSON(
		ctx,
		strings.TrimRight(providerInstance.config.BaseURL, "/")+"/messages",
		map[string]string{
			"x-api-key":         providerInstance.config.APIKey,
			"anthropic-version": "2023-06-01",
		},
		payload,
		&responseBody,
	)
	if err != nil {
		return provider.ChatResponse{}, err
	}

	textParts := make([]string, 0, len(responseBody.Content))
	for _, block := range responseBody.Content {
		if block.Type == "text" || block.Type == "thinking" {
			if strings.TrimSpace(block.Text) != "" {
				textParts = append(textParts, block.Text)
			}
		}
	}

	if len(textParts) == 0 {
		return provider.ChatResponse{}, errors.New("minimax 返回 content 为空")
	}

	return provider.ChatResponse{
		Provider: "minimax",
		Model:    responseBody.Model,
		Text:     strings.Join(textParts, "\n"),
	}, nil
}

// StreamChat 通过拆文本的方式对外提供流式能力。
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

// Embedding 表示当前 Provider 默认不支持 embedding。
func (providerInstance *ProviderInstance) Embedding(ctx context.Context, request provider.EmbeddingRequest) (provider.EmbeddingResponse, error) {
	_ = ctx
	_ = request
	return provider.EmbeddingResponse{}, errors.New("minimax 默认实现不支持 embedding")
}

// Moderate 表示当前 Provider 默认不支持 moderate。
func (providerInstance *ProviderInstance) Moderate(ctx context.Context, request provider.ModerateRequest) (provider.ModerateResponse, error) {
	_ = ctx
	_ = request
	return provider.ModerateResponse{}, errors.New("minimax 默认实现不支持 moderate")
}

// ImageGenerate 表示当前 Provider 默认不支持图片生成。
func (providerInstance *ProviderInstance) ImageGenerate(ctx context.Context, request provider.ImageRequest) (provider.ImageResponse, error) {
	_ = ctx
	_ = request
	return provider.ImageResponse{}, errors.New("minimax 默认实现不支持 image_generate")
}

// TextToSpeech 先返回一个教学版占位 URL。
func (providerInstance *ProviderInstance) TextToSpeech(ctx context.Context, request provider.TTSRequest) (provider.TTSResponse, error) {
	if providerInstance.config.APIKey == "" {
		return provider.TTSResponse{}, errors.New("minimax api_key 不能为空")
	}

	_ = ctx
	_ = request

	return provider.TTSResponse{
		Provider: "minimax",
		URL:      "https://placehold.co/600x120?text=MiniMax+TTS",
	}, nil
}

func chooseTemperature(value float64) float64 {
	if value <= 0 {
		return 1
	}

	if value > 1 {
		return 1
	}

	return value
}

func chooseMaxTokens(value int) int {
	if value <= 0 {
		return 1000
	}

	return value
}

func chooseModel(requestModel string, fallbackModel string) string {
	if strings.TrimSpace(requestModel) != "" {
		return requestModel
	}

	return fallbackModel
}

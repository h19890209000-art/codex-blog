package openai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/provider"
	"ai-blog/backend/internal/provider/base"
	"ai-blog/backend/internal/stream"
)

// CompatibleProvider 用来实现 OpenAI 兼容协议。
type CompatibleProvider struct {
	name   string
	config config.ProviderConfig
	client *base.JSONClient
}

// NewCompatibleProvider 创建一个 OpenAI 兼容 Provider。
func NewCompatibleProvider(name string, cfg config.ProviderConfig) *CompatibleProvider {
	return &CompatibleProvider{
		name:   name,
		config: cfg,
		client: base.NewJSONClient(),
	}
}

// Name 返回 Provider 名称。
func (providerInstance *CompatibleProvider) Name() string {
	return providerInstance.name
}

// Supports 返回当前 Provider 是否支持某类能力。
func (providerInstance *CompatibleProvider) Supports(capability provider.Capability) bool {
	switch capability {
	case provider.CapabilityChat, provider.CapabilityStreamChat, provider.CapabilityEmbedding, provider.CapabilityModerate, provider.CapabilityImageGenerate, provider.CapabilityTextToSpeech:
		return true
	default:
		return false
	}
}

// Chat 调用 OpenAI 兼容聊天接口。
func (providerInstance *CompatibleProvider) Chat(ctx context.Context, request provider.ChatRequest) (provider.ChatResponse, error) {
	if providerInstance.config.BaseURL == "" {
		return provider.ChatResponse{}, errors.New("base_url 不能为空")
	}

	if providerInstance.config.APIKey == "" {
		return provider.ChatResponse{}, errors.New("api_key 不能为空")
	}

	payload := map[string]any{
		"model":       chooseModel(request.Model, providerInstance.config.Model),
		"messages":    request.Messages,
		"temperature": request.Temperature,
		"max_tokens":  request.MaxTokens,
	}

	responseBody := struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content          string `json:"content"`
				ReasoningContent string `json:"reasoning_content"`
			} `json:"message"`
		} `json:"choices"`
	}{}

	err := providerInstance.client.PostJSON(
		ctx,
		strings.TrimRight(providerInstance.config.BaseURL, "/")+"/chat/completions",
		map[string]string{
			"Authorization": "Bearer " + providerInstance.config.APIKey,
		},
		payload,
		&responseBody,
	)
	if err != nil {
		return provider.ChatResponse{}, err
	}

	if len(responseBody.Choices) == 0 {
		return provider.ChatResponse{}, errors.New("choices 为空")
	}

	text := strings.TrimSpace(responseBody.Choices[0].Message.Content)
	if text == "" {
		text = strings.TrimSpace(responseBody.Choices[0].Message.ReasoningContent)
	}

	return provider.ChatResponse{
		Provider: providerInstance.name,
		Model:    responseBody.Model,
		Text:     text,
	}, nil
}

// StreamChat 在兼容接口上先调用普通聊天，再拆成流式片段输出。
func (providerInstance *CompatibleProvider) StreamChat(ctx context.Context, request provider.ChatRequest) (<-chan string, <-chan error) {
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

// Embedding 调用向量接口。
func (providerInstance *CompatibleProvider) Embedding(ctx context.Context, request provider.EmbeddingRequest) (provider.EmbeddingResponse, error) {
	if providerInstance.config.BaseURL == "" || providerInstance.config.APIKey == "" {
		return provider.EmbeddingResponse{}, errors.New("embedding 需要 base_url 和 api_key")
	}

	payload := map[string]any{
		"model": chooseModel(request.Model, providerInstance.config.Model),
		"input": request.Input,
	}

	responseBody := struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}{}

	err := providerInstance.client.PostJSON(
		ctx,
		strings.TrimRight(providerInstance.config.BaseURL, "/")+"/embeddings",
		map[string]string{
			"Authorization": "Bearer " + providerInstance.config.APIKey,
		},
		payload,
		&responseBody,
	)
	if err != nil {
		return provider.EmbeddingResponse{}, err
	}

	if len(responseBody.Data) == 0 {
		return provider.EmbeddingResponse{}, errors.New("embedding 结果为空")
	}

	return provider.EmbeddingResponse{
		Provider: providerInstance.name,
		Vector:   responseBody.Data[0].Embedding,
	}, nil
}

// Moderate 调用审核接口。
func (providerInstance *CompatibleProvider) Moderate(ctx context.Context, request provider.ModerateRequest) (provider.ModerateResponse, error) {
	if providerInstance.config.BaseURL == "" || providerInstance.config.APIKey == "" {
		return provider.ModerateResponse{}, errors.New("moderate 需要 base_url 和 api_key")
	}

	payload := map[string]any{
		"input": request.Input,
	}

	responseBody := struct {
		Results []struct {
			Flagged bool `json:"flagged"`
		} `json:"results"`
	}{}

	err := providerInstance.client.PostJSON(
		ctx,
		strings.TrimRight(providerInstance.config.BaseURL, "/")+"/moderations",
		map[string]string{
			"Authorization": "Bearer " + providerInstance.config.APIKey,
		},
		payload,
		&responseBody,
	)
	if err != nil {
		return provider.ModerateResponse{}, err
	}

	flagged := len(responseBody.Results) > 0 && responseBody.Results[0].Flagged

	return provider.ModerateResponse{
		Provider:    providerInstance.name,
		Flagged:     flagged,
		Reason:      chooseModerationReason(flagged),
		RawDecision: fmt.Sprintf("flagged=%v", flagged),
	}, nil
}

// ImageGenerate 调用图片生成接口。
func (providerInstance *CompatibleProvider) ImageGenerate(ctx context.Context, request provider.ImageRequest) (provider.ImageResponse, error) {
	if providerInstance.config.BaseURL == "" || providerInstance.config.APIKey == "" {
		return provider.ImageResponse{}, errors.New("image_generate 需要 base_url 和 api_key")
	}

	payload := map[string]any{
		"model":  providerInstance.config.Model,
		"prompt": request.Prompt,
		"size":   request.Size,
	}

	responseBody := struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}{}

	err := providerInstance.client.PostJSON(
		ctx,
		strings.TrimRight(providerInstance.config.BaseURL, "/")+"/images/generations",
		map[string]string{
			"Authorization": "Bearer " + providerInstance.config.APIKey,
		},
		payload,
		&responseBody,
	)
	if err != nil {
		return provider.ImageResponse{}, err
	}

	if len(responseBody.Data) == 0 {
		return provider.ImageResponse{}, errors.New("图片生成结果为空")
	}

	return provider.ImageResponse{
		Provider: providerInstance.name,
		URL:      responseBody.Data[0].URL,
	}, nil
}

// TextToSpeech 用一个占位地址演示 TTS 流程。
func (providerInstance *CompatibleProvider) TextToSpeech(ctx context.Context, request provider.TTSRequest) (provider.TTSResponse, error) {
	if providerInstance.config.BaseURL == "" || providerInstance.config.APIKey == "" {
		return provider.TTSResponse{}, errors.New("text_to_speech 需要 base_url 和 api_key")
	}

	_ = ctx
	_ = request

	return provider.TTSResponse{
		Provider: providerInstance.name,
		URL:      "https://placehold.co/600x120?text=TTS+" + providerInstance.name,
	}, nil
}

func chooseModel(requestModel string, fallbackModel string) string {
	if strings.TrimSpace(requestModel) != "" {
		return requestModel
	}

	return fallbackModel
}

func chooseModerationReason(flagged bool) string {
	if flagged {
		return "模型判断该内容可能存在风险"
	}

	return "模型判断该内容可以通过"
}

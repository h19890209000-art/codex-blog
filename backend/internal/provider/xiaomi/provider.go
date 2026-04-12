package xiaomi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/provider"
	"ai-blog/backend/internal/provider/openai"
)

type Provider struct {
	config     config.ProviderConfig
	compatible *openai.CompatibleProvider
}

func New(cfg config.ProviderConfig) provider.Provider {
	return &Provider{
		config:     cfg,
		compatible: openai.NewCompatibleProvider("xiaomi", cfg),
	}
}

func (providerInstance *Provider) Name() string {
	return "xiaomi"
}

func (providerInstance *Provider) Supports(capability provider.Capability) bool {
	switch capability {
	case provider.CapabilityChat, provider.CapabilityStreamChat, provider.CapabilityModerate:
		return true
	default:
		return false
	}
}

func (providerInstance *Provider) Chat(ctx context.Context, request provider.ChatRequest) (provider.ChatResponse, error) {
	return providerInstance.compatible.Chat(ctx, request)
}

func (providerInstance *Provider) StreamChat(ctx context.Context, request provider.ChatRequest) (<-chan string, <-chan error) {
	return providerInstance.compatible.StreamChat(ctx, request)
}

func (providerInstance *Provider) Embedding(ctx context.Context, request provider.EmbeddingRequest) (provider.EmbeddingResponse, error) {
	_ = ctx
	_ = request
	return provider.EmbeddingResponse{}, errors.New("xiaomi provider does not expose embeddings in this project")
}

func (providerInstance *Provider) Moderate(ctx context.Context, request provider.ModerateRequest) (provider.ModerateResponse, error) {
	response, err := providerInstance.compatible.Chat(ctx, provider.ChatRequest{
		Model:       moderationModel(providerInstance.config.Model),
		Temperature: 0.1,
		MaxTokens:   80,
		Messages: []provider.ChatMessage{
			{
				Role:    "system",
				Content: "You are a moderation classifier. Reply with minified JSON only.",
			},
			{
				Role: "user",
				Content: "Return exactly one minified JSON object with keys flagged,reason,category. " +
					"flagged must be boolean. Classify the following user-generated content for spam, scams, illegal content, explicit sexual content, hateful abuse, violent threats, malware, or dangerous instructions. " +
					"If safe, set flagged to false. Content: " + request.Input,
			},
		},
	})
	if err != nil {
		return provider.ModerateResponse{}, err
	}

	flagged, reason, rawDecision := parseModerationDecision(response.Text)
	return provider.ModerateResponse{
		Provider:    providerInstance.Name(),
		Flagged:     flagged,
		Reason:      reason,
		RawDecision: rawDecision,
	}, nil
}

func (providerInstance *Provider) ImageGenerate(ctx context.Context, request provider.ImageRequest) (provider.ImageResponse, error) {
	_ = ctx
	_ = request
	return provider.ImageResponse{}, errors.New("xiaomi provider does not expose image generation in this project")
}

func (providerInstance *Provider) TextToSpeech(ctx context.Context, request provider.TTSRequest) (provider.TTSResponse, error) {
	_ = ctx
	_ = request
	return provider.TTSResponse{}, errors.New("xiaomi provider does not expose text-to-speech in this project")
}

func parseModerationDecision(text string) (bool, string, string) {
	raw := strings.TrimSpace(text)
	if raw == "" {
		return false, "empty moderation decision", raw
	}

	candidate := raw
	if strings.Contains(candidate, "```") {
		candidate = strings.ReplaceAll(candidate, "```json", "")
		candidate = strings.ReplaceAll(candidate, "```", "")
		candidate = strings.TrimSpace(candidate)
	}

	start := strings.Index(candidate, "{")
	end := strings.LastIndex(candidate, "}")
	if start >= 0 && end > start {
		candidate = candidate[start : end+1]
	}

	var parsed struct {
		Flagged  bool   `json:"flagged"`
		Reason   string `json:"reason"`
		Category string `json:"category"`
	}
	if err := json.Unmarshal([]byte(candidate), &parsed); err == nil {
		reason := strings.TrimSpace(parsed.Reason)
		if reason == "" {
			if parsed.Flagged {
				reason = "flagged by xiaomi moderation classifier"
			} else {
				reason = "content is safe"
			}
		}
		if parsed.Category != "" {
			reason = reason + " (" + strings.TrimSpace(parsed.Category) + ")"
		}
		return parsed.Flagged, reason, raw
	}

	lower := strings.ToLower(raw)
	if strings.Contains(lower, `"flagged":false`) || strings.Contains(lower, `"flagged": false`) {
		return false, "content is safe", raw
	}

	if strings.Contains(lower, "safe") && !strings.Contains(lower, "unsafe") {
		return false, "content is safe", raw
	}

	flagged := strings.Contains(lower, `"flagged":true`) ||
		strings.Contains(lower, `"flagged": true`) ||
		strings.Contains(lower, "unsafe") ||
		strings.Contains(lower, "reject") ||
		strings.Contains(lower, "violation")

	if flagged {
		return true, "flagged by heuristic parse of xiaomi moderation response", raw
	}

	return false, "content is safe", raw
}

func moderationModel(currentModel string) string {
	if strings.TrimSpace(currentModel) == "" {
		return "mimo-v2-flash"
	}

	if strings.EqualFold(strings.TrimSpace(currentModel), "mimo-v2-pro") {
		return "mimo-v2-flash"
	}

	return currentModel
}

package service

import (
	"errors"
	"strings"

	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/provider"
	"ai-blog/backend/internal/provider/glm"
	"ai-blog/backend/internal/provider/minimax"
	"ai-blog/backend/internal/provider/ollama"
	"ai-blog/backend/internal/provider/openai"
	"ai-blog/backend/internal/provider/xiaomi"
)

// ProviderStatus 用来给后台页面展示每个模型厂商的状态。
type ProviderStatus struct {
	Alias        string                `json:"alias"`
	Type         string                `json:"type"`
	Ready        bool                  `json:"ready"`
	Model        string                `json:"model"`
	BaseURL      string                `json:"base_url"`
	Capabilities []provider.Capability `json:"capabilities"`
	Reason       string                `json:"reason"`
}

// ProviderRegistry 负责管理所有 Provider。
type ProviderRegistry struct {
	providers map[string]provider.Provider
	configs   map[string]config.ProviderConfig
	routing   config.AIRouting
}

// NewProviderRegistry 创建 Provider 注册表。
func NewProviderRegistry(cfg config.AppConfig) (*ProviderRegistry, error) {
	registry := &ProviderRegistry{
		providers: make(map[string]provider.Provider),
		configs:   make(map[string]config.ProviderConfig),
		routing:   cfg.AI.Routing,
	}

	for alias, providerConfig := range cfg.AI.Providers {
		instance, err := buildProvider(alias, providerConfig)
		if err != nil {
			return nil, err
		}

		registry.providers[alias] = instance
		registry.configs[alias] = providerConfig
	}

	return registry, nil
}

func buildProvider(alias string, providerConfig config.ProviderConfig) (provider.Provider, error) {
	switch providerConfig.Type {
	case "glm":
		return glm.New(providerConfig), nil
	case "minimax":
		return minimax.New(providerConfig), nil
	case "xiaomi":
		return xiaomi.New(providerConfig), nil
	case "openai":
		return openai.NewCompatibleProvider(alias, providerConfig), nil
	case "ollama":
		return ollama.New(providerConfig), nil
	default:
		return nil, errors.New("不支持的 provider 类型: " + providerConfig.Type)
	}
}

// ResolveByFeature 会根据功能名和能力要求选择一个当前可用的 Provider。
func (registry *ProviderRegistry) ResolveByFeature(feature string, capability provider.Capability) (provider.Provider, error) {
	candidates := registry.CandidatesByFeature(feature, capability)
	if len(candidates) == 0 {
		return nil, errors.New("没有找到支持该能力且配置完整的 provider")
	}

	return candidates[0], nil
}

// CandidatesByFeature 返回一个“按优先级排序”的可用 Provider 列表。
// 业务层可以依次尝试这些 Provider，这样当某个厂商临时失败时，能自动降级到下一个。
func (registry *ProviderRegistry) CandidatesByFeature(feature string, capability provider.Capability) []provider.Provider {
	orderedAliases := registry.featureCandidates(feature)
	results := make([]provider.Provider, 0, len(orderedAliases))

	for _, alias := range orderedAliases {
		instance, ok := registry.providers[alias]
		if !ok {
			continue
		}

		if !instance.Supports(capability) {
			continue
		}

		if !registry.isReady(alias, capability) {
			continue
		}

		results = append(results, instance)
	}

	return results
}

// StatusList 返回所有 Provider 的展示信息。
func (registry *ProviderRegistry) StatusList() []ProviderStatus {
	orderedAliases := uniqueAliases("glm", "minimax", "xiaomi", "openai", "ollama")
	results := make([]ProviderStatus, 0, len(orderedAliases))

	for _, alias := range orderedAliases {
		cfg, exists := registry.configs[alias]
		if !exists {
			continue
		}

		status := ProviderStatus{
			Alias:        alias,
			Type:         cfg.Type,
			Model:        cfg.Model,
			BaseURL:      cfg.BaseURL,
			Capabilities: supportedCapabilities(registry.providers[alias]),
		}

		if registry.isReady(alias, provider.CapabilityChat) || registry.isReady(alias, provider.CapabilityImageGenerate) || registry.isReady(alias, provider.CapabilityModerate) {
			status.Ready = true
			status.Reason = "基础配置完整，可参与路由和降级。"
		} else {
			status.Ready = false
			status.Reason = readinessReason(cfg)
		}

		results = append(results, status)
	}

	return results
}

func (registry *ProviderRegistry) featureCandidates(feature string) []string {
	switch feature {
	case "analyze_title":
		return uniqueAliases(registry.routing.AnalyzeTitle, "glm", "minimax", "openai", "ollama")
	case "summary":
		return uniqueAliases(registry.routing.Summary, "glm", "minimax", "openai", "ollama")
	case "chat":
		return uniqueAliases(registry.routing.Chat, "minimax", "glm", "openai", "ollama")
	case "moderate":
		return uniqueAliases(registry.routing.Moderate, "xiaomi", "openai", "glm")
	case "image":
		return uniqueAliases(registry.routing.Image, "openai", "glm")
	case "embedding":
		return uniqueAliases(registry.routing.Embedding, "ollama", "openai", "glm")
	case "tts":
		return uniqueAliases(registry.routing.TTS, "minimax", "openai", "glm")
	default:
		return uniqueAliases("openai", "glm", "minimax", "ollama", "xiaomi")
	}
}

func (registry *ProviderRegistry) isReady(alias string, capability provider.Capability) bool {
	cfg, exists := registry.configs[alias]
	if !exists {
		return false
	}

	switch cfg.Type {
	case "ollama":
		if !isRealValue(cfg.BaseURL) {
			return false
		}

		if capability == provider.CapabilityEmbedding || capability == provider.CapabilityChat || capability == provider.CapabilityStreamChat {
			return true
		}

		return false
	default:
		if !isRealValue(cfg.BaseURL) || !isRealValue(cfg.APIKey) {
			return false
		}

		// 大部分能力都至少需要 base_url + api_key。
		// model 对 moderate 不是强依赖，所以这里单独放宽。
		if capability == provider.CapabilityModerate {
			return true
		}

		return isRealValue(cfg.Model)
	}
}

func supportedCapabilities(instance provider.Provider) []provider.Capability {
	allCapabilities := []provider.Capability{
		provider.CapabilityChat,
		provider.CapabilityStreamChat,
		provider.CapabilityEmbedding,
		provider.CapabilityModerate,
		provider.CapabilityImageGenerate,
		provider.CapabilityTextToSpeech,
	}

	results := make([]provider.Capability, 0, len(allCapabilities))
	for _, capability := range allCapabilities {
		if instance != nil && instance.Supports(capability) {
			results = append(results, capability)
		}
	}

	return results
}

func readinessReason(cfg config.ProviderConfig) string {
	if !isRealValue(cfg.BaseURL) {
		return "缺少 base_url，当前不会参与调用。"
	}

	if cfg.Type != "ollama" && !isRealValue(cfg.APIKey) {
		return "缺少 api_key，当前不会参与调用。"
	}

	if cfg.Type != "ollama" && !isRealValue(cfg.Model) {
		return "缺少 model，当前不会参与调用。"
	}

	if cfg.Type == "ollama" {
		return "等待本地 Ollama 服务可用。"
	}

	return "配置待完善。"
}

func isRealValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}

	lowerValue := strings.ToLower(trimmed)
	if strings.Contains(lowerValue, "replace-with-your") {
		return false
	}

	if strings.Contains(lowerValue, "your-") {
		return false
	}

	if strings.Contains(lowerValue, "example") {
		return false
	}

	return true
}

func uniqueAliases(values ...string) []string {
	seen := make(map[string]bool)
	results := make([]string, 0, len(values))

	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}

		seen[value] = true
		results = append(results, value)
	}

	return results
}

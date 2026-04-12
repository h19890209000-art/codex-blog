package service

import (
	"testing"

	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/provider"
)

func TestProviderRegistryResolveByFeature(t *testing.T) {
	cfg := config.Default()

	// 这里主动给 MiniMax 一个测试 key。
	// 这样测试就能验证“优先路由会先命中配置完整的模型”。
	minimax := cfg.AI.Providers["minimax"]
	minimax.APIKey = "test-minimax-key"
	cfg.AI.Providers["minimax"] = minimax

	registry, err := NewProviderRegistry(cfg)
	if err != nil {
		t.Fatalf("创建 ProviderRegistry 失败: %v", err)
	}

	selectedProvider, err := registry.ResolveByFeature("chat", provider.CapabilityChat)
	if err != nil {
		t.Fatalf("按功能选择 Provider 失败: %v", err)
	}

	if selectedProvider.Name() != "minimax" {
		t.Fatalf("期望默认聊天 Provider 是 minimax，实际是 %s", selectedProvider.Name())
	}
}

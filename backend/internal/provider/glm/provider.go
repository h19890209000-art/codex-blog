package glm

import (
	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/provider"
	"ai-blog/backend/internal/provider/openai"
)

// New 创建智谱 GLM Provider。
// 根据官方文档，GLM 支持 OpenAI 兼容接入，所以这里直接复用兼容实现。
func New(cfg config.ProviderConfig) provider.Provider {
	return openai.NewCompatibleProvider("glm", cfg)
}

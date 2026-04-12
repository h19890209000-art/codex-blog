package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"ai-blog/backend/internal/provider"
	"ai-blog/backend/internal/repository"
	"ai-blog/backend/internal/support"
)

// AIService 负责 AI 相关业务逻辑。
// 这里尽量把每个功能拆成“小步骤”，方便你顺着读。
type AIService struct {
	registry    *ProviderRegistry
	articleRepo repository.ArticleRepository
	commentRepo repository.CommentRepository
}

// NewAIService 创建 AI 服务。
func NewAIService(registry *ProviderRegistry, articleRepo repository.ArticleRepository, commentRepo repository.CommentRepository) *AIService {
	return &AIService{
		registry:    registry,
		articleRepo: articleRepo,
		commentRepo: commentRepo,
	}
}

// AnalyzeTitle 分析标题含义。
func (service *AIService) AnalyzeTitle(ctx context.Context, title string) (map[string]any, error) {
	prompt := "请用中文分析这个博客标题的含义、适合的读者，以及文章可能会展开的核心内容。\n标题：" + title

	text, providerName, err := service.chatWithFallback(ctx, "analyze_title", []provider.ChatMessage{
		{Role: "system", Content: "你是一位擅长拆解技术标题的内容顾问，请直接输出清晰、简洁、可读的中文分析。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return map[string]any{
			"provider": "local-fallback",
			"result":   buildLocalTitleAnalysis(title),
			"hint":     "上游模型暂时不可用，已切换到本地兜底：" + err.Error(),
		}, nil
	}

	return map[string]any{
		"provider": providerName,
		"result":   strings.TrimSpace(text),
	}, nil
}

// GenerateSummary 生成摘要、关键词和 meta 描述。
func (service *AIService) GenerateSummary(ctx context.Context, content string) (map[string]any, error) {
	prompt := "请基于下面这篇技术文章内容，输出三部分内容：\n1. 一段 120 字以内的中文摘要\n2. 5 个关键词，用中文逗号分隔\n3. 一段适合 SEO 的 meta 描述\n\n文章内容：\n" + content

	text, providerName, err := service.chatWithFallback(ctx, "summary", []provider.ChatMessage{
		{Role: "system", Content: "你是一位技术博客编辑，请直接输出摘要、关键词和 meta 描述，不要写多余寒暄。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		result := buildLocalSummary(content)
		result["hint"] = "上游模型暂时不可用，已切换到本地兜底：" + err.Error()
		return result, nil
	}

	summary, keywords, meta := parseSummaryResponse(text, content)

	return map[string]any{
		"provider": providerName,
		"summary":  summary,
		"keywords": keywords,
		"meta":     meta,
		"raw":      strings.TrimSpace(text),
	}, nil
}

// SuggestTags 生成标签建议。
func (service *AIService) SuggestTags(ctx context.Context, content string) (map[string]any, error) {
	prompt := "请根据下面这篇文章内容，给我 5 个简短的中文标签。只返回标签本身，每行一个，不要解释。\n\n文章内容：\n" + content

	text, providerName, err := service.chatWithFallback(ctx, "summary", []provider.ChatMessage{
		{Role: "system", Content: "你是一位技术内容运营助手，输出标签时要短、准、清晰。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return map[string]any{
			"provider": "local-fallback",
			"tags":     []string{"Go", "AI", "博客", "开发实践", "后端"},
			"hint":     "上游模型暂时不可用，已切换到本地兜底：" + err.Error(),
		}, nil
	}

	return map[string]any{
		"provider": providerName,
		"tags":     parseTagLines(text),
	}, nil
}

// Brainstorm 根据关键词生成选题灵感。
func (service *AIService) Brainstorm(ctx context.Context, keyword string) (map[string]any, error) {
	prompt := "围绕这个关键词，生成 5 个适合技术博客的选题。每个选题占一行，格式为：标题 - 一句说明。\n关键词：" + keyword

	text, providerName, err := service.chatWithFallback(ctx, "summary", []provider.ChatMessage{
		{Role: "system", Content: "你是一位技术博客选题策划助手，请给出具体、可写、可落地的标题灵感。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return map[string]any{
			"provider": "local-fallback",
			"items": []string{
				keyword + " 入门避坑清单",
				keyword + " 从 0 到 1 实战指南",
				keyword + " 在真实项目里的使用方式",
				keyword + " 最常见错误和修复办法",
				keyword + " 面向新手的完整上手路线",
			},
			"hint": "上游模型暂时不可用，已切换到本地兜底：" + err.Error(),
		}, nil
	}

	items := support.SplitLines(text)
	return map[string]any{
		"provider": providerName,
		"items":    items,
		"result":   strings.TrimSpace(text),
	}, nil
}

// Rewrite 按指定风格改写文本。
func (service *AIService) Rewrite(ctx context.Context, content string, style string) (map[string]any, error) {
	prompt := fmt.Sprintf("请把下面内容改写成“%s”风格，要求语义不变、表达更自然、结构更清晰。\n\n原文：\n%s", style, content)

	text, providerName, err := service.chatWithFallback(ctx, "summary", []provider.ChatMessage{
		{Role: "system", Content: "你是一位擅长中文润色的技术编辑，请直接输出改写后的结果。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return map[string]any{
			"provider": "local-fallback",
			"result":   "【" + style + "风格改写】\n" + content,
			"hint":     "上游模型暂时不可用，已切换到本地兜底：" + err.Error(),
		}, nil
	}

	return map[string]any{
		"provider": providerName,
		"result":   strings.TrimSpace(text),
	}, nil
}

// GenerateCover 生成封面图。
// 如果当前没有可用的图像模型，就立即返回一个本地占位图，避免长时间等待。
func (service *AIService) GenerateCover(ctx context.Context, title string) (map[string]any, error) {
	imageProvider, err := service.registry.ResolveByFeature("image", provider.CapabilityImageGenerate)
	if err != nil {
		return localCoverResult(title, "当前没有可用的图像模型，已返回占位封面图。"), nil
	}

	imageResponse, err := imageProvider.ImageGenerate(ctx, provider.ImageRequest{
		Prompt: "请为这篇技术博客标题生成一张干净、现代、适合文章封面的横版图片：" + title,
		Size:   "1200x630",
	})
	if err != nil {
		return localCoverResult(title, "图像模型调用失败，已自动切回占位封面图："+err.Error()), nil
	}

	return map[string]any{
		"provider": imageResponse.Provider,
		"url":      imageResponse.URL,
		"hint":     "已成功从图像模型生成封面地址。",
	}, nil
}

// ReplySuggestions 生成评论回复建议。
func (service *AIService) ReplySuggestions(ctx context.Context, commentID int64) (map[string]any, error) {
	comment, err := service.commentRepo.FindByID(commentID)
	if err != nil {
		return nil, err
	}

	text, providerName, err := service.chatWithFallback(ctx, "chat", []provider.ChatMessage{
		{Role: "system", Content: "你是一位友好、专业的技术博主，请输出适合直接发送给读者的中文回复建议。"},
		{Role: "user", Content: "请基于这条评论，生成 3 条适合博主直接回复的短句：\n" + comment.Content},
	})
	if err != nil {
		return map[string]any{
			"provider": "local-fallback",
			"replies": []string{
				"这个问题提得很好，我后面会补一段更具体的示例。",
				"你已经抓住重点了，关键就在职责边界的划分。",
				"如果你愿意，我后面可以专门再写一篇把这块展开讲。",
			},
			"hint": "上游模型暂时不可用，已切换到本地兜底：" + err.Error(),
		}, nil
	}

	return map[string]any{
		"provider": providerName,
		"replies":  parseTagLines(text),
		"result":   strings.TrimSpace(text),
	}, nil
}

// ArticleQA 负责文章级问答。
func (service *AIService) ArticleQA(ctx context.Context, articleID int64, question string) (map[string]any, error) {
	article, err := service.articleRepo.FindByID(articleID)
	if err != nil {
		return nil, err
	}

	userPrompt := "文章标题：" + article.Title + "\n文章内容：" + article.Content + "\n读者问题：" + question + "\n请只基于文章内容回答，并在结尾说明这是根据当前文章得出的结论。"
	text, providerName, err := service.chatWithFallback(ctx, "chat", []provider.ChatMessage{
		{Role: "system", Content: "你是一位只基于给定文章内容回答问题的技术助手，不要编造文章里没有的事实。"},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		return map[string]any{
			"provider": "local-fallback",
			"answer":   "根据文章《" + article.Title + "》，核心答案是：" + support.TrimText(article.Content, 90),
			"citation": "article:" + strconv.FormatInt(article.ID, 10),
			"hint":     "上游模型暂时不可用，已切换到本地兜底：" + err.Error(),
		}, nil
	}

	return map[string]any{
		"provider": providerName,
		"answer":   strings.TrimSpace(text),
		"citation": "article:" + strconv.FormatInt(article.ID, 10),
	}, nil
}

// SiteQA 负责全站知识库问答。
func (service *AIService) SiteQA(ctx context.Context, question string) (map[string]any, error) {
	matchedArticles, err := service.articleRepo.Search(question)
	if err != nil {
		return nil, err
	}

	contextLines := make([]string, 0, len(matchedArticles))
	citations := make([]string, 0, len(matchedArticles))

	for _, article := range matchedArticles {
		contextLines = append(contextLines, article.Title+"："+article.Content)
		citations = append(citations, "article:"+strconv.FormatInt(article.ID, 10))
	}

	if len(contextLines) == 0 {
		contextLines = append(contextLines, "没有检索到强相关站内文章，请明确说明不确定。")
	}

	userPrompt := "站内检索内容如下：\n" + strings.Join(contextLines, "\n\n") + "\n\n用户问题：" + question + "\n请优先引用站内内容来回答。"
	text, providerName, err := service.chatWithFallback(ctx, "chat", []provider.ChatMessage{
		{Role: "system", Content: "你是一位博客知识库助手，回答时优先引用站内内容，不确定时要明确说明。"},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		return map[string]any{
			"provider":  "local-fallback",
			"answer":    "我当前根据站内文章检索到的线索是：" + strings.Join(contextLines, "；"),
			"citations": citations,
			"hint":      "上游模型暂时不可用，已切换到本地兜底：" + err.Error(),
		}, nil
	}

	return map[string]any{
		"provider":  providerName,
		"answer":    strings.TrimSpace(text),
		"citations": citations,
	}, nil
}

// ModerateComment 审核评论内容。
func (service *AIService) ModerateComment(ctx context.Context, content string) (map[string]any, error) {
	moderateProvider, err := service.registry.ResolveByFeature("moderate", provider.CapabilityModerate)
	if err != nil {
		return localModeration(content), nil
	}

	result, err := moderateProvider.Moderate(ctx, provider.ModerateRequest{Input: content})
	if err != nil {
		return localModeration(content), nil
	}

	return map[string]any{
		"provider": result.Provider,
		"flagged":  result.Flagged,
		"reason":   result.Reason,
	}, nil
}

// chatWithFallback 会依次尝试多个候选 Provider。
// 这样某个模型调用失败时，系统可以自动降级，而不是直接报错。
func (service *AIService) chatWithFallback(ctx context.Context, feature string, messages []provider.ChatMessage) (string, string, error) {
	candidates := service.registry.CandidatesByFeature(feature, provider.CapabilityChat)
	if len(candidates) == 0 {
		return "", "", fmt.Errorf("feature %s 没有可用的聊天模型", feature)
	}

	errorMessages := make([]string, 0, len(candidates))
	for _, selectedProvider := range candidates {
		response, err := selectedProvider.Chat(ctx, provider.ChatRequest{
			Messages:    messages,
			Temperature: 0.7,
			MaxTokens:   800,
		})
		if err != nil {
			errorMessages = append(errorMessages, selectedProvider.Name()+": "+err.Error())
			continue
		}

		if strings.TrimSpace(response.Text) == "" {
			errorMessages = append(errorMessages, selectedProvider.Name()+": 返回了空文本")
			continue
		}

		return response.Text, response.Provider, nil
	}

	if len(errorMessages) == 0 {
		return "", "", fmt.Errorf("feature %s 调用失败", feature)
	}

	return "", "", fmt.Errorf("所有候选模型都调用失败：%s", strings.Join(errorMessages, " | "))
}

func parseSummaryResponse(text string, originalContent string) (string, []string, string) {
	cleanText := strings.TrimSpace(text)
	lines := support.SplitLines(cleanText)

	if len(lines) == 0 {
		fallback := buildLocalSummary(originalContent)
		return fallback["summary"].(string), fallback["keywords"].([]string), fallback["meta"].(string)
	}

	summary := lines[0]
	keywords := make([]string, 0)
	meta := ""

	for _, line := range lines[1:] {
		lowerLine := strings.ToLower(line)

		if strings.Contains(lowerLine, "关键词") || strings.Contains(lowerLine, "keywords") {
			candidate := strings.NewReplacer("关键词：", "", "关键词:", "", "keywords:", "", "Keywords:", "").Replace(line)
			keywords = parseCommaSeparatedTags(candidate)
			continue
		}

		if strings.Contains(lowerLine, "meta") || strings.Contains(lowerLine, "描述") {
			meta = strings.NewReplacer("meta：", "", "meta:", "", "描述：", "", "描述:", "").Replace(line)
			meta = strings.TrimSpace(meta)
			continue
		}

		if meta == "" {
			meta = strings.TrimSpace(line)
		}
	}

	if len(keywords) == 0 {
		keywords = parseTagLines(cleanText)
		if len(keywords) > 5 {
			keywords = keywords[:5]
		}
	}

	if meta == "" {
		meta = support.TrimText(summary, 90)
	}

	return strings.TrimSpace(summary), keywords, meta
}

func parseTagLines(text string) []string {
	lines := support.SplitLines(text)
	results := make([]string, 0, len(lines))

	for _, line := range lines {
		cleanLine := strings.TrimSpace(line)
		cleanLine = strings.TrimLeft(cleanLine, "-•0123456789.、 ")
		cleanLine = strings.TrimSpace(cleanLine)
		if cleanLine == "" {
			continue
		}

		// 遇到逗号分隔时，拆开后再继续清洗。
		if strings.Contains(cleanLine, "，") || strings.Contains(cleanLine, ",") {
			results = append(results, parseCommaSeparatedTags(cleanLine)...)
			continue
		}

		results = append(results, cleanLine)
	}

	if len(results) == 0 {
		return []string{"Go", "AI", "博客"}
	}

	return uniqueStrings(results)
}

func parseCommaSeparatedTags(text string) []string {
	normalized := strings.NewReplacer("，", ",", "、", ",", "；", ",", ";", ",").Replace(text)
	parts := strings.Split(normalized, ",")
	results := make([]string, 0, len(parts))

	for _, part := range parts {
		cleanPart := strings.TrimSpace(part)
		cleanPart = strings.TrimLeft(cleanPart, "-•0123456789.、 ")
		cleanPart = strings.TrimSpace(cleanPart)
		if cleanPart == "" {
			continue
		}
		results = append(results, cleanPart)
	}

	return uniqueStrings(results)
}

func uniqueStrings(values []string) []string {
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

func buildLocalTitleAnalysis(title string) string {
	return "这个标题的核心主题是“" + title + "”。\n\n它大概率会围绕问题背景、实现思路和落地细节展开。\n\n对于读者来说，最大的收获是能快速判断这篇文章是否值得继续阅读。"
}

func buildLocalSummary(content string) map[string]any {
	short := support.TrimText(content, 120)

	return map[string]any{
		"provider": "local-fallback",
		"summary":  short,
		"keywords": []string{"技术博客", "AI", "Go", "Vue3", "教程"},
		"meta":     "这是一篇关于技术实现与 AI 集成实践的文章。",
	}
}

func localModeration(content string) map[string]any {
	lowerContent := strings.ToLower(content)
	flagged := strings.Contains(lowerContent, "spam") || strings.Contains(lowerContent, "广告")

	reason := "内容正常"
	if flagged {
		reason = "命中了简单的本地审核规则"
	}

	return map[string]any{
		"provider": "local-fallback",
		"flagged":  flagged,
		"reason":   reason,
	}
}

func localCoverResult(title string, hint string) map[string]any {
	return map[string]any{
		"provider": "local-fallback",
		"url":      "https://placehold.co/1200x630?text=" + title,
		"hint":     hint,
	}
}

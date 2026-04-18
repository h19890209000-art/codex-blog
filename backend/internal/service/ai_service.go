package service

import (
	"context"
	"encoding/json"
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
func (service *AIService) TranslateBriefingForStudy(ctx context.Context, title string, content string) (map[string]any, error) {
	prompt := "请把下面这篇英文 AI 新闻正文翻译成自然、准确、适合中国英语初学者阅读的中文。要求保留原文段落结构，不要额外添加小标题，不要省略信息。\n\n标题：" + title + "\n\n正文：\n" + content

	text, providerName, err := service.chatWithFallbackWithOptions(ctx, "chat", []provider.ChatMessage{
		{Role: "system", Content: "你是一位双语新闻编辑，请输出忠实、清晰、适合中国用户学习英语的中文译文，并保留段落结构。"},
		{Role: "user", Content: prompt},
	}, 0.4, 1800)
	if err != nil {
		return map[string]any{
			"provider":    "local-fallback",
			"translation": "暂时无法生成完整译文，你可以先阅读左侧原文并结合标题、摘要学习。",
			"hint":        "上游模型暂时不可用，已切换到本地兜底：" + err.Error(),
		}, nil
	}

	return map[string]any{
		"provider":    providerName,
		"translation": strings.TrimSpace(text),
	}, nil
}

func (service *AIService) ExplainEnglishWord(ctx context.Context, title string, sentence string, word string) (map[string]any, error) {
	prompt := "请结合下面这句英文，解释指定单词在当前语境里的意思。请严格返回 JSON，不要使用 Markdown 代码块。\n" +
		`{"word":"","meaning":"","part_of_speech":"","phonetic":"","usage":""}` +
		"\n\n标题：" + title + "\n句子：" + sentence + "\n单词：" + word

	text, providerName, err := service.chatWithFallbackWithOptions(ctx, "chat", []provider.ChatMessage{
		{Role: "system", Content: "你是一位帮助中国初学者学英语的老师，请用中文解释单词，并严格返回 JSON。"},
		{Role: "user", Content: prompt},
	}, 0.2, 320)
	if err != nil {
		result := buildLocalWordExplanation(word, sentence)
		result["hint"] = "上游模型暂时不可用，已切换到本地兜底：" + err.Error()
		return result, nil
	}

	parsed := make(map[string]any)
	if parseJSONObject(text, &parsed) {
		if isEmptyWordExplanation(parsed) {
			result := buildLocalWordExplanation(word, sentence)
			result["provider"] = providerName
			result["hint"] = "model returned an empty structured answer, switched to local fallback"
			return result, nil
		}
		parsed["provider"] = providerName
		if parsed["word"] == nil || strings.TrimSpace(fmt.Sprint(parsed["word"])) == "" {
			parsed["word"] = word
		}
		return parsed, nil
	}

	return map[string]any{
		"provider":       providerName,
		"word":           word,
		"meaning":        strings.TrimSpace(text),
		"part_of_speech": "",
		"phonetic":       "",
		"usage":          "请结合当前句子理解该词含义。",
	}, nil
}

func (service *AIService) AnalyzeEnglishSentence(ctx context.Context, title string, sentence string) (map[string]any, error) {
	prompt := "请分析下面这句英文新闻句子，帮助中文母语的英语初学者理解。请严格返回 JSON，不要使用 Markdown 代码块。\n" +
		`{"sentence":"","translation":"","explanation":"","structure":"","subject":"","predicate":"","object":"","grammar_points":[]}` +
		"\n\n标题：" + title + "\n句子：" + sentence

	text, providerName, err := service.chatWithFallbackWithOptions(ctx, "chat", []provider.ChatMessage{
		{Role: "system", Content: "你是一位英语精读老师，请用中文解释句子结构、主谓宾和语法点，并严格返回 JSON。"},
		{Role: "user", Content: prompt},
	}, 0.2, 800)
	if err != nil {
		result := buildLocalSentenceAnalysis(sentence)
		result["hint"] = "上游模型暂时不可用，已切换到本地兜底：" + err.Error()
		return result, nil
	}

	parsed := make(map[string]any)
	if parseJSONObject(text, &parsed) {
		if isEmptySentenceAnalysis(parsed) {
			result := buildLocalSentenceAnalysis(sentence)
			result["provider"] = providerName
			result["hint"] = "model returned an empty structured answer, switched to local fallback"
			return result, nil
		}
		parsed["provider"] = providerName
		if parsed["sentence"] == nil || strings.TrimSpace(fmt.Sprint(parsed["sentence"])) == "" {
			parsed["sentence"] = sentence
		}
		return parsed, nil
	}

	return map[string]any{
		"provider":       providerName,
		"sentence":       sentence,
		"translation":    "",
		"explanation":    strings.TrimSpace(text),
		"structure":      "请结合说明理解句子结构",
		"subject":        "",
		"predicate":      "",
		"object":         "",
		"grammar_points": []string{},
	}, nil
}

func (service *AIService) BuildBriefingLearningPlan(ctx context.Context, title string, summary string, sourceContent string, translatedContent string, goal string) (map[string]any, error) {
	goal = strings.TrimSpace(goal)
	if goal == "" {
		goal = "我是中国初学者，想用每天的 AI 资讯练英语，重点是能听懂、能复述、能在技术交流里说简单英语。"
	}

	prompt := "Return strict JSON only for a 10-minute English micro-lesson for a Chinese beginner. " +
		"Use the AI news material below and keep the lesson practical, scene-first, and reusable. " +
		"Schema: " +
		`{"goal_profile":{"track":"","level":"","first_scene":"","reason_cn":""},"daily_flow":[{"step":"","title_cn":"","instruction_cn":"","output_en":""}],"chunks":[{"phrase":"","translation_cn":"","why_it_works_cn":"","example_en":"","substitution_options":[],"coach_tip_cn":""}],"roleplay":{"scene":"","goal_cn":"","target_output_en":"","keywords":[],"opening_en":"","opening_cn":"","help_cn":""},"review_cards":[{"front":"","back":"","note_cn":"","category":""}]}` +
		" Keep daily_flow at 4 items, chunks at 5 to 8 items, review_cards at 4 to 6 items. " +
		" All explanations should be concise Chinese, while phrase/example/opening/target_output should be English. " +
		"\n\nUser goal in Chinese:\n" + goal +
		"\n\nNews title:\n" + title +
		"\n\nNews summary:\n" + support.TrimText(summary, 260) +
		"\n\nEnglish source:\n" + support.TrimText(sourceContent, 2400) +
		"\n\nChinese translation:\n" + support.TrimText(translatedContent, 1800)

	text, providerName, err := service.chatWithFallbackWithOptions(ctx, "chat", []provider.ChatMessage{
		{Role: "system", Content: "You are an English coach for Chinese beginners. Return strict JSON only."},
		{Role: "user", Content: prompt},
	}, 0.2, 1200)
	if err != nil {
		result := buildLocalBriefingLearningPlan(title, summary, sourceContent, goal)
		result["hint"] = "upstream model unavailable, switched to local fallback: " + err.Error()
		return result, nil
	}

	parsed := make(map[string]any)
	if parseJSONObject(text, &parsed) {
		if isEmptyLearningPlan(parsed) {
			result := buildLocalBriefingLearningPlan(title, summary, sourceContent, goal)
			result["provider"] = providerName
			result["hint"] = "model returned an empty lesson plan, switched to local fallback"
			return result, nil
		}
		parsed["provider"] = providerName
		return parsed, nil
	}

	result := buildLocalBriefingLearningPlan(title, summary, sourceContent, goal)
	result["provider"] = providerName
	result["hint"] = "model response was not valid JSON, switched to local fallback"
	return result, nil
}

func (service *AIService) RunBriefingRoleplay(ctx context.Context, title string, summary string, goal string, scene string, learnerReply string) (map[string]any, error) {
	goal = strings.TrimSpace(goal)
	scene = strings.TrimSpace(scene)
	learnerReply = strings.TrimSpace(learnerReply)
	if goal == "" {
		goal = "我想练到能在技术沟通里用简单英语说明重点。"
	}
	if scene == "" {
		scene = "technical meeting"
	}

	prompt := "Return strict JSON only for a semi-open roleplay coach response. " +
		"The learner is a Chinese beginner and may answer in Chinese, broken English, or mixed language. " +
		"You must first judge whether the learner can be understood, then give a short Chinese correction and a better English reply. " +
		"Schema: " +
		`{"scene":"","can_be_understood_score":0,"coach_reply_en":"","coach_reply_cn":"","correction_cn":"","better_reply_en":"","next_prompt_en":"","next_prompt_cn":"","key_mistakes":[]}` +
		"\n\nNews title:\n" + title +
		"\n\nNews summary:\n" + support.TrimText(summary, 260) +
		"\n\nLearner goal in Chinese:\n" + goal +
		"\n\nRoleplay scene:\n" + scene +
		"\n\nLearner reply:\n" + learnerReply

	text, providerName, err := service.chatWithFallbackWithOptions(ctx, "chat", []provider.ChatMessage{
		{Role: "system", Content: "You are an English speaking coach for Chinese beginners. Return strict JSON only."},
		{Role: "user", Content: prompt},
	}, 0.2, 700)
	if err != nil {
		result := buildLocalRoleplayResult(scene, learnerReply)
		result["hint"] = "upstream model unavailable, switched to local fallback: " + err.Error()
		return result, nil
	}

	parsed := make(map[string]any)
	if parseJSONObject(text, &parsed) {
		if isEmptyRoleplayResult(parsed) {
			result := buildLocalRoleplayResult(scene, learnerReply)
			result["provider"] = providerName
			result["hint"] = "model returned an empty roleplay result, switched to local fallback"
			return result, nil
		}
		if scoreValue, ok := parsed["can_be_understood_score"]; ok {
			if score, err := strconv.Atoi(strings.TrimSpace(fmt.Sprint(scoreValue))); err == nil && score >= 0 && score <= 10 {
				parsed["can_be_understood_score"] = score * 10
			}
		}
		parsed["provider"] = providerName
		return parsed, nil
	}

	result := buildLocalRoleplayResult(scene, learnerReply)
	result["provider"] = providerName
	result["hint"] = "model response was not valid JSON, switched to local fallback"
	return result, nil
}

func (service *AIService) chatWithFallback(ctx context.Context, feature string, messages []provider.ChatMessage) (string, string, error) {
	return service.chatWithFallbackWithOptions(ctx, feature, messages, 0.7, 800)
}

func (service *AIService) chatWithFallbackWithOptions(ctx context.Context, feature string, messages []provider.ChatMessage, temperature float64, maxTokens int) (string, string, error) {
	candidates := service.registry.CandidatesByFeature(feature, provider.CapabilityChat)
	if len(candidates) == 0 {
		return "", "", fmt.Errorf("feature %s 没有可用的聊天模型", feature)
	}

	errorMessages := make([]string, 0, len(candidates))
	for _, selectedProvider := range candidates {
		response, err := selectedProvider.Chat(ctx, provider.ChatRequest{
			Messages:    messages,
			Temperature: temperature,
			MaxTokens:   maxTokens,
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

func parseJSONObject(text string, target any) bool {
	cleanText := strings.TrimSpace(text)
	cleanText = strings.TrimPrefix(cleanText, "```json")
	cleanText = strings.TrimPrefix(cleanText, "```")
	cleanText = strings.TrimSuffix(cleanText, "```")
	cleanText = strings.TrimSpace(cleanText)

	start := strings.Index(cleanText, "{")
	end := strings.LastIndex(cleanText, "}")
	if start >= 0 && end > start {
		cleanText = cleanText[start : end+1]
	}

	return json.Unmarshal([]byte(cleanText), target) == nil
}

func isEmptyWordExplanation(parsed map[string]any) bool {
	return strings.TrimSpace(fmt.Sprint(parsed["meaning"])) == "" &&
		strings.TrimSpace(fmt.Sprint(parsed["part_of_speech"])) == "" &&
		strings.TrimSpace(fmt.Sprint(parsed["phonetic"])) == "" &&
		strings.TrimSpace(fmt.Sprint(parsed["usage"])) == ""
}

func isEmptySentenceAnalysis(parsed map[string]any) bool {
	return strings.TrimSpace(fmt.Sprint(parsed["translation"])) == "" &&
		strings.TrimSpace(fmt.Sprint(parsed["explanation"])) == "" &&
		strings.TrimSpace(fmt.Sprint(parsed["structure"])) == "" &&
		strings.TrimSpace(fmt.Sprint(parsed["subject"])) == "" &&
		strings.TrimSpace(fmt.Sprint(parsed["predicate"])) == "" &&
		strings.TrimSpace(fmt.Sprint(parsed["object"])) == "" &&
		len(toStringSlice(parsed["grammar_points"])) == 0
}

func isEmptyLearningPlan(parsed map[string]any) bool {
	return len(toMapSlice(parsed["daily_flow"])) == 0 &&
		len(toMapSlice(parsed["chunks"])) == 0 &&
		len(toMapSlice(parsed["review_cards"])) == 0
}

func isEmptyRoleplayResult(parsed map[string]any) bool {
	return strings.TrimSpace(fmt.Sprint(parsed["coach_reply_en"])) == "" &&
		strings.TrimSpace(fmt.Sprint(parsed["better_reply_en"])) == "" &&
		strings.TrimSpace(fmt.Sprint(parsed["correction_cn"])) == ""
}

func toStringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		if value == nil {
			return []string{}
		}

		text := strings.TrimSpace(fmt.Sprint(value))
		if text == "" {
			return []string{}
		}

		return []string{text}
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		text := strings.TrimSpace(fmt.Sprint(item))
		if text != "" {
			result = append(result, text)
		}
	}

	return result
}

func toMapSlice(value any) []map[string]any {
	items, ok := value.([]any)
	if !ok {
		return []map[string]any{}
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		parsed, ok := item.(map[string]any)
		if ok {
			result = append(result, parsed)
		}
	}

	return result
}

func buildLocalWordExplanation(word string, sentence string) map[string]any {
	return map[string]any{
		"provider":       "local-fallback",
		"word":           word,
		"meaning":        "请结合当前句子理解这个单词的大致含义。",
		"part_of_speech": "待判断",
		"phonetic":       "",
		"usage":          "句子语境：" + support.TrimText(sentence, 120),
	}
}

func buildLocalBriefingLearningPlan(title string, summary string, sourceContent string, goal string) map[string]any {
	track, firstScene, reason := classifyGoalProfile(goal, title)
	sentences := splitSourceSentences(sourceContent, 6)
	if len(sentences) == 0 {
		sentences = []string{
			title,
			summary,
			"The key point is that this AI update matters in real work.",
			"I can explain the main idea in simple English.",
			"This helps me join a short technical discussion.",
		}
	}

	chunks := make([]map[string]any, 0, 6)
	for index, sentence := range sentences {
		if strings.TrimSpace(sentence) == "" {
			continue
		}
		chunks = append(chunks, map[string]any{
			"phrase":               support.TrimText(sentence, 120),
			"translation_cn":       "先把这句话当成整块输入，重点练到能听懂、能复述，而不是逐词翻译。",
			"why_it_works_cn":      "这是资讯场景里最常见的说明句，学会后可以直接拿去做复述。",
			"example_en":           buildExampleSentence(sentence, index),
			"substitution_options": buildSubstitutions(track),
			"coach_tip_cn":         "先听一句，再跟一句，然后把其中一个关键词替换掉自己再说一次。",
		})
		if len(chunks) >= 6 {
			break
		}
	}

	reviewCards := make([]map[string]any, 0, 4)
	for index, chunk := range chunks {
		reviewCards = append(reviewCards, map[string]any{
			"front":    strings.TrimSpace(fmt.Sprint(chunk["phrase"])),
			"back":     strings.TrimSpace(fmt.Sprint(chunk["example_en"])),
			"note_cn":  "回忆中文意思，再用自己的话复述一遍。",
			"category": fmt.Sprintf("lesson-%d", index+1),
		})
		if len(reviewCards) >= 4 {
			break
		}
	}

	return map[string]any{
		"provider": "local-fallback",
		"goal_profile": map[string]any{
			"track":       track,
			"level":       "beginner",
			"first_scene": firstScene,
			"reason_cn":   reason,
		},
		"daily_flow": []map[string]any{
			{"step": "1", "title_cn": "听一句", "instruction_cn": "先点句块播放，先追求能不能听懂，不用急着抠口音。", "output_en": "I can catch the main idea."},
			{"step": "2", "title_cn": "跟一句", "instruction_cn": "跟读 2 到 3 次，保持整块输出，不拆成单词。", "output_en": "Let me say that again."},
			{"step": "3", "title_cn": "换一句", "instruction_cn": "把一个关键词换掉，做最小替换练习。", "output_en": "The key point is that..."},
			{"step": "4", "title_cn": "说一句", "instruction_cn": "结合今天资讯，用一句自己的英语总结重点。", "output_en": "My short summary is..."},
		},
		"chunks": chunks,
		"roleplay": map[string]any{
			"scene":            firstScene,
			"goal_cn":          "你要用 1 到 2 句英语，把今天的资讯重点说给对方听。",
			"target_output_en": "The key point is that this AI update improves real work.",
			"keywords":         buildSubstitutions(track),
			"opening_en":       buildRoleplayOpening(firstScene),
			"opening_cn":       "先不用追求复杂表达，能让对方听懂就算赢。",
			"help_cn":          "你可以先用中文想意思，再压缩成简单英语，比如：The key point is... / This means... / In practice...",
		},
		"review_cards": reviewCards,
	}
}

func buildLocalRoleplayResult(scene string, learnerReply string) map[string]any {
	score := 68
	keyMistakes := []string{"先缩短句子，优先保证对方能听懂", "尽量用一个主句表达一个重点"}
	betterReply := "The key point is that this AI news shows a practical improvement."
	if hasEnglishLetters(learnerReply) {
		score = 78
		keyMistakes = []string{"表达方向是对的，可以再更短一点", "把关键词提前，句子会更稳"}
		betterReply = support.TrimText(strings.TrimSpace(learnerReply), 120)
		if !strings.HasSuffix(betterReply, ".") {
			betterReply += "."
		}
	}

	return map[string]any{
		"provider":                "local-fallback",
		"scene":                   scene,
		"can_be_understood_score": score,
		"coach_reply_en":          "I understood your main idea. Let us make it shorter and clearer.",
		"coach_reply_cn":          "我大致能听懂你的意思，下一步把句子压短一点会更自然。",
		"correction_cn":           "先说结论，再补充细节。对初学者来说，短句比复杂长句更有效。",
		"better_reply_en":         betterReply,
		"next_prompt_en":          "Now say it again in one short sentence.",
		"next_prompt_cn":          "现在请你再用一句更短的英语重复一遍。",
		"key_mistakes":            keyMistakes,
	}
}

func classifyGoalProfile(goal string, title string) (string, string, string) {
	lowerGoal := strings.ToLower(goal + " " + title)
	switch {
	case strings.Contains(goal, "面试") || strings.Contains(lowerGoal, "interview"):
		return "interview", "job interview", "系统判断你更需要面试场景，所以会优先训练简短自我表达、说明项目和回答追问。"
	case strings.Contains(goal, "出差") || strings.Contains(goal, "旅行") || strings.Contains(lowerGoal, "travel"):
		return "travel", "business trip", "系统判断你更需要出行沟通，所以会优先训练问路、确认安排、请求帮助这类句块。"
	case strings.Contains(goal, "会议") || strings.Contains(goal, "技术") || strings.Contains(goal, "工作") || strings.Contains(lowerGoal, "ai") || strings.Contains(lowerGoal, "model"):
		return "tech-work", "technical meeting", "系统判断你更偏职场和技术交流，所以会优先训练说明重点、复述结论和提问题。"
	default:
		return "daily-life", "daily conversation", "系统先按通用日常沟通来安排，重点是高频句块、听懂和开口。"
	}
}

func splitSourceSentences(source string, limit int) []string {
	result := make([]string, 0, limit)
	for _, part := range strings.Split(strings.ReplaceAll(source, "\r", ""), "\n") {
		text := strings.TrimSpace(part)
		if len(text) < 36 {
			continue
		}
		if strings.Count(text, "/") >= 2 {
			continue
		}
		if strings.Count(text, " ") < 4 {
			continue
		}
		text = support.TrimText(text, 140)
		result = append(result, text)
		if len(result) >= limit {
			break
		}
	}

	return result
}

func buildExampleSentence(sentence string, index int) string {
	examples := []string{
		"The key point is that this update changes real work.",
		"This means the model can do more useful tasks.",
		"In practice, teams may use it in daily work.",
		"I want a simple explanation of this result.",
		"My short summary is that the benchmark is not enough.",
		"I have one question about the real impact.",
	}

	if index >= 0 && index < len(examples) {
		return examples[index]
	}

	return support.TrimText(sentence, 100)
}

func buildSubstitutions(track string) []string {
	switch track {
	case "interview":
		return []string{"project", "experience", "result", "challenge"}
	case "travel":
		return []string{"schedule", "hotel", "ticket", "help"}
	case "tech-work":
		return []string{"model", "workflow", "result", "risk"}
	default:
		return []string{"idea", "plan", "problem", "help"}
	}
}

func buildRoleplayOpening(scene string) string {
	switch scene {
	case "job interview":
		return "Please explain this AI update in one simple English answer, like you are in an interview."
	case "business trip":
		return "Your colleague asks what this AI news means for work during a business trip."
	case "technical meeting":
		return "You are in a short technical meeting. Please summarize today's AI briefing in simple English."
	default:
		return "Tell me the main point of today's AI news in one simple English sentence."
	}
}

func hasEnglishLetters(text string) bool {
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}

	return false
}

func buildLocalSentenceAnalysis(sentence string) map[string]any {
	return map[string]any{
		"provider":       "local-fallback",
		"sentence":       sentence,
		"translation":    "",
		"explanation":    "请先结合上下文理解这句话，系统暂时无法给出更完整的语法解析。",
		"structure":      "请重点观察主语、谓语和宾语的位置",
		"subject":        "",
		"predicate":      "",
		"object":         "",
		"grammar_points": []string{"先识别句子主干，再看修饰成分"},
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

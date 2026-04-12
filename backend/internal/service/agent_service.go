package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"ai-blog/backend/internal/provider"
	"ai-blog/backend/internal/support"
)

// ExtractAgentSource 会把“粘贴文本”和“上传文件”统一整理成一份可继续加工的素材文本。
// 这是 Agent 工作台的第一步：先抽内容，再决定要不要生成文章草稿。
func (service *AIService) ExtractAgentSource(inputText string, fileName string, fileData []byte) (map[string]any, error) {
	parts := make([]string, 0, 2)
	sources := make([]string, 0, 2)

	trimmedInput := strings.TrimSpace(inputText)
	if trimmedInput != "" {
		parts = append(parts, trimmedInput)
		sources = append(sources, "手动粘贴文本")
	}

	if len(fileData) > 0 {
		extractedText, err := extractTextFromUpload(fileName, fileData)
		if err != nil {
			return nil, err
		}

		if strings.TrimSpace(extractedText) == "" {
			return nil, errors.New("上传文件里没有提取到可用内容，请检查文件是否为空")
		}

		parts = append(parts, extractedText)
		sources = append(sources, fileName)
	}

	if len(parts) == 0 {
		return nil, errors.New("请先上传文件，或者粘贴一段文字")
	}

	combinedText := normalizeExtractedText(strings.Join(parts, "\n\n"))
	lineCount := len(strings.Split(combinedText, "\n"))

	return map[string]any{
		"content":    combinedText,
		"preview":    support.TrimText(combinedText, 600),
		"char_count": utf8.RuneCountInString(combinedText),
		"line_count": lineCount,
		"sources":    sources,
		"message":    "素材抽取完成，现在可以继续生成文章草稿。",
	}, nil
}

// GenerateArticleDraft 会把抽取好的素材整理成适合博客后台确认的文章草稿。
// 返回值尽量结构化，这样前端可以直接填充到文章编辑表单里。
func (service *AIService) GenerateArticleDraft(ctx context.Context, sourceText string, goal string, tone string, categoryHint string) (map[string]any, error) {
	trimmedSource := strings.TrimSpace(sourceText)
	if trimmedSource == "" {
		return nil, errors.New("请先提供要整理的素材内容")
	}

	trimmedGoal := strings.TrimSpace(goal)
	if trimmedGoal == "" {
		trimmedGoal = "整理成一篇适合技术博客后台确认的文章草稿"
	}

	trimmedTone := strings.TrimSpace(tone)
	if trimmedTone == "" {
		trimmedTone = "清晰、自然、适合发布前审核"
	}

	prompt := fmt.Sprintf(
		"请根据下面的素材，%s。\n"+
			"要求：\n"+
			"1. 全文使用中文。\n"+
			"2. 内容要适合博客后台先审稿再发布。\n"+
			"3. 如果素材像课件或提纲，请主动补成可读性更好的文章结构。\n"+
			"4. 语气要求：%s。\n"+
			"5. 如果看得出分类，请给一个最合适的分类名；如果给不准，就输出“教程”。\n"+
			"6. 标签输出 3 到 5 个，使用中文逗号分隔。\n"+
			"7. 正文部分请输出 Markdown。\n\n"+
			"请严格按下面格式返回：\n"+
			"标题：...\n"+
			"摘要：...\n"+
			"分类：...\n"+
			"标签：标签1，标签2，标签3\n"+
			"正文：\n"+
			"# 标题\n"+
			"## 小节\n"+
			"...\n\n"+
			"可选分类提示：%s\n\n"+
			"素材内容：\n%s",
		trimmedGoal,
		trimmedTone,
		strings.TrimSpace(categoryHint),
		support.TrimText(trimmedSource, 12000),
	)

	text, providerName, err := service.chatWithFallback(ctx, "chat", []provider.ChatMessage{
		{
			Role:    "system",
			Content: "你是一位擅长把课件、会议纪要、培训资料和长文本整理成博客草稿的中文内容助手，请直接输出结构化结果。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	})
	if err != nil {
		result := buildLocalArticleDraft(trimmedSource, trimmedGoal, trimmedTone, categoryHint)
		result["hint"] = "上游模型暂时不可用，已切换到本地兜底：" + err.Error()
		return result, nil
	}

	return parseArticleDraft(text, trimmedSource, providerName, trimmedGoal, trimmedTone, categoryHint), nil
}

// AgentChat 提供后台 Agent 的日常聊天能力。
// 如果前端把素材内容一起带过来，Agent 就能围绕该素材继续回答问题。
func (service *AIService) AgentChat(ctx context.Context, message string, contextText string) (map[string]any, error) {
	trimmedMessage := strings.TrimSpace(message)
	if trimmedMessage == "" {
		return nil, errors.New("请输入你想让 Agent 帮你处理的问题")
	}

	contextPreview := support.TrimText(strings.TrimSpace(contextText), 6000)
	userPrompt := "用户问题：" + trimmedMessage
	if contextPreview != "" {
		userPrompt = "以下是当前对话可参考的素材：\n" + contextPreview + "\n\n" + userPrompt
	}

	text, providerName, err := service.chatWithFallback(ctx, "chat", []provider.ChatMessage{
		{
			Role:    "system",
			Content: "你是博客后台里的日常 AI 助手。请优先给出直接可执行、适合内容创作和整理工作的中文建议。",
		},
		{
			Role:    "user",
			Content: userPrompt,
		},
	})
	if err != nil {
		return map[string]any{
			"provider":     "local-fallback",
			"answer":       buildLocalAgentChatAnswer(trimmedMessage, contextPreview),
			"context_used": contextPreview != "",
			"hint":         "上游模型暂时不可用，已切换到本地兜底：" + err.Error(),
		}, nil
	}

	return map[string]any{
		"provider":     providerName,
		"answer":       strings.TrimSpace(text),
		"context_used": contextPreview != "",
	}, nil
}

func extractTextFromUpload(fileName string, fileData []byte) (string, error) {
	extension := strings.ToLower(filepath.Ext(fileName))

	switch extension {
	case ".txt", ".md", ".markdown":
		return normalizeExtractedText(string(fileData)), nil
	case ".docx":
		return extractTextFromDocx(fileData)
	case ".pptx":
		return extractTextFromPptx(fileData)
	case ".doc", ".ppt":
		return "", errors.New("当前先支持 docx 和 pptx，老格式 doc/ppt 还没有接入")
	default:
		return "", fmt.Errorf("暂不支持 %s 文件，请上传 txt、md、docx 或 pptx", extension)
	}
}

func extractTextFromDocx(fileData []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
	if err != nil {
		return "", errors.New("docx 文件解析失败，请确认文件没有损坏")
	}

	for _, file := range reader.File {
		if file.Name != "word/document.xml" {
			continue
		}

		content, err := readZipFile(file)
		if err != nil {
			return "", err
		}

		return extractTextFromXML(content), nil
	}

	return "", errors.New("docx 文件里没有找到正文内容")
}

func extractTextFromPptx(fileData []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
	if err != nil {
		return "", errors.New("pptx 文件解析失败，请确认文件没有损坏")
	}

	slideNames := make([]string, 0)
	slideMap := make(map[string]*zip.File)

	for _, file := range reader.File {
		if !strings.HasPrefix(file.Name, "ppt/slides/slide") || !strings.HasSuffix(file.Name, ".xml") {
			continue
		}

		slideNames = append(slideNames, file.Name)
		slideMap[file.Name] = file
	}

	sort.Strings(slideNames)
	if len(slideNames) == 0 {
		return "", errors.New("pptx 文件里没有找到幻灯片内容")
	}

	slideTexts := make([]string, 0, len(slideNames))
	for _, slideName := range slideNames {
		content, err := readZipFile(slideMap[slideName])
		if err != nil {
			return "", err
		}

		extracted := extractTextFromXML(content)
		if strings.TrimSpace(extracted) == "" {
			continue
		}

		slideTexts = append(slideTexts, "## "+filepath.Base(slideName)+"\n"+extracted)
	}

	if len(slideTexts) == 0 {
		return "", errors.New("pptx 文件里没有提取到可读文字")
	}

	return normalizeExtractedText(strings.Join(slideTexts, "\n\n")), nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func extractTextFromXML(content []byte) string {
	decoder := xml.NewDecoder(bytes.NewReader(content))
	var builder strings.Builder

	appendNewLine := func() {
		current := builder.String()
		if strings.HasSuffix(current, "\n\n") {
			return
		}

		if strings.HasSuffix(current, "\n") {
			builder.WriteString("\n")
			return
		}

		builder.WriteString("\n")
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch typedToken := token.(type) {
		case xml.StartElement:
			switch typedToken.Name.Local {
			case "br":
				appendNewLine()
			case "tab":
				builder.WriteString(" ")
			}
		case xml.EndElement:
			switch typedToken.Name.Local {
			case "p", "tr", "txBody", "body", "slide":
				appendNewLine()
			}
		case xml.CharData:
			text := strings.TrimSpace(string(typedToken))
			if text == "" {
				continue
			}

			current := builder.String()
			if current != "" && !strings.HasSuffix(current, "\n") && !strings.HasSuffix(current, " ") {
				builder.WriteString(" ")
			}

			builder.WriteString(text)
		}
	}

	return normalizeExtractedText(builder.String())
}

func normalizeExtractedText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")
	cleanLines := make([]string, 0, len(lines))
	lastWasBlank := true

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			if lastWasBlank {
				continue
			}

			cleanLines = append(cleanLines, "")
			lastWasBlank = true
			continue
		}

		cleanLines = append(cleanLines, trimmedLine)
		lastWasBlank = false
	}

	return strings.TrimSpace(strings.Join(cleanLines, "\n"))
}

func parseArticleDraft(text string, sourceText string, providerName string, goal string, tone string, categoryHint string) map[string]any {
	result := buildLocalArticleDraft(sourceText, goal, tone, categoryHint)
	cleanText := strings.TrimSpace(text)

	title := findLineValue(cleanText, []string{"标题：", "标题:", "Title:", "title:"})
	summary := findLineValue(cleanText, []string{"摘要：", "摘要:", "Summary:", "summary:"})
	categoryName := findLineValue(cleanText, []string{"分类：", "分类:", "Category:", "category:"})
	tagLine := findLineValue(cleanText, []string{"标签：", "标签:", "Tags:", "tags:"})
	content := findSectionValue(cleanText, []string{"正文：", "正文:", "内容：", "内容:", "Content:", "content:"})

	if title != "" {
		result["title"] = title
	}

	if summary != "" {
		result["summary"] = summary
	}

	if categoryName != "" {
		result["category_name"] = categoryName
	}

	if tagLine != "" {
		result["tag_names"] = parseCommaSeparatedTags(tagLine)
	}

	if strings.TrimSpace(content) != "" {
		result["content"] = stripMarkdownFence(content)
	}

	result["provider"] = providerName
	result["raw"] = cleanText
	return result
}

func findLineValue(text string, prefixes []string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmedLine := strings.TrimSpace(line)
		for _, prefix := range prefixes {
			if strings.HasPrefix(trimmedLine, prefix) {
				return strings.TrimSpace(strings.TrimPrefix(trimmedLine, prefix))
			}
		}
	}

	return ""
}

func findSectionValue(text string, prefixes []string) string {
	lines := strings.Split(text, "\n")

	for index, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		for _, prefix := range prefixes {
			if !strings.HasPrefix(trimmedLine, prefix) {
				continue
			}

			inlineValue := strings.TrimSpace(strings.TrimPrefix(trimmedLine, prefix))
			if inlineValue != "" {
				return inlineValue
			}

			if index+1 >= len(lines) {
				return ""
			}

			return strings.TrimSpace(strings.Join(lines[index+1:], "\n"))
		}
	}

	return ""
}

func stripMarkdownFence(text string) string {
	trimmedText := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmedText, "```") {
		return trimmedText
	}

	trimmedText = strings.TrimPrefix(trimmedText, "```markdown")
	trimmedText = strings.TrimPrefix(trimmedText, "```md")
	trimmedText = strings.TrimPrefix(trimmedText, "```")
	trimmedText = strings.TrimSuffix(trimmedText, "```")
	return strings.TrimSpace(trimmedText)
}

func buildLocalArticleDraft(sourceText string, goal string, tone string, categoryHint string) map[string]any {
	title := guessDraftTitle(sourceText)
	summary := support.TrimText(strings.ReplaceAll(sourceText, "\n", " "), 120)
	categoryName := guessCategoryName(sourceText, categoryHint)
	tagNames := guessTagNames(sourceText)
	bodySource := support.TrimText(sourceText, 2000)

	content := strings.TrimSpace(fmt.Sprintf(
		"# %s\n\n"+
			"> 这是一份由 Agent 先整理出来的草稿，适合你在后台确认后再发布。\n\n"+
			"## 一、素材核心信息\n\n"+
			"%s\n\n"+
			"## 二、可以继续补充的重点\n\n"+
			"- 请结合你的真实业务场景补充案例。\n"+
			"- 请检查是否需要加入图片、代码块或流程图。\n"+
			"- 请确认专业术语是否符合你的目标读者。\n\n"+
			"## 三、原始素材整理稿\n\n"+
			"%s",
		title,
		summary,
		bodySource,
	))

	return map[string]any{
		"provider":       "local-fallback",
		"title":          title,
		"summary":        summary,
		"category_name":  categoryName,
		"tag_names":      tagNames,
		"content":        content,
		"goal":           goal,
		"tone":           tone,
		"source_preview": support.TrimText(sourceText, 400),
	}
}

func guessDraftTitle(sourceText string) string {
	for _, line := range strings.Split(sourceText, "\n") {
		trimmedLine := strings.TrimSpace(line)
		trimmedLine = strings.TrimLeft(trimmedLine, "#-*0123456789. ")
		if trimmedLine == "" {
			continue
		}

		if utf8.RuneCountInString(trimmedLine) > 32 {
			return support.TrimText(trimmedLine, 32)
		}

		return trimmedLine
	}

	return "待确认的文章草稿"
}

func guessCategoryName(sourceText string, categoryHint string) string {
	if strings.TrimSpace(categoryHint) != "" {
		return strings.TrimSpace(categoryHint)
	}

	lowerText := strings.ToLower(sourceText)
	switch {
	case strings.Contains(sourceText, "安装"), strings.Contains(sourceText, "部署"), strings.Contains(sourceText, "教程"), strings.Contains(sourceText, "指南"):
		return "教程"
	case strings.Contains(lowerText, "go"), strings.Contains(lowerText, "gin"), strings.Contains(lowerText, "gorm"):
		return "Go"
	case strings.Contains(lowerText, "php"), strings.Contains(lowerText, "laravel"):
		return "PHP"
	case strings.Contains(lowerText, "vue"), strings.Contains(lowerText, "vite"), strings.Contains(lowerText, "javascript"):
		return "前端"
	case strings.Contains(lowerText, "mysql"), strings.Contains(lowerText, "redis"), strings.Contains(lowerText, "sql"):
		return "数据库"
	case strings.Contains(lowerText, "ai"), strings.Contains(lowerText, "大模型"), strings.Contains(lowerText, "agent"):
		return "AI"
	default:
		return "教程"
	}
}

func guessTagNames(sourceText string) []string {
	lowerText := strings.ToLower(sourceText)
	tagNames := make([]string, 0, 5)

	appendIfContains := func(keyword string, tagName string) {
		if strings.Contains(lowerText, keyword) {
			tagNames = append(tagNames, tagName)
		}
	}

	appendIfContains("go", "Go")
	appendIfContains("gin", "Gin")
	appendIfContains("gorm", "GORM")
	appendIfContains("php", "PHP")
	appendIfContains("vue", "Vue3")
	appendIfContains("mysql", "MySQL")
	appendIfContains("redis", "Redis")
	appendIfContains("ai", "AI")
	appendIfContains("agent", "Agent")
	appendIfContains("oss", "OSS")
	appendIfContains("markdown", "Markdown")
	appendIfContains("部署", "部署")
	appendIfContains("安装", "安装")
	appendIfContains("教程", "教程")
	appendIfContains("windows", "Windows")
	appendIfContains("linux", "Linux")

	if len(tagNames) == 0 {
		tagNames = []string{"博客草稿", "内容整理", "AI"}
	}

	if len(tagNames) > 5 {
		tagNames = tagNames[:5]
	}

	return uniqueStrings(tagNames)
}

func buildLocalAgentChatAnswer(message string, contextPreview string) string {
	answer := "我先给你一个可直接执行的建议：\n" +
		"1. 先确认这次是要整理文章、总结材料，还是日常问答。\n" +
		"2. 如果你已经有素材，建议先点“抽取内容”，确认文本是否完整。\n" +
		"3. 再点“生成草稿”，最后去文章管理里微调后发布。"

	if contextPreview != "" {
		answer += "\n\n我已经参考了你当前带过来的上下文素材，下一步你可以继续问我：帮你提炼标题、摘要、分类或标签。"
	}

	answer += "\n\n当前问题是：" + message
	return answer
}

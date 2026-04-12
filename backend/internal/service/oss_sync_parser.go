package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"ai-blog/backend/internal/config"

	"gopkg.in/yaml.v3"
)

// markdownFrontMatter 表示 Markdown 顶部可能存在的 YAML 元数据。
type markdownFrontMatter struct {
	Title   string   `yaml:"title"`
	Summary string   `yaml:"summary"`
	Tags    []string `yaml:"tags"`
	Cover   string   `yaml:"cover"`
}

// importedArticleData 是 Markdown 解析后的中间结果。
type importedArticleData struct {
	Title       string
	Content     string
	Summary     string
	CoverURL    string
	TagNames    []string
	SourceKey   string
	SourcePath  string
	SourceHash  string
	Category    string
	ArticleName string
}

var obsidianImagePattern = regexp.MustCompile(`!\[\[([^\]]+)\]\]`)
var markdownImagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

// parseImportedMarkdown 把 OSS 里的 Markdown 文件解析成系统可直接入库的数据。
func parseImportedMarkdown(ossConfig config.OSSConfig, objectKey string, rawMarkdown string) importedArticleData {
	normalizedMarkdown := strings.ReplaceAll(rawMarkdown, "\r\n", "\n")

	categoryName, articleName := extractCategoryAndArticleName(ossConfig.Prefix, objectKey)
	sourceKey := categoryName + "/" + articleName

	frontMatter, body := extractFrontMatter(normalizedMarkdown)
	title := chooseTitle(frontMatter.Title, body, articleName)

	// 这里把 Markdown 里的相对图片地址，统一改成可直接访问的 OSS URL。
	convertedContent := replaceRelativeImagesWithOSSURL(ossConfig, objectKey, body)
	summary := strings.TrimSpace(frontMatter.Summary)
	if summary == "" {
		summary = buildSummaryFromMarkdown(convertedContent)
	}

	coverURL := normalizeAssetPath(ossConfig, objectKey, frontMatter.Cover)
	if coverURL == "" {
		coverURL = findFirstImageURL(ossConfig, objectKey, convertedContent)
	}

	contentHash := sha256.Sum256([]byte(title + "\n" + convertedContent + "\n" + summary + "\n" + coverURL))

	return importedArticleData{
		Title:       title,
		Content:     convertedContent,
		Summary:     summary,
		CoverURL:    coverURL,
		TagNames:    frontMatter.Tags,
		SourceKey:   sourceKey,
		SourcePath:  objectKey,
		SourceHash:  hex.EncodeToString(contentHash[:]),
		Category:    categoryName,
		ArticleName: articleName,
	}
}

// extractCategoryAndArticleName 从 OSS 对象路径中解析“分类”和“文章名”。
// 例如：blog/Go/hello-world.md => Go / hello-world
func extractCategoryAndArticleName(prefix string, objectKey string) (string, string) {
	normalizedPrefix := strings.Trim(strings.TrimSpace(prefix), "/")
	normalizedKey := strings.Trim(strings.TrimSpace(objectKey), "/")

	if normalizedPrefix != "" && strings.HasPrefix(normalizedKey, normalizedPrefix+"/") {
		normalizedKey = strings.TrimPrefix(normalizedKey, normalizedPrefix+"/")
	}

	parts := strings.Split(normalizedKey, "/")
	if len(parts) < 2 {
		fileName := strings.TrimSuffix(path.Base(normalizedKey), path.Ext(normalizedKey))
		return "未分类", fileName
	}

	categoryName := strings.TrimSpace(parts[0])
	if categoryName == "" {
		categoryName = "未分类"
	}

	fileName := strings.TrimSuffix(path.Base(normalizedKey), path.Ext(normalizedKey))
	return categoryName, fileName
}

// extractFrontMatter 负责解析 Markdown 顶部的 YAML frontmatter。
// 如果文件没有 frontmatter，就直接返回空结构和原始正文。
func extractFrontMatter(markdown string) (markdownFrontMatter, string) {
	if !strings.HasPrefix(markdown, "---\n") {
		return markdownFrontMatter{}, markdown
	}

	parts := strings.SplitN(markdown, "\n---\n", 2)
	if len(parts) != 2 {
		return markdownFrontMatter{}, markdown
	}

	var matter markdownFrontMatter
	if err := yaml.Unmarshal([]byte(parts[0][4:]), &matter); err != nil {
		return markdownFrontMatter{}, markdown
	}

	return matter, parts[1]
}

// chooseTitle 决定最终文章标题。
// 优先级是：frontmatter 标题 > 第一行一级标题 > 文件名。
func chooseTitle(frontMatterTitle string, markdownBody string, fallback string) string {
	if strings.TrimSpace(frontMatterTitle) != "" {
		return strings.TrimSpace(frontMatterTitle)
	}

	lines := strings.Split(markdownBody, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		}
	}

	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(fallback)
	}

	return "未命名文章"
}

// replaceRelativeImagesWithOSSURL 把 Markdown 里的相对图片地址改成绝对 OSS URL。
func replaceRelativeImagesWithOSSURL(ossConfig config.OSSConfig, markdown string, content string) string {
	result := obsidianImagePattern.ReplaceAllStringFunc(content, func(match string) string {
		subMatches := obsidianImagePattern.FindStringSubmatch(match)
		if len(subMatches) < 2 {
			return match
		}

		imageURL := normalizeAssetPath(ossConfig, markdown, subMatches[1])
		if imageURL == "" {
			return match
		}

		return "![](" + imageURL + ")"
	})

	result = markdownImagePattern.ReplaceAllStringFunc(result, func(match string) string {
		subMatches := markdownImagePattern.FindStringSubmatch(match)
		if len(subMatches) < 3 {
			return match
		}

		imageAlt := subMatches[1]
		imagePath := subMatches[2]
		imageURL := normalizeAssetPath(ossConfig, markdown, imagePath)
		if imageURL == "" {
			return match
		}

		return "![" + imageAlt + "](" + imageURL + ")"
	})

	return result
}

// normalizeAssetPath 把 Markdown 里出现的图片路径转成可访问的 OSS URL。
func normalizeAssetPath(ossConfig config.OSSConfig, markdownObjectKey string, assetPath string) string {
	trimmedPath := strings.TrimSpace(assetPath)
	if trimmedPath == "" {
		return ""
	}

	if strings.HasPrefix(trimmedPath, "http://") || strings.HasPrefix(trimmedPath, "https://") {
		return trimmedPath
	}

	trimmedPath = strings.TrimPrefix(trimmedPath, "./")

	// Obsidian 里常见的 ![[图片名.jpg]] 会只给出文件名。
	// 这时我们默认图片和 Markdown 文件在同一个分类目录下。
	baseDir := path.Dir(strings.Trim(strings.TrimSpace(markdownObjectKey), "/"))
	objectKey := path.Clean(path.Join(baseDir, trimmedPath))

	return buildOSSObjectURL(ossConfig, objectKey)
}

// buildOSSObjectURL 把 OSS 对象路径拼成公开访问地址。
func buildOSSObjectURL(ossConfig config.OSSConfig, objectKey string) string {
	segments := strings.Split(strings.Trim(objectKey, "/"), "/")
	escapedSegments := make([]string, 0, len(segments))

	for _, segment := range segments {
		if strings.TrimSpace(segment) == "" {
			continue
		}
		escapedSegments = append(escapedSegments, url.PathEscape(segment))
	}

	if len(escapedSegments) == 0 || strings.TrimSpace(ossConfig.Bucket) == "" || strings.TrimSpace(ossConfig.Endpoint) == "" {
		return ""
	}

	return "https://" + ossConfig.Bucket + "." + ossConfig.Endpoint + "/" + strings.Join(escapedSegments, "/")
}

// buildSummaryFromMarkdown 用最直白的方式生成摘要。
// 它会跳过标题、图片和代码块，只截取第一段有意义的正文。
func buildSummaryFromMarkdown(markdown string) string {
	lines := strings.Split(markdown, "\n")
	candidates := make([]string, 0)
	insideCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "```") {
			insideCodeBlock = !insideCodeBlock
			continue
		}

		if insideCodeBlock {
			continue
		}

		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "![](") || strings.HasPrefix(trimmed, "![") {
			continue
		}

		cleaned := strings.NewReplacer("*", "", "`", "", ">", "", "-", " ").Replace(trimmed)
		cleaned = strings.TrimSpace(cleaned)
		if cleaned == "" {
			continue
		}

		candidates = append(candidates, cleaned)
		if len(strings.Join(candidates, " ")) >= 160 {
			break
		}
	}

	summary := strings.TrimSpace(strings.Join(candidates, " "))
	if summary == "" {
		return "这篇文章来自 OSS 自动同步。"
	}

	if utf8.RuneCountInString(summary) > 160 {
		runes := []rune(summary)
		return string(runes[:160]) + "..."
	}

	return summary
}

// findFirstImageURL 从 Markdown 正文里找第一张图，作为默认封面图。
func findFirstImageURL(ossConfig config.OSSConfig, markdownObjectKey string, content string) string {
	obsidianMatches := obsidianImagePattern.FindStringSubmatch(content)
	if len(obsidianMatches) >= 2 {
		return normalizeAssetPath(ossConfig, markdownObjectKey, obsidianMatches[1])
	}

	markdownMatches := markdownImagePattern.FindStringSubmatch(content)
	if len(markdownMatches) >= 3 {
		return normalizeAssetPath(ossConfig, markdownObjectKey, markdownMatches[2])
	}

	return ""
}

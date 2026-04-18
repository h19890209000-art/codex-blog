package support

import (
	"strings"

	"golang.org/x/net/html"
)

var blockedHTMLTags = map[string]bool{
	"script":   true,
	"style":    true,
	"noscript": true,
	"svg":      true,
	"path":     true,
	"form":     true,
	"button":   true,
	"input":    true,
	"nav":      true,
	"footer":   true,
	"header":   true,
	"aside":    true,
}

var blockTextTags = map[string]bool{
	"p":          true,
	"li":         true,
	"blockquote": true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
}

func ExtractReadableTextFromHTML(rawHTML string) string {
	document, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return ""
	}

	root := bestContentNode(document)
	if root == nil {
		root = document
	}

	blocks := make([]string, 0, 24)
	collectTextBlocks(root, &blocks)
	if len(blocks) < 4 && root != document {
		blocks = blocks[:0]
		collectTextBlocks(document, &blocks)
	}

	if len(blocks) == 0 {
		return ""
	}

	return strings.Join(uniqueTextBlocks(blocks), "\n\n")
}

func bestContentNode(root *html.Node) *html.Node {
	var best *html.Node
	bestScore := 0

	var visit func(node *html.Node)
	visit = func(node *html.Node) {
		if node == nil {
			return
		}

		if node.Type == html.ElementNode {
			score := contentNodeScore(node)
			if score > bestScore {
				bestScore = score
				best = node
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}

	visit(root)
	return best
}

func contentNodeScore(node *html.Node) int {
	if node == nil || node.Type != html.ElementNode {
		return 0
	}

	tag := strings.ToLower(node.Data)
	if blockedHTMLTags[tag] {
		return 0
	}

	score := 0
	if tag == "article" || tag == "main" {
		score += 60
	}
	if tag == "section" || tag == "div" {
		score += 10
	}

	attributes := strings.ToLower(nodeAttribute(node, "class") + " " + nodeAttribute(node, "id"))
	for _, keyword := range []string{"article", "content", "post", "entry", "story", "main", "body"} {
		if strings.Contains(attributes, keyword) {
			score += 18
		}
	}
	for _, keyword := range []string{"nav", "footer", "header", "sidebar", "comment", "related", "share", "social", "menu", "banner", "ad"} {
		if strings.Contains(attributes, keyword) {
			score -= 20
		}
	}

	paragraphCount := 0
	textLength := 0
	var visit func(current *html.Node)
	visit = func(current *html.Node) {
		if current == nil {
			return
		}

		if current.Type == html.ElementNode && blockTextTags[strings.ToLower(current.Data)] {
			paragraphCount++
		}
		if current.Type == html.TextNode {
			textLength += len(strings.Fields(current.Data))
		}

		for child := current.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(node)

	score += paragraphCount * 8
	score += textLength / 12
	return score
}

func collectTextBlocks(node *html.Node, blocks *[]string) {
	if node == nil {
		return
	}

	if node.Type == html.ElementNode {
		tag := strings.ToLower(node.Data)
		if blockedHTMLTags[tag] {
			return
		}

		if blockTextTags[tag] {
			text := cleanHTMLText(nodeText(node))
			if shouldKeepBlock(tag, text) {
				*blocks = append(*blocks, text)
			}
			return
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		collectTextBlocks(child, blocks)
	}
}

func nodeText(node *html.Node) string {
	if node == nil {
		return ""
	}

	var builder strings.Builder
	var visit func(current *html.Node)
	visit = func(current *html.Node) {
		if current == nil {
			return
		}

		if current.Type == html.ElementNode && blockedHTMLTags[strings.ToLower(current.Data)] {
			return
		}
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
			builder.WriteString(" ")
		}

		for child := current.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}

	visit(node)
	return builder.String()
}

func cleanHTMLText(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func shouldKeepBlock(tag string, text string) bool {
	if text == "" {
		return false
	}
	if len([]rune(text)) < 18 && tag != "h1" && tag != "h2" && tag != "h3" {
		return false
	}
	return true
}

func uniqueTextBlocks(blocks []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(blocks))

	for _, block := range blocks {
		normalized := strings.ToLower(strings.TrimSpace(block))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		result = append(result, block)
	}

	return result
}

func nodeAttribute(node *html.Node, name string) string {
	for _, attribute := range node.Attr {
		if strings.EqualFold(attribute.Key, name) {
			return attribute.Val
		}
	}
	return ""
}

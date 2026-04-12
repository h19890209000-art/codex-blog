package support

import (
	"strings"
	"unicode/utf8"
)

// TrimText 会把文本裁剪到指定长度，避免返回太长。
func TrimText(text string, maxLength int) string {
	if utf8.RuneCountInString(text) <= maxLength {
		return text
	}

	runes := []rune(text)
	return string(runes[:maxLength]) + "..."
}

// SplitLines 会把多行文本拆成清爽的字符串切片。
func SplitLines(text string) []string {
	rawLines := strings.Split(text, "\n")
	lines := make([]string, 0, len(rawLines))

	for _, line := range rawLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		lines = append(lines, trimmed)
	}

	return lines
}

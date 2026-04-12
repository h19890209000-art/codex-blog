package service

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestExtractTextFromDocx(t *testing.T) {
	fileData := buildZipFile(t, map[string]string{
		"word/document.xml": `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Go 后台入门</w:t></w:r></w:p><w:p><w:r><w:t>这一段来自 docx。</w:t></w:r></w:p></w:body></w:document>`,
	})

	text, err := extractTextFromDocx(fileData)
	if err != nil {
		t.Fatalf("extractTextFromDocx returned error: %v", err)
	}

	if !strings.Contains(text, "Go 后台入门") {
		t.Fatalf("expected extracted text to contain title, got %q", text)
	}

	if !strings.Contains(text, "这一段来自 docx。") {
		t.Fatalf("expected extracted text to contain body, got %q", text)
	}
}

func TestExtractTextFromPptx(t *testing.T) {
	fileData := buildZipFile(t, map[string]string{
		"ppt/slides/slide1.xml": `<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:cSld><p:spTree><p:sp><p:txBody><a:p><a:r><a:t>第一页标题</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld></p:sld>`,
		"ppt/slides/slide2.xml": `<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:cSld><p:spTree><p:sp><p:txBody><a:p><a:r><a:t>第二页内容</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld></p:sld>`,
	})

	text, err := extractTextFromPptx(fileData)
	if err != nil {
		t.Fatalf("extractTextFromPptx returned error: %v", err)
	}

	if !strings.Contains(text, "slide1.xml") {
		t.Fatalf("expected extracted text to contain slide name, got %q", text)
	}

	if !strings.Contains(text, "第二页内容") {
		t.Fatalf("expected extracted text to contain slide text, got %q", text)
	}
}

func TestParseArticleDraft(t *testing.T) {
	result := parseArticleDraft(
		"标题：Agent 生成草稿\n摘要：这是一段摘要\n分类：AI\n标签：Agent，博客，草稿\n正文：\n# Agent 生成草稿\n\n正文内容",
		"源文本",
		"minimax",
		"整理文章",
		"自然",
		"",
	)

	if result["title"] != "Agent 生成草稿" {
		t.Fatalf("unexpected title: %#v", result["title"])
	}

	tagNames, ok := result["tag_names"].([]string)
	if !ok || len(tagNames) == 0 {
		t.Fatalf("expected tag_names to be []string, got %#v", result["tag_names"])
	}
}

func buildZipFile(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)

	for name, content := range files {
		fileWriter, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create returned error: %v", err)
		}

		if _, err := fileWriter.Write([]byte(content)); err != nil {
			t.Fatalf("Write returned error: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	return buffer.Bytes()
}

package stream

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SplitText 把一段长文本切成较小的流式片段。
func SplitText(text string) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	chunks := make([]string, 0, len(words))
	builder := strings.Builder{}

	for index, word := range words {
		if builder.Len() > 0 {
			builder.WriteString(" ")
		}

		builder.WriteString(word)

		if (index+1)%8 == 0 {
			chunks = append(chunks, builder.String())
			builder.Reset()
		}
	}

	if builder.Len() > 0 {
		chunks = append(chunks, builder.String())
	}

	return chunks
}

// WriteChunks 会把文本片段一段一段写成 SSE 响应。
func WriteChunks(ctx *gin.Context, chunks []string) {
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.WriteHeader(http.StatusOK)

	for _, chunk := range chunks {
		_, _ = fmt.Fprintf(ctx.Writer, "data: %s\n\n", strings.ReplaceAll(chunk, "\n", " "))
		ctx.Writer.Flush()
		time.Sleep(70 * time.Millisecond)
	}

	_, _ = fmt.Fprint(ctx.Writer, "event: done\ndata: [DONE]\n\n")
	ctx.Writer.Flush()
}

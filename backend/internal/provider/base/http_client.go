package base

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// JSONClient 是所有 HTTP Provider 共用的基础客户端。
type JSONClient struct {
	HTTPClient *http.Client
}

// NewJSONClient 创建一个带超时控制的 HTTP 客户端。
func NewJSONClient() *JSONClient {
	return &JSONClient{
		HTTPClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

// PostJSON 负责发送 JSON 请求，并把响应解码到目标结构体中。
func (client *JSONClient) PostJSON(ctx context.Context, url string, headers map[string]string, payload any, target any) error {
	rawBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(rawBytes))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("provider 返回错误: status=%d body=%s", response.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	if target == nil {
		return nil
	}

	return json.Unmarshal(bodyBytes, target)
}

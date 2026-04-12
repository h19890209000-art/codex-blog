package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-blog/backend/internal/bootstrap"
)

func TestHealthEndpoint(t *testing.T) {
	app, err := bootstrap.NewApp()
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	app.TestHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际是 %d", recorder.Code)
	}
}

package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-blog/backend/internal/router"

	"github.com/gin-gonic/gin"
)

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router.RegisterBaseRoutes(engine)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", recorder.Code)
	}
}

package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// registerBaseRoutes 注册最基础的公共路由。
// 这些路由通常不区分前台后台，比如健康检查。
func registerBaseRoutes(engine *gin.Engine) {
	engine.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "ok",
		})
	})
}

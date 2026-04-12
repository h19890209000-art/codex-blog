package response

import "github.com/gin-gonic/gin"

// Success 返回统一成功结构。
func Success(ctx *gin.Context, data any) {
	ctx.JSON(200, gin.H{
		"success": true,
		"data":    data,
	})
}

// Error 返回统一失败结构。
func Error(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
	})
}

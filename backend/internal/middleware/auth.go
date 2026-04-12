package middleware

import (
	"net/http"
	"strings"

	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/response"
	"ai-blog/backend/internal/support"

	"github.com/gin-gonic/gin"
)

// AdminAuth 用来校验后台登录令牌。
func AdminAuth(cfg config.AuthConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodOptions {
			ctx.Next()
			return
		}

		authHeader := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(ctx, http.StatusUnauthorized, "请先登录后台")
			ctx.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := support.ParseToken(cfg.TokenSecret, token)
		if err != nil {
			response.Error(ctx, http.StatusUnauthorized, "登录状态无效，请重新登录")
			ctx.Abort()
			return
		}

		ctx.Set("auth_user_id", claims.UserID)
		ctx.Set("auth_username", claims.Username)
		ctx.Set("auth_role", claims.Role)
		ctx.Next()
	}
}

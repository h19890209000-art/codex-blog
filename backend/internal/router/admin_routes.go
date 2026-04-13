package router

import (
	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/controller"
	"ai-blog/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerAdminRoutes 注册后台管理端路由。
// 我们把后台路由拆成两部分：
// 1. 不需要登录的公开路由，比如登录接口。
// 2. 登录后才能访问的受保护路由，比如文章管理和同步管理。
func registerAdminRoutes(engine *gin.Engine, appConfig config.AppConfig, adminController *controller.AdminController) {
	registerAdminPublicRoutes(engine, adminController)
	registerAdminProtectedRoutes(engine, appConfig, adminController)
}

// registerAdminPublicRoutes 注册后台公开路由。
func registerAdminPublicRoutes(engine *gin.Engine, adminController *controller.AdminController) {
	adminGroup := engine.Group("/api/admin")
	{
		adminGroup.POST("/auth/login", adminController.Login)
	}
}

// registerAdminProtectedRoutes 注册后台受保护路由。
func registerAdminProtectedRoutes(engine *gin.Engine, appConfig config.AppConfig, adminController *controller.AdminController) {
	adminProtectedGroup := engine.Group("/api/admin")
	adminProtectedGroup.Use(middleware.AdminAuth(appConfig.Auth))
	{
		adminProtectedGroup.GET("/me", adminController.Me)
		adminProtectedGroup.POST("/auth/change-password", adminController.ChangePassword)

		adminProtectedGroup.GET("/dashboard", adminController.Dashboard)
		adminProtectedGroup.GET("/system-configs", adminController.ListSystemConfigs)
		adminProtectedGroup.PUT("/system-configs", adminController.SaveSystemConfigs)

		adminProtectedGroup.GET("/articles", adminController.ListArticles)
		adminProtectedGroup.POST("/articles", adminController.CreateArticle)
		adminProtectedGroup.PUT("/articles/:id", adminController.UpdateArticle)
		adminProtectedGroup.DELETE("/articles/:id", adminController.DeleteArticle)

		adminProtectedGroup.GET("/daily-briefings", adminController.ListDailyBriefings)
		adminProtectedGroup.POST("/daily-briefings", adminController.CreateDailyBriefing)
		adminProtectedGroup.PUT("/daily-briefings/:id", adminController.UpdateDailyBriefing)
		adminProtectedGroup.DELETE("/daily-briefings/:id", adminController.DeleteDailyBriefing)
		adminProtectedGroup.GET("/daily-briefings/fetch-status", adminController.DailyBriefingFetchStatus)
		adminProtectedGroup.POST("/daily-briefings/fetch", adminController.RunDailyBriefingFetch)

		adminProtectedGroup.GET("/categories", adminController.ListCategories)
		adminProtectedGroup.POST("/categories", adminController.CreateCategory)

		adminProtectedGroup.GET("/tags", adminController.ListTags)
		adminProtectedGroup.POST("/tags", adminController.CreateTag)

		adminProtectedGroup.GET("/comments", adminController.ListComments)
		adminProtectedGroup.PUT("/comments/:id/status", adminController.UpdateCommentStatus)
		adminProtectedGroup.DELETE("/comments/:id", adminController.DeleteComment)

		adminProtectedGroup.GET("/users", adminController.ListUsers)
		adminProtectedGroup.POST("/users", adminController.CreateUser)
		adminProtectedGroup.PUT("/users/:id/role", adminController.UpdateUserRole)
		adminProtectedGroup.DELETE("/users/:id", adminController.DeleteUser)

		adminProtectedGroup.GET("/ai/providers", adminController.ProviderOverview)
		adminProtectedGroup.POST("/ai/generate-summary", adminController.GenerateSummary)
		adminProtectedGroup.POST("/ai/suggest-tags", adminController.SuggestTags)
		adminProtectedGroup.POST("/ai/brainstorm", adminController.Brainstorm)
		adminProtectedGroup.POST("/ai/rewrite", adminController.Rewrite)
		adminProtectedGroup.POST("/ai/generate-cover", adminController.GenerateCover)
		adminProtectedGroup.POST("/agent/extract", adminController.ExtractAgentSource)
		adminProtectedGroup.POST("/agent/generate-draft", adminController.GenerateArticleDraft)
		adminProtectedGroup.POST("/agent/chat", adminController.AgentChat)

		adminProtectedGroup.GET("/sync/oss/status", adminController.OSSSyncStatus)
		adminProtectedGroup.POST("/sync/oss/run", adminController.RunOSSSync)

		adminProtectedGroup.POST("/comments/:id/reply-suggestions", adminController.ReplySuggestions)
		adminProtectedGroup.POST("/comments/moderate", adminController.ModerateComment)
	}
}

package router

import (
	"ai-blog/backend/internal/controller"

	"github.com/gin-gonic/gin"
)

// registerPublicRoutes 注册前台读者端路由。
// 这一组路由不需要登录，主要服务普通访客。
func registerPublicRoutes(engine *gin.Engine, publicController *controller.PublicController) {
	publicGroup := engine.Group("/api/public")
	{
		publicGroup.GET("/articles", publicController.ListArticles)
		publicGroup.GET("/articles/:id", publicController.GetArticle)
		publicGroup.GET("/articles/:id/navigation", publicController.GetArticleNavigation)
		publicGroup.GET("/categories", publicController.ListCategories)
		publicGroup.GET("/tags", publicController.ListTags)
		publicGroup.GET("/archives", publicController.ListArchives)
		publicGroup.GET("/daily-briefings", publicController.ListDailyBriefings)
		publicGroup.GET("/system-configs", publicController.ListSystemConfigs)
		publicGroup.GET("/articles/:id/comments", publicController.ListComments)
		publicGroup.POST("/articles/:id/comments", publicController.CreateComment)
		publicGroup.POST("/ai/analyze-title", publicController.AnalyzeTitle)
		publicGroup.POST("/qa/article/:id", publicController.ArticleQA)
		publicGroup.POST("/qa/site", publicController.SiteQA)
	}
}

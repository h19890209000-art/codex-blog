package router

import (
	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/controller"

	"github.com/gin-gonic/gin"
)

// Register 是整个路由模块的总入口。
// 你可以把它理解成“路由调度中心”，只负责调用各个分类路由的注册函数。
func Register(engine *gin.Engine, appConfig config.AppConfig, publicController *controller.PublicController, adminController *controller.AdminController) {
	RegisterBaseRoutes(engine)
	registerPublicRoutes(engine, publicController)
	registerAdminRoutes(engine, appConfig, adminController)
}

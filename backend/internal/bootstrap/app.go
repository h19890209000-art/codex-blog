package bootstrap

import (
	"fmt"
	"net/http"
	"time"

	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/controller"
	"ai-blog/backend/internal/middleware"
	"ai-blog/backend/internal/repository"
	"ai-blog/backend/internal/router"
	"ai-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// App 表示整个后端应用。
// 你可以把它理解成“启动完成后的总装配结果”。
type App struct {
	config     config.AppConfig
	httpServer *http.Server
	engine     *gin.Engine
}

// NewApp 创建应用。
// 这个函数就是后端的总装配流程，建议你以后读 Go 项目时优先看这种入口。
func NewApp() (*App, error) {
	// 第一步：读取配置文件。
	appConfig, err := config.Load()
	if err != nil {
		return nil, err
	}

	// 第二步：初始化数据库。
	// 这里会自动建库、建表、创建默认管理员和演示数据。
	db, err := InitDatabase(appConfig)
	if err != nil {
		return nil, err
	}

	// 第三步：创建 Repository 层。
	// Repository 只做一件事：和数据库打交道。
	articleRepo := repository.NewGormArticleRepository(db)
	commentRepo := repository.NewGormCommentRepository(db)
	userRepo := repository.NewGormUserRepository(db)
	categoryRepo := repository.NewGormCategoryRepository(db)
	tagRepo := repository.NewGormTagRepository(db)
	dailyBriefingRepo := repository.NewGormDailyBriefingRepository(db)
	systemConfigRepo := repository.NewGormSystemConfigRepository(db)

	// 第四步：创建 AI Provider 注册表。
	// 这里会把 GLM、MiniMax、小米、OpenAI、Ollama 这些厂商接起来。
	providerRegistry, err := service.NewProviderRegistry(appConfig)
	if err != nil {
		return nil, err
	}

	// 第五步：创建业务 Service。
	// Service 负责真正的业务处理逻辑。
	articleService := service.NewArticleService(articleRepo)
	authService := service.NewAuthService(appConfig.Auth, userRepo)
	adminService := service.NewAdminService(articleRepo, categoryRepo, tagRepo, commentRepo, userRepo)
	portalService := service.NewPortalService(articleRepo, categoryRepo, tagRepo, commentRepo)
	aiService := service.NewAIService(providerRegistry, articleRepo, commentRepo)
	dailyBriefingService := service.NewDailyBriefingService(dailyBriefingRepo)
	systemConfigService := service.NewSystemConfigService(systemConfigRepo)
	ossSyncService, err := service.NewOSSSyncService(appConfig, articleRepo, categoryRepo, tagRepo)
	if err != nil {
		return nil, err
	}

	// 第六步：创建 Controller。
	// Controller 负责接收 HTTP 请求，然后把请求交给 Service。
	publicController := controller.NewPublicController(articleService, portalService, aiService, dailyBriefingService, systemConfigService)
	adminController := controller.NewAdminController(authService, adminService, aiService, providerRegistry, ossSyncService, dailyBriefingService, systemConfigService)

	// 第七步：创建 Gin 引擎。
	engine := gin.Default()

	// 第八步：挂载全局中间件。
	// 这里先启用跨域，方便本地前后端分开调试。
	engine.Use(middleware.CORS())

	// 第九步：注册所有路由。
	router.Register(engine, appConfig, publicController, adminController)

	// 第十步：如果开启了 OSS 自动同步，就在后台启动定时任务。
	ossSyncService.StartAutoSync()

	// 第十一步：创建标准 HTTP Server。
	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port),
		Handler:           engine,
		ReadHeaderTimeout: time.Duration(appConfig.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout:      time.Duration(appConfig.Server.WriteTimeoutSeconds) * time.Second,
	}

	return &App{
		config:     appConfig,
		httpServer: httpServer,
		engine:     engine,
	}, nil
}

// Run 启动 HTTP 服务。
func (app *App) Run() error {
	return app.httpServer.ListenAndServe()
}

// TestHandler 返回测试用的 HTTP Handler。
// 单元测试时不需要真的监听端口，只要拿到 Gin 引擎就够了。
func (app *App) TestHandler() http.Handler {
	return app.engine
}

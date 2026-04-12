package controller

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"ai-blog/backend/internal/dto"
	"ai-blog/backend/internal/response"
	"ai-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminController 负责后台管理端接口。
type AdminController struct {
	authService          *service.AuthService
	adminService         *service.AdminService
	aiService            *service.AIService
	providers            *service.ProviderRegistry
	syncService          *service.OSSSyncService
	dailyBriefingService *service.DailyBriefingService
}

// NewAdminController 创建后台控制器。
func NewAdminController(
	authService *service.AuthService,
	adminService *service.AdminService,
	aiService *service.AIService,
	providers *service.ProviderRegistry,
	syncService *service.OSSSyncService,
	dailyBriefingService *service.DailyBriefingService,
) *AdminController {
	return &AdminController{
		authService:          authService,
		adminService:         adminService,
		aiService:            aiService,
		providers:            providers,
		syncService:          syncService,
		dailyBriefingService: dailyBriefingService,
	}
}

// Login 处理后台登录。
func (controller *AdminController) Login(ctx *gin.Context) {
	var request dto.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	result, err := controller.authService.Login(request.Username, request.Password)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(ctx, result)
}

// Me 返回当前登录用户信息。
func (controller *AdminController) Me(ctx *gin.Context) {
	userID, ok := controller.getAuthUserID(ctx)
	if !ok {
		return
	}

	result, err := controller.authService.Profile(userID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// ChangePassword 修改当前管理员密码。
func (controller *AdminController) ChangePassword(ctx *gin.Context) {
	userID, ok := controller.getAuthUserID(ctx)
	if !ok {
		return
	}

	var request dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	if err := controller.authService.ChangePassword(userID, request.OldPassword, request.NewPassword); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"message": "密码修改成功，请使用新密码重新登录。",
	})
}

// Dashboard 返回后台首页统计数据。
func (controller *AdminController) Dashboard(ctx *gin.Context) {
	result, err := controller.adminService.Dashboard()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// ListArticles 返回后台文章列表。
func (controller *AdminController) ListArticles(ctx *gin.Context) {
	var query dto.ArticleListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.Error(ctx, http.StatusBadRequest, "查询参数格式不正确")
		return
	}

	result, err := controller.adminService.ListArticles(query.Keyword, query.Status, query.Page, query.PageSize)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// CreateArticle 创建文章。
func (controller *AdminController) CreateArticle(ctx *gin.Context) {
	var request dto.SaveArticleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	article, err := controller.adminService.SaveArticle(
		0,
		request.Title,
		request.Content,
		request.Summary,
		request.CoverURL,
		request.Status,
		request.CategoryName,
		request.TagNames,
	)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, article)
}

// UpdateArticle 更新文章。
func (controller *AdminController) UpdateArticle(ctx *gin.Context) {
	articleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "文章 ID 不合法")
		return
	}

	var request dto.SaveArticleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	article, err := controller.adminService.SaveArticle(
		articleID,
		request.Title,
		request.Content,
		request.Summary,
		request.CoverURL,
		request.Status,
		request.CategoryName,
		request.TagNames,
	)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, article)
}

// DeleteArticle 删除文章。
func (controller *AdminController) DeleteArticle(ctx *gin.Context) {
	articleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "文章 ID 不合法")
		return
	}

	if err := controller.adminService.DeleteArticle(articleID); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"message": "文章删除成功",
	})
}

// ListCategories 返回分类列表。
func (controller *AdminController) ListCategories(ctx *gin.Context) {
	categories, err := controller.adminService.ListCategories()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, categories)
}

// CreateCategory 创建分类。
func (controller *AdminController) CreateCategory(ctx *gin.Context) {
	var request dto.SaveCategoryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	category, err := controller.adminService.CreateCategory(request.Name)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, category)
}

// ListTags 返回标签列表。
func (controller *AdminController) ListTags(ctx *gin.Context) {
	tags, err := controller.adminService.ListTags()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, tags)
}

// CreateTag 创建标签。
func (controller *AdminController) CreateTag(ctx *gin.Context) {
	var request dto.SaveTagRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	tag, err := controller.adminService.CreateTag(request.Name)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, tag)
}

// ListComments 返回评论列表。
func (controller *AdminController) ListComments(ctx *gin.Context) {
	var query dto.CommentListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.Error(ctx, http.StatusBadRequest, "查询参数格式不正确")
		return
	}

	result, err := controller.adminService.ListComments(query.Keyword, query.Status, query.Page, query.PageSize)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// UpdateCommentStatus 修改评论状态。
func (controller *AdminController) UpdateCommentStatus(ctx *gin.Context) {
	commentID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "评论 ID 不合法")
		return
	}

	var request dto.UpdateCommentStatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	if err := controller.adminService.UpdateCommentStatus(commentID, request.Status); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"message": "评论状态更新成功",
	})
}

// DeleteComment 删除评论。
func (controller *AdminController) DeleteComment(ctx *gin.Context) {
	commentID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "评论 ID 不合法")
		return
	}

	if err := controller.adminService.DeleteComment(commentID); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"message": "评论删除成功",
	})
}

// ListUsers 返回用户列表。
func (controller *AdminController) ListUsers(ctx *gin.Context) {
	var query dto.UserListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.Error(ctx, http.StatusBadRequest, "查询参数格式不正确")
		return
	}

	result, err := controller.adminService.ListUsers(query.Keyword, query.Role, query.Page, query.PageSize)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// CreateUser 创建用户。
func (controller *AdminController) CreateUser(ctx *gin.Context) {
	var request dto.SaveUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	result, err := controller.adminService.CreateUser(request.Username, request.Password, request.Avatar, request.Role)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, result)
}

// UpdateUserRole 修改用户角色。
func (controller *AdminController) UpdateUserRole(ctx *gin.Context) {
	userID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "用户 ID 不合法")
		return
	}

	var request dto.UpdateUserRoleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	if err := controller.adminService.UpdateUserRole(userID, request.Role); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"message": "用户角色更新成功",
	})
}

// DeleteUser 删除用户。
func (controller *AdminController) DeleteUser(ctx *gin.Context) {
	currentUserID, ok := controller.getAuthUserID(ctx)
	if !ok {
		return
	}

	targetUserID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "用户 ID 不合法")
		return
	}

	if err := controller.adminService.DeleteUser(currentUserID, targetUserID); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"message": "用户删除成功",
	})
}

// ProviderOverview 返回 AI Provider 概览。
func (controller *AdminController) ProviderOverview(ctx *gin.Context) {
	response.Success(ctx, controller.adminService.ProviderOverview(controller.providers))
}

// OSSSyncStatus 返回 OSS 同步状态。
func (controller *AdminController) OSSSyncStatus(ctx *gin.Context) {
	response.Success(ctx, controller.syncService.Status())
}

// RunOSSSync 手动触发一次 OSS 同步。
func (controller *AdminController) RunOSSSync(ctx *gin.Context) {
	result, err := controller.syncService.RunOnce(ctx.Request.Context(), "manual")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, result)
}

// ListDailyBriefings returns briefing items for the admin console.
func (controller *AdminController) ListDailyBriefings(ctx *gin.Context) {
	var query dto.DailyBriefingListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid query parameters")
		return
	}

	result, err := controller.dailyBriefingService.ListAdmin(query.Date, query.Keyword, query.Status, query.Page, query.PageSize)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// CreateDailyBriefing creates a manual briefing item.
func (controller *AdminController) CreateDailyBriefing(ctx *gin.Context) {
	var request dto.SaveDailyBriefingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request payload")
		return
	}

	item, err := controller.dailyBriefingService.SaveBriefing(
		0,
		request.BriefingDate,
		request.Title,
		request.Summary,
		request.SourceName,
		request.SourceURL,
		request.Status,
		request.SortOrder,
		request.SourcePublishedAt,
	)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, item)
}

// UpdateDailyBriefing updates an existing briefing item.
func (controller *AdminController) UpdateDailyBriefing(ctx *gin.Context) {
	itemID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid briefing id")
		return
	}

	var request dto.SaveDailyBriefingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request payload")
		return
	}

	item, err := controller.dailyBriefingService.SaveBriefing(
		itemID,
		request.BriefingDate,
		request.Title,
		request.Summary,
		request.SourceName,
		request.SourceURL,
		request.Status,
		request.SortOrder,
		request.SourcePublishedAt,
	)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, item)
}

// DeleteDailyBriefing deletes a briefing item.
func (controller *AdminController) DeleteDailyBriefing(ctx *gin.Context) {
	itemID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid briefing id")
		return
	}

	if err := controller.dailyBriefingService.DeleteBriefing(itemID); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"message": "daily briefing deleted",
	})
}

// DailyBriefingFetchStatus returns the latest fetch status.
func (controller *AdminController) DailyBriefingFetchStatus(ctx *gin.Context) {
	response.Success(ctx, controller.dailyBriefingService.Status())
}

// RunDailyBriefingFetch triggers a manual fetch for daily AI news.
func (controller *AdminController) RunDailyBriefingFetch(ctx *gin.Context) {
	var request dto.FetchDailyBriefingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		response.Error(ctx, http.StatusBadRequest, "invalid request payload")
		return
	}

	result, err := controller.dailyBriefingService.FetchNow(ctx.Request.Context(), request.Date, request.Limit, "admin")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, result)
}

// GenerateSummary 生成摘要。
func (controller *AdminController) GenerateSummary(ctx *gin.Context) {
	var request dto.GenerateSummaryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	result, err := controller.aiService.GenerateSummary(ctx.Request.Context(), request.Content)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// SuggestTags 生成标签建议。
func (controller *AdminController) SuggestTags(ctx *gin.Context) {
	var request dto.SuggestTagsRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	result, err := controller.aiService.SuggestTags(ctx.Request.Context(), request.Content)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// Brainstorm 生成标题和大纲灵感。
func (controller *AdminController) Brainstorm(ctx *gin.Context) {
	var request dto.BrainstormRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	result, err := controller.aiService.Brainstorm(ctx.Request.Context(), request.Keyword)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// Rewrite 润色文本。
func (controller *AdminController) Rewrite(ctx *gin.Context) {
	var request dto.RewriteRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	result, err := controller.aiService.Rewrite(ctx.Request.Context(), request.Content, request.Style)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// GenerateCover 生成封面图。
func (controller *AdminController) GenerateCover(ctx *gin.Context) {
	var request dto.GenerateCoverRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	result, err := controller.aiService.GenerateCover(ctx.Request.Context(), request.Title)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// ReplySuggestions 返回评论回复建议。
func (controller *AdminController) ReplySuggestions(ctx *gin.Context) {
	commentID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "评论 ID 不合法")
		return
	}

	result, err := controller.aiService.ReplySuggestions(ctx.Request.Context(), commentID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// ModerateComment 演示评论审核能力。
func (controller *AdminController) ModerateComment(ctx *gin.Context) {
	var request struct {
		Content string `json:"content"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	result, err := controller.aiService.ModerateComment(ctx.Request.Context(), request.Content)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// getAuthUserID 从中间件里取出当前登录用户 ID。
func (controller *AdminController) getAuthUserID(ctx *gin.Context) (int64, bool) {
	userIDValue, exists := ctx.Get("auth_user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "登录状态无效")
		return 0, false
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "登录状态无效")
		return 0, false
	}

	return userID, true
}

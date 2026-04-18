package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ai-blog/backend/internal/dto"
	"ai-blog/backend/internal/response"
	"ai-blog/backend/internal/service"
	"ai-blog/backend/internal/stream"

	"github.com/gin-gonic/gin"
)

type PublicController struct {
	articleService       *service.ArticleService
	portalService        *service.PortalService
	aiService            *service.AIService
	dailyBriefingService *service.DailyBriefingService
	systemConfigService  *service.SystemConfigService
}

func NewPublicController(
	articleService *service.ArticleService,
	portalService *service.PortalService,
	aiService *service.AIService,
	dailyBriefingService *service.DailyBriefingService,
	systemConfigService *service.SystemConfigService,
) *PublicController {
	return &PublicController{
		articleService:       articleService,
		portalService:        portalService,
		aiService:            aiService,
		dailyBriefingService: dailyBriefingService,
		systemConfigService:  systemConfigService,
	}
}

func (controller *PublicController) ListArticles(ctx *gin.Context) {
	articles, err := controller.articleService.ListPublic(ctx.Query("keyword"))
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, articles)
}

func (controller *PublicController) GetArticle(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid article id")
		return
	}

	article, err := controller.articleService.Detail(id)
	if err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	response.Success(ctx, article)
}

func (controller *PublicController) GetArticleNavigation(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid article id")
		return
	}

	previousArticle, nextArticle, err := controller.articleService.Navigation(id)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	var previousPayload any
	var nextPayload any
	if previousArticle.ID > 0 {
		previousPayload = previousArticle
	}
	if nextArticle.ID > 0 {
		nextPayload = nextArticle
	}

	response.Success(ctx, gin.H{
		"prev_article": previousPayload,
		"next_article": nextPayload,
	})
}

func (controller *PublicController) ListCategories(ctx *gin.Context) {
	categories, err := controller.portalService.ListCategories()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, categories)
}

func (controller *PublicController) ListTags(ctx *gin.Context) {
	tags, err := controller.portalService.ListTags()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, tags)
}

func (controller *PublicController) ListArchives(ctx *gin.Context) {
	archives, err := controller.portalService.ListArchives()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, archives)
}

func (controller *PublicController) ListDailyBriefings(ctx *gin.Context) {
	result, err := controller.dailyBriefingService.ListPublic(ctx.Query("date"))
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

func (controller *PublicController) ListSystemConfigs(ctx *gin.Context) {
	result, err := controller.systemConfigService.PublicMap()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

func (controller *PublicController) ListComments(ctx *gin.Context) {
	articleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid article id")
		return
	}

	comments, err := controller.portalService.ListComments(articleID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, comments)
}

func (controller *PublicController) CreateComment(ctx *gin.Context) {
	articleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid article id")
		return
	}

	var request dto.CreateCommentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request payload")
		return
	}

	comment, err := controller.portalService.CreateComment(articleID, request.Author, request.Content)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, comment)
}

func (controller *PublicController) AnalyzeTitle(ctx *gin.Context) {
	var request dto.AnalyzeTitleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request payload")
		return
	}

	result, err := controller.aiService.AnalyzeTitle(ctx.Request.Context(), request.Title)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if request.Stream {
		stream.WriteChunks(ctx, stream.SplitText(result["result"].(string)))
		return
	}

	response.Success(ctx, result)
}

func (controller *PublicController) ArticleQA(ctx *gin.Context) {
	articleID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid article id")
		return
	}

	var request dto.QARequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request payload")
		return
	}

	result, err := controller.aiService.ArticleQA(ctx.Request.Context(), articleID, request.Question)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if request.Stream {
		stream.WriteChunks(ctx, stream.SplitText(result["answer"].(string)))
		return
	}

	response.Success(ctx, result)
}

func (controller *PublicController) SiteQA(ctx *gin.Context) {
	var request dto.QARequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request payload")
		return
	}

	result, err := controller.aiService.SiteQA(ctx.Request.Context(), request.Question)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if request.Stream {
		stream.WriteChunks(ctx, stream.SplitText(result["answer"].(string)))
		return
	}

	response.Success(ctx, result)
}

func (controller *PublicController) GetDailyBriefingStudy(ctx *gin.Context) {
	briefingID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid briefing id")
		return
	}

	item, err := controller.dailyBriefingService.EnsureSourceContent(ctx.Request.Context(), briefingID)
	if err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	translationProvider := "cached"
	translationHint := ""
	if strings.TrimSpace(item.TranslatedContent) == "" {
		translated, err := controller.aiService.TranslateBriefingForStudy(ctx.Request.Context(), item.Title, item.SourceContent)
		if err != nil {
			response.Error(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		item, err = controller.dailyBriefingService.SaveTranslatedContent(briefingID, strings.TrimSpace(toStringValue(translated["translation"])))
		if err != nil {
			response.Error(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		translationProvider = toStringValue(translated["provider"])
		translationHint = toStringValue(translated["hint"])
	}

	response.Success(ctx, gin.H{
		"id":                   item.ID,
		"briefing_date":        item.BriefingDate,
		"title":                item.Title,
		"summary":              item.Summary,
		"source_name":          item.SourceName,
		"source_url":           item.SourceURL,
		"source_published_at":  item.SourcePublishedAt,
		"source_content":       item.SourceContent,
		"translated_content":   item.TranslatedContent,
		"translation_provider": translationProvider,
		"translation_hint":     translationHint,
	})
}

func (controller *PublicController) BuildDailyBriefingLearningPlan(ctx *gin.Context) {
	briefingID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid briefing id")
		return
	}

	var request dto.BriefingLearningPlanRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request payload")
		return
	}

	item, err := controller.dailyBriefingService.EnsureSourceContent(ctx.Request.Context(), briefingID)
	if err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	if strings.TrimSpace(item.TranslatedContent) == "" {
		translated, err := controller.aiService.TranslateBriefingForStudy(ctx.Request.Context(), item.Title, item.SourceContent)
		if err != nil {
			response.Error(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		item, err = controller.dailyBriefingService.SaveTranslatedContent(briefingID, strings.TrimSpace(toStringValue(translated["translation"])))
		if err != nil {
			response.Error(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	}

	result, err := controller.aiService.BuildBriefingLearningPlan(
		ctx.Request.Context(),
		item.Title,
		item.Summary,
		item.SourceContent,
		item.TranslatedContent,
		request.Goal,
	)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

func (controller *PublicController) RunDailyBriefingRoleplay(ctx *gin.Context) {
	briefingID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid briefing id")
		return
	}

	var request dto.BriefingRoleplayRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request payload")
		return
	}
	if strings.TrimSpace(request.LearnerReply) == "" {
		response.Error(ctx, http.StatusBadRequest, "learner_reply is required")
		return
	}

	item, err := controller.dailyBriefingService.GetPublishedByID(briefingID)
	if err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	result, err := controller.aiService.RunBriefingRoleplay(
		ctx.Request.Context(),
		item.Title,
		item.Summary,
		request.Goal,
		request.Scene,
		request.LearnerReply,
	)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

func (controller *PublicController) ExplainDailyBriefingWord(ctx *gin.Context) {
	briefingID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid briefing id")
		return
	}

	var request dto.StudyWordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request payload")
		return
	}
	if strings.TrimSpace(request.Word) == "" || strings.TrimSpace(request.Sentence) == "" {
		response.Error(ctx, http.StatusBadRequest, "word and sentence are required")
		return
	}

	item, err := controller.dailyBriefingService.GetPublishedByID(briefingID)
	if err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	result, err := controller.aiService.ExplainEnglishWord(ctx.Request.Context(), item.Title, request.Sentence, request.Word)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

func (controller *PublicController) AnalyzeDailyBriefingSentence(ctx *gin.Context) {
	briefingID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid briefing id")
		return
	}

	var request dto.StudySentenceRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request payload")
		return
	}
	if strings.TrimSpace(request.Sentence) == "" {
		response.Error(ctx, http.StatusBadRequest, "sentence is required")
		return
	}

	item, err := controller.dailyBriefingService.GetPublishedByID(briefingID)
	if err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	result, err := controller.aiService.AnalyzeEnglishSentence(ctx.Request.Context(), item.Title, request.Sentence)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

func toStringValue(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

package controller

import (
	"io"
	"net/http"

	"ai-blog/backend/internal/dto"
	"ai-blog/backend/internal/response"

	"github.com/gin-gonic/gin"
)

// ExtractAgentSource 负责处理 Agent 工作台的素材抽取。
// 这里同时支持两种来源：
// 1. 手动粘贴长文本。
// 2. 上传 docx、pptx、md、txt 文件。
func (controller *AdminController) ExtractAgentSource(ctx *gin.Context) {
	inputText := ctx.PostForm("text")

	var fileName string
	var fileData []byte

	file, err := ctx.FormFile("file")
	if err == nil && file != nil {
		fileName = file.Filename

		openedFile, openErr := file.Open()
		if openErr != nil {
			response.Error(ctx, http.StatusBadRequest, "上传文件打开失败")
			return
		}
		defer openedFile.Close()

		content, readErr := io.ReadAll(openedFile)
		if readErr != nil {
			response.Error(ctx, http.StatusBadRequest, "上传文件读取失败")
			return
		}

		fileData = content
	}

	result, err := controller.aiService.ExtractAgentSource(inputText, fileName, fileData)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, result)
}

// GenerateArticleDraft 负责根据素材生成可确认的文章草稿。
func (controller *AdminController) GenerateArticleDraft(ctx *gin.Context) {
	var request dto.AgentDraftRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	result, err := controller.aiService.GenerateArticleDraft(
		ctx.Request.Context(),
		request.SourceText,
		request.Goal,
		request.Tone,
		request.CategoryHint,
	)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, result)
}

// AgentChat 负责后台 Agent 的日常聊天。
func (controller *AdminController) AgentChat(ctx *gin.Context) {
	var request dto.AgentChatRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数格式不正确")
		return
	}

	result, err := controller.aiService.AgentChat(ctx.Request.Context(), request.Message, request.Context)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, result)
}

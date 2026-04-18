package dto

// CreateCommentRequest 是前台提交评论的请求结构。
type CreateCommentRequest struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type StudyWordRequest struct {
	Word     string `json:"word"`
	Sentence string `json:"sentence"`
}

type StudySentenceRequest struct {
	Sentence string `json:"sentence"`
}

type BriefingLearningPlanRequest struct {
	Goal string `json:"goal"`
}

type BriefingRoleplayRequest struct {
	Goal         string `json:"goal"`
	Scene        string `json:"scene"`
	LearnerReply string `json:"learner_reply"`
}

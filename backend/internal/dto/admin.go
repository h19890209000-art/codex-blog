package dto

// LoginRequest 是管理员登录请求。
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ChangePasswordRequest 是修改密码请求。
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// SaveArticleRequest 是保存文章请求。
type SaveArticleRequest struct {
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Summary      string   `json:"summary"`
	CoverURL     string   `json:"cover_url"`
	Status       int      `json:"status"`
	CategoryName string   `json:"category_name"`
	TagNames     []string `json:"tag_names"`
}

// ArticleListQuery 是后台文章分页查询参数。
type ArticleListQuery struct {
	Keyword  string `form:"keyword"`
	Status   *int   `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// CommentListQuery 是后台评论分页查询参数。
type CommentListQuery struct {
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// UpdateCommentStatusRequest 是修改评论状态请求。
type UpdateCommentStatusRequest struct {
	Status string `json:"status"`
}

// UserListQuery 是后台用户分页查询参数。
type UserListQuery struct {
	Keyword  string `form:"keyword"`
	Role     string `form:"role"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// SaveUserRequest 是创建用户请求。
type SaveUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}

// UpdateUserRoleRequest 是修改用户角色请求。
type UpdateUserRoleRequest struct {
	Role string `json:"role"`
}

// SaveCategoryRequest 是创建分类请求。
type SaveCategoryRequest struct {
	Name string `json:"name"`
}

// SaveTagRequest 是创建标签请求。
type SaveTagRequest struct {
	Name string `json:"name"`
}

type DailyBriefingListQuery struct {
	Date     string `form:"date"`
	Keyword  string `form:"keyword"`
	Status   *int   `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type SaveDailyBriefingRequest struct {
	BriefingDate      string `json:"briefing_date"`
	Title             string `json:"title"`
	Summary           string `json:"summary"`
	SourceName        string `json:"source_name"`
	SourceURL         string `json:"source_url"`
	Status            int    `json:"status"`
	SortOrder         int    `json:"sort_order"`
	SourcePublishedAt string `json:"source_published_at"`
}

type FetchDailyBriefingRequest struct {
	Date  string `json:"date"`
	Limit int    `json:"limit"`
}

type SaveSystemConfigsRequest struct {
	Items []SystemConfigItem `json:"items"`
}

type SystemConfigItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

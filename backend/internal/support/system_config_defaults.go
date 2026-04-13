package support

// SystemConfigDefinition describes a configurable system field.
type SystemConfigDefinition struct {
	Key         string
	Group       string
	Name        string
	Value       string
	InputType   string
	Description string
	IsPublic    bool
}

// DefaultSystemConfigDefinitions keeps the built-in system copy in one place.
func DefaultSystemConfigDefinitions() []SystemConfigDefinition {
	return []SystemConfigDefinition{
		{
			Key:         "reader_home_title",
			Group:       "reader_home",
			Name:        "读者端首页主标题",
			Value:       "AI 智能博客读者端",
			InputType:   "text",
			Description: "显示在读者端首页第一屏的大标题。",
			IsPublic:    true,
		},
		{
			Key:         "reader_home_intro",
			Group:       "reader_home",
			Name:        "读者端首页主说明",
			Value:       "这里除了文章列表，现在还新增了每日简讯和顶部快速内容切换。点分类或标签后，顶部会立即切成对应内容，不用再往下翻。",
			InputType:   "textarea",
			Description: "显示在读者端首页主标题下方的介绍文案。",
			IsPublic:    true,
		},
		{
			Key:         "reader_home_search_placeholder",
			Group:       "reader_home",
			Name:        "搜索框占位文案",
			Value:       "搜索文章标题或正文",
			InputType:   "text",
			Description: "首页搜索框里的占位提示。",
			IsPublic:    true,
		},
		{
			Key:         "reader_home_search_button_label",
			Group:       "reader_home",
			Name:        "搜索按钮文案",
			Value:       "搜索文章",
			InputType:   "text",
			Description: "首页搜索按钮文案。",
			IsPublic:    true,
		},
		{
			Key:         "reader_home_clear_button_label",
			Group:       "reader_home",
			Name:        "清空按钮文案",
			Value:       "清空筛选",
			InputType:   "text",
			Description: "首页清空筛选按钮文案。",
			IsPublic:    true,
		},
		{
			Key:         "reader_home_about_title",
			Group:       "reader_home",
			Name:        "博客简介标题",
			Value:       "博客简介",
			InputType:   "text",
			Description: "首页侧边栏简介卡片标题。",
			IsPublic:    true,
		},
		{
			Key:         "reader_home_about_intro",
			Group:       "reader_home",
			Name:        "博客简介说明",
			Value:       "这个读者端包含文章、分类、标签、全站问答和每日简讯。现在点左侧筛选后，主内容区会优先展示对应结果。",
			InputType:   "textarea",
			Description: "首页侧边栏简介卡片说明文案。",
			IsPublic:    true,
		},
		{
			Key:         "reader_home_site_qa_title",
			Group:       "reader_home",
			Name:        "全站问答标题",
			Value:       "全站知识问答",
			InputType:   "text",
			Description: "首页全站问答模块标题。",
			IsPublic:    true,
		},
		{
			Key:         "reader_home_site_qa_placeholder",
			Group:       "reader_home",
			Name:        "全站问答占位文案",
			Value:       "例如：这个博客里有哪些 Go 入门内容？",
			InputType:   "text",
			Description: "首页全站问答输入框占位文案。",
			IsPublic:    true,
		},
		{
			Key:         "reader_home_site_qa_button_label",
			Group:       "reader_home",
			Name:        "全站问答按钮文案",
			Value:       "提问全站 AI",
			InputType:   "text",
			Description: "首页全站问答按钮文案。",
			IsPublic:    true,
		},
		{
			Key:         "reader_home_article_list_title",
			Group:       "reader_home",
			Name:        "文章列表标题",
			Value:       "文章列表",
			InputType:   "text",
			Description: "首页文章列表模块标题。",
			IsPublic:    true,
		},
	}
}

// DefaultSystemConfigMap builds a key-based lookup for config definitions.
func DefaultSystemConfigMap() map[string]SystemConfigDefinition {
	result := make(map[string]SystemConfigDefinition)
	for _, item := range DefaultSystemConfigDefinitions() {
		result[item.Key] = item
	}
	return result
}

<script setup>
import { computed, onMounted, ref } from 'vue'
import DailyBriefingAdminPanel from './components/DailyBriefingAdminPanel.vue'
import { adminApiUrl } from './lib/api'

const token = ref(localStorage.getItem('admin_token') || '')
const activeTab = ref('dashboard')
const loading = ref(false)
const aiLoading = ref(false)
const agentLoading = ref(false)
const loginError = ref('')
const systemMessage = ref('')

const user = ref(null)
const dashboard = ref({})
const providers = ref([])
const syncStatus = ref(null)
const categories = ref([])
const tags = ref([])
const articles = ref([])
const comments = ref([])
const users = ref([])
const articleTotal = ref(0)
const commentTotal = ref(0)
const userTotal = ref(0)
const aiResult = ref(null)
const extractedSource = ref(null)
const agentDraft = ref(null)
const agentMessages = ref([])
const selectedAgentFile = ref(null)
const categoryName = ref('')
const tagName = ref('')

const loginForm = ref({ username: 'admin', password: 'admin123456' })
const passwordForm = ref({ old_password: '', new_password: '' })
const articleForm = ref({ id: 0, title: '', content: '', summary: '', cover_url: '', status: 0, category_name: '', tag_names: '' })
const articleQuery = ref({ keyword: '', status: '', page: 1, page_size: 6 })
const commentQuery = ref({ keyword: '', status: '', page: 1, page_size: 6 })
const userQuery = ref({ keyword: '', role: '', page: 1, page_size: 6 })
const userForm = ref({ username: '', password: '', avatar: '', role: 'user' })
const aiForm = ref({ content: '请把这段后台开发说明整理成更适合新手阅读的文字。', title: 'Go 新手博客后台', keyword: 'Go 博客后台', style: '通俗易懂' })
const agentForm = ref({ text: '', goal: '整理成一篇适合技术博客后台确认的文章草稿', tone: '清晰、自然、适合发布前审核', category_hint: '' })
const agentChatForm = ref({ message: '' })
const mobileMenuOpen = ref(false)

const navTabs = [
  { key: 'dashboard', label: '仪表盘', mobileLabel: '首页', icon: '盘', hint: '查看整体概览' },
  { key: 'agent', label: 'Agent 工作台', mobileLabel: 'Agent', icon: '稿', hint: '整理素材与草稿' },
  { key: 'articles', label: '文章管理', mobileLabel: '文章', icon: '文', hint: '编辑与发布内容' },
  { key: 'briefings', label: '每日简讯', mobileLabel: '简讯', icon: '讯', hint: '管理每日快讯' },
  { key: 'taxonomy', label: '分类标签', mobileLabel: '分类', icon: '类', hint: '维护分类和标签' },
  { key: 'comments', label: '评论管理', mobileLabel: '评论', icon: '评', hint: '审核互动内容' },
  { key: 'users', label: '用户管理', mobileLabel: '用户', icon: '人', hint: '查看用户与权限' },
  { key: 'ai', label: 'AI 工具台', mobileLabel: 'AI', icon: '智', hint: '生成摘要和优化' },
  { key: 'settings', label: '系统设置', mobileLabel: '设置', icon: '设', hint: '配置系统与 OSS' }
]
const mobilePrimaryTabKeys = ['dashboard', 'articles', 'briefings', 'ai']

const isLoggedIn = computed(() => Boolean(token.value))
const articlePages = computed(() => Math.max(1, Math.ceil(articleTotal.value / articleQuery.value.page_size)))
const commentPages = computed(() => Math.max(1, Math.ceil(commentTotal.value / commentQuery.value.page_size)))
const userPages = computed(() => Math.max(1, Math.ceil(userTotal.value / userQuery.value.page_size)))
const readyProviderCount = computed(() => providers.value.filter((item) => item.ready).length)
const draftTagText = computed(() => (agentDraft.value?.tag_names || []).join(', '))
const agentContextText = computed(() => agentDraft.value?.content || extractedSource.value?.content || '')
const activeTabMeta = computed(() => navTabs.find((item) => item.key === activeTab.value) || navTabs[0])
const mobilePrimaryTabs = computed(() => navTabs.filter((item) => mobilePrimaryTabKeys.includes(item.key)))
const mobileSecondaryTabs = computed(() => navTabs.filter((item) => !mobilePrimaryTabKeys.includes(item.key)))
const mobileNavActiveKey = computed(() => (mobilePrimaryTabKeys.includes(activeTab.value) ? activeTab.value : 'more'))
const aiBlocks = computed(() => {
  const value = normalizeAIResult(aiResult.value)
  if (!value) return []
  const labels = { provider: '使用模型', summary: '摘要', keywords: '关键词', meta: 'Meta 描述', tags: '标签建议', items: '灵感列表', result: '生成结果', url: '链接地址', hint: '说明', replies: '回复建议', flagged: '是否拦截', reason: '原因', raw: '原始输出', answer: '回答' }
  return Object.entries(value)
    .filter(([, value]) => value !== '' && value !== null && value !== undefined)
    .map(([key, value]) => ({ label: labels[key] || key, content: Array.isArray(value) ? value.join('\n') : typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value) }))
})

function showMessage(text) {
  systemMessage.value = text
  setTimeout(() => {
    if (systemMessage.value === text) systemMessage.value = ''
  }, 2500)
}

function cleanAIText(value) {
  return String(value || '')
    .replace(/\*\*/g, '')
    .replace(/__/g, '')
    .replace(/`/g, '')
    .replace(/[：]/g, ':')
    .trim()
}

function looksLikeSectionLabel(value) {
  const normalized = cleanAIText(value).toLowerCase()
  return ['摘要', 'summary', '关键词', '关键字', 'keywords', 'meta', 'meta 描述', 'meta描述', '描述'].includes(normalized)
}

function parseSummaryFromRaw(raw) {
  const lines = String(raw || '')
    .split(/\r?\n/)
    .map((line) => cleanAIText(line).replace(/^[#\-*\d.\s]+/, '').trim())
    .filter(Boolean)

  const result = { summary: '', keywords: [], meta: '' }
  let section = ''

  for (const line of lines) {
    const lower = line.toLowerCase()

    if (lower === '摘要' || lower === 'summary') {
      section = 'summary'
      continue
    }
    if (lower === '关键词' || lower === '关键字' || lower === 'keywords') {
      section = 'keywords'
      continue
    }
    if (lower === 'meta 描述' || lower === 'meta描述' || lower === 'meta' || lower === '描述') {
      section = 'meta'
      continue
    }

    const match = line.match(/^(摘要|summary|关键词|关键字|keywords|meta 描述|meta描述|meta|描述):\s*(.+)$/i)
    if (match) {
      const label = match[1].toLowerCase()
      const content = match[2].trim()
      if (label === '摘要' || label === 'summary') result.summary = content
      if (label === '关键词' || label === '关键字' || label === 'keywords') result.keywords = content.split(/[,\uff0c]/).map((item) => cleanAIText(item)).filter(Boolean)
      if (label === 'meta 描述' || label === 'meta描述' || label === 'meta' || label === '描述') result.meta = content
      continue
    }

    if (section === 'summary' && !result.summary) {
      result.summary = line
      continue
    }
    if (section === 'keywords' && result.keywords.length === 0) {
      result.keywords = line.split(/[,\uff0c]/).map((item) => cleanAIText(item)).filter(Boolean)
      continue
    }
    if (section === 'meta' && !result.meta) {
      result.meta = line
      continue
    }

    if (!result.summary) {
      result.summary = line
    } else if (!result.meta) {
      result.meta = line
    }
  }

  return result
}

function normalizeAIResult(payload) {
  if (!payload || typeof payload !== 'object') return payload

  const result = { ...payload }

  if (typeof result.summary === 'string') result.summary = cleanAIText(result.summary)
  if (typeof result.meta === 'string') result.meta = cleanAIText(result.meta)
  if (typeof result.result === 'string') result.result = cleanAIText(result.result)
  if (typeof result.answer === 'string') result.answer = cleanAIText(result.answer)
  if (Array.isArray(result.keywords)) result.keywords = result.keywords.map((item) => cleanAIText(item)).filter(Boolean)
  if (Array.isArray(result.tags)) result.tags = result.tags.map((item) => cleanAIText(item)).filter(Boolean)
  if (Array.isArray(result.items)) result.items = result.items.map((item) => cleanAIText(item)).filter(Boolean)
  if (Array.isArray(result.replies)) result.replies = result.replies.map((item) => cleanAIText(item)).filter(Boolean)

  if (typeof result.raw === 'string') {
    const parsed = parseSummaryFromRaw(result.raw)
    if ((!result.summary || looksLikeSectionLabel(result.summary)) && parsed.summary) result.summary = parsed.summary
    if ((!result.meta || looksLikeSectionLabel(result.meta)) && parsed.meta) result.meta = parsed.meta
    if ((!Array.isArray(result.keywords) || result.keywords.length === 0 || result.keywords.every(looksLikeSectionLabel)) && parsed.keywords.length > 0) {
      result.keywords = parsed.keywords
    }
  }

  return result
}

function authHeaders(asJson = true) {
  const headers = { Authorization: `Bearer ${token.value}` }
  if (asJson) headers['Content-Type'] = 'application/json'
  return headers
}

function buildQueryString(data) {
  const params = new URLSearchParams()
  Object.entries(data).forEach(([key, value]) => {
    if (value !== '' && value !== null && value !== undefined) params.set(key, String(value))
  })
  return params.toString()
}

async function request(path, options = {}) {
  const response = await fetch(adminApiUrl(path), options)
  const result = await response.json()
  if (!response.ok || result.success === false) throw new Error(result.error || '请求失败')
  return result.data
}

async function login() {
  loginError.value = ''
  try {
    const data = await request('/auth/login', { method: 'POST', headers: authHeaders(true), body: JSON.stringify(loginForm.value) })
    token.value = data.token
    localStorage.setItem('admin_token', data.token)
    await loadAll()
    showMessage('登录成功')
  } catch (error) {
    loginError.value = error.message
  }
}

function logout() {
  mobileMenuOpen.value = false
  token.value = ''
  user.value = null
  localStorage.removeItem('admin_token')
}

function setActiveTab(tabKey) {
  activeTab.value = tabKey
  mobileMenuOpen.value = false
}

function toggleMobileMenu() {
  mobileMenuOpen.value = !mobileMenuOpen.value
}

async function loadAll() {
  if (!token.value) return
  loading.value = true
  try {
    const [me, dash, categoryData, tagData, providerData, syncData] = await Promise.all([
      request('/me', { headers: authHeaders() }),
      request('/dashboard', { headers: authHeaders() }),
      request('/categories', { headers: authHeaders() }),
      request('/tags', { headers: authHeaders() }),
      request('/ai/providers', { headers: authHeaders() }),
      request('/sync/oss/status', { headers: authHeaders() })
    ])
    user.value = me.user
    dashboard.value = dash
    categories.value = categoryData || []
    tags.value = tagData || []
    providers.value = providerData.providers || []
    syncStatus.value = syncData
    await Promise.all([loadArticles(), loadComments(), loadUsers()])
  } finally {
    loading.value = false
  }
}

async function loadArticles() {
  const data = await request(`/articles?${buildQueryString(articleQuery.value)}`, { headers: authHeaders() })
  articles.value = data.items || []
  articleTotal.value = data.total || 0
}

async function loadComments() {
  const data = await request(`/comments?${buildQueryString(commentQuery.value)}`, { headers: authHeaders() })
  comments.value = data.items || []
  commentTotal.value = data.total || 0
}

async function loadUsers() {
  const data = await request(`/users?${buildQueryString(userQuery.value)}`, { headers: authHeaders() })
  users.value = data.items || []
  userTotal.value = data.total || 0
}

function resetArticleForm() {
  articleForm.value = { id: 0, title: '', content: '', summary: '', cover_url: '', status: 0, category_name: '', tag_names: '' }
}

function editArticle(article) {
  articleForm.value = { id: article.id, title: article.title, content: article.content, summary: article.summary, cover_url: article.cover_url, status: article.status, category_name: article.category?.name || '', tag_names: (article.tags || []).map((tag) => tag.name).join(', ') }
  activeTab.value = 'articles'
}

async function saveArticle() {
  const payload = { ...articleForm.value, tag_names: articleForm.value.tag_names.split(',').map((item) => item.trim()).filter(Boolean) }
  await request(payload.id ? `/articles/${payload.id}` : '/articles', { method: payload.id ? 'PUT' : 'POST', headers: authHeaders(), body: JSON.stringify(payload) })
  resetArticleForm()
  await loadArticles()
  showMessage('文章保存成功')
}

async function deleteArticle(article) {
  if (!window.confirm(`确定删除《${article.title}》吗？`)) return
  await request(`/articles/${article.id}`, { method: 'DELETE', headers: authHeaders() })
  await loadArticles()
  showMessage('文章删除成功')
}

async function createCategory() {
  if (!categoryName.value.trim()) return
  await request('/categories', { method: 'POST', headers: authHeaders(), body: JSON.stringify({ name: categoryName.value }) })
  categoryName.value = ''
  categories.value = await request('/categories', { headers: authHeaders() })
  showMessage('分类创建成功')
}

async function createTag() {
  if (!tagName.value.trim()) return
  await request('/tags', { method: 'POST', headers: authHeaders(), body: JSON.stringify({ name: tagName.value }) })
  tagName.value = ''
  tags.value = await request('/tags', { headers: authHeaders() })
  showMessage('标签创建成功')
}

async function updateCommentStatus(comment, status) {
  await request(`/comments/${comment.id}/status`, { method: 'PUT', headers: authHeaders(), body: JSON.stringify({ status }) })
  await loadComments()
}

async function deleteComment(comment) {
  if (!window.confirm('确定删除这条评论吗？')) return
  await request(`/comments/${comment.id}`, { method: 'DELETE', headers: authHeaders() })
  await loadComments()
}

async function createUser() {
  await request('/users', { method: 'POST', headers: authHeaders(), body: JSON.stringify(userForm.value) })
  userForm.value = { username: '', password: '', avatar: '', role: 'user' }
  await loadUsers()
  showMessage('用户创建成功')
}

async function updateUserRole(targetUser, role) {
  await request(`/users/${targetUser.id}/role`, { method: 'PUT', headers: authHeaders(), body: JSON.stringify({ role }) })
  await loadUsers()
}

async function deleteUser(targetUser) {
  if (!window.confirm(`确定删除用户 ${targetUser.username} 吗？`)) return
  await request(`/users/${targetUser.id}`, { method: 'DELETE', headers: authHeaders() })
  await loadUsers()
}

async function changePassword() {
  await request('/auth/change-password', { method: 'POST', headers: authHeaders(), body: JSON.stringify(passwordForm.value) })
  passwordForm.value = { old_password: '', new_password: '' }
  showMessage('密码修改成功')
}

async function runAITool(path, payload) {
  aiLoading.value = true
  aiResult.value = { provider: '处理中', result: 'AI 正在处理，请稍候...' }
  try {
    aiResult.value = normalizeAIResult(await request(path, { method: 'POST', headers: authHeaders(), body: JSON.stringify(payload) }))
  } catch (error) {
    aiResult.value = { provider: 'error', reason: error.message }
  } finally {
    aiLoading.value = false
  }
}

async function runSync() {
  const data = await request('/sync/oss/run', { method: 'POST', headers: authHeaders() })
  syncStatus.value = await request('/sync/oss/status', { headers: authHeaders() })
  await loadArticles()
  showMessage(data.message || '同步完成')
}

function handleAgentFileChange(event) {
  selectedAgentFile.value = event.target.files?.[0] || null
}

async function extractAgentSource() {
  agentLoading.value = true
  try {
    const formData = new FormData()
    if (agentForm.value.text.trim()) formData.append('text', agentForm.value.text)
    if (selectedAgentFile.value) formData.append('file', selectedAgentFile.value)
    extractedSource.value = await request('/agent/extract', { method: 'POST', headers: authHeaders(false), body: formData })
    showMessage('素材抽取完成')
  } catch (error) {
    extractedSource.value = null
    showMessage(error.message)
  } finally {
    agentLoading.value = false
  }
}

async function generateAgentDraft() {
  const sourceText = extractedSource.value?.content || agentForm.value.text.trim()
  if (!sourceText) {
    showMessage('请先上传文件或粘贴素材，再生成草稿')
    return
  }
  agentLoading.value = true
  try {
    agentDraft.value = await request('/agent/generate-draft', { method: 'POST', headers: authHeaders(), body: JSON.stringify({ source_text: sourceText, goal: agentForm.value.goal, tone: agentForm.value.tone, category_hint: agentForm.value.category_hint }) })
    showMessage('文章草稿已生成')
  } finally {
    agentLoading.value = false
  }
}

function fillDraftIntoArticleForm() {
  if (!agentDraft.value) return
  articleForm.value = { id: 0, title: agentDraft.value.title || '', content: agentDraft.value.content || '', summary: agentDraft.value.summary || '', cover_url: '', status: 0, category_name: agentDraft.value.category_name || '', tag_names: (agentDraft.value.tag_names || []).join(', ') }
  activeTab.value = 'articles'
  showMessage('草稿已经填充到文章编辑器')
}

async function saveAgentDraftAsArticle() {
  if (!agentDraft.value) return
  await request('/articles', { method: 'POST', headers: authHeaders(), body: JSON.stringify({ title: agentDraft.value.title || '待确认草稿', content: agentDraft.value.content || '', summary: agentDraft.value.summary || '', cover_url: '', status: 0, category_name: agentDraft.value.category_name || '', tag_names: agentDraft.value.tag_names || [] }) })
  await loadArticles()
  showMessage('Agent 草稿已保存到文章列表')
}

async function sendAgentChat() {
  if (!agentChatForm.value.message.trim()) return
  const question = agentChatForm.value.message.trim()
  agentMessages.value.push({ role: 'user', content: question })
  agentChatForm.value.message = ''
  agentLoading.value = true
  try {
    const data = await request('/agent/chat', { method: 'POST', headers: authHeaders(), body: JSON.stringify({ message: question, context: agentContextText.value }) })
    agentMessages.value.push({ role: 'assistant', content: data.answer, provider: data.provider, hint: data.hint })
  } catch (error) {
    agentMessages.value.push({ role: 'assistant', content: error.message, provider: 'error' })
  } finally {
    agentLoading.value = false
  }
}

function clearAgentWorkspace() {
  extractedSource.value = null
  agentDraft.value = null
  agentMessages.value = []
  selectedAgentFile.value = null
  agentForm.value = { text: '', goal: '整理成一篇适合技术博客后台确认的文章草稿', tone: '清晰、自然、适合发布前审核', category_hint: '' }
}

function turnPage(query, pages, step, loader) {
  const nextPage = query.value.page + step
  if (nextPage < 1 || nextPage > pages.value) return
  query.value.page = nextPage
  loader()
}

function providerCapabilityText(capabilities) {
  const map = { chat: '聊天', stream_chat: '流式聊天', embedding: '向量', moderate: '审核', image_generate: '图片', text_to_speech: '语音' }
  return (capabilities || []).map((item) => map[item] || item).join(' / ')
}

onMounted(() => {
  if (token.value) loadAll()
})
</script>

<template>
  <main class="page-shell">
    <section v-if="!isLoggedIn" class="login-layout">
      <article class="hero-card">
        <p class="eyebrow">AI BLOG CONSOLE</p>
        <h1>一个更像正式产品的博客后台</h1>
        <p class="hero-text">文章管理、评论审核、用户管理、AI 工具和 Agent 工作台都放在同一个后台里。你可以上传 Word、PPT、Markdown 或长文本，让 Agent 先整理成文章草稿。</p>
        <div class="tip-list">
          <div class="tip-card"><strong>内容创作</strong><span>支持摘要、标签、改写、选题和草稿整理。</span></div>
          <div class="tip-card"><strong>资料转文章</strong><span>支持 `docx / pptx / md / txt` 内容抽取。</span></div>
          <div class="tip-card"><strong>日常 AI</strong><span>后台 Agent 平时也可以当成日常助手来用。</span></div>
        </div>
      </article>

      <article class="panel-card login-card">
        <p class="eyebrow">Admin Login</p>
        <h2>后台登录</h2>
        <p class="muted">默认账号是 `admin`，默认密码是 `admin123456`。</p>
        <label class="field"><span>用户名</span><input v-model="loginForm.username" type="text" /></label>
        <label class="field"><span>密码</span><input v-model="loginForm.password" type="password" @keyup.enter="login" /></label>
        <button class="primary large-button" @click="login">进入管理后台</button>
        <p v-if="loginError" class="error-text">{{ loginError }}</p>
      </article>
    </section>

    <template v-else>
      <aside class="panel-card sidebar">
        <div class="sidebar-top">
          <p class="eyebrow">管理后台</p>
          <h2>{{ user?.username || '管理员' }}</h2>
          <p class="muted">当前角色：{{ user?.role || 'admin' }}</p>
          <p v-if="systemMessage" class="success-text">{{ systemMessage }}</p>
        </div>

        <nav class="menu-list">
          <button v-for="tab in navTabs" :key="tab.key" :class="{ active: activeTab === tab.key }" @click="setActiveTab(tab.key)">{{ tab.label }}</button>
        </nav>

        <button class="ghost" @click="logout">退出登录</button>
      </aside>

      <section class="content-area">
        <article class="panel-card mobile-header">
          <div class="mobile-header-copy">
            <p class="eyebrow">管理后台</p>
            <h2>{{ activeTabMeta.label }}</h2>
            <p class="muted">{{ user?.username || '管理员' }} · {{ user?.role || 'admin' }}</p>
            <p v-if="systemMessage" class="success-text">{{ systemMessage }}</p>
          </div>
          <button class="ghost" @click="logout">退出</button>
        </article>

        <section v-if="activeTab === 'dashboard'" class="stats-grid">
          <article class="panel-card stat-card"><p>文章</p><strong>{{ dashboard.article_count || 0 }}</strong></article>
          <article class="panel-card stat-card"><p>已发布</p><strong>{{ dashboard.published_count || 0 }}</strong></article>
          <article class="panel-card stat-card"><p>阅读量</p><strong>{{ dashboard.view_count || 0 }}</strong></article>
          <article class="panel-card stat-card"><p>分类</p><strong>{{ dashboard.category_count || 0 }}</strong></article>
          <article class="panel-card stat-card"><p>标签</p><strong>{{ dashboard.tag_count || 0 }}</strong></article>
          <article class="panel-card stat-card"><p>评论</p><strong>{{ dashboard.comment_count || 0 }}</strong></article>
          <article class="panel-card stat-card"><p>用户</p><strong>{{ dashboard.user_count || 0 }}</strong></article>
          <article class="panel-card stat-card"><p>可用模型</p><strong>{{ readyProviderCount }}</strong></article>
        </section>

        <section v-if="activeTab === 'agent'" class="two-column">
          <article class="panel-card">
            <div class="section-head"><div><h2>Agent 工作台</h2><p class="muted">先抽素材，再生成草稿，最后决定是否进入文章管理。</p></div><button class="ghost" @click="clearAgentWorkspace">清空工作台</button></div>
            <label class="field"><span>粘贴素材</span><textarea v-model="agentForm.text" rows="8" placeholder="你可以直接粘贴会议纪要、培训内容、需求说明或长文本。"></textarea></label>
            <label class="field"><span>上传文件</span><input type="file" accept=".txt,.md,.markdown,.docx,.pptx" @change="handleAgentFileChange" /></label>
            <p v-if="selectedAgentFile" class="muted">已选择文件：{{ selectedAgentFile.name }}</p>
            <div class="inline-grid">
              <label class="field"><span>生成目标</span><input v-model="agentForm.goal" type="text" /></label>
              <label class="field"><span>语气风格</span><input v-model="agentForm.tone" type="text" /></label>
            </div>
            <label class="field"><span>分类提示</span><input v-model="agentForm.category_hint" type="text" placeholder="比如：Go、AI、教程、数据库" /></label>
            <div class="button-grid"><button class="primary" :disabled="agentLoading" @click="extractAgentSource">1. 抽取内容</button><button class="primary" :disabled="agentLoading" @click="generateAgentDraft">2. 生成草稿</button></div>
            <div class="result-panel">
              <div class="section-head"><h3>抽取结果</h3><span class="muted">{{ extractedSource ? `共 ${extractedSource.char_count} 字` : '还没有抽取素材' }}</span></div>
              <div v-if="!extractedSource" class="empty-box">上传文件或粘贴长文本后，这里会展示 Agent 抽取出来的正文。</div>
              <div v-else class="stack"><p class="muted">来源：{{ (extractedSource.sources || []).join(' / ') }}</p><pre class="result-pre">{{ extractedSource.preview }}</pre></div>
            </div>
          </article>

          <article class="panel-card">
            <div class="section-head"><div><h2>草稿结果</h2><p class="muted">这里会生成一篇适合博客后台确认的 Markdown 草稿。</p></div></div>
            <div v-if="!agentDraft" class="empty-box">先完成“抽取内容”或直接粘贴素材，然后点击“生成草稿”。</div>
            <div v-else class="stack">
              <div class="draft-meta">
                <div class="meta-card"><span>标题</span><strong>{{ agentDraft.title }}</strong></div>
                <div class="meta-card"><span>分类</span><strong>{{ agentDraft.category_name }}</strong></div>
                <div class="meta-card"><span>标签</span><strong>{{ draftTagText || '暂无标签' }}</strong></div>
                <div class="meta-card"><span>模型</span><strong>{{ agentDraft.provider }}</strong></div>
              </div>
              <label class="field"><span>摘要</span><textarea :value="agentDraft.summary" rows="3" readonly></textarea></label>
              <label class="field"><span>正文草稿</span><textarea :value="agentDraft.content" rows="15" readonly></textarea></label>
              <p v-if="agentDraft.hint" class="muted">提示：{{ agentDraft.hint }}</p>
              <div class="button-grid"><button class="primary" @click="fillDraftIntoArticleForm">填充到文章编辑器</button><button class="ghost" @click="saveAgentDraftAsArticle">一键保存为草稿</button></div>
            </div>
          </article>

          <article class="panel-card full-width">
            <div class="section-head"><div><h2>日常 AI 助手</h2><p class="muted">如果上面已经抽取了素材或生成了草稿，这里会自动把它们当作上下文参考。</p></div></div>
            <div class="chat-box">
              <div v-if="agentMessages.length === 0" class="empty-box">你可以直接问：帮我把这份 PPT 变成大纲、帮我想个更好的标题、这篇草稿还缺什么。</div>
              <div v-for="(message, index) in agentMessages" :key="index" class="chat-item" :class="message.role">
                <p class="chat-role">{{ message.role === 'user' ? '你' : `Agent · ${message.provider || 'assistant'}` }}</p>
                <pre class="chat-content">{{ message.content }}</pre>
                <p v-if="message.hint" class="muted">{{ message.hint }}</p>
              </div>
            </div>
            <div class="chat-form"><textarea v-model="agentChatForm.message" rows="3" placeholder="输入你要问 Agent 的问题，比如：请把这份材料整理成适合技术读者的结构。"></textarea><button class="primary" :disabled="agentLoading" @click="sendAgentChat">发送</button></div>
          </article>
        </section>

        <section v-if="activeTab === 'articles'" class="two-column">
          <article class="panel-card">
            <div class="section-head"><h2>{{ articleForm.id ? '编辑文章' : '新建文章' }}</h2><button class="ghost" @click="resetArticleForm">清空</button></div>
            <label class="field"><span>标题</span><input v-model="articleForm.title" type="text" /></label>
            <label class="field"><span>摘要</span><textarea v-model="articleForm.summary" rows="3"></textarea></label>
            <label class="field"><span>封面地址</span><input v-model="articleForm.cover_url" type="text" /></label>
            <div class="inline-grid">
              <label class="field"><span>分类</span><input v-model="articleForm.category_name" type="text" /></label>
              <label class="field"><span>状态</span><select v-model="articleForm.status"><option :value="0">草稿</option><option :value="1">发布</option></select></label>
            </div>
            <label class="field"><span>标签</span><input v-model="articleForm.tag_names" type="text" placeholder="多个标签用英文逗号分隔" /></label>
            <label class="field"><span>正文</span><textarea v-model="articleForm.content" rows="14"></textarea></label>
            <button class="primary" @click="saveArticle">保存文章</button>
          </article>

          <article class="panel-card">
            <div class="section-head"><h2>文章列表</h2><span class="muted">共 {{ articleTotal }} 篇</span></div>
            <div class="search-grid"><input v-model="articleQuery.keyword" type="text" placeholder="搜索标题或正文" /><select v-model="articleQuery.status"><option value="">全部</option><option value="0">草稿</option><option value="1">发布</option></select><button class="primary" @click="articleQuery.page = 1; loadArticles()">搜索</button></div>
            <div class="stack">
              <div v-for="article in articles" :key="article.id" class="list-item">
                <div class="item-main"><strong>{{ article.title }}</strong><p>{{ article.summary || '暂无摘要' }}</p><small>分类：{{ article.category?.name || '未分类' }} / 来源：{{ article.source_type || 'manual' }}</small></div>
                <div class="item-actions"><button class="ghost" @click="editArticle(article)">编辑</button><button class="danger" @click="deleteArticle(article)">删除</button></div>
              </div>
            </div>
            <div class="pager"><button class="ghost" @click="turnPage(articleQuery, articlePages, -1, loadArticles)">上一页</button><span>{{ articleQuery.page }} / {{ articlePages }}</span><button class="ghost" @click="turnPage(articleQuery, articlePages, 1, loadArticles)">下一页</button></div>
          </article>
        </section>

        <DailyBriefingAdminPanel
          v-if="activeTab === 'briefings'"
          :token="token"
          @notify="showMessage"
        />

        <section v-if="activeTab === 'taxonomy'" class="two-column">
          <article class="panel-card"><h2>分类管理</h2><div class="search-grid compact"><input v-model="categoryName" type="text" placeholder="输入分类名称" /><button class="primary" @click="createCategory">创建分类</button></div><div class="chip-list"><span v-for="category in categories" :key="category.id" class="chip">{{ category.name }}</span></div></article>
          <article class="panel-card"><h2>标签管理</h2><div class="search-grid compact"><input v-model="tagName" type="text" placeholder="输入标签名称" /><button class="primary" @click="createTag">创建标签</button></div><div class="chip-list"><span v-for="tag in tags" :key="tag.id" class="chip">{{ tag.name }}</span></div></article>
        </section>

        <section v-if="activeTab === 'comments'" class="panel-card">
          <div class="section-head"><h2>评论管理</h2><span class="muted">共 {{ commentTotal }} 条</span></div>
          <div class="search-grid"><input v-model="commentQuery.keyword" type="text" placeholder="搜索作者或内容" /><select v-model="commentQuery.status"><option value="">全部</option><option value="approved">已通过</option><option value="pending">待审核</option><option value="rejected">已拒绝</option></select><button class="primary" @click="commentQuery.page = 1; loadComments()">搜索</button></div>
          <div class="stack">
            <div v-for="comment in comments" :key="comment.id" class="list-item">
              <div class="item-main"><strong>{{ comment.author }}</strong><p>{{ comment.content }}</p><small>文章：{{ comment.article?.title || '未知文章' }} / 状态：{{ comment.status }}</small></div>
              <div class="item-actions vertical"><button class="ghost" @click="updateCommentStatus(comment, 'approved')">通过</button><button class="ghost" @click="updateCommentStatus(comment, 'pending')">待审核</button><button class="ghost" @click="updateCommentStatus(comment, 'rejected')">拒绝</button><button class="danger" @click="deleteComment(comment)">删除</button></div>
            </div>
          </div>
          <div class="pager"><button class="ghost" @click="turnPage(commentQuery, commentPages, -1, loadComments)">上一页</button><span>{{ commentQuery.page }} / {{ commentPages }}</span><button class="ghost" @click="turnPage(commentQuery, commentPages, 1, loadComments)">下一页</button></div>
        </section>

        <section v-if="activeTab === 'users'" class="two-column">
          <article class="panel-card">
            <h2>新增用户</h2>
            <label class="field"><span>用户名</span><input v-model="userForm.username" type="text" /></label>
            <label class="field"><span>密码</span><input v-model="userForm.password" type="password" /></label>
            <label class="field"><span>头像地址</span><input v-model="userForm.avatar" type="text" /></label>
            <label class="field"><span>角色</span><select v-model="userForm.role"><option value="user">普通用户</option><option value="admin">管理员</option></select></label>
            <button class="primary" @click="createUser">创建用户</button>
          </article>

          <article class="panel-card">
            <div class="section-head"><h2>用户管理</h2><span class="muted">共 {{ userTotal }} 个</span></div>
            <div class="search-grid"><input v-model="userQuery.keyword" type="text" placeholder="搜索用户名" /><select v-model="userQuery.role"><option value="">全部</option><option value="user">普通用户</option><option value="admin">管理员</option></select><button class="primary" @click="userQuery.page = 1; loadUsers()">搜索</button></div>
            <div class="stack">
              <div v-for="targetUser in users" :key="targetUser.id" class="list-item">
                <div class="item-main"><strong>{{ targetUser.username }}</strong><p>角色：{{ targetUser.role }}</p></div>
                <div class="item-actions vertical"><button class="ghost" @click="updateUserRole(targetUser, 'user')">设为用户</button><button class="ghost" @click="updateUserRole(targetUser, 'admin')">设为管理员</button><button class="danger" @click="deleteUser(targetUser)">删除</button></div>
              </div>
            </div>
            <div class="pager"><button class="ghost" @click="turnPage(userQuery, userPages, -1, loadUsers)">上一页</button><span>{{ userQuery.page }} / {{ userPages }}</span><button class="ghost" @click="turnPage(userQuery, userPages, 1, loadUsers)">下一页</button></div>
          </article>
        </section>

        <section v-if="activeTab === 'ai'" class="two-column">
          <article class="panel-card">
            <div class="section-head"><div><h2>AI 工具台</h2><p class="muted">{{ aiLoading ? 'AI 正在处理中...' : '支持自动降级和本地兜底说明' }}</p></div></div>
            <label class="field"><span>示例内容</span><textarea v-model="aiForm.content" rows="7"></textarea></label>
            <label class="field"><span>示例标题</span><input v-model="aiForm.title" type="text" /></label>
            <div class="inline-grid">
              <label class="field"><span>关键词</span><input v-model="aiForm.keyword" type="text" /></label>
              <label class="field"><span>改写风格</span><input v-model="aiForm.style" type="text" /></label>
            </div>
            <div class="button-grid three"><button class="primary" @click="runAITool('/ai/generate-summary', { content: aiForm.content })">生成摘要</button><button class="primary" @click="runAITool('/ai/suggest-tags', { content: aiForm.content })">推荐标签</button><button class="primary" @click="runAITool('/ai/brainstorm', { keyword: aiForm.keyword })">灵感风暴</button><button class="primary" @click="runAITool('/ai/rewrite', { content: aiForm.content, style: aiForm.style })">润色改写</button><button class="primary" @click="runAITool('/ai/generate-cover', { title: aiForm.title })">生成封面</button><button class="primary" @click="runAITool('/comments/moderate', { content: aiForm.content })">评论审核</button></div>
            <div class="result-panel">
              <div class="section-head"><h3>本次输出</h3><span class="muted">{{ aiResult?.provider ? `来源：${aiResult.provider}` : '等待调用' }}</span></div>
              <div v-if="aiBlocks.length === 0" class="empty-box">点击上面的任意一个 AI 工具后，这里会展示结构化结果。</div>
              <div v-else class="stack"><div v-for="block in aiBlocks" :key="block.label" class="result-block"><p class="muted">{{ block.label }}</p><pre class="result-pre">{{ block.content }}</pre></div></div>
            </div>
          </article>

          <article class="panel-card">
            <div class="section-head"><h2>模型状态</h2><span class="muted">可用 {{ readyProviderCount }} / {{ providers.length }}</span></div>
            <div class="stack">
              <div v-for="item in providers" :key="item.alias" class="provider-card" :class="{ ready: item.ready }">
                <div class="section-head"><strong>{{ item.alias }}</strong><span class="chip">{{ item.ready ? '可用' : '未就绪' }}</span></div>
                <p class="muted">模型：{{ item.model || '未配置' }}</p>
                <p class="muted">能力：{{ providerCapabilityText(item.capabilities) || '暂无' }}</p>
                <p>{{ item.reason }}</p>
              </div>
            </div>
          </article>
        </section>

        <section v-if="activeTab === 'settings'" class="two-column">
          <article class="panel-card">
            <h2>管理员密码</h2>
            <label class="field"><span>旧密码</span><input v-model="passwordForm.old_password" type="password" /></label>
            <label class="field"><span>新密码</span><input v-model="passwordForm.new_password" type="password" /></label>
            <button class="primary" @click="changePassword">修改密码</button>
          </article>
          <article class="panel-card">
            <div class="section-head"><div><h2>OSS 同步</h2><p class="muted">每小时自动同步一次，也支持手动立即同步。</p></div><button class="primary" :disabled="loading || syncStatus?.is_running" @click="runSync">{{ syncStatus?.is_running ? '同步中...' : '立即同步' }}</button></div>
            <div class="stack"><p>Bucket：{{ syncStatus?.bucket || '-' }}</p><p>Endpoint：{{ syncStatus?.endpoint || '-' }}</p><p>Prefix：{{ syncStatus?.prefix || '-' }}</p><p>同步间隔：{{ syncStatus?.interval_minutes || 60 }} 分钟</p><p>最近结果：{{ syncStatus?.last_result?.message || '还没有执行过同步任务。' }}</p></div>
          </article>
        </section>
      </section>

      <div v-if="mobileMenuOpen" class="mobile-sheet-backdrop" @click="mobileMenuOpen = false">
        <section class="panel-card mobile-sheet" @click.stop>
          <div class="section-head mobile-sheet-head">
            <div>
              <h3>更多功能</h3>
              <p class="muted">切换到其它后台模块，或者直接退出登录。</p>
            </div>
            <button class="ghost" @click="mobileMenuOpen = false">收起</button>
          </div>
          <div class="mobile-sheet-grid">
            <button
              v-for="tab in mobileSecondaryTabs"
              :key="tab.key"
              class="mobile-sheet-card"
              :class="{ active: activeTab === tab.key }"
              @click="setActiveTab(tab.key)"
            >
              <span class="mobile-tab-icon">{{ tab.icon }}</span>
              <span class="mobile-tab-copy">
                <strong>{{ tab.label }}</strong>
                <small>{{ tab.hint }}</small>
              </span>
            </button>
          </div>
          <button class="ghost mobile-logout" @click="logout">退出登录</button>
        </section>
      </div>

      <nav class="mobile-tabbar" aria-label="移动端底部导航">
        <button
          v-for="tab in mobilePrimaryTabs"
          :key="tab.key"
          :class="{ active: mobileNavActiveKey === tab.key }"
          @click="setActiveTab(tab.key)"
        >
          <span class="mobile-tab-icon">{{ tab.icon }}</span>
          <span class="mobile-tab-label">{{ tab.mobileLabel }}</span>
        </button>
        <button :class="{ active: mobileNavActiveKey === 'more' || mobileMenuOpen }" @click="toggleMobileMenu">
          <span class="mobile-tab-icon">更</span>
          <span class="mobile-tab-label">更多</span>
        </button>
      </nav>
    </template>
  </main>
</template>

<style scoped>
:global(body) {
  margin: 0;
  font-family: "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
  color: #223040;
  background:
    radial-gradient(circle at 12% 18%, rgba(255, 210, 146, 0.34), transparent 0 18%),
    radial-gradient(circle at 88% 20%, rgba(134, 191, 255, 0.28), transparent 0 16%),
    radial-gradient(circle at 76% 80%, rgba(255, 255, 255, 0.28), transparent 0 18%),
    linear-gradient(145deg, #f7efe4 0%, #edf2f8 46%, #eef5fb 100%);
  min-height: 100vh;
  overflow-x: hidden;
}

.page-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 20px;
  padding: 20px;
  box-sizing: border-box;
  perspective: 1500px;
  align-items: start;
}

.login-layout {
  min-height: calc(100vh - 40px);
  display: grid;
  grid-template-columns: 1.2fr 0.8fr;
  gap: 24px;
  grid-column: 1 / -1;
}

.panel-card,
.hero-card {
  position: relative;
  overflow: hidden;
  isolation: isolate;
  border-radius: 32px;
  padding: 24px;
  box-sizing: border-box;
  backdrop-filter: blur(28px) saturate(150%);
  -webkit-backdrop-filter: blur(28px) saturate(150%);
  transform-style: preserve-3d;
  transition: transform 0.34s ease, box-shadow 0.34s ease, border-color 0.34s ease;
}

.panel-card::before,
.hero-card::before {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background:
    radial-gradient(circle at 14% 0%, rgba(255, 255, 255, 0.88), transparent 30%),
    radial-gradient(circle at 84% 10%, rgba(255, 255, 255, 0.42), transparent 24%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.74) 0%, rgba(255, 255, 255, 0.18) 28%, rgba(255, 255, 255, 0.08) 62%, rgba(255, 255, 255, 0.12) 100%);
  pointer-events: none;
  z-index: 0;
}

.panel-card::after,
.hero-card::after {
  content: "";
  position: absolute;
  width: 220px;
  height: 220px;
  top: -90px;
  right: -70px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.54), transparent 68%);
  pointer-events: none;
  z-index: 0;
}

.panel-card > *,
.hero-card > * {
  position: relative;
  z-index: 1;
}

.panel-card {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.62), rgba(255, 255, 255, 0.34)),
    linear-gradient(150deg, rgba(255, 255, 255, 0.14), rgba(170, 203, 234, 0.08));
  border: 1px solid rgba(255, 255, 255, 0.6);
  box-shadow:
    0 28px 70px rgba(45, 67, 92, 0.14),
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    inset 0 -18px 36px rgba(110, 151, 194, 0.08);
}

.panel-card:hover {
  transform: translateY(-6px) rotateX(4deg) rotateY(-3deg);
  box-shadow:
    0 34px 84px rgba(45, 67, 92, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.76),
    inset 0 -22px 38px rgba(110, 151, 194, 0.1);
  border-color: rgba(255, 255, 255, 0.72);
}

.hero-card {
  padding: 42px;
  background:
    linear-gradient(145deg, rgba(17, 34, 54, 0.96), rgba(33, 63, 92, 0.92) 48%, rgba(105, 139, 162, 0.84)),
    radial-gradient(circle at top right, rgba(255, 255, 255, 0.18), transparent 32%);
  color: #fff;
  box-shadow:
    0 34px 84px rgba(35, 52, 74, 0.24),
    inset 0 1px 0 rgba(255, 255, 255, 0.22);
}

.hero-card:hover {
  transform: translateY(-8px) rotateX(5deg) rotateY(-3deg);
}

.hero-text {
  line-height: 1.9;
  color: rgba(255, 255, 255, 0.88);
}

.tip-list {
  margin-top: 28px;
  display: grid;
  gap: 14px;
}

.tip-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 18px 20px;
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.12), rgba(255, 255, 255, 0.06));
  border: 1px solid rgba(255, 255, 255, 0.16);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.16);
}

.login-card {
  width: min(460px, 100%);
  justify-self: center;
  align-self: center;
}

.sidebar {
  display: flex;
  flex-direction: column;
  gap: 18px;
  align-self: start;
  position: sticky;
  top: 20px;
  min-height: calc(100vh - 40px);
  padding: 16px 14px 14px;
  border-radius: 34px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.48), rgba(255, 255, 255, 0.26)),
    linear-gradient(155deg, rgba(255, 255, 255, 0.18), rgba(186, 214, 239, 0.08));
  border: 1px solid rgba(255, 255, 255, 0.56);
  box-shadow:
    0 24px 56px rgba(45, 67, 92, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.74);
}

.sidebar-top,
.menu-list,
.stack,
.chip-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.menu-list {
  margin: 4px 0 0;
  gap: 10px;
}

.chip-list {
  flex-direction: row;
  flex-wrap: wrap;
}

.content-area {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

.mobile-header,
.mobile-tabbar,
.mobile-sheet-backdrop {
  display: none;
}

.mobile-header {
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.mobile-header-copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.mobile-header h2 {
  margin: 0;
  color: #23374c;
}

.mobile-tabbar,
.mobile-sheet-grid,
.mobile-sheet-card,
.mobile-tab-copy {
  display: flex;
}

.mobile-tabbar {
  align-items: stretch;
}

.mobile-sheet-grid {
  flex-wrap: wrap;
}

.mobile-tab-copy {
  flex-direction: column;
}

.mobile-tabbar button {
  text-align: center;
}

.mobile-sheet-card {
  text-align: left;
}

.mobile-tab-icon {
  width: 30px;
  height: 30px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  background: rgba(32, 48, 68, 0.08);
  color: #23374c;
}

.mobile-tab-label {
  font-size: 12px;
  line-height: 1.2;
}

.mobile-tab-copy {
  gap: 2px;
  text-align: left;
}

.mobile-tab-copy strong,
.mobile-tab-copy small {
  display: block;
}

.mobile-tab-copy small {
  color: #6a7683;
}

.mobile-sheet-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.mobile-sheet-card {
  flex-direction: row;
  gap: 10px;
  align-items: flex-start;
  padding: 14px;
  border-radius: 18px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.72), rgba(255, 255, 255, 0.4)),
    linear-gradient(145deg, rgba(255, 255, 255, 0.14), rgba(170, 203, 234, 0.08));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.78),
    0 12px 24px rgba(45, 67, 92, 0.08);
}

.mobile-sheet-card.active {
  background: linear-gradient(160deg, #223040, #355775);
  color: #fff;
}

.mobile-sheet-card.active .mobile-tab-icon,
.mobile-tabbar button.active .mobile-tab-icon {
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
}

.mobile-sheet-card.active .mobile-tab-copy small {
  color: rgba(255, 255, 255, 0.72);
}

.mobile-logout {
  width: 100%;
  margin-top: 12px;
}

.sidebar-top {
  padding: 6px 10px 2px;
}

.sidebar-top h2 {
  margin: 0;
  font-size: 2rem;
  line-height: 1.05;
  color: #23374c;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 20px;
}

.two-column {
  display: grid;
  grid-template-columns: 1.02fr 0.98fr;
  gap: 20px;
}

.full-width {
  grid-column: 1 / -1;
}

.section-head {
  display: flex;
  justify-content: space-between;
  align-items: start;
  gap: 12px;
  margin-bottom: 12px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 14px;
}

.inline-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.button-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.button-grid.three {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.search-grid {
  display: grid;
  grid-template-columns: 1fr 180px 120px;
  gap: 12px;
  margin-bottom: 14px;
}

.search-grid.compact {
  grid-template-columns: 1fr 140px;
}

.list-item,
.result-block,
.provider-card,
.meta-card,
.chat-item {
  border-radius: 22px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.7), rgba(255, 255, 255, 0.36)),
    linear-gradient(145deg, rgba(255, 255, 255, 0.14), rgba(170, 203, 234, 0.06));
  border: 1px solid rgba(255, 255, 255, 0.54);
  padding: 16px;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.76),
    0 16px 32px rgba(45, 67, 92, 0.08);
}

.list-item {
  display: flex;
  justify-content: space-between;
  gap: 14px;
}

.item-main {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.item-main p,
.item-main small {
  margin: 0;
  line-height: 1.7;
}

.item-actions {
  display: flex;
  gap: 8px;
}

.item-actions.vertical {
  flex-direction: column;
}

.draft-meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.meta-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.result-panel {
  margin-top: 18px;
  padding: 18px;
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.7), rgba(255, 255, 255, 0.36));
  border: 1px solid rgba(255, 255, 255, 0.54);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.76),
    0 16px 32px rgba(45, 67, 92, 0.08);
}

.empty-box {
  border-radius: 20px;
  padding: 18px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.72), rgba(247, 250, 252, 0.46));
  color: #6a7683;
  line-height: 1.8;
  border: 1px solid rgba(255, 255, 255, 0.54);
}

.chat-box {
  max-height: 460px;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.chat-item.user {
  background:
    linear-gradient(180deg, rgba(225, 235, 246, 0.86), rgba(255, 255, 255, 0.34));
}

.chat-item.assistant {
  background:
    linear-gradient(180deg, rgba(238, 246, 255, 0.92), rgba(255, 255, 255, 0.42));
}

.chat-role {
  margin: 0 0 8px;
  font-size: 13px;
  color: #6a7683;
}

.chat-content,
.result-pre {
  margin: 0;
  white-space: pre-wrap;
  line-height: 1.8;
  font-family: "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
}

.chat-form {
  margin-top: 14px;
  display: grid;
  grid-template-columns: 1fr 120px;
  gap: 12px;
  align-items: end;
}

.provider-card.ready {
  background:
    linear-gradient(180deg, rgba(237, 249, 241, 0.96), rgba(255, 255, 255, 0.42));
}

.pager {
  margin-top: 16px;
  display: grid;
  grid-template-columns: 120px 1fr 120px;
  gap: 12px;
  align-items: center;
}

.chip,
.eyebrow {
  display: inline-flex;
}

.chip {
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.44);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    0 8px 16px rgba(45, 67, 92, 0.08);
}

.eyebrow {
  margin: 0;
  font-size: 12px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: #977247;
}

.muted {
  color: #6a7683;
}

.success-text {
  color: #1a7a48;
}

.error-text {
  color: #c23f3f;
}

.stat-card strong {
  font-size: 28px;
}

input,
textarea,
select,
button {
  border-radius: 18px;
  font-size: 14px;
  box-sizing: border-box;
  font: inherit;
}

input,
textarea,
select {
  width: 100%;
  border: 1px solid rgba(255, 255, 255, 0.56);
  padding: 14px 16px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.88), rgba(255, 255, 255, 0.58));
  color: #223040;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.84),
    0 10px 24px rgba(45, 67, 92, 0.06);
}

textarea {
  resize: vertical;
}

button {
  border: none;
  padding: 12px 18px;
  cursor: pointer;
  transition: transform 0.22s ease, opacity 0.22s ease, box-shadow 0.22s ease, background 0.22s ease;
}

button:hover {
  transform: translateY(-2px);
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.65;
  transform: none;
}

.primary {
  background: linear-gradient(160deg, #223040, #355775);
  color: #fff;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.22),
    0 12px 20px rgba(34, 48, 64, 0.18);
}

.ghost {
  background: rgba(255, 255, 255, 0.42);
  color: #223040;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.74),
    0 8px 16px rgba(45, 67, 92, 0.08);
}

.danger {
  background: rgba(194, 63, 63, 0.12);
  color: #c23f3f;
}

.large-button {
  width: 100%;
}

.sidebar > .ghost:last-child {
  margin-top: auto;
}

.menu-list button {
  justify-content: flex-start;
  padding: 13px 16px;
  border-radius: 20px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.64), rgba(255, 255, 255, 0.34));
  color: #223040;
  text-align: left;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    0 10px 18px rgba(45, 67, 92, 0.07);
}

.menu-list button.active {
  background: linear-gradient(160deg, #223040, #355775);
  color: #fff;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.24),
    0 12px 20px rgba(34, 48, 64, 0.16);
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .button-grid.three {
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 960px) {
  .page-shell,
  .login-layout,
  .two-column,
  .stats-grid,
  .inline-grid,
  .button-grid,
  .button-grid.three,
  .search-grid,
  .chat-form,
  .pager,
  .draft-meta {
    grid-template-columns: 1fr;
  }

  .page-shell {
    display: flex;
    flex-direction: column;
  }

  .sidebar {
    position: static;
    min-height: auto;
    top: auto;
  }

  .list-item {
    flex-direction: column;
  }

  .item-actions {
    flex-direction: row;
    flex-wrap: wrap;
  }
}

@media (max-width: 760px) {
  .page-shell {
    min-height: 100vh;
    gap: 12px;
    padding: 0 0 20px;
    perspective: none;
  }

  .login-layout {
    min-height: auto;
    gap: 12px;
    padding: 12px;
  }

  .panel-card,
  .hero-card {
    padding: 16px;
    border-radius: 20px;
    backdrop-filter: none;
    -webkit-backdrop-filter: none;
    transform-style: flat;
  }

  .panel-card:hover,
  .hero-card:hover,
  .list-item:hover,
  .provider-card:hover,
  .chat-item:hover,
  .stat-card:hover {
    transform: none;
  }

  .hero-card {
    padding: 22px 18px;
  }

  .sidebar {
    display: none;
  }

  .content-area {
    gap: 12px;
    padding: 0 12px calc(112px + env(safe-area-inset-bottom));
  }

  .mobile-header {
    display: flex;
  }

  .mobile-header .ghost {
    width: auto;
    min-width: 76px;
  }

  .mobile-tabbar {
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    gap: 8px;
    position: fixed;
    left: 12px;
    right: 12px;
    bottom: calc(12px + env(safe-area-inset-bottom));
    z-index: 40;
    padding: 10px;
    border-radius: 24px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(242, 247, 252, 0.88)),
      linear-gradient(145deg, rgba(255, 255, 255, 0.18), rgba(180, 208, 232, 0.1));
    border: 1px solid rgba(255, 255, 255, 0.82);
    box-shadow:
      0 20px 36px rgba(27, 43, 63, 0.18),
      inset 0 1px 0 rgba(255, 255, 255, 0.86);
  }

  .mobile-tabbar button {
    min-height: 60px;
    padding: 8px 6px;
    border-radius: 18px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    background: transparent;
    color: #526272;
    box-shadow: none;
  }

  .mobile-tabbar button:hover {
    transform: none;
  }

  .mobile-tabbar button.active {
    background: linear-gradient(160deg, #223040, #355775);
    color: #fff;
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.22),
      0 12px 20px rgba(34, 48, 64, 0.18);
  }

  .mobile-sheet-backdrop {
    display: block;
    position: fixed;
    inset: 0;
    z-index: 35;
    background: rgba(20, 31, 45, 0.24);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
  }

  .mobile-sheet {
    position: absolute;
    left: 12px;
    right: 12px;
    bottom: calc(96px + env(safe-area-inset-bottom));
    margin: 0;
    padding: 16px;
    border-radius: 24px;
    max-height: calc(100vh - 160px);
    overflow: auto;
  }

  .mobile-sheet-head .ghost {
    width: auto;
  }

  .section-head {
    flex-direction: column;
    align-items: stretch;
  }

  .section-head button,
  .section-head .primary,
  .section-head .ghost {
    width: 100%;
  }

  .stats-grid {
    gap: 12px;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .list-item,
  .provider-card,
  .meta-card,
  .result-block,
  .chat-item,
  .empty-box {
    padding: 14px;
    border-radius: 18px;
  }

  .item-actions {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    width: 100%;
  }

  .item-actions button {
    min-height: 42px;
  }

  .chat-box {
    max-height: 58vh;
  }

  .chat-form {
    gap: 10px;
  }

  .chat-content,
  .result-pre {
    overflow-x: auto;
    word-break: break-word;
  }

  input,
  textarea,
  select,
  button {
    min-height: 44px;
  }

  textarea {
    min-height: 110px;
  }
}

@media (max-width: 520px) {
  .panel-card,
  .hero-card {
    padding: 14px;
    border-radius: 16px;
  }

  .content-area {
    padding: 0 10px calc(104px + env(safe-area-inset-bottom));
  }

  .stats-grid {
    gap: 10px;
  }

  .button-grid,
  .button-grid.three,
  .item-actions {
    grid-template-columns: 1fr;
  }

  .mobile-tabbar {
    left: 10px;
    right: 10px;
    bottom: calc(10px + env(safe-area-inset-bottom));
    gap: 6px;
    padding: 8px;
    border-radius: 20px;
  }

  .mobile-tabbar button {
    min-height: 56px;
    padding: 8px 4px;
  }

  .mobile-tab-icon {
    width: 26px;
    height: 26px;
    font-size: 11px;
  }

  .mobile-tab-label {
    font-size: 11px;
  }

  .mobile-sheet {
    left: 10px;
    right: 10px;
    bottom: calc(88px + env(safe-area-inset-bottom));
    padding: 14px;
    border-radius: 20px;
  }

  .mobile-sheet-grid {
    grid-template-columns: 1fr;
  }

  .pager {
    grid-template-columns: 1fr auto 1fr;
    gap: 8px;
  }

  .stat-card strong {
    font-size: 24px;
  }
}
</style>

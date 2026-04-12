<script setup>
import { computed, onMounted, ref, watch } from 'vue'

const props = defineProps({
  token: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['notify'])

const apiBase = 'http://127.0.0.1:8080/api/admin'
const loading = ref(false)
const fetchLoading = ref(false)
const briefings = ref([])
const total = ref(0)
const fetchStatus = ref(null)

const query = ref({
  date: '',
  keyword: '',
  status: '',
  page: 1,
  page_size: 8
})

const form = ref({
  id: 0,
  briefing_date: '',
  title: '',
  summary: '',
  source_name: '',
  source_url: '',
  status: 1,
  sort_order: 1,
  source_published_at: ''
})

const fetchForm = ref({
  date: '',
  limit: 10
})

const pages = computed(() => Math.max(1, Math.ceil(total.value / query.value.page_size)))

function authHeaders(asJson = true) {
  const headers = { Authorization: `Bearer ${props.token}` }
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
  const response = await fetch(`${apiBase}${path}`, options)
  const result = await response.json()
  if (!response.ok || result.success === false) {
    throw new Error(result.error || '请求失败')
  }
  return result.data
}

function formatDateTime(value) {
  if (!value) return ''
  const date = new Date(value)
  const pad = (part) => String(part).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function resetForm() {
  form.value = {
    id: 0,
    briefing_date: fetchForm.value.date || new Date().toISOString().slice(0, 10),
    title: '',
    summary: '',
    source_name: '',
    source_url: '',
    status: 1,
    sort_order: 1,
    source_published_at: ''
  }
}

function editBriefing(item) {
  form.value = {
    id: item.id,
    briefing_date: item.briefing_date,
    title: item.title,
    summary: item.summary || '',
    source_name: item.source_name || '',
    source_url: item.source_url || '',
    status: item.status,
    sort_order: item.sort_order || 1,
    source_published_at: formatDateTime(item.source_published_at)
  }
}

async function loadBriefings() {
  loading.value = true
  try {
    const data = await request(`/daily-briefings?${buildQueryString(query.value)}`, {
      headers: authHeaders()
    })
    briefings.value = data.items || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

async function loadFetchStatus() {
  fetchStatus.value = await request('/daily-briefings/fetch-status', {
    headers: authHeaders()
  })
}

async function saveBriefing() {
  const payload = { ...form.value }
  await request(payload.id ? `/daily-briefings/${payload.id}` : '/daily-briefings', {
    method: payload.id ? 'PUT' : 'POST',
    headers: authHeaders(),
    body: JSON.stringify(payload)
  })
  emit('notify', '每日简讯已保存')
  resetForm()
  await loadBriefings()
}

async function deleteBriefing(item) {
  if (!window.confirm(`确定删除《${item.title}》吗？`)) return
  await request(`/daily-briefings/${item.id}`, {
    method: 'DELETE',
    headers: authHeaders()
  })
  emit('notify', '每日简讯已删除')
  await loadBriefings()
}

async function runFetch() {
  fetchLoading.value = true
  try {
    const result = await request('/daily-briefings/fetch', {
      method: 'POST',
      headers: authHeaders(),
      body: JSON.stringify(fetchForm.value)
    })
    emit('notify', result.message || '已完成简讯抓取')
    query.value.date = fetchForm.value.date || query.value.date
    await Promise.all([loadBriefings(), loadFetchStatus()])
  } finally {
    fetchLoading.value = false
  }
}

function turnPage(step) {
  const nextPage = query.value.page + step
  if (nextPage < 1 || nextPage > pages.value) return
  query.value.page = nextPage
  loadBriefings()
}

onMounted(async () => {
  fetchForm.value.date = new Date().toISOString().slice(0, 10)
  resetForm()
  await Promise.all([loadBriefings(), loadFetchStatus()])
})

watch(
  () => props.token,
  () => {
    if (props.token) {
      loadBriefings()
      loadFetchStatus()
    }
  }
)
</script>

<template>
  <section class="two-column">
    <article class="panel-card">
      <div class="section-head">
        <div>
          <h2>{{ form.id ? '编辑简讯条目' : '新增简讯条目' }}</h2>
          <p class="muted">自动抓取会生成发布状态的条目，你也可以手动补录或修订。</p>
        </div>
        <button class="ghost" @click="resetForm">清空</button>
      </div>

      <label class="field">
        <span>简讯日期</span>
        <input v-model="form.briefing_date" type="date" />
      </label>
      <label class="field">
        <span>标题</span>
        <input v-model="form.title" type="text" />
      </label>
      <label class="field">
        <span>摘要</span>
        <textarea v-model="form.summary" rows="5"></textarea>
      </label>
      <div class="inline-grid">
        <label class="field">
          <span>来源名称</span>
          <input v-model="form.source_name" type="text" />
        </label>
        <label class="field">
          <span>排序</span>
          <input v-model.number="form.sort_order" type="number" min="1" />
        </label>
      </div>
      <label class="field">
        <span>原文链接</span>
        <input v-model="form.source_url" type="url" />
      </label>
      <div class="inline-grid">
        <label class="field">
          <span>来源发布时间</span>
          <input v-model="form.source_published_at" type="datetime-local" />
        </label>
        <label class="field">
          <span>状态</span>
          <select v-model="form.status">
            <option :value="0">草稿</option>
            <option :value="1">发布</option>
          </select>
        </label>
      </div>
      <button class="primary" @click="saveBriefing">保存条目</button>
    </article>

    <article class="panel-card">
      <div class="section-head">
        <div>
          <h2>自动抓取</h2>
          <p class="muted">会从公开 RSS 新闻源抓取全球 AI 时讯，并生成当天 10 条内简讯。</p>
        </div>
        <button class="primary" :disabled="fetchLoading" @click="runFetch">
          {{ fetchLoading ? '抓取中...' : '立即抓取' }}
        </button>
      </div>

      <div class="inline-grid">
        <label class="field">
          <span>抓取日期</span>
          <input v-model="fetchForm.date" type="date" />
        </label>
        <label class="field">
          <span>抓取条数</span>
          <input v-model.number="fetchForm.limit" type="number" min="1" max="20" />
        </label>
      </div>

      <div class="status-card">
        <p><strong>运行状态：</strong>{{ fetchStatus?.is_running ? '进行中' : '空闲' }}</p>
        <p><strong>默认脚本：</strong>{{ fetchStatus?.script_command || 'go run ./cmd/dailybriefing_fetcher' }}</p>
        <p><strong>最近结果：</strong>{{ fetchStatus?.last_result?.message || '暂无' }}</p>
      </div>

      <div class="script-card">
        <p class="muted">你也可以在项目根目录直接运行脚本：</p>
        <pre class="code">.\scripts\fetch-daily-briefings.ps1 -date {{ fetchForm.date }} -limit {{ fetchForm.limit }}</pre>
      </div>
    </article>

    <article class="panel-card full-width">
      <div class="section-head">
        <div>
          <h2>简讯列表</h2>
          <p class="muted">支持按日期、关键词和发布状态筛选。</p>
        </div>
        <span class="muted">共 {{ total }} 条</span>
      </div>

      <div class="search-grid">
        <input v-model="query.keyword" type="text" placeholder="搜索标题、摘要或来源" />
        <input v-model="query.date" type="date" />
        <select v-model="query.status">
          <option value="">全部状态</option>
          <option value="0">草稿</option>
          <option value="1">发布</option>
        </select>
        <button class="primary" @click="query.page = 1; loadBriefings()">搜索</button>
      </div>

      <div v-if="loading" class="empty-box">正在加载简讯列表...</div>
      <div v-else-if="briefings.length === 0" class="empty-box">还没有简讯数据，可以先抓取一次。</div>
      <div v-else class="stack">
        <div v-for="item in briefings" :key="item.id" class="list-item">
          <div class="item-main">
            <strong>{{ item.sort_order }}. {{ item.title }}</strong>
            <p>{{ item.summary || '暂无摘要' }}</p>
            <small>
              {{ item.briefing_date }} / {{ item.source_name || '未知来源' }} / {{ item.source_type || 'manual' }} /
              {{ item.status === 1 ? '已发布' : '草稿' }}
            </small>
          </div>
          <div class="item-actions vertical">
            <button class="ghost" @click="editBriefing(item)">编辑</button>
            <button class="ghost" @click="window.open(item.source_url, '_blank')">原文</button>
            <button class="danger" @click="deleteBriefing(item)">删除</button>
          </div>
        </div>
      </div>

      <div class="pager">
        <button class="ghost" @click="turnPage(-1)">上一页</button>
        <span>{{ query.page }} / {{ pages }}</span>
        <button class="ghost" @click="turnPage(1)">下一页</button>
      </div>
    </article>
  </section>
</template>

<style scoped>
.two-column {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.full-width {
  grid-column: 1 / -1;
}

.section-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: start;
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

.search-grid {
  display: grid;
  grid-template-columns: 1.3fr 180px 140px 120px;
  gap: 12px;
  margin-bottom: 14px;
}

.status-card,
.script-card {
  border-radius: 22px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.7), rgba(255, 255, 255, 0.36)),
    linear-gradient(145deg, rgba(255, 255, 255, 0.14), rgba(170, 203, 234, 0.06));
  border: 1px solid rgba(255, 255, 255, 0.54);
  padding: 16px;
  line-height: 1.8;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.76),
    0 16px 32px rgba(45, 67, 92, 0.08);
}

.code {
  margin: 0;
  white-space: pre-wrap;
}

.stack {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list-item {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  padding: 16px;
  border-radius: 22px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.7), rgba(255, 255, 255, 0.36)),
    linear-gradient(145deg, rgba(255, 255, 255, 0.14), rgba(170, 203, 234, 0.06));
  border: 1px solid rgba(255, 255, 255, 0.54);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.76),
    0 16px 32px rgba(45, 67, 92, 0.08);
  transition: transform 0.24s ease, box-shadow 0.24s ease;
}

.list-item:hover {
  transform: translateY(-4px) rotateX(3deg);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.76),
    0 22px 38px rgba(45, 67, 92, 0.14);
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

.vertical {
  flex-direction: column;
}

.pager {
  margin-top: 16px;
  display: grid;
  grid-template-columns: 120px 1fr 120px;
  gap: 12px;
  align-items: center;
}

.empty-box,
.muted {
  color: #6a7683;
}

input,
textarea,
select,
button {
  border-radius: 18px;
  font-size: 14px;
  box-sizing: border-box;
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
  transition: transform 0.22s ease, box-shadow 0.22s ease;
}

button:hover {
  transform: translateY(-2px);
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

@media (max-width: 960px) {
  .two-column,
  .inline-grid,
  .search-grid,
  .pager {
    grid-template-columns: 1fr;
  }

  .list-item {
    flex-direction: column;
  }

  .item-actions {
    flex-direction: row;
    flex-wrap: wrap;
  }
}
</style>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import DailyBriefingPanel from '../components/DailyBriefingPanel.vue'

const apiBase = 'http://127.0.0.1:8080/api/public'
const router = useRouter()

const articles = ref([])
const categories = ref([])
const tags = ref([])
const archives = ref([])
const searchKeyword = ref('')
const activeCategory = ref('')
const activeTag = ref('')
const siteQuestion = ref('')
const siteAnswer = ref('')
const loading = ref(false)

const filteredArticles = computed(() => {
  return articles.value.filter((article) => {
    const matchCategory = !activeCategory.value || article.category?.name === activeCategory.value
    const matchTag = !activeTag.value || (article.tags || []).some((tag) => tag.name === activeTag.value)
    return matchCategory && matchTag
  })
})

const activeFilterLabel = computed(() => {
  if (activeCategory.value) return `分类：${activeCategory.value}`
  if (activeTag.value) return `标签：${activeTag.value}`
  return ''
})

const topFocusArticles = computed(() => filteredArticles.value.slice(0, 4))
const isFilterMode = computed(() => Boolean(activeCategory.value || activeTag.value))

async function request(path, options = {}) {
  const response = await fetch(`${apiBase}${path}`, options)
  const result = await response.json()
  if (!response.ok || result.success === false) {
    throw new Error(result.error || '请求失败')
  }
  return result.data
}

async function loadSidebarData() {
  const [categoryData, tagData, archiveData] = await Promise.all([
    request('/categories'),
    request('/tags'),
    request('/archives')
  ])

  categories.value = categoryData
  tags.value = tagData
  archives.value = archiveData
}

async function loadArticles() {
  loading.value = true
  try {
    const keyword = searchKeyword.value.trim()
    articles.value = await request(`/articles?keyword=${encodeURIComponent(keyword)}`)
  } finally {
    loading.value = false
  }
}

async function askSiteQuestion() {
  if (!siteQuestion.value.trim()) return

  siteAnswer.value = 'AI 正在整理全站内容...'
  const data = await request('/qa/site', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      question: siteQuestion.value
    })
  })
  siteAnswer.value = data.answer || '暂时没有拿到回答'
}

function openArticle(articleId) {
  if (!articleId) return
  router.push(`/articles/${articleId}`)
}

function chooseCategory(categoryName) {
  activeCategory.value = categoryName
  activeTag.value = ''
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function chooseTag(tagName) {
  activeTag.value = tagName
  activeCategory.value = ''
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function clearFilters() {
  activeCategory.value = ''
  activeTag.value = ''
}

onMounted(async () => {
  await Promise.all([loadSidebarData(), loadArticles()])
})
</script>

<template>
  <main class="page">
    <div class="ambient-field" aria-hidden="true">
      <span class="orb orb-a"></span>
      <span class="orb orb-b"></span>
      <span class="orb orb-c"></span>
      <span class="orb orb-d"></span>
    </div>

    <section class="hero glass hero-stable">
      <div>
        <p class="eyebrow">Reader Experience</p>
        <h1>AI 智能博客读者端</h1>
        <p class="intro">
          这里除了文章列表，现在还新增了每日简讯和顶部快速内容切换。点分类或标签后，顶部会立即切成对应内容，不用再往下翻。
        </p>
      </div>

      <div class="hero-actions">
        <input v-model="searchKeyword" type="text" placeholder="搜索文章标题或正文" />
        <button @click="loadArticles">搜索文章</button>
        <button class="ghost-button" @click="clearFilters">清空筛选</button>
      </div>
    </section>

    <section class="layout">
      <aside class="sidebar">
        <article class="glass panel">
          <p class="eyebrow">About</p>
          <h2>博客简介</h2>
          <p class="intro-small">
            这个读者端包含文章、分类、标签、全站问答和每日简讯。现在点击侧边栏后，主内容区会优先展示对应结果。
          </p>
        </article>

        <article class="glass panel">
          <p class="eyebrow">Categories</p>
          <h2>分类</h2>
          <div class="chip-list">
            <button
              v-for="category in categories"
              :key="category.id"
              class="chip-button"
              :class="{ selected: activeCategory === category.name }"
              @click="chooseCategory(category.name)"
            >
              {{ category.name }}
            </button>
          </div>
        </article>

        <article class="glass panel">
          <p class="eyebrow">Tags</p>
          <h2>标签</h2>
          <div class="chip-list">
            <button
              v-for="tag in tags"
              :key="tag.id"
              class="chip-button"
              :class="{ selected: activeTag === tag.name }"
              @click="chooseTag(tag.name)"
            >
              {{ tag.name }}
            </button>
          </div>
        </article>

        <article class="glass panel">
          <p class="eyebrow">Archive</p>
          <h2>归档</h2>
          <div class="archive-list">
            <span v-for="archive in archives" :key="archive.label">
              {{ archive.label }} · {{ archive.count }} 篇
            </span>
          </div>
        </article>

        <article class="glass panel">
          <p class="eyebrow">Site QA</p>
          <h2>全站知识问答</h2>
          <textarea v-model="siteQuestion" rows="4" placeholder="例如：这个博客里有哪些 Go 入门内容？"></textarea>
          <button @click="askSiteQuestion">提问全站 AI</button>
          <pre class="result">{{ siteAnswer }}</pre>
        </article>
      </aside>

      <section class="content">
        <article v-if="isFilterMode" class="glass panel focus-panel">
          <div class="section-head">
            <div>
              <p class="eyebrow">Focused Content</p>
              <h2>{{ activeFilterLabel }}</h2>
              <p class="hint">已为你切换到对应内容预览，点击卡片可直接进入详情页。</p>
            </div>
            <span class="focus-count">共 {{ filteredArticles.length }} 篇相关文章</span>
          </div>

          <div v-if="topFocusArticles.length === 0" class="result">
            当前筛选下还没有匹配文章，可以试试别的分类或标签。
          </div>

          <div v-else class="card-list focus-list">
            <article
              v-for="article in topFocusArticles"
              :key="article.id"
              class="card article-card focus-card"
              role="button"
              tabindex="0"
              @click="openArticle(article.id)"
              @keyup.enter="openArticle(article.id)"
            >
              <div class="card-top">
                <div>
                  <p class="tagline">
                    {{ article.category?.name || '未分类' }}
                    ·
                    {{ (article.tags || []).map((tag) => tag.name).join(' / ') || '无标签' }}
                  </p>
                  <h3>{{ article.title }}</h3>
                </div>
                <button class="inline-action" @click.stop="openArticle(article.id)">立即阅读</button>
              </div>
              <p>{{ article.summary || '暂无摘要' }}</p>
              <small>阅读量：{{ article.view_count }}</small>
            </article>
          </div>
        </article>

        <DailyBriefingPanel v-else compact />

        <article class="glass panel">
          <div class="section-head">
            <div>
              <p class="eyebrow">Article List</p>
              <h2>文章列表</h2>
            </div>
            <p class="hint">
              {{ loading ? '正在加载文章...' : `当前共 ${filteredArticles.length} 篇文章` }}
            </p>
          </div>

          <div class="card-list">
            <article
              v-for="article in filteredArticles"
              :key="article.id"
              class="card article-card"
              role="button"
              tabindex="0"
              @click="openArticle(article.id)"
              @keyup.enter="openArticle(article.id)"
            >
              <div class="card-top">
                <div>
                  <p class="tagline">
                    {{ article.category?.name || '未分类' }}
                    ·
                    {{ (article.tags || []).map((tag) => tag.name).join(' / ') || '无标签' }}
                  </p>
                  <h3>{{ article.title }}</h3>
                </div>
                <button class="inline-action" @click.stop="openArticle(article.id)">进入详情页</button>
              </div>
              <p>{{ article.summary }}</p>
              <small>阅读量：{{ article.view_count }}</small>
            </article>
          </div>
        </article>
      </section>
    </section>
  </main>
</template>

<style scoped>
.page {
  position: relative;
  max-width: 1320px;
  margin: 0 auto;
  padding: 28px 20px 88px;
  isolation: isolate;
}

.glass {
  position: relative;
  overflow: hidden;
  isolation: isolate;
  backdrop-filter: blur(28px) saturate(150%);
  -webkit-backdrop-filter: blur(28px) saturate(150%);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.52), rgba(255, 255, 255, 0.24)),
    linear-gradient(145deg, rgba(255, 255, 255, 0.3), rgba(175, 210, 240, 0.08));
  border: 1px solid rgba(255, 255, 255, 0.56);
  box-shadow:
    0 30px 70px rgba(32, 64, 93, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.7),
    inset 0 -18px 40px rgba(98, 144, 189, 0.08);
  border-radius: 30px;
  transform-style: preserve-3d;
}

.glass::before,
.glass::after {
  content: "";
  position: absolute;
  pointer-events: none;
}

.glass > * {
  position: relative;
  z-index: 1;
}

.glass::before {
  inset: 0;
  border-radius: inherit;
  background:
    radial-gradient(circle at 16% 0%, rgba(255, 255, 255, 0.88), transparent 34%),
    radial-gradient(circle at 84% 8%, rgba(255, 255, 255, 0.52), transparent 30%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.66) 0%, rgba(255, 255, 255, 0.16) 26%, rgba(255, 255, 255, 0.06) 62%, rgba(255, 255, 255, 0.14) 100%);
  opacity: 0.9;
  animation: liquid-highlight 10s ease-in-out infinite alternate;
  z-index: 0;
}

.glass::after {
  width: 180px;
  height: 180px;
  right: -42px;
  top: -62px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.52), transparent 68%);
  opacity: 0.78;
  animation: liquid-glow 14s ease-in-out infinite;
  z-index: 0;
}

.ambient-field {
  position: absolute;
  inset: -40px -100px auto;
  height: 520px;
  pointer-events: none;
  z-index: -1;
  overflow: hidden;
}

.orb {
  position: absolute;
  display: block;
  border-radius: 999px;
  filter: blur(10px);
  opacity: 0.72;
  transform: translate3d(0, 0, 0);
}

.orb-a {
  width: 240px;
  height: 240px;
  left: -40px;
  top: 26px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.95), rgba(149, 204, 255, 0.32) 48%, transparent 72%);
  animation: orb-float-a 15s ease-in-out infinite;
}

.orb-b {
  width: 320px;
  height: 320px;
  right: 2%;
  top: 12px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.88), rgba(255, 220, 178, 0.32) 48%, transparent 74%);
  animation: orb-float-b 18s ease-in-out infinite;
}

.orb-c {
  width: 180px;
  height: 180px;
  left: 42%;
  top: 120px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.72), rgba(171, 214, 249, 0.24) 54%, transparent 74%);
  animation: orb-float-c 13s ease-in-out infinite;
}

.orb-d {
  width: 260px;
  height: 260px;
  right: 22%;
  top: 250px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.7), rgba(186, 227, 255, 0.2) 52%, transparent 72%);
  animation: orb-float-d 17s ease-in-out infinite;
}

.hero {
  padding: 24px 26px;
  display: grid;
  grid-template-columns: minmax(0, 1.05fr) minmax(340px, 0.95fr);
  gap: 18px;
  margin-bottom: 20px;
  transform: translateZ(0);
  align-items: start;
  min-height: 264px;
}

.hero-stable {
  box-shadow:
    0 36px 82px rgba(32, 64, 93, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.74),
    inset 0 -22px 46px rgba(98, 144, 189, 0.08);
}

.hero-stable::before {
  background:
    radial-gradient(circle at 10% 0%, rgba(255, 255, 255, 0.94), transparent 32%),
    radial-gradient(circle at 82% 10%, rgba(255, 255, 255, 0.48), transparent 26%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.86) 0%, rgba(255, 255, 255, 0.18) 34%, rgba(255, 255, 255, 0.08) 70%, rgba(255, 255, 255, 0.16) 100%);
}

.hero-stable::after {
  width: 260px;
  height: 260px;
  right: -30px;
  top: -90px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.62), transparent 66%);
}

.hero-actions,
.archive-list {
  display: grid;
  gap: 12px;
}

.hero-actions {
  align-content: start;
  justify-items: stretch;
  padding-top: 4px;
}

.layout {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 20px;
}

.sidebar,
.content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.panel {
  padding: 24px;
  transition: transform 0.35s ease, box-shadow 0.35s ease, border-color 0.35s ease;
}

.panel:hover {
  transform: translateY(-6px) rotateX(4deg) rotateY(-3deg);
  box-shadow:
    0 34px 82px rgba(32, 64, 93, 0.22),
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    inset 0 -20px 42px rgba(98, 144, 189, 0.1);
  border-color: rgba(255, 255, 255, 0.74);
}

.section-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: center;
  margin-bottom: 16px;
}

.eyebrow {
  text-transform: uppercase;
  letter-spacing: 0.24em;
  font-size: 12px;
  color: #51779a;
}

.intro,
.intro-small,
.hint,
.result {
  line-height: 1.8;
}

.intro {
  max-width: 760px;
  margin: 0;
}

.intro-small,
.hint,
.archive-list,
.focus-count {
  color: var(--reader-muted);
}

.card-list {
  display: grid;
  gap: 14px;
}

.focus-panel {
  display: grid;
  gap: 12px;
}

.focus-list {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.card {
  padding: 18px;
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.7), rgba(255, 255, 255, 0.36)),
    linear-gradient(150deg, rgba(255, 255, 255, 0.16), rgba(149, 193, 231, 0.08));
  border: 1px solid rgba(255, 255, 255, 0.54);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.76),
    0 18px 34px rgba(35, 67, 97, 0.08);
}

.article-card {
  transition: transform 0.28s ease, box-shadow 0.28s ease, border-color 0.28s ease;
  transform-style: preserve-3d;
  cursor: pointer;
  outline: none;
}

.article-card:hover,
.article-card:focus-visible {
  transform: translateY(-8px) rotateX(5deg) rotateY(-4deg);
  box-shadow:
    0 26px 46px rgba(35, 67, 97, 0.15),
    inset 0 1px 0 rgba(255, 255, 255, 0.78);
  border-color: rgba(255, 255, 255, 0.72);
}

.card-top {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.chip-button,
button {
  border: none;
  border-radius: 999px;
  padding: 11px 16px;
  background: linear-gradient(160deg, #173452, #244f73);
  color: white;
  cursor: pointer;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.24),
    0 10px 18px rgba(18, 45, 68, 0.18);
  transition: transform 0.24s ease, box-shadow 0.24s ease, background 0.24s ease;
}

button:hover {
  transform: translateY(-2px) translateZ(10px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.26),
    0 16px 24px rgba(18, 45, 68, 0.22);
}

.hero-actions button:hover,
.hero-actions .ghost-button:hover {
  transform: none;
}

.chip-button {
  background: rgba(255, 255, 255, 0.44);
  color: #183149;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.66),
    0 8px 18px rgba(35, 67, 97, 0.08);
}

.chip-button.selected {
  background: linear-gradient(160deg, #173452, #244f73);
  color: #fff;
}

.ghost-button {
  background: rgba(255, 255, 255, 0.42);
  color: #183149;
}

.hero-actions input,
.hero-actions button {
  min-height: 54px;
}

.hero-actions input {
  padding-inline: 18px;
}

.hero-actions button {
  width: 100%;
}

.inline-action {
  flex-shrink: 0;
}

input,
textarea {
  width: 100%;
  border: 1px solid rgba(255, 255, 255, 0.56);
  border-radius: 20px;
  padding: 14px 16px;
  box-sizing: border-box;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.78), rgba(255, 255, 255, 0.56));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.84),
    0 10px 24px rgba(35, 67, 97, 0.06);
}

.tagline {
  color: #4f7aa4;
  font-size: 13px;
}

.result {
  white-space: pre-wrap;
  min-height: 72px;
  padding: 18px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.46);
  border: 1px solid rgba(255, 255, 255, 0.5);
}

@keyframes liquid-highlight {
  0% {
    transform: translate3d(-4%, 0, 0) scaleX(0.98);
    opacity: 0.9;
  }
  50% {
    transform: translate3d(3%, 2%, 0) scaleX(1.02);
    opacity: 1;
  }
  100% {
    transform: translate3d(6%, 4%, 0) scaleX(1.04);
    opacity: 0.84;
  }
}

@keyframes liquid-glow {
  0% {
    transform: translate3d(0, 0, 0) scale(1);
    opacity: 0.62;
  }
  50% {
    transform: translate3d(-20px, 14px, 0) scale(1.08);
    opacity: 0.82;
  }
  100% {
    transform: translate3d(8px, 26px, 0) scale(0.96);
    opacity: 0.58;
  }
}

@keyframes orb-float-a {
  0%, 100% {
    transform: translate3d(0, 0, 0);
  }
  50% {
    transform: translate3d(36px, 20px, 0);
  }
}

@keyframes orb-float-b {
  0%, 100% {
    transform: translate3d(0, 0, 0);
  }
  50% {
    transform: translate3d(-28px, 24px, 0);
  }
}

@keyframes orb-float-c {
  0%, 100% {
    transform: translate3d(0, 0, 0);
  }
  50% {
    transform: translate3d(18px, -16px, 0);
  }
}

@keyframes orb-float-d {
  0%, 100% {
    transform: translate3d(0, 0, 0);
  }
  50% {
    transform: translate3d(-22px, 18px, 0);
  }
}

@media (max-width: 980px) {
  .hero,
  .layout,
  .card-top,
  .section-head,
  .focus-list {
    grid-template-columns: 1fr;
    display: grid;
  }

  .hero {
    min-height: auto;
  }

  .ambient-field {
    inset: -20px -60px auto;
    height: 420px;
  }
}
</style>

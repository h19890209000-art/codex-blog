<script setup>
import MarkdownIt from 'markdown-it'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const apiBase = 'http://127.0.0.1:8080/api/public'
const route = useRoute()
const router = useRouter()

// 这里创建一个 Markdown 渲染器。
// 首版先保持配置简单，方便你以后继续看懂和扩展。
const markdown = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: true
})

const article = ref(null)
const comments = ref([])
const allArticles = ref([])
const titleAnalysis = ref('')
const qaAnswer = ref('')
const loading = ref(false)
const articleQuestion = ref('')
const commentForm = ref({
  author: '',
  content: ''
})

const articleId = computed(() => Number(route.params.id))

// markdownContent 用来把 Markdown 正文转成可直接插入页面的 HTML。
const markdownContent = computed(() => {
  if (!article.value?.content) {
    return ''
  }

  return markdown.render(article.value.content)
})

// currentArticleIndex 用来找到当前文章在文章列表里的位置。
const currentArticleIndex = computed(() => {
  return allArticles.value.findIndex((item) => item.id === article.value?.id)
})

// prevArticle 表示上一篇文章。
const prevArticle = computed(() => {
  if (currentArticleIndex.value <= 0) {
    return null
  }

  return allArticles.value[currentArticleIndex.value - 1]
})

// nextArticle 表示下一篇文章。
const nextArticle = computed(() => {
  if (currentArticleIndex.value < 0 || currentArticleIndex.value >= allArticles.value.length - 1) {
    return null
  }

  return allArticles.value[currentArticleIndex.value + 1]
})

async function request(path, options = {}) {
  const response = await fetch(`${apiBase}${path}`, options)
  const result = await response.json()

  if (!response.ok || result.success === false) {
    throw new Error(result.error || '请求失败')
  }

  return result.data
}

async function loadArticleDetail() {
  if (!articleId.value) {
    return
  }

  loading.value = true

  try {
    // 这里多拉一次文章列表，是为了做“上一篇 / 下一篇”导航。
    const [articleData, commentData, articleList] = await Promise.all([
      request(`/articles/${articleId.value}`),
      request(`/articles/${articleId.value}/comments`),
      request('/articles')
    ])

    article.value = articleData
    comments.value = commentData
    allArticles.value = articleList || []
    titleAnalysis.value = ''
    qaAnswer.value = ''
    articleQuestion.value = ''
  } finally {
    loading.value = false
  }
}

async function analyzeTitle() {
  if (!article.value) {
    return
  }

  titleAnalysis.value = 'AI 正在思考这个标题的含义...'

  const data = await request('/ai/analyze-title', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      title: article.value.title,
      stream: false
    })
  })

  titleAnalysis.value = data.result || '暂时没有结果'
}

async function askArticleQuestion() {
  if (!article.value || !articleQuestion.value.trim()) {
    return
  }

  qaAnswer.value = 'AI 正在基于当前文章回答...'

  const data = await request(`/qa/article/${article.value.id}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      question: articleQuestion.value
    })
  })

  qaAnswer.value = data.answer || '没有拿到回答'
}

async function submitComment() {
  if (!article.value || !commentForm.value.content.trim()) {
    return
  }

  await request(`/articles/${article.value.id}/comments`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(commentForm.value)
  })

  commentForm.value = {
    author: '',
    content: ''
  }

  await loadArticleDetail()
}

function goBackHome() {
  router.push('/')
}

function openSiblingArticle(targetArticle) {
  if (!targetArticle?.id) {
    return
  }

  router.push(`/articles/${targetArticle.id}`)
}

onMounted(loadArticleDetail)
watch(() => route.params.id, loadArticleDetail)
</script>

<template>
  <main class="page">
    <section v-if="loading" class="detail-shell">
      <article class="glass skeleton hero-skeleton"></article>
      <article class="glass skeleton content-skeleton"></article>
      <section class="double-panel">
        <article class="glass skeleton panel-skeleton"></article>
        <article class="glass skeleton panel-skeleton"></article>
      </section>
      <article class="glass skeleton comment-skeleton"></article>
    </section>

    <section v-else class="detail-shell">
      <article
        class="hero-cover glass reveal delay-1"
        :style="article?.cover_url ? { backgroundImage: `linear-gradient(135deg, rgba(15,32,48,.72), rgba(46,79,106,.46)), url(${article.cover_url})` } : {}"
      >
        <button class="ghost-button back-button" @click="goBackHome">返回首页</button>
        <p class="eyebrow">Article Detail</p>
        <h1>{{ article?.title }}</h1>
        <p class="tagline">
          分类：{{ article?.category?.name || '未分类' }}
          · 标签：{{ (article?.tags || []).map((tag) => tag.name).join(' / ') || '暂无标签' }}
          · 阅读量：{{ article?.view_count || 0 }}
        </p>
        <p class="article-summary">{{ article?.summary }}</p>
      </article>

      <article v-if="article" class="glass panel reveal delay-2">
        <div class="section-head">
          <div>
            <p class="eyebrow">Markdown Article</p>
            <h2>正文内容</h2>
          </div>
          <button @click="analyzeTitle">AI 解析标题</button>
        </div>

        <div class="markdown-body" v-html="markdownContent"></div>
      </article>

      <section v-if="article" class="double-panel">
        <article class="glass panel reveal delay-3">
          <h3>标题 AI 解析</h3>
          <pre class="result">{{ titleAnalysis }}</pre>
        </article>

        <article class="glass panel reveal delay-4">
          <h3>文章智能问答</h3>
          <textarea v-model="articleQuestion" rows="4" placeholder="例如：这篇文章主要讲了什么？"></textarea>
          <button @click="askArticleQuestion">提问当前文章</button>
          <pre class="result">{{ qaAnswer }}</pre>
        </article>
      </section>

      <section v-if="article" class="double-panel nav-panel reveal delay-5">
        <article class="glass sibling-card" :class="{ disabled: !prevArticle }" @click="openSiblingArticle(prevArticle)">
          <p class="eyebrow">Previous</p>
          <h3>{{ prevArticle?.title || '已经是第一篇了' }}</h3>
          <p>{{ prevArticle?.summary || '当前文章前面没有更多文章。' }}</p>
        </article>

        <article class="glass sibling-card" :class="{ disabled: !nextArticle }" @click="openSiblingArticle(nextArticle)">
          <p class="eyebrow">Next</p>
          <h3>{{ nextArticle?.title || '已经是最后一篇了' }}</h3>
          <p>{{ nextArticle?.summary || '当前文章后面没有更多文章。' }}</p>
        </article>
      </section>

      <article v-if="article" class="glass panel reveal delay-6">
        <h3>评论区</h3>
        <div class="comment-form">
          <input v-model="commentForm.author" type="text" placeholder="你的昵称，可留空" />
          <textarea v-model="commentForm.content" rows="3" placeholder="写下你的评论"></textarea>
          <button @click="submitComment">提交评论</button>
        </div>

        <div class="comment-list">
          <article v-for="comment in comments" :key="comment.id" class="comment-card">
            <strong>{{ comment.author }}</strong>
            <p>{{ comment.content }}</p>
            <small>{{ comment.created_at }}</small>
          </article>
        </div>
      </article>
    </section>
  </main>
</template>

<style scoped>
.page {
  max-width: 1100px;
  margin: 0 auto;
  padding: 36px 20px 92px;
}

.detail-shell {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.glass {
  position: relative;
  overflow: hidden;
  isolation: isolate;
  backdrop-filter: blur(28px) saturate(155%);
  -webkit-backdrop-filter: blur(28px) saturate(155%);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.56), rgba(255, 255, 255, 0.24)),
    linear-gradient(145deg, rgba(255, 255, 255, 0.22), rgba(173, 206, 240, 0.1));
  border: 1px solid rgba(255, 255, 255, 0.58);
  box-shadow:
    0 28px 70px rgba(32, 64, 93, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    inset 0 -18px 36px rgba(98, 144, 189, 0.08);
  border-radius: 30px;
  transform-style: preserve-3d;
}

.glass::before {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background:
    radial-gradient(circle at 15% 0%, rgba(255, 255, 255, 0.84), transparent 30%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.72), rgba(255, 255, 255, 0.16) 34%, rgba(255, 255, 255, 0.08) 100%);
  pointer-events: none;
  z-index: 0;
}

.glass > * {
  position: relative;
  z-index: 1;
}

.hero-cover {
  padding: 28px;
  color: #fff;
  background:
    linear-gradient(135deg, rgba(20, 39, 58, 0.86), rgba(61, 98, 130, 0.64)),
    radial-gradient(circle at top right, rgba(255, 218, 156, 0.34), transparent 28%);
  background-size: cover;
  background-position: center;
  transform: translateZ(0);
}

.panel {
  padding: 26px;
  transition: transform 0.34s ease, box-shadow 0.34s ease;
}

.panel:hover,
.hero-cover:hover {
  transform: translateY(-6px) rotateX(4deg) rotateY(-3deg);
  box-shadow:
    0 34px 82px rgba(32, 64, 93, 0.22),
    inset 0 1px 0 rgba(255, 255, 255, 0.76);
}

.back-button {
  margin-bottom: 14px;
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
  color: #7ca8cf;
}

.hero-cover .eyebrow {
  color: rgba(255, 255, 255, 0.78);
}

.tagline {
  color: rgba(255, 255, 255, 0.88);
  font-size: 13px;
  line-height: 1.8;
}

.article-summary,
.result,
.comment-card p,
.sibling-card p {
  line-height: 1.9;
}

.markdown-body {
  padding: 24px;
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.72), rgba(255, 255, 255, 0.38)),
    linear-gradient(145deg, rgba(255, 255, 255, 0.16), rgba(173, 206, 240, 0.08));
  border: 1px solid rgba(255, 255, 255, 0.54);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.76),
    0 16px 30px rgba(35, 67, 97, 0.08);
  line-height: 1.9;
  color: #1a3148;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  color: #16314b;
  line-height: 1.35;
}

.markdown-body :deep(h1) {
  font-size: 34px;
  margin: 0 0 20px;
}

.markdown-body :deep(h2) {
  font-size: 26px;
  margin: 30px 0 14px;
}

.markdown-body :deep(h3) {
  font-size: 20px;
  margin: 24px 0 10px;
}

.markdown-body :deep(p),
.markdown-body :deep(li) {
  font-size: 16px;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  padding-left: 22px;
}

.markdown-body :deep(code) {
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(24, 49, 73, 0.08);
}

.markdown-body :deep(pre) {
  overflow: auto;
  padding: 18px;
  border-radius: 20px;
  background: #183149;
  color: #fff;
}

.markdown-body :deep(pre code) {
  padding: 0;
  background: transparent;
}

.markdown-body :deep(img) {
  max-width: 100%;
  display: block;
  margin: 18px auto;
  border-radius: 22px;
  box-shadow: 0 18px 36px rgba(44, 88, 122, 0.16);
}

.double-panel {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.nav-panel {
  align-items: stretch;
}

.sibling-card {
  padding: 24px;
  cursor: pointer;
  transition: transform 0.28s ease, box-shadow 0.28s ease, opacity 0.25s ease;
}

.sibling-card:hover {
  transform: translateY(-8px) rotateX(5deg);
  box-shadow:
    0 26px 46px rgba(35, 67, 97, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.78);
}

.sibling-card.disabled {
  cursor: default;
  opacity: 0.72;
}

.sibling-card.disabled:hover {
  transform: none;
  box-shadow: 0 20px 56px rgba(44, 88, 122, 0.14);
}

.comment-form,
.comment-list {
  display: grid;
  gap: 12px;
}

.comment-card {
  padding: 18px;
  border-radius: 20px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.7), rgba(255, 255, 255, 0.36));
  border: 1px solid rgba(255, 255, 255, 0.54);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.74),
    0 16px 30px rgba(35, 67, 97, 0.08);
}

.comment-card small {
  color: #5d7388;
}

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
  transition: transform 0.24s ease, box-shadow 0.24s ease;
}

button:hover {
  transform: translateY(-2px) translateZ(10px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.28),
    0 16px 24px rgba(18, 45, 68, 0.22);
}

.ghost-button {
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
  backdrop-filter: blur(12px);
}

input,
textarea {
  width: 100%;
  border: 1px solid rgba(255, 255, 255, 0.58);
  border-radius: 20px;
  padding: 14px 16px;
  box-sizing: border-box;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.8), rgba(255, 255, 255, 0.56));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.84),
    0 10px 24px rgba(35, 67, 97, 0.06);
}

.result {
  min-height: 72px;
  white-space: pre-wrap;
  padding: 18px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.44);
  border: 1px solid rgba(255, 255, 255, 0.52);
}

.skeleton {
  position: relative;
  overflow: hidden;
}

.skeleton::after {
  content: "";
  position: absolute;
  inset: 0;
  transform: translateX(-100%);
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.56), transparent);
  animation: shimmer 1.4s infinite;
}

.hero-skeleton {
  height: 260px;
}

.content-skeleton {
  height: 420px;
}

.panel-skeleton {
  height: 220px;
}

.comment-skeleton {
  height: 320px;
}

.reveal {
  opacity: 0;
  transform: translateY(20px);
  animation: rise-in 0.65s ease forwards;
}

.delay-1 { animation-delay: 0.06s; }
.delay-2 { animation-delay: 0.14s; }
.delay-3 { animation-delay: 0.22s; }
.delay-4 { animation-delay: 0.3s; }
.delay-5 { animation-delay: 0.38s; }
.delay-6 { animation-delay: 0.46s; }

@keyframes rise-in {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes shimmer {
  to {
    transform: translateX(100%);
  }
}

@media (max-width: 980px) {
  .double-panel,
  .section-head {
    display: grid;
    grid-template-columns: 1fr;
  }
}
</style>

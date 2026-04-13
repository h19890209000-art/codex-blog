<script setup>
import MarkdownIt from 'markdown-it'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { publicApiUrl } from '../lib/api'

const route = useRoute()
const router = useRouter()

const markdown = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: true
})

const article = ref(null)
const comments = ref([])
const prevArticle = ref(null)
const nextArticle = ref(null)
const titleAnalysis = ref('')
const qaAnswer = ref('')
const loading = ref(false)
const shareNotice = ref('')
const articleQuestion = ref('')
const commentForm = ref({
  author: '',
  content: ''
})

const articleId = computed(() => Number(route.params.id))

const markdownContent = computed(() => {
  if (!article.value?.content) return ''
  return markdown.render(article.value.content)
})

async function request(path, options = {}) {
  const response = await fetch(publicApiUrl(path), options)
  const result = await response.json()

  if (!response.ok || result.success === false) {
    throw new Error(result.error || '请求失败')
  }

  return result.data
}

async function loadArticleDetail() {
  if (!articleId.value) return

  loading.value = true

  try {
    const [articleData, commentData, navigationData] = await Promise.all([
      request(`/articles/${articleId.value}`),
      request(`/articles/${articleId.value}/comments`),
      request(`/articles/${articleId.value}/navigation`)
    ])

    article.value = articleData
    comments.value = commentData
    prevArticle.value = navigationData?.prev_article?.id ? navigationData.prev_article : null
    nextArticle.value = navigationData?.next_article?.id ? navigationData.next_article : null
    titleAnalysis.value = ''
    qaAnswer.value = ''
    articleQuestion.value = ''
  } finally {
    loading.value = false
  }
}

async function analyzeTitle() {
  if (!article.value) return

  titleAnalysis.value = 'AI 正在分析标题...'

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
  if (!article.value || !articleQuestion.value.trim()) return

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
  if (!article.value || !commentForm.value.content.trim()) return

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
  if (!targetArticle?.id) return
  router.push(`/articles/${targetArticle.id}`)
}

function buildShareLink(targetArticle = article.value) {
  if (!targetArticle?.id) return ''
  const resolved = router.resolve(`/articles/${targetArticle.id}`)
  return new URL(resolved.href, window.location.origin).toString()
}

function showShareNotice(text) {
  shareNotice.value = text
  window.clearTimeout(showShareNotice.timer)
  showShareNotice.timer = window.setTimeout(() => {
    if (shareNotice.value === text) {
      shareNotice.value = ''
    }
  }, 2200)
}

async function shareCurrentArticle() {
  if (!article.value?.id) return

  const url = buildShareLink()
  const title = article.value.title || '文章分享'

  try {
    if (navigator.share) {
      await navigator.share({
        title,
        text: article.value.summary || title,
        url
      })
      showShareNotice('已调起系统分享')
      return
    }

    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(url)
      showShareNotice('分享链接已复制')
      return
    }

    showShareNotice(url)
  } catch (error) {
    if (error?.name === 'AbortError') return
    showShareNotice('分享失败，请重试')
  }
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
        <div class="hero-actions">
          <button class="action-button action-button--soft" @click="goBackHome">返回首页</button>
          <button class="action-button action-button--soft" @click="shareCurrentArticle">分享链接</button>
        </div>
        <p class="eyebrow">Article Detail</p>
        <h1>{{ article?.title }}</h1>
        <p class="tagline">
          分类：{{ article?.category?.name || '未分类' }}
          /
          标签：{{ (article?.tags || []).map((tag) => tag.name).join(' / ') || '暂无标签' }}
          /
          阅读量：{{ article?.view_count || 0 }}
        </p>
        <p class="article-summary">{{ article?.summary }}</p>
        <p v-if="shareNotice" class="share-notice">{{ shareNotice }}</p>
      </article>

      <article v-if="article" class="glass panel reveal delay-2">
        <div class="section-head">
          <div>
            <p class="eyebrow">Markdown Article</p>
            <h2>正文内容</h2>
          </div>
          <button class="action-button action-button--primary" @click="analyzeTitle">AI 解析标题</button>
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
          <button class="action-button action-button--primary" @click="askArticleQuestion">提问当前文章</button>
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
          <button class="action-button action-button--primary" @click="submitComment">提交评论</button>
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
  --button-blue-top: #355f82;
  --button-blue-bottom: #274a67;
  --button-blue-soft-top: rgba(53, 95, 130, 0.88);
  --button-blue-soft-bottom: rgba(39, 74, 103, 0.8);
  --button-border: rgba(255, 255, 255, 0.28);
  --button-shadow: 0 12px 24px rgba(34, 63, 90, 0.18);
  --button-shadow-hover: 0 18px 30px rgba(34, 63, 90, 0.24);
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
  backdrop-filter: none;
  -webkit-backdrop-filter: none;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(255, 255, 255, 0.8)),
    linear-gradient(145deg, rgba(255, 255, 255, 0.3), rgba(173, 206, 240, 0.12));
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

.hero-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 14px;
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

.tagline,
.article-summary,
.result,
.comment-card p,
.sibling-card p {
  line-height: 1.9;
}

.tagline {
  color: rgba(255, 255, 255, 0.88);
  font-size: 13px;
}

.share-notice {
  margin: 12px 0 0;
  color: rgba(255, 255, 255, 0.82);
  font-size: 13px;
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

.action-button {
  position: relative;
  overflow: hidden;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 118px;
  min-height: 46px;
  padding: 0 18px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 999px;
  cursor: pointer;
  box-sizing: border-box;
  transition: transform 0.22s ease, box-shadow 0.22s ease, background 0.22s ease, color 0.22s ease, border-color 0.22s ease;
}

.action-button::before {
  content: "";
  position: absolute;
  inset: -18% auto -18% -30%;
  width: 42%;
  transform: translateX(-190%) skewX(-24deg);
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.86), transparent);
  opacity: 0;
}

.action-button::after {
  content: "";
  position: absolute;
  inset: 1px 1px auto 1px;
  height: 52%;
  border-radius: inherit;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.24), rgba(255, 255, 255, 0.04));
  pointer-events: none;
}

.action-button:hover {
  transform: translateY(-2px);
}

.action-button:hover::before {
  opacity: 1;
  animation: liquid-sweep 0.78s ease forwards;
}

.action-button--primary {
  background: linear-gradient(180deg, var(--button-blue-top), var(--button-blue-bottom));
  color: #f8fbff;
  border-color: var(--button-border);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.24),
    var(--button-shadow);
}

.action-button--primary:hover {
  background: linear-gradient(180deg, var(--button-blue-top), var(--button-blue-bottom));
  color: #ffffff;
  border-color: rgba(255, 255, 255, 0.36);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.28),
    inset 0 14px 20px rgba(255, 255, 255, 0.06),
    var(--button-shadow-hover);
}

.action-button--soft {
  background: linear-gradient(180deg, var(--button-blue-soft-top), var(--button-blue-soft-bottom));
  color: #f6fbff;
  border-color: var(--button-border);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.2),
    0 10px 22px rgba(35, 67, 97, 0.14);
}

.action-button--soft:hover {
  background: linear-gradient(180deg, var(--button-blue-soft-top), var(--button-blue-soft-bottom));
  color: #ffffff;
  border-color: rgba(255, 255, 255, 0.34);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.24),
    inset 0 14px 20px rgba(255, 255, 255, 0.05),
    var(--button-shadow-hover);
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

.hero-skeleton { height: 260px; }
.content-skeleton { height: 420px; }
.panel-skeleton { height: 220px; }
.comment-skeleton { height: 320px; }

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

@keyframes liquid-sweep {
  0% { transform: translateX(-190%) skewX(-24deg); }
  100% { transform: translateX(380%) skewX(-24deg); }
}

@media (max-width: 980px) {
  .double-panel,
  .section-head {
    display: grid;
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .page {
    padding: 16px 12px calc(104px + env(safe-area-inset-bottom));
  }

  .detail-shell,
  .double-panel {
    gap: 14px;
  }

  .hero-cover,
  .panel,
  .sibling-card {
    padding: 18px;
    border-radius: 24px;
  }

  .panel:hover,
  .hero-cover:hover,
  .sibling-card:hover {
    transform: none;
  }

  .hero-actions {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
    margin-bottom: 16px;
  }

  .hero-actions .action-button {
    width: 100%;
  }

  .section-head {
    gap: 10px;
    margin-bottom: 14px;
  }

  .section-head .action-button {
    width: 100%;
  }

  .markdown-body {
    padding: 18px;
    border-radius: 20px;
    overflow-x: auto;
  }

  .markdown-body :deep(h1) {
    font-size: 28px;
    margin-bottom: 16px;
  }

  .markdown-body :deep(h2) {
    font-size: 24px;
    margin: 24px 0 12px;
  }

  .markdown-body :deep(h3) {
    font-size: 19px;
  }

  .markdown-body :deep(p),
  .markdown-body :deep(li) {
    font-size: 15px;
  }

  .markdown-body :deep(pre) {
    padding: 14px;
    border-radius: 16px;
  }

  .comment-form,
  .comment-list {
    gap: 10px;
  }

  .comment-card {
    padding: 16px;
    border-radius: 18px;
  }

  .result,
  input,
  textarea,
  .action-button {
    min-height: 44px;
  }
}

@media (max-width: 540px) {
  .page {
    padding: 14px 10px calc(98px + env(safe-area-inset-bottom));
  }

  .hero-cover,
  .panel,
  .sibling-card {
    padding: 16px;
    border-radius: 20px;
  }

  .hero-actions {
    grid-template-columns: 1fr;
  }

  .article-summary,
  .result,
  .comment-card p,
  .sibling-card p {
    font-size: 14px;
  }

  .markdown-body {
    padding: 16px;
  }

  .markdown-body :deep(h1) {
    font-size: 25px;
  }

  .markdown-body :deep(h2) {
    font-size: 21px;
  }

  .markdown-body :deep(h3) {
    font-size: 18px;
  }
}
</style>

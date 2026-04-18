<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import BriefingLearningLab from '../components/BriefingLearningLab.vue'
import { publicApiRequest, publicApiUrl } from '../lib/api'

const route = useRoute()

const loading = ref(false)
const errorText = ref('')
const studyItem = ref(null)

const activeSentence = ref('')
const sentenceLoading = ref(false)
const sentenceError = ref('')
const sentenceAnalysis = ref(null)

const wordPanel = ref(null)
const wordPanelLoading = ref(false)
const wordPanelError = ref('')

const wordTooltip = reactive({
  visible: false,
  x: 0,
  y: 0,
  word: '',
  loading: false,
  error: '',
  data: null
})

const wordCache = reactive({})
const sentenceCache = reactive({})
const briefingId = computed(() => Number(route.params.id))

let tooltipHideTimer = 0
let latestWordKey = ''

function splitParagraphs(text) {
  return String(text || '')
    .split(/\n{2,}/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function splitIntoSentences(paragraph) {
  return (paragraph.match(/[^.!?]+(?:[.!?]+["')\]]*)?|.+$/g) || [paragraph])
    .map((item) => item.trim())
    .filter(Boolean)
}

function tokenizeSentence(sentence) {
  return (sentence.match(/[A-Za-z]+(?:'[A-Za-z]+)?|[0-9]+(?:\.[0-9]+)?|\s+|[^\s]/g) || [sentence]).map((token) => ({
    text: token,
    isWord: /^[A-Za-z]/.test(token),
    isSpace: /^\s+$/.test(token)
  }))
}

const originalParagraphs = computed(() =>
  splitParagraphs(studyItem.value?.source_content).map((paragraph, paragraphIndex) => ({
    id: `paragraph-${paragraphIndex}`,
    sentences: splitIntoSentences(paragraph).map((sentence, sentenceIndex) => ({
      id: `sentence-${paragraphIndex}-${sentenceIndex}`,
      raw: sentence,
      tokens: tokenizeSentence(sentence)
    }))
  }))
)

const translatedParagraphs = computed(() => splitParagraphs(studyItem.value?.translated_content))

async function request(path, options = {}) {
  const response = await fetch(publicApiUrl(path), options)
  const result = await response.json()

  if (!response.ok || result.success === false) {
    throw new Error(result.error || '请求失败')
  }

  return result.data
}

async function loadStudyPage() {
  if (!briefingId.value) return

  loading.value = true
  errorText.value = ''
  sentenceAnalysis.value = null
  sentenceError.value = ''
  wordPanel.value = null
  wordPanelError.value = ''
  wordTooltip.visible = false
  activeSentence.value = ''
  Object.keys(wordCache).forEach((key) => delete wordCache[key])
  Object.keys(sentenceCache).forEach((key) => delete sentenceCache[key])

  try {
    studyItem.value = await publicApiRequest(`/daily-briefings/${briefingId.value}/study`)
  } catch (error) {
    errorText.value = error.message
  } finally {
    loading.value = false
  }
}

async function fetchWordExplanation(word, sentence) {
  const key = `${word.toLowerCase()}|${sentence}`
  if (wordCache[key]) {
    return wordCache[key]
  }

  const data = await publicApiRequest(`/daily-briefings/${briefingId.value}/word-explanation`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      word,
      sentence
    })
  })

  wordCache[key] = data
  return data
}

async function fetchSentenceAnalysis(sentence) {
  if (sentenceCache[sentence]) {
    return sentenceCache[sentence]
  }

  const data = await publicApiRequest(`/daily-briefings/${briefingId.value}/sentence-analysis`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      sentence
    })
  })

  sentenceCache[sentence] = data
  return data
}

function updateTooltipPosition(event) {
  wordTooltip.x = Math.max(12, Math.min(window.innerWidth - 320, event.clientX + 18))
  wordTooltip.y = Math.max(12, Math.min(window.innerHeight - 220, event.clientY + 18))
}

async function handleWordEnter(event, word, sentence) {
  window.clearTimeout(tooltipHideTimer)
  updateTooltipPosition(event)

  const key = `${word.toLowerCase()}|${sentence}`
  latestWordKey = key

  wordTooltip.visible = true
  wordTooltip.word = word
  wordTooltip.loading = true
  wordTooltip.error = ''
  wordTooltip.data = null

  wordPanelLoading.value = true
  wordPanelError.value = ''

  try {
    const data = await fetchWordExplanation(word, sentence)
    if (latestWordKey !== key) return

    wordTooltip.data = data
    wordTooltip.loading = false
    wordPanel.value = data
    wordPanelLoading.value = false
  } catch (error) {
    if (latestWordKey !== key) return

    wordTooltip.error = error.message
    wordTooltip.loading = false
    wordPanelError.value = error.message
    wordPanelLoading.value = false
  }
}

function handleWordMove(event) {
  if (!wordTooltip.visible) return
  updateTooltipPosition(event)
}

function handleWordLeave() {
  tooltipHideTimer = window.setTimeout(() => {
    wordTooltip.visible = false
  }, 120)
}

function keepTooltipOpen() {
  window.clearTimeout(tooltipHideTimer)
}

async function analyzeSentence(sentence) {
  activeSentence.value = sentence
  sentenceLoading.value = true
  sentenceError.value = ''

  try {
    sentenceAnalysis.value = await fetchSentenceAnalysis(sentence)
  } catch (error) {
    sentenceError.value = error.message
  } finally {
    sentenceLoading.value = false
  }
}

function openOriginalSource() {
  if (!studyItem.value?.source_url) return
  window.open(studyItem.value.source_url, '_blank', 'noopener,noreferrer')
}

onMounted(loadStudyPage)
watch(() => route.params.id, loadStudyPage)
onBeforeUnmount(() => {
  window.clearTimeout(tooltipHideTimer)
})
</script>

<template>
  <main class="page">
    <section class="hero glass">
      <div class="hero-copy">
        <p class="eyebrow">English Briefing Study</p>
        <h1>{{ studyItem?.title || '每日资讯英语精读' }}</h1>
        <p class="meta-line">
          {{ studyItem?.briefing_date || '今日资讯' }}
          <span v-if="studyItem?.source_name"> / {{ studyItem.source_name }}</span>
        </p>
        <p class="intro">
          这里会先抓取每日资讯原文，再生成中文译文。你可以把鼠标移到英文单词上查看当前语境下的意思，也可以直接点击一句英文查看句子结构、主谓宾和写法说明。
        </p>
        <p v-if="studyItem?.translation_hint" class="hint-text">{{ studyItem.translation_hint }}</p>
      </div>

      <div class="hero-actions">
        <RouterLink class="action-button action-button--soft" to="/briefings">返回每日资讯</RouterLink>
        <button class="action-button action-button--primary" @click="openOriginalSource">查看原文链接</button>
      </div>
    </section>

    <section v-if="loading" class="loading-card glass">
      <h2>首次打开正在准备学习内容</h2>
      <p>系统会自动抓取原文并生成译文，通常需要几秒钟。</p>
    </section>

    <section v-else-if="errorText" class="loading-card glass">
      <h2>学习页暂时打不开</h2>
      <p>{{ errorText }}</p>
    </section>

    <template v-else-if="studyItem">
      <section class="study-layout">
        <article class="glass panel">
          <div class="section-head">
            <div>
              <p class="eyebrow">Original</p>
              <h2>英文原文</h2>
            </div>
            <p class="muted">点击一句查看句法，悬浮单词看词义</p>
          </div>

          <div class="article-scroll">
            <div
              v-for="paragraph in originalParagraphs"
              :key="paragraph.id"
              class="sentence-paragraph"
            >
              <div
                v-for="sentence in paragraph.sentences"
                :key="sentence.id"
                class="sentence-block"
                :class="{ active: activeSentence === sentence.raw }"
                @click="analyzeSentence(sentence.raw)"
              >
                <template v-for="(token, tokenIndex) in sentence.tokens" :key="`${sentence.id}-${tokenIndex}`">
                  <button
                    v-if="token.isWord"
                    class="word-token"
                    @mouseenter.stop="handleWordEnter($event, token.text, sentence.raw)"
                    @mousemove.stop="handleWordMove"
                    @mouseleave.stop="handleWordLeave"
                    @click.stop="handleWordEnter($event, token.text, sentence.raw)"
                  >
                    {{ token.text }}
                  </button>
                  <span v-else-if="token.isSpace" class="token-space">{{ token.text }}</span>
                  <span v-else class="token-punctuation">{{ token.text }}</span>
                </template>
              </div>
            </div>
          </div>
        </article>

        <article class="glass panel">
          <div class="section-head">
            <div>
              <p class="eyebrow">Translation</p>
              <h2>中文译文</h2>
            </div>
            <p class="muted">{{ studyItem.translation_provider || 'cached' }}</p>
          </div>

          <div class="article-scroll translation-copy">
            <p v-for="(paragraph, index) in translatedParagraphs" :key="`translated-${index}`">
              {{ paragraph }}
            </p>
          </div>
        </article>
      </section>

      <section class="analysis-layout">
        <article class="glass panel info-panel">
          <p class="eyebrow">Word Focus</p>
          <h3>单词释义</h3>
          <p v-if="wordPanelLoading" class="muted">正在解释当前单词...</p>
          <p v-else-if="wordPanelError" class="muted">{{ wordPanelError }}</p>
          <div v-else-if="wordPanel" class="info-stack">
            <div class="info-card">
              <span>单词</span>
              <strong>{{ wordPanel.word }}</strong>
            </div>
            <div class="info-card">
              <span>中文意思</span>
              <strong>{{ wordPanel.meaning || '暂无' }}</strong>
            </div>
            <div class="info-grid">
              <div class="info-card">
                <span>词性</span>
                <strong>{{ wordPanel.part_of_speech || '待判断' }}</strong>
              </div>
              <div class="info-card">
                <span>音标</span>
                <strong>{{ wordPanel.phonetic || '暂无' }}</strong>
              </div>
            </div>
            <div class="info-card wide">
              <span>当前句子里的用法</span>
              <p>{{ wordPanel.usage || '请结合句子整体理解。' }}</p>
            </div>
          </div>
          <p v-else class="muted">把鼠标移到左侧原文的单词上，这里会显示该词在当前句子里的意思。</p>
        </article>

        <article class="glass panel info-panel">
          <p class="eyebrow">Sentence Focus</p>
          <h3>句子结构解析</h3>
          <p v-if="sentenceLoading" class="muted">正在分析当前句子...</p>
          <p v-else-if="sentenceError" class="muted">{{ sentenceError }}</p>
          <div v-else-if="sentenceAnalysis" class="info-stack">
            <div class="info-card wide">
              <span>原句</span>
              <p>{{ sentenceAnalysis.sentence }}</p>
            </div>
            <div class="info-card wide">
              <span>中文翻译</span>
              <p>{{ sentenceAnalysis.translation || '暂无译文说明' }}</p>
            </div>
            <div class="info-card wide">
              <span>为什么这样写</span>
              <p>{{ sentenceAnalysis.explanation || '暂无说明' }}</p>
            </div>
            <div class="info-card">
              <span>句子结构</span>
              <strong>{{ sentenceAnalysis.structure || '待分析' }}</strong>
            </div>
            <div class="info-grid">
              <div class="info-card">
                <span>主语</span>
                <strong>{{ sentenceAnalysis.subject || '待分析' }}</strong>
              </div>
              <div class="info-card">
                <span>谓语</span>
                <strong>{{ sentenceAnalysis.predicate || '待分析' }}</strong>
              </div>
              <div class="info-card">
                <span>宾语</span>
                <strong>{{ sentenceAnalysis.object || '待分析' }}</strong>
              </div>
            </div>
            <div v-if="sentenceAnalysis.grammar_points?.length" class="info-card wide">
              <span>语法点</span>
              <ul class="grammar-list">
                <li v-for="(point, index) in sentenceAnalysis.grammar_points" :key="`point-${index}`">{{ point }}</li>
              </ul>
            </div>
          </div>
          <p v-else class="muted">点击左侧任意一句英文，这里会解释句子结构、主谓宾和写法。</p>
        </article>
      </section>

      <BriefingLearningLab
        :briefing-id="briefingId"
        :briefing-title="studyItem.title"
      />
    </template>

    <Teleport to="body">
      <div
        v-if="wordTooltip.visible"
        class="word-tooltip"
        :style="{ left: `${wordTooltip.x}px`, top: `${wordTooltip.y}px` }"
        @mouseenter="keepTooltipOpen"
        @mouseleave="handleWordLeave"
      >
        <strong>{{ wordTooltip.word }}</strong>
        <p v-if="wordTooltip.loading">正在解释...</p>
        <p v-else-if="wordTooltip.error">{{ wordTooltip.error }}</p>
        <template v-else-if="wordTooltip.data">
          <p>{{ wordTooltip.data.meaning || '暂无释义' }}</p>
          <small>{{ wordTooltip.data.part_of_speech || '待判断' }}<span v-if="wordTooltip.data.phonetic"> / {{ wordTooltip.data.phonetic }}</span></small>
        </template>
      </div>
    </Teleport>
  </main>
</template>

<style scoped>
.page {
  max-width: 1240px;
  margin: 0 auto;
  padding: 28px 20px 88px;
  display: grid;
  gap: 20px;
}

.glass {
  position: relative;
  overflow: hidden;
  isolation: isolate;
  border-radius: 30px;
  border: 1px solid rgba(255, 255, 255, 0.58);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.72), rgba(255, 255, 255, 0.42)),
    linear-gradient(145deg, rgba(255, 255, 255, 0.18), rgba(170, 205, 234, 0.12));
  box-shadow:
    0 28px 70px rgba(32, 64, 93, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.76);
}

.glass::before {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background:
    radial-gradient(circle at 16% 0%, rgba(255, 255, 255, 0.86), transparent 30%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.66), rgba(255, 255, 255, 0.12) 34%, rgba(255, 255, 255, 0.06) 100%);
  pointer-events: none;
}

.glass > * {
  position: relative;
  z-index: 1;
}

.hero,
.panel,
.loading-card {
  padding: 24px;
}

.hero {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  align-items: start;
}

.hero-copy {
  display: grid;
  gap: 10px;
  max-width: 800px;
}

.hero-actions {
  display: grid;
  gap: 12px;
  min-width: 190px;
}

.eyebrow {
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 0.24em;
  font-size: 12px;
  color: #567a9b;
}

.meta-line,
.intro,
.muted,
.hint-text,
.loading-card p {
  margin: 0;
  color: #5d7388;
  line-height: 1.8;
}

.hint-text {
  color: #466985;
}

.study-layout,
.analysis-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.section-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: start;
  margin-bottom: 16px;
}

.article-scroll {
  max-height: 720px;
  overflow: auto;
  display: grid;
  gap: 16px;
  padding-right: 6px;
}

.translation-copy p,
.sentence-paragraph {
  margin: 0;
}

.sentence-paragraph {
  display: grid;
  gap: 10px;
}

.sentence-block {
  padding: 14px 16px;
  border-radius: 22px;
  cursor: pointer;
  line-height: 2;
  background: rgba(255, 255, 255, 0.42);
  border: 1px solid rgba(255, 255, 255, 0.58);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.7),
    0 12px 24px rgba(35, 67, 97, 0.08);
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}

.sentence-block:hover,
.sentence-block.active {
  transform: translateY(-2px);
  border-color: rgba(139, 184, 219, 0.68);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.84),
    0 16px 30px rgba(35, 67, 97, 0.12);
}

.word-token {
  display: inline;
  padding: 0;
  margin: 0;
  border: none;
  background: none;
  color: #1f4767;
  cursor: pointer;
  font-weight: 600;
}

.word-token:hover {
  color: #0d5a8f;
}

.token-space,
.token-punctuation {
  white-space: pre-wrap;
}

.info-panel h3,
.loading-card h2,
.hero h1,
.section-head h2 {
  margin: 0;
}

.info-stack {
  display: grid;
  gap: 12px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.info-card {
  display: grid;
  gap: 8px;
  padding: 14px 16px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.44);
  border: 1px solid rgba(255, 255, 255, 0.58);
}

.info-card.wide {
  grid-column: 1 / -1;
}

.info-card span {
  font-size: 12px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: #6e8697;
}

.info-card p,
.info-card strong {
  margin: 0;
  color: #18354d;
  line-height: 1.8;
}

.grammar-list {
  margin: 0;
  padding-left: 18px;
  color: #18354d;
  line-height: 1.8;
}

.action-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 48px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 999px;
  padding: 0 18px;
  cursor: pointer;
  color: #18354d;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.88), rgba(227, 239, 248, 0.62)),
    linear-gradient(135deg, rgba(255, 255, 255, 0.2), rgba(157, 194, 223, 0.08));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.88),
    0 12px 24px rgba(35, 67, 97, 0.12);
}

.action-button--primary {
  background: linear-gradient(180deg, #355f82, #274a67);
  color: #f8fbff;
}

.word-tooltip {
  position: fixed;
  z-index: 30;
  width: min(280px, calc(100vw - 24px));
  padding: 14px 16px;
  border-radius: 20px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(235, 244, 251, 0.88)),
    linear-gradient(135deg, rgba(255, 255, 255, 0.26), rgba(164, 198, 226, 0.12));
  border: 1px solid rgba(255, 255, 255, 0.82);
  box-shadow:
    0 18px 38px rgba(35, 67, 97, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
}

.word-tooltip strong,
.word-tooltip p,
.word-tooltip small {
  display: block;
  margin: 0;
  color: #173d5c;
  line-height: 1.7;
}

.word-tooltip p {
  margin-top: 6px;
}

@media (max-width: 980px) {
  .study-layout,
  .analysis-layout {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .page {
    padding: 18px 12px 96px;
  }

  .hero {
    display: grid;
    grid-template-columns: 1fr;
  }

  .hero-actions {
    min-width: 0;
  }

  .hero-actions .action-button {
    width: 100%;
  }

  .panel,
  .hero,
  .loading-card {
    padding: 18px;
  }

  .article-scroll {
    max-height: none;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }
}
</style>

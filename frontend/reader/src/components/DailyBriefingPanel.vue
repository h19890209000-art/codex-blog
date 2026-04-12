<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'

const props = defineProps({
  compact: {
    type: Boolean,
    default: false
  }
})

const apiBase = 'http://127.0.0.1:8080/api/public'
const loading = ref(false)
const errorText = ref('')
const activeDate = ref('')
const availableDates = ref([])
const items = ref([])

const visibleItems = computed(() => {
  if (!props.compact) return items.value
  return items.value.slice(0, 4)
})

async function request(path) {
  const response = await fetch(`${apiBase}${path}`)
  const result = await response.json()
  if (!response.ok || result.success === false) {
    throw new Error(result.error || '请求失败')
  }
  return result.data
}

function formatDateLabel(value) {
  if (!value) return '最新'
  return value
}

function formatTime(value) {
  if (!value) return ''
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

async function loadBriefings(date = '') {
  loading.value = true
  errorText.value = ''
  try {
    const query = date ? `?date=${encodeURIComponent(date)}` : ''
    const data = await request(`/daily-briefings${query}`)
    activeDate.value = data.date || ''
    availableDates.value = data.available_dates || []
    items.value = data.items || []
  } catch (error) {
    errorText.value = error.message
  } finally {
    loading.value = false
  }
}

function openSource(url) {
  if (!url) return
  window.open(url, '_blank', 'noopener,noreferrer')
}

onMounted(() => {
  loadBriefings()
})

watch(
  () => props.compact,
  () => {
    if (!items.value.length) loadBriefings(activeDate.value)
  }
)
</script>

<template>
  <article class="glass panel briefing-panel">
    <div class="section-head">
      <div>
        <p class="eyebrow">Daily AI Briefing</p>
        <h2>{{ compact ? '每日简讯' : '全球 AI 每日简讯' }}</h2>
        <p class="muted">
          {{ activeDate ? `${activeDate} · 自动聚合 10 条内全球 AI 时讯` : '聚合全球 AI 新闻源，生成今日简讯' }}
        </p>
      </div>
      <RouterLink v-if="compact" class="link-chip" to="/briefings">查看全部</RouterLink>
    </div>

    <div class="date-list">
      <button
        v-for="date in availableDates"
        :key="date"
        class="date-chip"
        :class="{ active: date === activeDate }"
        @click="loadBriefings(date)"
      >
        {{ formatDateLabel(date) }}
      </button>
    </div>

    <div v-if="loading" class="empty-state">正在加载今日简讯...</div>
    <div v-else-if="errorText" class="empty-state">{{ errorText }}</div>
    <div v-else-if="visibleItems.length === 0" class="empty-state">
      今日简讯还没有生成，可以稍后再来，或者先去后台手动抓取一次。
    </div>

    <div v-else class="briefing-list">
      <article
        v-for="item in visibleItems"
        :key="item.id"
        class="briefing-card"
      >
        <div class="briefing-top">
          <div>
            <p class="source-line">
              <span>{{ item.source_name || '未知来源' }}</span>
              <span v-if="item.source_published_at">· {{ formatTime(item.source_published_at) }}</span>
            </p>
            <h3>{{ item.sort_order }}. {{ item.title }}</h3>
          </div>
          <button class="ghost-button small" @click="openSource(item.source_url)">原文</button>
        </div>
        <p class="summary">{{ item.summary || '点击原文查看完整报道。' }}</p>
      </article>
    </div>
  </article>
</template>

<style scoped>
.glass {
  position: relative;
  overflow: hidden;
  isolation: isolate;
  backdrop-filter: blur(28px) saturate(155%);
  -webkit-backdrop-filter: blur(28px) saturate(155%);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.56), rgba(255, 255, 255, 0.24)),
    linear-gradient(135deg, rgba(255, 255, 255, 0.18), rgba(153, 196, 233, 0.1));
  border: 1px solid rgba(255, 255, 255, 0.58);
  box-shadow:
    0 28px 68px rgba(32, 64, 93, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    inset 0 -18px 32px rgba(113, 162, 206, 0.08);
  border-radius: 30px;
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

.panel {
  padding: 24px;
}

.briefing-panel {
  display: grid;
  gap: 18px;
  transition: transform 0.34s ease, box-shadow 0.34s ease;
  transform-style: preserve-3d;
}

.briefing-panel:hover {
  transform: translateY(-6px) rotateX(4deg) rotateY(-3deg);
  box-shadow:
    0 34px 82px rgba(32, 64, 93, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.76);
}

.section-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: start;
}

.date-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.date-chip,
.link-chip,
.ghost-button {
  border: none;
  border-radius: 999px;
  padding: 10px 16px;
  cursor: pointer;
  transition: transform 0.22s ease, box-shadow 0.22s ease;
}

.date-chip,
.link-chip {
  background: rgba(255, 255, 255, 0.42);
  color: #183149;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    0 8px 18px rgba(35, 67, 97, 0.08);
}

.date-chip.active {
  background: linear-gradient(160deg, #173452, #244f73);
  color: #fff;
}

.link-chip {
  text-decoration: none;
  display: inline-flex;
  align-items: center;
}

.briefing-list {
  display: grid;
  gap: 14px;
}

.briefing-card {
  padding: 18px;
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.7), rgba(255, 255, 255, 0.36)),
    linear-gradient(150deg, rgba(255, 255, 255, 0.16), rgba(149, 193, 231, 0.08));
  border: 1px solid rgba(255, 255, 255, 0.54);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.76),
    0 18px 34px rgba(35, 67, 97, 0.08);
  display: grid;
  gap: 10px;
  transition: transform 0.26s ease, box-shadow 0.26s ease;
}

.briefing-card:hover {
  transform: translateY(-6px) rotateX(4deg);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.78),
    0 24px 42px rgba(35, 67, 97, 0.14);
}

.briefing-top {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.source-line,
.muted,
.summary,
.empty-state {
  color: #5d7388;
  line-height: 1.8;
}

.source-line {
  font-size: 13px;
  margin: 0 0 6px;
}

.summary,
h3 {
  margin: 0;
}

.small {
  padding: 8px 14px;
  background: linear-gradient(160deg, #173452, #244f73);
  color: #fff;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.24),
    0 10px 18px rgba(18, 45, 68, 0.16);
}

.eyebrow {
  text-transform: uppercase;
  letter-spacing: 0.24em;
  font-size: 12px;
  color: #51779a;
  margin: 0 0 8px;
}

@media (max-width: 760px) {
  .section-head,
  .briefing-top {
    display: grid;
    grid-template-columns: 1fr;
  }
}
</style>

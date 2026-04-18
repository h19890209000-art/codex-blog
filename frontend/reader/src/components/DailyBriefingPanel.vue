<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { publicApiUrl } from '../lib/api'

const props = defineProps({
  compact: {
    type: Boolean,
    default: false
  }
})

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
  const response = await fetch(publicApiUrl(path))
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
          {{ activeDate ? `${activeDate} / 自动聚合 10 条全球 AI 时讯` : '聚合全球 AI 新闻源，生成今日简讯' }}
        </p>
      </div>
      <RouterLink v-if="compact" class="action-button action-button--soft action-link" to="/briefings">查看全部</RouterLink>
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
              <span v-if="item.source_published_at"> / {{ formatTime(item.source_published_at) }}</span>
            </p>
            <h3>{{ item.sort_order }}. {{ item.title }}</h3>
          </div>
          <div class="inline-actions">
            <RouterLink class="action-button action-button--soft inline-action" :to="`/briefings/${item.id}/study`">英语精读</RouterLink>
            <button class="action-button action-button--soft inline-action" @click="openSource(item.source_url)">原文</button>
          </div>
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
  backdrop-filter: none;
  -webkit-backdrop-filter: none;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(255, 255, 255, 0.8)),
    linear-gradient(135deg, rgba(255, 255, 255, 0.3), rgba(153, 196, 233, 0.12));
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
  --button-soft-top: rgba(255, 255, 255, 0.82);
  --button-soft-bottom: rgba(229, 240, 249, 0.58);
  --button-border: rgba(255, 255, 255, 0.8);
  --button-text: #2d5372;
  --button-shadow: 0 12px 22px rgba(35, 67, 97, 0.12);
  --button-shadow-hover: 0 18px 28px rgba(35, 67, 97, 0.16);
  --button-edge-left: rgba(255, 255, 255, 0.62);
  --button-edge-right: rgba(129, 169, 199, 0.24);
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

.date-chip {
  border: none;
  border-radius: 999px;
  padding: 10px 16px;
  cursor: pointer;
  background: rgba(255, 255, 255, 0.46);
  color: #21425d;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.74),
    0 8px 18px rgba(35, 67, 97, 0.08);
  transition: transform 0.22s ease, box-shadow 0.22s ease, background 0.22s ease, color 0.22s ease;
}

.date-chip:hover {
  transform: translateY(-1px);
  background: rgba(244, 251, 255, 0.82);
  color: #173d5c;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.86),
    0 12px 22px rgba(35, 67, 97, 0.1);
}

.date-chip.active {
  background: linear-gradient(160deg, #2a5376, #3d6f95);
  color: #fff;
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
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 18px;
  align-items: start;
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

.action-button {
  position: relative;
  overflow: hidden;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 108px;
  min-height: 46px;
  padding: 0 18px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  cursor: pointer;
  text-decoration: none;
  box-sizing: border-box;
  transition: transform 0.22s ease, box-shadow 0.22s ease, background 0.22s ease, color 0.22s ease, border-color 0.22s ease;
}

.action-button::before {
  content: "";
  position: absolute;
  inset: -26% auto -26% -16%;
  width: 24%;
  transform: translateX(-260%) skewX(-20deg);
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.18) 18%,
    rgba(255, 255, 255, 0.92) 50%,
    rgba(255, 255, 255, 0.26) 82%,
    transparent 100%
  );
  filter: blur(1.5px);
  opacity: 0;
}

.action-button::after {
  content: "";
  position: absolute;
  inset: 1px;
  border-radius: inherit;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.74), rgba(255, 255, 255, 0.22) 36%, rgba(255, 255, 255, 0.05) 64%, transparent 76%),
    linear-gradient(118deg, rgba(255, 255, 255, 0.18), transparent 28%, transparent 72%, rgba(129, 169, 199, 0.18));
  box-shadow:
    inset 1px 1px 0 var(--button-edge-left),
    inset -1px -1px 0 var(--button-edge-right);
  pointer-events: none;
}

.action-button:hover {
  transform: translateY(-2px);
}

.action-button:hover::before {
  opacity: 1;
  animation: liquid-sweep 0.78s ease forwards;
}

.action-button--soft {
  background:
    linear-gradient(180deg, var(--button-soft-top), var(--button-soft-bottom)),
    linear-gradient(135deg, rgba(255, 255, 255, 0.18), rgba(162, 198, 226, 0.08));
  color: var(--button-text);
  border-color: var(--button-border);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.86),
    inset 0 -8px 14px rgba(130, 168, 198, 0.06),
    var(--button-shadow);
}

.action-button--soft:hover {
  background:
    linear-gradient(180deg, var(--button-soft-top), var(--button-soft-bottom)),
    linear-gradient(135deg, rgba(255, 255, 255, 0.2), rgba(162, 198, 226, 0.1));
  color: #244763;
  border-color: rgba(255, 255, 255, 0.88);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.92),
    inset 0 -10px 16px rgba(130, 168, 198, 0.08),
    var(--button-shadow-hover);
}

.action-link {
  white-space: nowrap;
}

.inline-action {
  align-self: start;
}

.inline-actions {
  display: grid;
  gap: 10px;
  justify-items: end;
}

.eyebrow {
  text-transform: uppercase;
  letter-spacing: 0.24em;
  font-size: 12px;
  color: #51779a;
  margin: 0 0 8px;
}

@keyframes liquid-sweep {
  0% {
    transform: translateX(-260%) skewX(-20deg);
  }
  100% {
    transform: translateX(470%) skewX(-20deg);
  }
}

@media (max-width: 760px) {
  .section-head,
  .briefing-top {
    display: grid;
    grid-template-columns: 1fr;
  }

  .panel {
    padding: 18px;
  }

  .briefing-panel,
  .briefing-card {
    border-radius: 24px;
  }

  .briefing-panel:hover,
  .briefing-card:hover {
    transform: none;
  }

  .date-list {
    flex-wrap: nowrap;
    overflow-x: auto;
    padding-bottom: 4px;
    scrollbar-width: none;
  }

  .date-list::-webkit-scrollbar {
    display: none;
  }

  .date-chip {
    flex: 0 0 auto;
  }

  .briefing-card {
    padding: 16px;
    gap: 8px;
  }

  .inline-action {
    justify-self: stretch;
    width: 100%;
  }

  .inline-actions {
    justify-items: stretch;
  }

  .action-button {
    min-height: 44px;
  }
}

@media (max-width: 540px) {
  .panel {
    padding: 16px;
  }

  .briefing-panel,
  .briefing-card {
    border-radius: 20px;
  }

  .summary,
  .muted,
  .empty-state {
    font-size: 14px;
  }

  .action-link {
    width: 100%;
  }
}
</style>

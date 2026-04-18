<script setup>
import { computed, ref, watch } from 'vue'
import { publicApiRequest } from '../lib/api'

const props = defineProps({
  briefingId: { type: Number, required: true },
  briefingTitle: { type: String, default: '' }
})

const REVIEW_KEY = 'briefing-learning-review-v1'
const GOAL_KEY = 'briefing-learning-goal-v1'

const learningGoal = ref(loadStoredGoal())
const lessonLoading = ref(false)
const lessonError = ref('')
const lessonPlan = ref(null)
const roleplayReply = ref('')
const roleplayLoading = ref(false)
const roleplayError = ref('')
const roleplayResult = ref(null)
const reviewCards = ref(loadReviewCards())
const notice = ref('')

const roleplayBlock = computed(() => lessonPlan.value?.roleplay || null)
const reviewDueCards = computed(() => {
  const now = Date.now()
  const cards = [...reviewCards.value].sort((a, b) => Date.parse(a.nextReviewAt || 0) - Date.parse(b.nextReviewAt || 0))
  const due = cards.filter((item) => (Date.parse(item.nextReviewAt || 0) || 0) <= now)
  return due.length ? due : cards.slice(0, 6)
})

watch(
  () => props.briefingId,
  () => {
    lessonPlan.value = null
    lessonError.value = ''
    roleplayReply.value = ''
    roleplayResult.value = null
    roleplayError.value = ''
  }
)

function loadStoredGoal() {
  if (typeof window === 'undefined') return '我想用 AI 资讯练英语，重点是听懂、复述、技术沟通。'
  return window.localStorage.getItem(GOAL_KEY) || '我想用 AI 资讯练英语，重点是听懂、复述、技术沟通。'
}

function loadReviewCards() {
  if (typeof window === 'undefined') return []
  try {
    const raw = window.localStorage.getItem(REVIEW_KEY)
    const parsed = raw ? JSON.parse(raw) : []
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function persistGoal() {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(GOAL_KEY, learningGoal.value.trim())
}

function persistCards() {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(REVIEW_KEY, JSON.stringify(reviewCards.value))
}

function flashNotice(text) {
  notice.value = text
  window.clearTimeout(flashNotice.timer)
  flashNotice.timer = window.setTimeout(() => {
    notice.value = ''
  }, 1800)
}

function speak(text) {
  if (!('speechSynthesis' in window) || !text) return
  const utterance = new SpeechSynthesisUtterance(String(text))
  utterance.lang = 'en-US'
  utterance.rate = 0.92
  window.speechSynthesis.cancel()
  window.speechSynthesis.speak(utterance)
}

async function loadPlan() {
  if (!props.briefingId) return
  lessonLoading.value = true
  lessonError.value = ''
  persistGoal()
  try {
    lessonPlan.value = await publicApiRequest(`/daily-briefings/${props.briefingId}/learning-plan`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ goal: learningGoal.value })
    })
  } catch (error) {
    lessonError.value = error.message
  } finally {
    lessonLoading.value = false
  }
}

async function runRoleplay() {
  if (!props.briefingId || !roleplayReply.value.trim()) return
  roleplayLoading.value = true
  roleplayError.value = ''
  try {
    roleplayResult.value = await publicApiRequest(`/daily-briefings/${props.briefingId}/roleplay`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        goal: learningGoal.value,
        scene: roleplayBlock.value?.scene || '',
        learner_reply: roleplayReply.value
      })
    })
  } catch (error) {
    roleplayError.value = error.message
  } finally {
    roleplayLoading.value = false
  }
}

function addCard(front, back, noteCn, category) {
  if (!front || !back) return
  const now = new Date().toISOString()
  const existingIndex = reviewCards.value.findIndex((item) => item.front === front)
  const card = {
    id: existingIndex >= 0 ? reviewCards.value[existingIndex].id : `${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    front,
    back,
    noteCn: noteCn || '',
    category: category || 'briefing',
    sourceTitle: props.briefingTitle,
    nextReviewAt: new Date(Date.now() + 6 * 60 * 60 * 1000).toISOString(),
    reviewCount: existingIndex >= 0 ? reviewCards.value[existingIndex].reviewCount : 0,
    errorCount: existingIndex >= 0 ? reviewCards.value[existingIndex].errorCount : 0,
    updatedAt: now
  }
  if (existingIndex >= 0) {
    reviewCards.value.splice(existingIndex, 1, card)
  } else {
    reviewCards.value.unshift(card)
  }
  persistCards()
}

function importLessonCards() {
  const cards = Array.isArray(lessonPlan.value?.review_cards) ? lessonPlan.value.review_cards : []
  const chunks = Array.isArray(lessonPlan.value?.chunks) ? lessonPlan.value.chunks.slice(0, 3) : []
  cards.forEach((item) => addCard(item.front, item.back, item.note_cn, item.category))
  chunks.forEach((item) => addCard(item.phrase, item.example_en || item.translation_cn, item.coach_tip_cn, 'chunk'))
  flashNotice('已导入今日复习卡')
}

function addRoleplayCard() {
  if (!roleplayResult.value?.better_reply_en) return
  addCard(roleplayResult.value.better_reply_en, roleplayResult.value.correction_cn || roleplayResult.value.coach_reply_cn, '先听再复述', 'roleplay')
  flashNotice('已加入口语纠正卡')
}

function scheduleCard(card, days, isHard) {
  reviewCards.value = reviewCards.value.map((item) => item.id !== card.id ? item : ({
    ...item,
    nextReviewAt: new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString(),
    reviewCount: Number(item.reviewCount || 0) + 1,
    errorCount: Number(item.errorCount || 0) + (isHard ? 1 : 0),
    updatedAt: new Date().toISOString()
  }))
  persistCards()
}

function removeCard(cardId) {
  reviewCards.value = reviewCards.value.filter((item) => item.id !== cardId)
  persistCards()
}

function formatTime(value) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '现在' : date.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <section class="learning-lab glass panel">
    <div class="lab-head">
      <div>
        <p class="eyebrow">Learning MVP</p>
        <h2>场景英语学习流</h2>
      </div>
      <button class="action-button action-button--primary" @click="loadPlan">{{ lessonLoading ? '正在生成...' : '生成今日 10 分钟练习' }}</button>
    </div>

    <label class="field-label" for="learning-goal">中文目标</label>
    <textarea id="learning-goal" v-model="learningGoal" class="goal-input" rows="4" placeholder="比如：我想练到能在技术会议里用简单英语解释 AI 资讯重点。" />
    <p class="muted">允许先用中文定目标，系统会先判断场景，再把训练拉回到英语输出。</p>
    <p v-if="notice" class="hint-text">{{ notice }}</p>
    <p v-if="lessonError" class="muted">{{ lessonError }}</p>

    <template v-if="lessonPlan">
      <div class="goal-profile">
        <div class="profile-chip"><span>Track</span><strong>{{ lessonPlan.goal_profile?.track }}</strong></div>
        <div class="profile-chip"><span>Level</span><strong>{{ lessonPlan.goal_profile?.level }}</strong></div>
        <div class="profile-chip"><span>Scene</span><strong>{{ lessonPlan.goal_profile?.first_scene }}</strong></div>
      </div>
      <p class="muted">{{ lessonPlan.goal_profile?.reason_cn }}</p>

      <div class="flow-grid">
        <article v-for="item in lessonPlan.daily_flow || []" :key="item.step" class="flow-card">
          <span>STEP {{ item.step }}</span>
          <strong>{{ item.title_cn }}</strong>
          <p>{{ item.instruction_cn }}</p>
          <small>{{ item.output_en }}</small>
        </article>
      </div>

      <div class="lab-actions">
        <button class="action-button action-button--soft" @click="importLessonCards">导入复习卡</button>
        <button v-if="lessonPlan.chunks?.[0]?.phrase" class="action-button action-button--soft" @click="speak(lessonPlan.chunks[0].phrase)">先听第一句</button>
      </div>

      <div class="chunk-grid">
        <article v-for="chunk in lessonPlan.chunks || []" :key="chunk.phrase" class="chunk-card">
          <strong>{{ chunk.phrase }}</strong>
          <p>{{ chunk.translation_cn }}</p>
          <small>{{ chunk.why_it_works_cn }}</small>
          <p class="chunk-example">{{ chunk.example_en }}</p>
          <small>{{ (chunk.substitution_options || []).join(' / ') }}</small>
          <div class="mini-actions">
            <button class="mini-button" @click="speak(chunk.phrase)">听</button>
            <button class="mini-button" @click="addCard(chunk.phrase, chunk.example_en || chunk.translation_cn, chunk.coach_tip_cn, 'chunk')">复习</button>
          </div>
        </article>
      </div>

      <div v-if="roleplayBlock" class="roleplay-box">
        <h3>半开放角色扮演</h3>
        <p class="muted">{{ roleplayBlock.goal_cn }}</p>
        <p class="coach-line">{{ roleplayBlock.opening_en }}</p>
        <p class="muted">{{ roleplayBlock.help_cn }}</p>
        <textarea v-model="roleplayReply" class="goal-input" rows="4" placeholder="你可以先写中文，也可以写中英混合，AI 会先判断能不能被听懂。" />
        <div class="lab-actions">
          <button class="action-button action-button--primary" @click="runRoleplay">{{ roleplayLoading ? 'AI 正在纠正...' : '让 AI 纠正我' }}</button>
          <button class="action-button action-button--soft" :disabled="!roleplayResult?.better_reply_en" @click="addRoleplayCard">加入复习卡</button>
        </div>
        <p v-if="roleplayError" class="muted">{{ roleplayError }}</p>
        <div v-else-if="roleplayResult" class="roleplay-result">
          <p><strong>可懂度：</strong>{{ roleplayResult.can_be_understood_score }}/100</p>
          <p><strong>中文反馈：</strong>{{ roleplayResult.coach_reply_cn }}</p>
          <p><strong>为什么这样改：</strong>{{ roleplayResult.correction_cn }}</p>
          <p><strong>更好的表达：</strong>{{ roleplayResult.better_reply_en }}</p>
          <p><strong>下一轮：</strong>{{ roleplayResult.next_prompt_en }}</p>
        </div>
      </div>
    </template>

    <div class="review-box">
      <div class="lab-head compact">
        <div>
          <p class="eyebrow">Review</p>
          <h3>复习卡片</h3>
        </div>
        <p class="muted">{{ reviewCards.length }} 张卡 / {{ reviewDueCards.length }} 张待复习</p>
      </div>
      <div v-if="reviewDueCards.length" class="review-grid">
        <article v-for="card in reviewDueCards" :key="card.id" class="review-card">
          <span>{{ card.category }}</span>
          <strong>{{ card.front }}</strong>
          <p>{{ card.back }}</p>
          <small>{{ card.noteCn || '先理解，再脱稿复述。' }}</small>
          <small>下次复习：{{ formatTime(card.nextReviewAt) }}</small>
          <div class="mini-actions">
            <button class="mini-button" @click="scheduleCard(card, 1, true)">还不熟</button>
            <button class="mini-button" @click="scheduleCard(card, 3, false)">会了</button>
            <button class="mini-button mini-button--danger" @click="removeCard(card.id)">删除</button>
          </div>
        </article>
      </div>
      <p v-else class="muted">这里会保存句块卡、角色纠正卡和错题卡，后面可以按间隔复习。</p>
    </div>
  </section>
</template>

<style scoped>
.learning-lab,.roleplay-box,.review-box,.flow-card,.chunk-card,.review-card,.profile-chip{display:grid;gap:12px}.learning-lab{position:relative;overflow:hidden;isolation:isolate;padding:24px;border-radius:30px;border:1px solid rgba(255,255,255,.58);background:linear-gradient(180deg,rgba(255,255,255,.72),rgba(255,255,255,.42)),linear-gradient(145deg,rgba(255,255,255,.18),rgba(170,205,234,.12));box-shadow:0 28px 70px rgba(32,64,93,.16),inset 0 1px 0 rgba(255,255,255,.76)}.learning-lab::before{content:"";position:absolute;inset:0;border-radius:inherit;background:radial-gradient(circle at 16% 0%,rgba(255,255,255,.86),transparent 30%),linear-gradient(180deg,rgba(255,255,255,.66),rgba(255,255,255,.12) 34%,rgba(255,255,255,.06) 100%);pointer-events:none}.learning-lab>*{position:relative;z-index:1}.eyebrow,.field-label,.profile-chip span,.flow-card span,.review-card span{font-size:12px;letter-spacing:.12em;text-transform:uppercase;color:#6e8697}.eyebrow{margin:0;letter-spacing:.24em;color:#567a9b}.muted,.hint-text{margin:0;color:#5d7388;line-height:1.8}.hint-text{color:#466985}.lab-head,.lab-actions,.mini-actions{display:flex;gap:12px;justify-content:space-between;align-items:start;flex-wrap:wrap}.compact{align-items:center}.goal-input{width:100%;box-sizing:border-box;resize:vertical;border:1px solid rgba(255,255,255,.68);border-radius:24px;background:rgba(255,255,255,.5);padding:16px 18px;color:#193952;line-height:1.8}.action-button,.mini-button{display:inline-flex;align-items:center;justify-content:center;border:1px solid rgba(255,255,255,.72);border-radius:999px;cursor:pointer;transition:transform .2s ease,box-shadow .2s ease}.action-button{min-height:48px;padding:0 18px;color:#18354d;background:linear-gradient(180deg,rgba(255,255,255,.88),rgba(227,239,248,.62)),linear-gradient(135deg,rgba(255,255,255,.2),rgba(157,194,223,.08));box-shadow:inset 0 1px 0 rgba(255,255,255,.88),0 12px 24px rgba(35,67,97,.12)}.action-button--primary{background:linear-gradient(180deg,#355f82,#274a67);color:#f8fbff}.goal-profile,.flow-grid,.chunk-grid,.review-grid{display:grid;gap:12px;grid-template-columns:repeat(2,minmax(0,1fr))}.profile-chip,.flow-card,.chunk-card,.review-card,.roleplay-result{padding:14px 16px;border-radius:22px;background:rgba(255,255,255,.44);border:1px solid rgba(255,255,255,.58)}.profile-chip strong,.flow-card strong,.chunk-card strong,.review-card strong,.lab-head h2,.review-box h3,.roleplay-box h3{margin:0;color:#163a56}.chunk-card p,.chunk-card small,.review-card p,.review-card small,.flow-card p,.flow-card small,.coach-line{margin:0;color:#526d83;line-height:1.8}.chunk-example,.coach-line{color:#173d5c}.mini-button{min-height:36px;padding:0 14px;background:rgba(255,255,255,.7)}.mini-button--danger{color:#7f3243}@media (max-width:980px){.goal-profile,.flow-grid,.chunk-grid,.review-grid{grid-template-columns:1fr}}
</style>

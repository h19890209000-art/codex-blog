<script setup>
import DailyBriefingPanel from '../components/DailyBriefingPanel.vue'
</script>

<template>
  <main class="page">
    <section class="hero glass">
      <div>
        <p class="eyebrow">AI BRIEFING</p>
        <h1>全球 AI 每日简讯</h1>
        <p class="intro">
          这里会展示系统按天聚合的 AI 时讯。默认优先展示最新一天，支持回看最近几天的简讯记录。
        </p>
      </div>
      <RouterLink class="back-link" to="/">返回首页</RouterLink>
    </section>

    <DailyBriefingPanel />
  </main>
</template>

<style scoped>
.page {
  max-width: 1100px;
  margin: 0 auto;
  padding: 32px 20px 80px;
  display: grid;
  gap: 22px;
}

.glass {
  position: relative;
  overflow: hidden;
  backdrop-filter: blur(28px) saturate(150%);
  -webkit-backdrop-filter: blur(28px) saturate(150%);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.56), rgba(255, 255, 255, 0.24)),
    linear-gradient(145deg, rgba(255, 255, 255, 0.22), rgba(175, 210, 240, 0.1));
  border: 1px solid rgba(255, 255, 255, 0.58);
  box-shadow:
    0 28px 70px rgba(32, 64, 93, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.74);
  border-radius: 30px;
  transition: transform 0.34s ease, box-shadow 0.34s ease;
}

.glass::before {
  content: "";
  position: absolute;
  inset: 1px 1px auto;
  height: 46%;
  border-radius: 28px 28px 20px 20px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.78), rgba(255, 255, 255, 0.08));
  pointer-events: none;
}

.glass:hover {
  transform: translateY(-6px) rotateX(4deg) rotateY(-3deg);
  box-shadow:
    0 34px 82px rgba(32, 64, 93, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.78);
}

.hero {
  padding: 28px;
  display: flex;
  justify-content: space-between;
  gap: 20px;
  align-items: start;
}

.eyebrow {
  text-transform: uppercase;
  letter-spacing: 0.24em;
  font-size: 12px;
  color: #51779a;
}

.intro {
  color: #5d7388;
  line-height: 1.9;
  max-width: 720px;
}

.back-link {
  align-self: center;
  border-radius: 999px;
  padding: 10px 16px;
  background: rgba(255, 255, 255, 0.42);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.74),
    0 10px 18px rgba(35, 67, 97, 0.08);
}

@media (max-width: 760px) {
  .hero {
    display: grid;
    grid-template-columns: 1fr;
  }
}
</style>

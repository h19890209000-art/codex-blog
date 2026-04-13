<script setup>
import { RouterView } from 'vue-router'
</script>

<template>
  <RouterView v-slot="{ Component, route }">
    <Transition name="page" mode="out-in">
      <component :is="Component" :key="route.fullPath" />
    </Transition>
  </RouterView>
</template>

<style>
:root {
  --reader-bg: #e7edf5;
  --reader-ink: #16324b;
  --reader-muted: #597287;
  --glass-surface: rgba(255, 255, 255, 0.38);
  --glass-strong: rgba(255, 255, 255, 0.54);
  --glass-line: rgba(255, 255, 255, 0.58);
  --glass-shadow: 0 28px 70px rgba(29, 56, 84, 0.18);
  --glass-shadow-soft: 0 14px 36px rgba(29, 56, 84, 0.12);
  --glass-highlight: linear-gradient(135deg, rgba(255, 255, 255, 0.86), rgba(255, 255, 255, 0.12));
  --glass-depth: linear-gradient(160deg, rgba(130, 180, 223, 0.18), rgba(255, 214, 163, 0.1));
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family: "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
  min-height: 100vh;
  background-color: var(--reader-bg);
  background:
    radial-gradient(circle at 14% 18%, rgba(110, 188, 255, 0.42), transparent 0 24%),
    radial-gradient(circle at 85% 14%, rgba(255, 228, 178, 0.4), transparent 0 18%),
    radial-gradient(circle at 84% 74%, rgba(132, 166, 255, 0.26), transparent 0 20%),
    linear-gradient(145deg, #eef4fb 0%, #edf1f7 42%, #f9f2ea 100%);
  color: var(--reader-ink);
  overflow-x: hidden;
  perspective: 1400px;
}

body::before,
body::after {
  content: "";
  position: fixed;
  inset: auto;
  z-index: -1;
  border-radius: 999px;
  filter: blur(10px);
  pointer-events: none;
}

body::before {
  width: 320px;
  height: 320px;
  top: 72px;
  left: -90px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.8), rgba(139, 204, 255, 0.18) 52%, transparent 72%);
}

body::after {
  width: 360px;
  height: 360px;
  right: -110px;
  bottom: 40px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.74), rgba(255, 202, 148, 0.2) 50%, transparent 74%);
}

a {
  color: inherit;
  text-decoration: none;
}

button,
input,
textarea,
select {
  font: inherit;
}

.page-enter-active,
.page-leave-active {
  transition: opacity 0.24s ease-out, transform 0.24s ease-out;
  will-change: opacity, transform;
}

.page-enter-from,
.page-leave-to {
  opacity: 0;
  transform: translateY(12px);
}

@media (prefers-reduced-motion: reduce) {
  .page-enter-active,
  .page-leave-active {
    transition: none;
  }
}
</style>

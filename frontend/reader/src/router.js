import { createRouter, createWebHistory } from 'vue-router'
import HomeView from './views/HomeView.vue'
import ArticleDetailView from './views/ArticleDetailView.vue'
import DailyBriefingsView from './views/DailyBriefingsView.vue'
import DailyBriefingStudyView from './views/DailyBriefingStudyView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/articles/:id',
      name: 'article-detail',
      component: ArticleDetailView,
      props: true
    },
    {
      path: '/briefings',
      name: 'daily-briefings',
      component: DailyBriefingsView
    },
    {
      path: '/briefings/:id/study',
      name: 'daily-briefing-study',
      component: DailyBriefingStudyView,
      props: true
    }
  ],
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }

    return { top: 0, left: 0 }
  }
})

export default router

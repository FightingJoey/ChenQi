import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/api/client'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/today' },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { guest: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
      meta: { guest: true },
    },
    {
      path: '/today',
      name: 'today',
      component: () => import('@/views/TodayView.vue'),
      meta: { auth: true },
    },
    {
      path: '/calendar',
      name: 'calendar',
      component: () => import('@/views/CalendarView.vue'),
      meta: { auth: true },
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/SettingsView.vue'),
      meta: { auth: true },
    },
    {
      path: '/journals',
      name: 'journals',
      component: () => import('@/views/JournalsView.vue'),
      meta: { auth: true },
    },
  ],
})

router.beforeEach((to) => {
  const authed = !!getToken()
  if (to.meta.auth && !authed) return { name: 'login', query: { redirect: to.fullPath } }
  if (to.meta.guest && authed) return { name: 'today' }
  return true
})

export default router

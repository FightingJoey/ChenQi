<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter, RouterView } from 'vue-router'
import {
  NConfigProvider,
  NMessageProvider,
  NDialogProvider,
  NLayout,
  NLayoutHeader,
  NLayoutContent,
  NMenu,
  NButton,
  NSpace,
  zhCN,
  dateZhCN,
  type MenuOption,
} from 'naive-ui'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

onMounted(() => auth.bootstrap())

const showNav = computed(() => auth.isAuthenticated && !route.meta.guest)

const menuOptions = computed<MenuOption[]>(() => [
  { label: '今日会话', key: '/today' },
  { label: '打卡日历', key: '/calendar' },
  { label: '书写日记', key: '/journals' },
  { label: '肯定语', key: '/affirmations' },
  { label: '设置', key: '/settings' },
])

const bottomTabs = [
  { key: '/today', label: '今日', emoji: '☀️' },
  { key: '/calendar', label: '日历', emoji: '📅' },
  { key: '/affirmations', label: '肯定语', emoji: '💬' },
  { key: '/journals', label: '日记', emoji: '✍️' },
  { key: '/settings', label: '设置', emoji: '⚙️' },
]

function onMenu(key: string) {
  router.push(key)
}

function logout() {
  auth.logout()
  router.push('/login')
}

function isTabActive(key: string) {
  return route.path === key || route.path.startsWith(key + '/')
}
</script>

<template>
  <n-config-provider :locale="zhCN" :date-locale="dateZhCN">
    <n-message-provider>
      <n-dialog-provider>
        <n-layout class="min-h-screen min-h-[100dvh] bg-transparent">
          <n-layout-header
            v-if="showNav"
            class="safe-top px-3 sm:px-4 md:px-8 py-3 bg-white/90 backdrop-blur border-b border-orange-100 sticky top-0 z-20"
          >
            <div class="max-w-5xl mx-auto flex items-center justify-between gap-3">
              <div class="flex items-center gap-2 sm:gap-3 min-w-0">
                <div class="w-9 h-9 shrink-0 rounded-full bg-gradient-to-br from-orange-400 to-amber-500 text-white flex items-center justify-center font-bold shadow">
                  晨
                </div>
                <div class="min-w-0">
                  <div class="font-semibold text-gray-800 leading-tight truncate">晨启 · ChenQi</div>
                  <div class="text-xs text-gray-500 truncate">你好，{{ auth.user?.username }}</div>
                </div>
              </div>
              <n-space align="center" :wrap="false">
                <n-menu
                  mode="horizontal"
                  :value="route.path"
                  :options="menuOptions"
                  @update:value="onMenu"
                  class="hidden lg:flex"
                />
                <n-button quaternary size="medium" class="!px-3" @click="logout">退出</n-button>
              </n-space>
            </div>
          </n-layout-header>

          <n-layout-content
            class="px-3 sm:px-4 md:px-8 py-4 sm:py-6"
            :class="showNav ? 'pb-24 md:pb-6' : ''"
          >
            <div class="max-w-5xl mx-auto w-full">
              <router-view v-if="auth.ready || route.meta.guest" />
              <div v-else class="text-center text-gray-500 py-20">加载中…</div>
            </div>
          </n-layout-content>

          <!-- Mobile bottom nav -->
          <nav
            v-if="showNav"
            class="md:hidden fixed bottom-0 inset-x-0 z-30 bg-white/95 backdrop-blur border-t border-orange-100 safe-bottom"
            aria-label="主导航"
          >
            <div class="grid grid-cols-5 max-w-lg mx-auto">
              <button
                v-for="tab in bottomTabs"
                :key="tab.key"
                type="button"
                class="flex flex-col items-center justify-center gap-0.5 py-2 min-h-[56px] text-[11px] sm:text-xs transition-colors"
                :class="isTabActive(tab.key) ? 'text-orange-600 font-semibold' : 'text-gray-500'"
                @click="onMenu(tab.key)"
              >
                <span class="text-lg leading-none" aria-hidden="true">{{ tab.emoji }}</span>
                <span>{{ tab.label }}</span>
              </button>
            </div>
          </nav>
        </n-layout>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

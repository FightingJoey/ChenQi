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
  { label: '设置', key: '/settings' },
])

function onMenu(key: string) {
  router.push(key)
}

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <n-config-provider :locale="zhCN" :date-locale="dateZhCN">
    <n-message-provider>
      <n-dialog-provider>
        <n-layout class="min-h-screen bg-transparent">
          <n-layout-header v-if="showNav" class="px-4 md:px-8 py-3 bg-white/80 backdrop-blur border-b border-orange-100">
            <div class="max-w-5xl mx-auto flex items-center justify-between gap-4">
              <div class="flex items-center gap-3">
                <div class="w-9 h-9 rounded-full bg-gradient-to-br from-orange-400 to-amber-500 text-white flex items-center justify-center font-bold shadow">
                  S
                </div>
                <div>
                  <div class="font-semibold text-gray-800 leading-tight">早起奇迹 · SAVERS</div>
                  <div class="text-xs text-gray-500">你好，{{ auth.user?.username }}</div>
                </div>
              </div>
              <n-space align="center">
                <n-menu
                  mode="horizontal"
                  :value="route.path"
                  :options="menuOptions"
                  @update:value="onMenu"
                  class="hidden md:flex"
                />
                <n-button quaternary size="small" @click="logout">退出</n-button>
              </n-space>
            </div>
            <div class="md:hidden max-w-5xl mx-auto mt-2">
              <n-menu
                mode="horizontal"
                :value="route.path"
                :options="menuOptions"
                @update:value="onMenu"
              />
            </div>
          </n-layout-header>
          <n-layout-content class="px-4 md:px-8 py-6">
            <div class="max-w-5xl mx-auto">
              <router-view v-if="auth.ready || route.meta.guest" />
              <div v-else class="text-center text-gray-500 py-20">加载中…</div>
            </div>
          </n-layout-content>
        </n-layout>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

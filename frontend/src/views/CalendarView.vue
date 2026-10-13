<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { NCard, NStatistic, NGrid, NGi, NButton, NSpin, NSpace, useMessage } from 'naive-ui'
import { api } from '@/api/client'
import type { CalendarDay, Stats } from '@/types'

const message = useMessage()
const loading = ref(true)
const stats = ref<Stats | null>(null)
const days = ref<CalendarDay[]>([])
const cursor = ref(new Date())

const dayMap = computed(() => {
  const m = new Map<string, string>()
  for (const d of days.value) m.set(d.date, d.status)
  return m
})

const title = computed(() => {
  const d = cursor.value
  return `${d.getFullYear()} 年 ${d.getMonth() + 1} 月`
})

const cells = computed(() => {
  const y = cursor.value.getFullYear()
  const m = cursor.value.getMonth()
  const first = new Date(y, m, 1)
  const startPad = (first.getDay() + 6) % 7 // Monday-first
  const daysInMonth = new Date(y, m + 1, 0).getDate()
  const list: { label: number | null; date: string | null; status: string | null }[] = []
  for (let i = 0; i < startPad; i++) list.push({ label: null, date: null, status: null })
  for (let d = 1; d <= daysInMonth; d++) {
    const date = `${y}-${String(m + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`
    list.push({ label: d, date, status: dayMap.value.get(date) || null })
  }
  while (list.length % 7 !== 0) list.push({ label: null, date: null, status: null })
  return list
})

function fmt(d: Date) {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

async function load() {
  loading.value = true
  try {
    const from = new Date(cursor.value.getFullYear(), cursor.value.getMonth(), 1)
    const to = new Date(cursor.value.getFullYear(), cursor.value.getMonth() + 1, 0)
    const [s, cal] = await Promise.all([
      api<Stats>('/stats'),
      api<{ days: CalendarDay[] }>(`/calendar?from=${fmt(from)}&to=${fmt(to)}`),
    ])
    stats.value = s
    days.value = cal.days
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function shift(delta: number) {
  cursor.value = new Date(cursor.value.getFullYear(), cursor.value.getMonth() + delta, 1)
  load()
}

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-800">打卡日历</h1>
      <p class="text-gray-500 text-sm mt-1">查看连续天数与完成记录</p>
    </div>

    <n-spin :show="loading">
      <n-grid cols="1 s:3" :x-gap="16" :y-gap="16" responsive="screen">
        <n-gi>
          <n-card class="shadow-sm border-orange-100">
            <n-statistic label="当前连续" :value="stats?.current_streak ?? 0">
              <template #suffix>天</template>
            </n-statistic>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card class="shadow-sm border-orange-100">
            <n-statistic label="最长连续" :value="stats?.longest_streak ?? 0">
              <template #suffix>天</template>
            </n-statistic>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card class="shadow-sm border-orange-100">
            <n-statistic label="累计完成" :value="stats?.total_sessions ?? 0">
              <template #suffix>次</template>
            </n-statistic>
          </n-card>
        </n-gi>
      </n-grid>

      <n-card class="mt-6 shadow-sm border-orange-100">
        <div class="flex items-center justify-between mb-4">
          <n-button size="small" @click="shift(-1)">上一月</n-button>
          <div class="font-semibold">{{ title }}</div>
          <n-button size="small" @click="shift(1)">下一月</n-button>
        </div>
        <div class="grid grid-cols-7 gap-1 text-center text-xs text-gray-400 mb-2">
          <div v-for="w in ['一','二','三','四','五','六','日']" :key="w">{{ w }}</div>
        </div>
        <div class="grid grid-cols-7 gap-1">
          <div
            v-for="(c, i) in cells"
            :key="i"
            class="aspect-square rounded-lg flex flex-col items-center justify-center text-sm"
            :class="c.label ? 'bg-orange-50/60 border border-orange-100' : ''"
          >
            <template v-if="c.label">
              <div>{{ c.label }}</div>
              <div class="text-base leading-none">
                <span v-if="c.status === 'completed'">✅</span>
                <span v-else-if="c.status === 'in_progress'">🔶</span>
              </div>
            </template>
          </div>
        </div>
        <p class="text-xs text-gray-400 mt-3">✅ 已完成 · 🔶 进行中</p>
      </n-card>
    </n-spin>
  </div>
</template>

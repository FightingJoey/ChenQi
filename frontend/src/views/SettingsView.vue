<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  NCard, NButton, NSpace, NSwitch, NInputNumber, NText, useMessage,
} from 'naive-ui'
import { api } from '@/api/client'
import type { Routine, RoutineStep } from '@/types'

const message = useMessage()
const loading = ref(false)
const saving = ref(false)
const steps = ref<RoutineStep[]>([])

async function load() {
  loading.value = true
  try {
    const r = await api<Routine>('/routine')
    steps.value = r.steps.map((s) => ({ ...s }))
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    const r = await api<Routine>('/routine', {
      method: 'PUT',
      body: JSON.stringify({ steps: steps.value }),
    })
    steps.value = r.steps
    message.success('已保存例行设置')
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}

function minutesOf(s: RoutineStep) {
  return Math.round(s.duration_seconds / 60)
}

function setMinutes(s: RoutineStep, m: number | null) {
  s.duration_seconds = Math.max(1, m || 1) * 60
}

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-800">设置</h1>
        <p class="text-gray-500 text-sm mt-1">调整各步骤时长与开关（最短 1 分钟）</p>
      </div>
      <n-button type="primary" class="!bg-orange-500" :loading="saving" @click="save">保存</n-button>
    </div>

    <n-card class="shadow-sm border-orange-100">
      <div class="space-y-4">
        <div
          v-for="s in steps"
          :key="s.key"
          class="flex flex-wrap items-center justify-between gap-3 p-3 rounded-xl bg-orange-50/50 border border-orange-100"
        >
          <div class="min-w-[120px]">
            <div class="font-medium">{{ s.name }}</div>
            <n-text depth="3" class="text-xs">{{ s.key }}</n-text>
          </div>
          <n-space align="center">
            <span class="text-sm text-gray-500">启用</span>
            <n-switch v-model:value="s.enabled" />
            <span class="text-sm text-gray-500 ml-2">时长（分钟）</span>
            <n-input-number
              :value="minutesOf(s)"
              :min="1"
              :max="120"
              size="small"
              style="width: 110px"
              @update:value="(v) => setMinutes(s, v)"
            />
          </n-space>
        </div>
      </div>
      <p class="text-xs text-gray-400 mt-4">
        默认顺序：静默 → 肯定语 → 观想 → 运动 → 阅读 → 书写。关闭的步骤不会出现在今日会话中。
      </p>
    </n-card>
  </div>
</template>

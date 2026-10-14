<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  NCard, NButton, NSwitch, NInputNumber, NText, useMessage,
} from 'naive-ui'
import { api } from '@/api/client'
import type { Routine, RoutineStep } from '@/types'
import { isSoundEnabled, setSoundEnabled, playStepComplete, unlockAudio } from '@/utils/sounds'

const message = useMessage()
const loading = ref(false)
const saving = ref(false)
const steps = ref<RoutineStep[]>([])
const soundOn = ref(isSoundEnabled())

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

async function onSoundToggle(v: boolean) {
  soundOn.value = v
  setSoundEnabled(v)
  if (v) {
    await unlockAudio()
    playStepComplete()
    message.success('完成音效已开启')
  } else {
    message.info('完成音效已关闭')
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6 flex flex-col">
    <div class="min-w-0">
      <h1 class="text-xl sm:text-2xl font-bold text-gray-800">设置</h1>
      <p class="text-gray-500 text-sm mt-1">调整各步骤时长、开关与音效</p>
    </div>

    <n-card class="shadow-sm border-orange-100" title="反馈音效">
      <div class="flex items-center justify-between gap-3">
        <div class="min-w-0">
          <div class="font-medium">完成提示音</div>
          <p class="text-xs text-gray-500 mt-1 leading-relaxed">
            步骤完成 / 今日打卡时播放轻柔提示音（默认开启）。关闭则仅保留文字提示。
          </p>
        </div>
        <n-switch :value="soundOn" @update:value="onSoundToggle" />
      </div>
    </n-card>

    <n-card class="shadow-sm border-orange-100" title="例行步骤">
      <div class="space-y-3">
        <div
          v-for="s in steps"
          :key="s.key"
          class="flex flex-col gap-3 p-3 rounded-xl bg-orange-50/50 border border-orange-100 sm:flex-row sm:items-center sm:justify-between"
        >
          <div class="min-w-0">
            <div class="font-medium">{{ s.name }}</div>
            <n-text depth="3" class="text-xs">{{ s.key }}</n-text>
          </div>
          <div class="flex flex-wrap items-center gap-x-3 gap-y-2">
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-500">启用</span>
              <n-switch v-model:value="s.enabled" />
            </div>
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-500 whitespace-nowrap">时长（分钟）</span>
              <n-input-number
                :value="minutesOf(s)"
                :min="1"
                :max="120"
                size="medium"
                class="!w-[120px]"
                @update:value="(v) => setMinutes(s, v)"
              />
            </div>
          </div>
        </div>
      </div>
      <p class="text-xs text-gray-400 mt-4 leading-relaxed">
        默认顺序：静默 → 肯定语 → 观想 → 运动 → 阅读 → 书写。关闭的步骤不会出现在今日会话中。
      </p>
    </n-card>

    <n-button
      type="primary"
      size="large"
      class="!bg-orange-500 w-full sm:w-auto sm:min-w-[160px] sm:self-end"
      :loading="saving"
      @click="save"
    >
      保存
    </n-button>
  </div>
</template>

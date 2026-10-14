<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  NCard, NButton, NSpace, NTag, NProgress, NText, NSelect, NInput,
  NRate, NDivider, NAlert, useMessage,
} from 'naive-ui'
import { api } from '@/api/client'
import type { ContentItem, Session, SessionStep } from '@/types'

const message = useMessage()
const loading = ref(false)
const session = ref<Session | null>(null)
const date = ref('')
const contents = ref<ContentItem[]>([])
const activeKey = ref<string | null>(null)
const remaining = ref(0)
const elapsed = ref(0)
const timerId = ref<number | null>(null)
const journalBody = ref('')
const journalPrompt = ref('')
const mood = ref(0)
const note = ref('')
const selectedContent = ref<string | null>(null)

const stepMeta: Record<string, { emoji: string; hint: string }> = {
  silence: { emoji: '🧘', hint: '静坐、呼吸、感受当下' },
  affirmations: { emoji: '💬', hint: '大声或默念肯定语' },
  visualization: { emoji: '🌅', hint: '在脑中清晰预演理想画面' },
  exercise: { emoji: '🏃', hint: '活动身体，唤醒能量' },
  reading: { emoji: '📖', hint: '阅读成长类书籍或摘录' },
  scribing: { emoji: '✍️', hint: '写下感恩、目标与反思' },
}

const currentStep = computed(() =>
  session.value?.steps.find((s) => s.key === activeKey.value) || null,
)

const progress = computed(() => {
  if (!session.value?.steps.length) return 0
  const done = session.value.steps.filter((s) => s.status === 'completed' || s.status === 'skipped').length
  return Math.round((done / session.value.steps.length) * 100)
})

const contentOptions = computed(() =>
  contents.value
    .filter((c) => !currentStep.value || c.step_key === currentStep.value.key)
    .map((c) => ({ label: c.title, value: c.id })),
)

const activeContent = computed(() =>
  contents.value.find((c) => c.id === selectedContent.value) || null,
)

function formatTime(sec: number) {
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

async function loadToday() {
  loading.value = true
  try {
    const res = await api<{ date: string; session: Session | null }>('/sessions/today')
    date.value = res.date
    session.value = res.session
    if (res.session?.status === 'in_progress') {
      const active = res.session.steps.find((s) => s.status === 'active')
      const next = res.session.steps.find((s) => s.status === 'pending')
      if (active) await enterStep(active, false)
      else if (next) await enterStep(next, true)
    }
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function loadContent(stepKey?: string) {
  const q = stepKey ? `?step_key=${stepKey}` : ''
  contents.value = await api<ContentItem[]>(`/content${q}`)
}

async function startSession() {
  loading.value = true
  try {
    session.value = await api<Session>('/sessions/start', { method: 'POST' })
    message.success('今日会话已开始')
    const first = session.value.steps.find((s) => s.status === 'pending')
    if (first) await enterStep(first, true)
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '开始失败')
  } finally {
    loading.value = false
  }
}

function stopTimer() {
  if (timerId.value != null) {
    window.clearInterval(timerId.value)
    timerId.value = null
  }
}

function startTimer(seconds: number) {
  stopTimer()
  remaining.value = seconds
  elapsed.value = 0
  timerId.value = window.setInterval(() => {
    elapsed.value += 1
    if (remaining.value > 0) remaining.value -= 1
  }, 1000)
}

async function enterStep(step: SessionStep, markActive: boolean) {
  activeKey.value = step.key
  await loadContent(step.key)
  const defaults = contents.value.filter((c) => c.step_key === step.key)
  selectedContent.value = defaults[0]?.id || null
  if (step.key === 'scribing') {
    journalPrompt.value = defaults[0]?.body || '写下今天的感恩与目标'
    journalBody.value = ''
  }
  if (markActive && step.status === 'pending') {
    await api(`/sessions/${session.value!.id}/steps/${step.key}`, {
      method: 'PATCH',
      body: JSON.stringify({ status: 'active', content_id: selectedContent.value }),
    })
    step.status = 'active'
  }
  startTimer(step.planned_duration_seconds)
}

async function finishStep(status: 'completed' | 'skipped') {
  if (!session.value || !currentStep.value) return
  stopTimer()
  try {
    if (currentStep.value.key === 'scribing' && status === 'completed' && journalBody.value.trim()) {
      await api('/journals', {
        method: 'POST',
        body: JSON.stringify({
          session_id: session.value.id,
          prompt: journalPrompt.value,
          body: journalBody.value,
        }),
      })
    }
    await api(`/sessions/${session.value.id}/steps/${currentStep.value.key}`, {
      method: 'PATCH',
      body: JSON.stringify({
        status,
        actual_duration_seconds: elapsed.value,
        content_id: selectedContent.value,
      }),
    })
    currentStep.value.status = status
    currentStep.value.actual_duration_seconds = elapsed.value
    const next = session.value.steps.find((s) => s.status === 'pending')
    if (next) {
      await enterStep(next, true)
    } else {
      activeKey.value = null
      message.success('所有步骤已完成，可以结束今日会话')
    }
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '更新失败')
  }
}

async function completeSession() {
  if (!session.value) return
  stopTimer()
  try {
    session.value = await api<Session>(`/sessions/${session.value.id}/complete`, {
      method: 'POST',
      body: JSON.stringify({ mood: mood.value || null, note: note.value }),
    })
    activeKey.value = null
    message.success('太棒了！今日 SAVERS 已打卡')
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '完成失败')
  }
}

function stepStatusType(s: string) {
  if (s === 'completed') return 'success'
  if (s === 'skipped') return 'warning'
  if (s === 'active') return 'info'
  return 'default'
}

onMounted(loadToday)
onUnmounted(stopTimer)

watch(selectedContent, (id) => {
  const c = contents.value.find((x) => x.id === id)
  if (c && currentStep.value?.key === 'scribing') journalPrompt.value = c.body
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-800">今日会话</h1>
        <p class="text-gray-500 text-sm mt-1">{{ date || '—' }} · 按顺序完成 S.A.V.E.R.S.</p>
      </div>
      <n-tag v-if="session" :type="session.status === 'completed' ? 'success' : 'info'">
        {{ session.status === 'completed' ? '已完成' : '进行中' }}
      </n-tag>
    </div>

    <n-card v-if="!session" class="shadow-sm border-orange-100 overflow-hidden">
      <div class="text-center py-8 space-y-4">
        <div class="text-5xl">☀️</div>
        <h2 class="text-lg sm:text-xl font-semibold px-1">准备好开始你的奇迹早晨了吗？</h2>
        <p class="text-gray-500 max-w-md mx-auto">
          依次完成静默、肯定语、观想、运动、阅读、书写。每一步都有计时与引导内容。
        </p>
        <n-button type="primary" size="large" block :loading="loading" class="!bg-orange-500 max-w-sm mx-auto" @click="startSession">
          开始今日会话
        </n-button>
      </div>
    </n-card>

    <template v-else>
      <n-card class="shadow-sm border-orange-100">
        <div class="mb-3 flex justify-between text-sm text-gray-500">
          <span>进度</span><span>{{ progress }}%</span>
        </div>
        <n-progress type="line" :percentage="progress" :show-indicator="false" status="success" />
        <div class="mt-4 flex flex-wrap gap-2">
          <n-tag
            v-for="s in session.steps"
            :key="s.key"
            :type="stepStatusType(s.status)"
            :bordered="s.key === activeKey"
            class="cursor-default"
          >
            {{ stepMeta[s.key]?.emoji || '' }} {{ s.name }}
          </n-tag>
        </div>
      </n-card>

      <n-card v-if="session.status === 'completed'" class="shadow-sm border-green-100 bg-green-50/50">
        <div class="text-center py-6 space-y-2">
          <div class="text-4xl">🎉</div>
          <h2 class="text-xl font-semibold text-green-800">今日打卡完成</h2>
          <p class="text-gray-600">坚持就是奇迹。明天见！</p>
          <p v-if="session.note" class="text-sm text-gray-500">备注：{{ session.note }}</p>
        </div>
      </n-card>

      <n-card v-else-if="currentStep" class="shadow-sm border-orange-100">
        <div class="flex flex-col md:flex-row md:items-start gap-8">
          <div class="flex-1 space-y-4">
            <div>
              <div class="text-sm text-orange-600 font-medium">
                步骤 {{ currentStep.order }} / {{ session.steps.length }}
              </div>
              <h2 class="text-2xl font-bold mt-1">
                {{ stepMeta[currentStep.key]?.emoji }} {{ currentStep.name }}
              </h2>
              <p class="text-gray-500 text-sm mt-1">{{ stepMeta[currentStep.key]?.hint }}</p>
            </div>

            <div class="rounded-2xl bg-gradient-to-br from-orange-50 to-amber-50 border border-orange-100 p-4 sm:p-6 text-center">
              <div class="text-4xl sm:text-5xl font-mono font-bold text-orange-700 tracking-wider">
                {{ formatTime(remaining) }}
              </div>
              <div class="text-xs text-gray-500 mt-2">
                计划 {{ formatTime(currentStep.planned_duration_seconds) }} · 已进行 {{ formatTime(elapsed) }}
              </div>
              <n-progress
                class="mt-4"
                type="line"
                :percentage="Math.min(100, Math.round((elapsed / currentStep.planned_duration_seconds) * 100))"
                :show-indicator="false"
              />
            </div>

            <div class="flex flex-col sm:flex-row gap-2 sm:gap-3">
              <n-button type="primary" class="!bg-orange-500 flex-1" size="large" @click="finishStep('completed')">完成此步</n-button>
              <n-button class="flex-1" size="large" @click="finishStep('skipped')">跳过</n-button>
            </div>
          </div>

          <div class="flex-1 space-y-3">
            <div class="font-medium text-gray-700">引导内容</div>
            <n-select
              v-if="contentOptions.length"
              v-model:value="selectedContent"
              :options="contentOptions"
              placeholder="选择引导文案"
            />
            <div v-if="activeContent" class="rounded-xl bg-white border border-orange-100 p-4 whitespace-pre-wrap text-gray-700 leading-relaxed text-sm">
              <div class="font-semibold mb-2">{{ activeContent.title }}</div>
              {{ activeContent.body }}
            </div>
            <n-alert v-else type="info" title="暂无内容">可在设置中调整时长，系统已内置默认引导。</n-alert>

            <template v-if="currentStep.key === 'scribing'">
              <n-divider />
              <div class="font-medium">今日书写</div>
              <n-input
                v-model:value="journalBody"
                type="textarea"
                :rows="8"
                placeholder="在这里写下你的想法…"
              />
            </template>
          </div>
        </div>
      </n-card>

      <n-card v-else-if="session.status === 'in_progress'" class="shadow-sm border-orange-100">
        <h3 class="font-semibold mb-3">结束今日会话</h3>
        <p class="text-sm text-gray-500 mb-4">所有步骤已处理完毕，给今天打个分并保存备注吧。</p>
        <div class="space-y-3 max-w-md">
          <div>
            <div class="text-sm mb-1">今日心情</div>
            <n-rate v-model:value="mood" />
          </div>
          <n-input v-model:value="note" type="textarea" :rows="3" placeholder="可选备注" />
          <n-button type="primary" class="!bg-orange-500" @click="completeSession">完成打卡</n-button>
        </div>
      </n-card>
    </template>
  </div>
</template>

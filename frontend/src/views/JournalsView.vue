<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NCard, NEmpty, NSpin, useMessage } from 'naive-ui'
import { api } from '@/api/client'
import type { Journal } from '@/types'

const message = useMessage()
const loading = ref(true)
const journals = ref<Journal[]>([])

async function load() {
  loading.value = true
  try {
    journals.value = await api<Journal[]>('/journals')
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-xl sm:text-2xl font-bold text-gray-800">书写日记</h1>
      <p class="text-gray-500 text-sm mt-1">来自 Scribing 步骤的记录</p>
    </div>
    <n-spin :show="loading">
      <n-empty v-if="!journals.length && !loading" description="还没有日记，完成一次书写步骤吧" />
      <div class="space-y-4">
        <n-card v-for="j in journals" :key="j.id" class="shadow-sm border-orange-100">
          <div class="text-xs text-gray-400 mb-2">{{ new Date(j.created_at).toLocaleString('zh-CN') }}</div>
          <div v-if="j.prompt" class="text-sm text-orange-700 whitespace-pre-wrap mb-3 bg-orange-50 p-3 rounded-lg">
            {{ j.prompt }}
          </div>
          <div class="whitespace-pre-wrap text-gray-800 leading-relaxed">{{ j.body || '（空）' }}</div>
        </n-card>
      </div>
    </n-spin>
  </div>
</template>

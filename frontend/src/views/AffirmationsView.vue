<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  NCard, NButton, NSpace, NTag, NInput, NModal, NForm, NFormItem,
  NEmpty, NSpin, NPopconfirm, useMessage,
} from 'naive-ui'
import { api } from '@/api/client'
import type { ContentItem } from '@/types'

const message = useMessage()
const loading = ref(true)
const items = ref<ContentItem[]>([])
const showForm = ref(false)
const saving = ref(false)
const editingId = ref<string | null>(null)
const formTitle = ref('')
const formBody = ref('')

const mine = computed(() => items.value.filter((i) => !!i.user_id))
const system = computed(() => items.value.filter((i) => !i.user_id))

async function load() {
  loading.value = true
  try {
    items.value = await api<ContentItem[]>('/content?step_key=affirmations')
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function openCreate(prefill?: { title?: string; body?: string }) {
  editingId.value = null
  formTitle.value = prefill?.title || ''
  formBody.value = prefill?.body || ''
  showForm.value = true
}

function openEdit(item: ContentItem) {
  if (!item.user_id) {
    message.warning('系统文案请先「复制为我的」再编辑')
    return
  }
  editingId.value = item.id
  formTitle.value = item.title
  formBody.value = item.body
  showForm.value = true
}

async function save() {
  const title = formTitle.value.trim()
  const body = formBody.value.trim()
  if (!title || !body) {
    message.warning('请填写标题和正文')
    return
  }
  saving.value = true
  try {
    if (editingId.value) {
      await api(`/content/${editingId.value}`, {
        method: 'PUT',
        body: JSON.stringify({ title, body }),
      })
      message.success('已更新肯定语')
    } else {
      await api('/content', {
        method: 'POST',
        body: JSON.stringify({ step_key: 'affirmations', title, body }),
      })
      message.success('已添加肯定语')
    }
    showForm.value = false
    await load()
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}

async function remove(item: ContentItem) {
  try {
    await api(`/content/${item.id}`, { method: 'DELETE' })
    message.success('已删除')
    await load()
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '删除失败')
  }
}

async function cloneItem(item: ContentItem) {
  try {
    await api(`/content/${item.id}/clone`, { method: 'POST' })
    message.success('已复制为我的肯定语，可继续编辑')
    await load()
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '复制失败')
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-800">肯定语</h1>
        <p class="text-gray-500 text-sm mt-1">
          管理系统与个人肯定语。今日会话的「肯定语」步骤会列出你创建的文案。
        </p>
      </div>
      <n-button type="primary" class="!bg-orange-500 w-full sm:w-auto" @click="openCreate()">新建肯定语</n-button>
    </div>

    <n-spin :show="loading">
      <n-card class="shadow-sm border-orange-100 mb-6" title="我的肯定语">
        <n-empty v-if="!mine.length && !loading" description="还没有个人肯定语，点击右上角新建，或复制系统文案" />
        <div class="space-y-3">
          <div
            v-for="item in mine"
            :key="item.id"
            class="rounded-xl border border-orange-100 bg-orange-50/40 p-4"
          >
            <div class="flex flex-wrap items-start justify-between gap-2">
              <div class="min-w-0 flex-1">
                <div class="font-semibold text-gray-800">{{ item.title }}</div>
                <div class="mt-2 text-sm text-gray-700 whitespace-pre-wrap leading-relaxed">{{ item.body }}</div>
              </div>
              <n-space>
                <n-tag size="small" type="success">我的</n-tag>
                <n-button size="small" @click="openEdit(item)">编辑</n-button>
                <n-popconfirm @positive-click="remove(item)">
                  <template #trigger>
                    <n-button size="small" type="error" quaternary>删除</n-button>
                  </template>
                  确定删除这条肯定语？
                </n-popconfirm>
              </n-space>
            </div>
          </div>
        </div>
      </n-card>

      <n-card class="shadow-sm border-orange-100" title="系统内置（只读）">
        <p class="text-xs text-gray-400 mb-3">可复制为我的副本后再修改，原系统文案不会被改动。</p>
        <div class="space-y-3">
          <div
            v-for="item in system"
            :key="item.id"
            class="rounded-xl border border-gray-100 bg-white p-4"
          >
            <div class="flex flex-wrap items-start justify-between gap-2">
              <div class="min-w-0 flex-1">
                <div class="font-semibold text-gray-800">{{ item.title }}</div>
                <div class="mt-2 text-sm text-gray-700 whitespace-pre-wrap leading-relaxed">{{ item.body }}</div>
              </div>
              <n-space>
                <n-tag size="small">系统</n-tag>
                <n-button size="small" type="primary" ghost @click="cloneItem(item)">复制为我的</n-button>
              </n-space>
            </div>
          </div>
        </div>
      </n-card>
    </n-spin>

    <n-modal
      v-model:show="showForm"
      preset="card"
      :title="editingId ? '编辑肯定语' : '新建肯定语'"
      class="w-[calc(100vw-1.5rem)] max-w-lg"
      :style="{ margin: '0.75rem' }"
      :mask-closable="false"
    >
      <n-form>
        <n-form-item label="标题">
          <n-input v-model:value="formTitle" placeholder="例如：自信与成长" maxlength="80" show-count />
        </n-form-item>
        <n-form-item label="正文">
          <n-input
            v-model:value="formBody"
            type="textarea"
            :rows="6"
            placeholder="写下你要每天对自己说的话…"
            maxlength="2000"
            show-count
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showForm = false">取消</n-button>
          <n-button type="primary" class="!bg-orange-500" :loading="saving" @click="save">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

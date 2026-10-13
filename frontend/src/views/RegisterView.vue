<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const message = useMessage()

const username = ref('')
const email = ref('')
const password = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    await auth.register(username.value.trim(), email.value.trim(), password.value)
    message.success('注册成功，已为你创建默认 SAVERS 例程')
    router.push('/today')
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-[80vh] flex items-center justify-center">
    <n-card class="w-full max-w-md shadow-lg border-orange-100" title="注册 · 开启你的奇迹早晨">
      <n-form @submit.prevent="submit">
        <n-form-item label="用户名">
          <n-input v-model:value="username" placeholder="例如 miracle" />
        </n-form-item>
        <n-form-item label="邮箱">
          <n-input v-model:value="email" placeholder="you@example.com" />
        </n-form-item>
        <n-form-item label="密码（至少 6 位）">
          <n-input v-model:value="password" type="password" show-password-on="click" />
        </n-form-item>
        <n-button type="primary" block :loading="loading" @click="submit" class="!bg-orange-500">
          创建账号
        </n-button>
      </n-form>
      <div class="mt-4 text-center text-sm text-gray-500">
        已有账号？
        <router-link to="/login" class="text-orange-600 hover:underline">去登录</router-link>
      </div>
    </n-card>
  </div>
</template>

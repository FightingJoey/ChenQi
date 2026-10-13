<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, NText, useMessage } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const message = useMessage()

const username = ref('')
const password = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    await auth.login(username.value.trim(), password.value)
    message.success('登录成功')
    const redirect = (route.query.redirect as string) || '/today'
    router.push(redirect)
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-[80vh] flex items-center justify-center">
    <n-card class="w-full max-w-md shadow-lg border-orange-100" title="登录 · 晨启">
      <n-form @submit.prevent="submit">
        <n-form-item label="用户名或邮箱">
          <n-input v-model:value="username" placeholder="用户名 / 邮箱" @keyup.enter="submit" />
        </n-form-item>
        <n-form-item label="密码">
          <n-input v-model:value="password" type="password" show-password-on="click" placeholder="密码" @keyup.enter="submit" />
        </n-form-item>
        <n-button type="primary" block :loading="loading" @click="submit" class="!bg-orange-500">
          登录
        </n-button>
      </n-form>
      <div class="mt-4 text-center text-sm text-gray-500">
        还没有账号？
        <router-link to="/register" class="text-orange-600 hover:underline">立即注册</router-link>
      </div>
      <n-text depth="3" class="block mt-6 text-xs leading-relaxed">
        晨启帮你践行《早起的奇迹》S.A.V.E.R.S.：静默、肯定语、观想、运动、阅读、书写。
      </n-text>
    </n-card>
  </div>
</template>

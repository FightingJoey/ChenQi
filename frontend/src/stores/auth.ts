import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, getToken, setToken } from '@/api/client'
import type { User } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(getToken())
  const ready = ref(false)

  const isAuthenticated = computed(() => !!token.value)

  async function bootstrap() {
    if (!token.value) {
      ready.value = true
      return
    }
    try {
      user.value = await api<User>('/auth/me')
    } catch {
      setToken(null)
      token.value = null
      user.value = null
    } finally {
      ready.value = true
    }
  }

  async function register(username: string, email: string, password: string) {
    const res = await api<{ token: string; user: User }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, email, password }),
    })
    setToken(res.token)
    token.value = res.token
    user.value = res.user
  }

  async function login(username: string, password: string) {
    const res = await api<{ token: string; user: User }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
    setToken(res.token)
    token.value = res.token
    user.value = res.user
  }

  function logout() {
    setToken(null)
    token.value = null
    user.value = null
  }

  return { user, token, ready, isAuthenticated, bootstrap, register, login, logout }
})

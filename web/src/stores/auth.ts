import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

const tokenKey = 'devkit.auth.token'
const userKey = 'devkit.auth.user'

interface AuthUser {
  email: string
  displayName: string
  avatarUrl: string
}

interface LoginPayload extends AuthUser {
  token: string
}

function savedToken(): string | null {
  return typeof localStorage === 'undefined' ? null : localStorage.getItem(tokenKey)
}

function savedUser(): AuthUser | null {
  if (typeof localStorage === 'undefined') {
    return null
  }

  const value = localStorage.getItem(userKey)
  if (!value) {
    return null
  }

  try {
    return JSON.parse(value) as AuthUser
  } catch {
    localStorage.removeItem(userKey)
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(savedToken())
  const user = ref<AuthUser | null>(savedUser())
  const isAuthenticated = computed(() => token.value !== null)

  function login(payload: LoginPayload) {
    const { token: newToken, ...userInfo } = payload
    token.value = newToken
    user.value = userInfo
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(tokenKey, newToken)
      localStorage.setItem(userKey, JSON.stringify(userInfo))
    }
  }

  function logout() {
    token.value = null
    user.value = null
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem(tokenKey)
      localStorage.removeItem(userKey)
    }
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    logout,
  }
})

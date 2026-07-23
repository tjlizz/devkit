import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

const tokenKey = 'devkit.auth.token'

function savedToken(): string | null {
  return typeof localStorage === 'undefined' ? null : localStorage.getItem(tokenKey)
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(savedToken())
  const isAuthenticated = computed(() => token.value !== null)

  function login(newToken: string) {
    token.value = newToken
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(tokenKey, newToken)
    }
  }

  function logout() {
    token.value = null
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem(tokenKey)
    }
  }

  return {
    token,
    isAuthenticated,
    login,
    logout,
  }
})

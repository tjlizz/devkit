import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from './auth'

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('tracks authentication state', () => {
    const auth = useAuthStore()

    expect(auth.isAuthenticated).toBe(false)
    auth.login({
      token: 'example-token',
      id: 1,
      email: 'developer@example.com',
      displayName: 'Developer',
      avatarUrl: 'https://example.com/avatar.png',
      role: 'admin',
      isDeveloper: false,
    })
    expect(auth.isAuthenticated).toBe(true)
    expect(auth.user?.email).toBe('developer@example.com')
    auth.updateUser({
      id: 1,
      email: 'developer@example.com',
      displayName: 'Developer',
      avatarUrl: 'https://example.com/custom-avatar.png',
      role: 'admin',
      isDeveloper: false,
    })
    expect(auth.user?.avatarUrl).toBe('https://example.com/custom-avatar.png')
    auth.logout()
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
  })
})

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
      email: 'developer@example.com',
      displayName: 'Developer',
      avatarUrl: 'https://example.com/avatar.png',
    })
    expect(auth.isAuthenticated).toBe(true)
    expect(auth.user?.email).toBe('developer@example.com')
    auth.logout()
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
  })
})

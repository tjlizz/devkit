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
    auth.login('example-token')
    expect(auth.isAuthenticated).toBe(true)
    auth.logout()
    expect(auth.isAuthenticated).toBe(false)
  })
})

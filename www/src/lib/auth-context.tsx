"use client"

import { createContext, useContext, useCallback, useEffect, useState, type ReactNode } from "react"

// Empty by default: requests go same-origin to /api/... and Next.js rewrites
// proxy them to the Go API (see next.config.ts), avoiding CORS entirely.
const API_BASE = process.env.NEXT_PUBLIC_API_URL || ""
const TOKEN_KEY = "devkit.auth.token"
const USER_KEY = "devkit.auth.user"

export interface AuthUser {
  id: number
  email: string
  displayName: string
  avatarUrl: string
  role: "user" | "admin"
  isDeveloper: boolean
  developerApplicationStatus?: "pending" | "approved" | "rejected"
}

interface AuthContextValue {
  user: AuthUser | null
  token: string | null
  loading: boolean
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, displayName: string) => Promise<void>
  forgotPassword: (email: string) => Promise<void>
  resetPassword: (token: string, newPassword: string) => Promise<void>
  changePassword: (oldPassword: string, newPassword: string) => Promise<void>
  updateAvatar: (avatar: File) => Promise<AuthUser>
  resetAvatar: () => Promise<AuthUser>
  logout: () => void
  refreshUser: () => Promise<AuthUser>
  upgradeToDeveloper: (payload: { displayName: string; profileUrl: string; reason: string }) => Promise<void>
  submitApp: (payload: AppPublishPayload) => Promise<void>
}

export interface AppPublishPayload {
  name: string
  slug: string
  tagline: string
  description: string
  category: string
  priceCents: number
  currency: string
  iconUrl: string
  coverImageUrl: string
  demoUrl: string
  docsUrl: string
  sourceUrl: string
  supportUrl: string
  tags: string[]
  version: string
  releaseNotes: string
}

const AuthContext = createContext<AuthContextValue | null>(null)

function getStoredToken(): string | null {
  if (typeof window === "undefined") return null
  return localStorage.getItem(TOKEN_KEY)
}

function getStoredUser(): AuthUser | null {
  if (typeof window === "undefined") return null
  try {
    const raw = localStorage.getItem(USER_KEY)
    return raw ? JSON.parse(raw) : null
  } catch {
    localStorage.removeItem(USER_KEY)
    return null
  }
}

function setStored(token: string, user: AuthUser) {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

function clearStored() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setToken(getStoredToken())
    setUser(getStoredUser())
    setLoading(false)
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const res = await fetch(`${API_BASE}/api/v1/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Login failed" }))
      throw { status: res.status, message: body.error || "Login failed" }
    }
    const data = await res.json()
    setToken(data.token)
    setUser(data.user)
    setStored(data.token, data.user)
  }, [])

  const register = useCallback(async (email: string, password: string, displayName: string) => {
    const res = await fetch(`${API_BASE}/api/v1/auth/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password, displayName }),
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Registration failed" }))
      throw { status: res.status, message: body.error || "Registration failed" }
    }
  }, [])

  const forgotPassword = useCallback(async (email: string) => {
    const res = await fetch(`${API_BASE}/api/v1/auth/forgot-password`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email }),
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Password reset request failed" }))
      throw { status: res.status, message: body.error || "Password reset request failed" }
    }
  }, [])

  const resetPassword = useCallback(async (token: string, newPassword: string) => {
    const res = await fetch(`${API_BASE}/api/v1/auth/reset-password`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token, newPassword }),
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Password reset failed" }))
      throw { status: res.status, message: body.error || "Password reset failed" }
    }
  }, [])

  const changePassword = useCallback(async (oldPassword: string, newPassword: string) => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/auth/change-password`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${currentToken}`,
      },
      body: JSON.stringify({ oldPassword, newPassword }),
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Password change failed" }))
      throw { status: res.status, message: body.error || "Password change failed" }
    }
  }, [token])

  const updateAvatar = useCallback(async (avatar: File) => {
    const currentToken = token || getStoredToken()
    const body = new FormData()
    body.append("avatar", avatar)
    const res = await fetch(`${API_BASE}/api/v1/auth/me/avatar`, {
      method: "PATCH",
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
      body,
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Avatar update failed" }))
      throw { status: res.status, message: body.error || "Avatar update failed" }
    }
    const data = await res.json()
    setUser(data.user)
    if (currentToken) {
      setStored(currentToken, data.user)
    }
    return data.user as AuthUser
  }, [token])

  const resetAvatar = useCallback(async () => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/auth/me/avatar`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Avatar reset failed" }))
      throw { status: res.status, message: body.error || "Avatar reset failed" }
    }
    const data = await res.json()
    setUser(data.user)
    if (currentToken) {
      setStored(currentToken, data.user)
    }
    return data.user as AuthUser
  }, [token])

  const logout = useCallback(() => {
    setToken(null)
    setUser(null)
    clearStored()
  }, [])

  const refreshUser = useCallback(async () => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/auth/me`, {
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Could not refresh user" }))
      throw { status: res.status, message: body.error || "Could not refresh user" }
    }
    const data = await res.json()
    setUser(data.user)
    if (currentToken) {
      setStored(currentToken, data.user)
    }
    return data.user as AuthUser
  }, [token])

  const upgradeToDeveloper = useCallback(async (payload: { displayName: string; profileUrl: string; reason: string }) => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/auth/upgrade-to-developer`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${currentToken}`,
      },
      body: JSON.stringify(payload),
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Upgrade failed" }))
      throw { status: res.status, message: body.error || "Upgrade failed" }
    }
    await refreshUser()
  }, [refreshUser, token])

  const submitApp = useCallback(async (payload: AppPublishPayload) => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/apps`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${currentToken}`,
      },
      body: JSON.stringify(payload),
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "App submission failed" }))
      throw { status: res.status, message: body.error || "App submission failed" }
    }
  }, [token])

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        loading,
        isAuthenticated: token !== null,
        login,
        register,
        forgotPassword,
        resetPassword,
        changePassword,
        updateAvatar,
        resetAvatar,
        logout,
        refreshUser,
        upgradeToDeveloper,
        submitApp,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error("useAuth must be used within AuthProvider")
  return ctx
}

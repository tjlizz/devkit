"use client"

import { createContext, useContext, useCallback, useEffect, useState, type ReactNode } from "react"

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"
const TOKEN_KEY = "devkit.auth.token"
const USER_KEY = "devkit.auth.user"

export interface AuthUser {
  id: number
  email: string
  displayName: string
  avatarUrl: string
}

interface AuthContextValue {
  user: AuthUser | null
  token: string | null
  loading: boolean
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, displayName: string) => Promise<void>
  logout: () => void
  upgradeToDeveloper: () => Promise<void>
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

  const logout = useCallback(() => {
    setToken(null)
    setUser(null)
    clearStored()
  }, [])

  const upgradeToDeveloper = useCallback(async () => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/auth/upgrade-to-developer`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${currentToken}`,
      },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Upgrade failed" }))
      throw { status: res.status, message: body.error || "Upgrade failed" }
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
        logout,
        upgradeToDeveloper,
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

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
  listDeveloperApps: () => Promise<DeveloperApp[]>
  getDeveloperApp: (id: number) => Promise<DeveloperApp>
  listDeveloperSales: () => Promise<{ sales: DeveloperSale[]; summary: DeveloperSalesSummary }>
  updateDeveloperApp: (id: number, payload: AppPublishPayload) => Promise<DeveloperApp>
  delistDeveloperApp: (id: number) => Promise<DeveloperApp>
  checkoutApp: (slug: string, planId?: number) => Promise<Order>
  confirmPayment: (orderId: number) => Promise<{ order: Order; entitlement: Entitlement }>
  refundOrder: (orderId: number) => Promise<Order>
  listMyOrders: () => Promise<Order[]>
  listMyEntitlements: () => Promise<Entitlement[]>
  getDelivery: (entitlementId: number) => Promise<Delivery>
  listAppArtifacts: (appId: number) => Promise<AppArtifact[]>
  uploadAppArtifact: (appId: number, file: File) => Promise<AppArtifact>
  deleteAppArtifact: (appId: number, artifactId: number) => Promise<void>
}

export interface Order {
  id: number
  buyerId: number
  appId: number
  appSlug: string
  appName: string
  planId?: number
  planName: string
  priceCents: number
  currency: string
  status: "pending" | "paid" | "refunded" | "cancelled"
  provider: string
  providerEventId: string
  paidAt?: string
  createdAt: string
  updatedAt: string
}

export interface Entitlement {
  id: number
  buyerId: number
  appId: number
  appSlug: string
  appName: string
  planId?: number
  planName: string
  orderId: number
  status: "active" | "revoked"
  grantedAt: string
  createdAt: string
  updatedAt: string
}

export interface AppArtifact {
  id: number
  appId: number
  fileName: string
  sizeBytes: number
  contentType: string
  checksumSha256: string
  createdAt: string
}

export interface Delivery {
  entitlementId: number
  appSlug: string
  appName: string
  version: string
  sourceUrl: string
  docsUrl: string
  demoUrl: string
  artifacts: AppArtifact[]
  deliveryToken: string
  expiresAt: string
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
  plans: AppPlan[]
}

export interface AppPlan {
  name: string
  priceCents: number
  currency: string
  description: string
  features: string[]
}

export interface DeveloperApp extends AppPublishPayload {
  id: number
  developerId: number
  developerName: string
  status: "pending_review" | "approved" | "rejected" | "delisted"
  reviewNote: string
  reviewedBy?: number
  reviewedAt?: string
  publishedAt?: string
  createdAt: string
  updatedAt: string
}

export interface DeveloperSale {
  orderId: number
  appId: number
  appSlug: string
  appName: string
  planName: string
  buyerEmail: string
  priceCents: number
  currency: string
  provider: string
  paidAt?: string
  createdAt: string
}

export interface DeveloperSalesSummary {
  totalOrders: number
  totalRevenueCents: number
  uniqueBuyers: number
  appsSold: number
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

  const getDeveloperApp = useCallback(async (id: number) => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/developer/apps/${id}`, {
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Could not load app" }))
      throw { status: res.status, message: body.error || "Could not load app" }
    }
    const data = await res.json()
    return data.app as DeveloperApp
  }, [token])

  const listDeveloperApps = useCallback(async () => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/developer/apps`, {
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Could not load apps" }))
      throw { status: res.status, message: body.error || "Could not load apps" }
    }
    const data = await res.json()
    return data.apps as DeveloperApp[]
  }, [token])

  const listDeveloperSales = useCallback(async () => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/developer/sales`, {
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Could not load sales" }))
      throw { status: res.status, message: body.error || "Could not load sales" }
    }
    const data = await res.json()
    return {
      sales: data.sales as DeveloperSale[],
      summary: data.summary as DeveloperSalesSummary,
    }
  }, [token])

  const updateDeveloperApp = useCallback(async (id: number, payload: AppPublishPayload) => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/developer/apps/${id}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${currentToken}`,
      },
      body: JSON.stringify(payload),
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Could not update app" }))
      throw { status: res.status, message: body.error || "Could not update app" }
    }
    const data = await res.json()
    return data.app as DeveloperApp
  }, [token])

  const delistDeveloperApp = useCallback(async (id: number) => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/developer/apps/${id}/delist`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Could not delist app" }))
      throw { status: res.status, message: body.error || "Could not delist app" }
    }
    const data = await res.json()
    return data.app as DeveloperApp
  }, [token])

  const checkoutApp = useCallback(async (slug: string, planId?: number) => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/marketplace/apps/${encodeURIComponent(slug)}/checkout`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${currentToken}`,
      },
      body: JSON.stringify(planId ? { planId } : {}),
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Checkout failed" }))
      throw { status: res.status, message: body.error || "Checkout failed" }
    }
    const data = await res.json()
    return data.order as Order
  }, [token])

  const confirmPayment = useCallback(async (orderId: number) => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/orders/${orderId}/confirm-payment`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Payment confirmation failed" }))
      throw { status: res.status, message: body.error || "Payment confirmation failed" }
    }
    const data = await res.json()
    return data as { order: Order; entitlement: Entitlement }
  }, [token])

  const refundOrder = useCallback(async (orderId: number) => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/orders/${orderId}/refund`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Refund failed" }))
      throw { status: res.status, message: body.error || "Refund failed" }
    }
    const data = await res.json()
    return data.order as Order
  }, [token])

  const listMyOrders = useCallback(async () => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/me/orders`, {
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Could not load orders" }))
      throw { status: res.status, message: body.error || "Could not load orders" }
    }
    const data = await res.json()
    return data.orders as Order[]
  }, [token])

  const listMyEntitlements = useCallback(async () => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/me/entitlements`, {
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Could not load entitlements" }))
      throw { status: res.status, message: body.error || "Could not load entitlements" }
    }
    const data = await res.json()
    return data.entitlements as Entitlement[]
  }, [token])

  const getDelivery = useCallback(async (entitlementId: number) => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/entitlements/${entitlementId}/delivery`, {
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Could not get delivery" }))
      throw { status: res.status, message: body.error || "Could not get delivery" }
    }
    return (await res.json()) as Delivery
  }, [token])

  const listAppArtifacts = useCallback(async (appId: number) => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/developer/apps/${appId}/artifacts`, {
      headers: { Authorization: `Bearer ${currentToken}` },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Could not list artifacts" }))
      throw { status: res.status, message: body.error || "Could not list artifacts" }
    }
    const data = await res.json()
    return data.artifacts as AppArtifact[]
  }, [token])

  const uploadAppArtifact = useCallback(async (appId: number, file: File) => {
    const currentToken = token || getStoredToken()
    const form = new FormData()
    form.append("artifact", file)
    const res = await fetch(`${API_BASE}/api/v1/developer/apps/${appId}/artifacts`, {
      method: "POST",
      headers: { Authorization: `Bearer ${currentToken}` },
      body: form,
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Could not upload artifact" }))
      throw { status: res.status, message: body.error || "Could not upload artifact" }
    }
    const data = await res.json()
    return data.artifact as AppArtifact
  }, [token])

  const deleteAppArtifact = useCallback(async (appId: number, artifactId: number) => {
    const currentToken = token || getStoredToken()
    const res = await fetch(`${API_BASE}/api/v1/developer/apps/${appId}/artifacts/${artifactId}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${currentToken}` },
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: "Could not delete artifact" }))
      throw { status: res.status, message: body.error || "Could not delete artifact" }
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
        listDeveloperApps,
        getDeveloperApp,
        updateDeveloperApp,
        delistDeveloperApp,
        listDeveloperSales,
        checkoutApp,
        confirmPayment,
        refundOrder,
        listMyOrders,
        listMyEntitlements,
        getDelivery,
        listAppArtifacts,
        uploadAppArtifact,
        deleteAppArtifact,
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

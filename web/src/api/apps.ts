import apiClient from './client'

export interface MarketplaceApp {
  id: number
  developerId: number
  developerName: string
  developerEmail?: string
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
  status: 'pending_review' | 'approved' | 'rejected'
  reviewNote: string
  reviewedBy?: number
  reviewedAt?: string
  publishedAt?: string
  createdAt: string
  updatedAt: string
}

export interface AdminAppsResponse {
  apps: MarketplaceApp[]
  limit: number
  offset: number
}

export interface AppResponse {
  app: MarketplaceApp
}

export async function listAdminApps(status: string) {
  const response = await apiClient.get<AdminAppsResponse>('/admin/apps', {
    params: { status },
  })
  return response.data
}

export async function approveApp(id: number, reviewNote: string) {
  const response = await apiClient.post<AppResponse>(`/admin/apps/${id}/approve`, {
    reviewNote,
  })
  return response.data
}

export async function rejectApp(id: number, reviewNote: string) {
  const response = await apiClient.post<AppResponse>(`/admin/apps/${id}/reject`, {
    reviewNote,
  })
  return response.data
}

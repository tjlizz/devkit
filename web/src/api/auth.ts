import apiClient from './client'

export interface AuthUser {
  id: number
  email: string
  displayName: string
  avatarUrl: string
  role: 'user' | 'admin'
  isDeveloper: boolean
  developerApplicationStatus?: 'pending' | 'approved' | 'rejected'
}

export interface DeveloperApplication {
  id: number
  userId: number
  email: string
  displayName: string
  profileUrl: string
  reason: string
  status: 'pending' | 'approved' | 'rejected'
  reviewNote: string
  reviewedBy?: number
  reviewedAt?: string
  createdAt: string
  updatedAt: string
}

export interface RegisterResponse {
  message: string
  user: AuthUser
}

export interface LoginResponse {
  token: string
  expiresAt: string
  user: AuthUser
}

export interface UserResponse {
  user: AuthUser
}

export interface ActivateResponse {
  status: 'activated'
}

export interface MessageResponse {
  message: string
}

export interface StatusResponse {
  status: string
}

export interface DeveloperApplicationsResponse {
  applications: DeveloperApplication[]
}

export interface DeveloperApplicationResponse {
  application: DeveloperApplication
}

export async function register(email: string, password: string, displayName: string) {
  const response = await apiClient.post<RegisterResponse>('/auth/register', {
    email,
    password,
    displayName,
  })
  return response.data
}

export async function login(email: string, password: string) {
  const response = await apiClient.post<LoginResponse>('/auth/login', {
    email,
    password,
  })
  return response.data
}

export async function activate(token: string) {
  const response = await apiClient.get<ActivateResponse>('/auth/activate', {
    params: { token },
  })
  return response.data
}

export async function forgotPassword(email: string) {
  const response = await apiClient.post<MessageResponse>('/auth/forgot-password', {
    email,
  })
  return response.data
}

export async function resetPassword(token: string, newPassword: string) {
  const response = await apiClient.post<StatusResponse>('/auth/reset-password', {
    token,
    newPassword,
  })
  return response.data
}

export async function changePassword(oldPassword: string, newPassword: string) {
  const response = await apiClient.post<StatusResponse>('/auth/change-password', {
    oldPassword,
    newPassword,
  })
  return response.data
}

export async function updateAvatar(avatar: File) {
  const body = new FormData()
  body.append('avatar', avatar)
  const response = await apiClient.patch<UserResponse>('/auth/me/avatar', body, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
  return response.data
}

export async function resetAvatar() {
  const response = await apiClient.delete<UserResponse>('/auth/me/avatar')
  return response.data
}

export async function listDeveloperApplications() {
  const response = await apiClient.get<DeveloperApplicationsResponse>(
    '/admin/developer-applications',
  )
  return response.data
}

export async function approveDeveloperApplication(id: number, reviewNote: string) {
  const response = await apiClient.post<DeveloperApplicationResponse>(
    `/admin/developer-applications/${id}/approve`,
    { reviewNote },
  )
  return response.data
}

export async function rejectDeveloperApplication(id: number, reviewNote: string) {
  const response = await apiClient.post<DeveloperApplicationResponse>(
    `/admin/developer-applications/${id}/reject`,
    { reviewNote },
  )
  return response.data
}

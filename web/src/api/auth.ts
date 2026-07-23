import apiClient from './client'

export interface AuthUser {
  id: number
  email: string
  displayName: string
  avatarUrl: string
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

export interface ActivateResponse {
  status: 'activated'
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

import client from './client'
import type { ApiResponse, LoginToken, CaptchaInfo, AuthUser, LoginRequest, RegisterRequest, UpdateCurrentUser } from '../types'

export async function getCaptchaId() {
  const res = await client.get<ApiResponse<CaptchaInfo>>('/captcha/id')
  return res.data
}

export async function login(data: LoginRequest) {
  const res = await client.post<ApiResponse<LoginToken>>('/login', data)
  return res.data
}

export async function register(data: RegisterRequest) {
  const res = await client.post<ApiResponse<LoginToken>>('/register', data)
  return res.data
}

export async function refreshToken() {
  const res = await client.post<ApiResponse<LoginToken>>('/current/refresh-token')
  return res.data
}

export async function getCurrentUser() {
  const res = await client.get<ApiResponse<AuthUser>>('/current/user')
  return res.data
}

export async function updateCurrentUser(data: UpdateCurrentUser) {
  const res = await client.put<ApiResponse<null>>('/current/user', data)
  return res.data
}

export async function uploadAvatar(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('category', 'avatar')
  const res = await client.post<ApiResponse<{ id: string; url: string }>>('/images/upload', formData)
  return res.data
}

export async function changePassword(data: { old_password: string; new_password: string }) {
  const res = await client.put<ApiResponse<null>>('/current/password', data)
  return res.data
}

export async function logout() {
  const res = await client.post<ApiResponse<null>>('/current/logout')
  return res.data
}

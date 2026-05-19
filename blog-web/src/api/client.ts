import axios from 'axios'
import { message } from 'antd'
import type { ApiResponse, LoginToken } from '../types'

const TOKEN_KEY = 'blog-token'
const USER_KEY = 'blog-user'
const storage = sessionStorage

const AUTH_PATHS = ['/login', '/register', '/captcha', '/current/refresh-token', '/public/subscribe']

// 账号已被封禁 — 强制踢下线
function handleAccountDisabled(detail: string) {
  storage.removeItem(TOKEN_KEY)
  storage.removeItem(USER_KEY)
  isRefreshing = false
  refreshQueue = []
  message.error(detail || '您的账号已被禁用，请联系运维人员')
  setTimeout(() => { window.location.href = '/login' }, 800)
}

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

client.interceptors.request.use((config) => {
  const isAuthEndpoint = AUTH_PATHS.some((p) => config.url?.includes(p))
  if (isAuthEndpoint) return config

  const token = storage.getItem(TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let isRefreshing = false
let refreshQueue: Array<(token: string) => void> = []

client.interceptors.response.use(
  (response) => {
    const body = response.data as ApiResponse
    if (!body.success && body.error) {
      // Preserve structured error details for better diagnostics
      const err = new Error(body.error.detail || 'Request failed') as any
      err.code = body.error.code
      err.id = body.error.id
      return Promise.reject(err)
    }
    return response
  },
  async (error) => {
    const originalRequest = error.config

    // Never intercept 401 on auth endpoints — their 401 means "bad credentials"
    const isAuthRequest = AUTH_PATHS.some((p) => originalRequest?.url?.includes(p))

    // 账号封禁：不尝试刷新 token，直接踢下线
    const detail401 = error.response?.data?.error?.detail || ''
    if (error.response?.status === 401 && detail401.includes('已被')) {
      handleAccountDisabled(detail401)
      return Promise.reject(error)
    }

    if (error.response?.status === 401 && !originalRequest._retry && !isAuthRequest) {
      if (isRefreshing) {
        return new Promise((resolve) => {
          refreshQueue.push((token: string) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            resolve(client(originalRequest))
          })
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        const currentToken = storage.getItem(TOKEN_KEY)
        const res = await axios.post<ApiResponse<LoginToken>>(
          '/api/v1/current/refresh-token',
          {},
          { headers: currentToken ? { Authorization: `Bearer ${currentToken}` } : {} },
        )
        const newToken = res.data.data.access_token
        storage.setItem(TOKEN_KEY, newToken)

        refreshQueue.forEach((cb) => cb(newToken))
        refreshQueue = []

        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return client(originalRequest)
      } catch {
        storage.removeItem(TOKEN_KEY)
        storage.removeItem(USER_KEY)
        refreshQueue = []
        isRefreshing = false
        return Promise.reject(error)
      } finally {
        isRefreshing = false
      }
    }

    if (error.response) {
      const body = error.response.data
      const detail = body?.error?.detail || body?.message || error.message
      const err = new Error(detail) as any
      err.code = body?.error?.code || error.response.status
      err.id = body?.error?.id
      return Promise.reject(err)
    }
    return Promise.reject(error)
  },
)

export default client

import axios from 'axios'
import { message } from 'antd'
import type { ApiResponse } from '../types'
import type { LoginToken } from '../types'

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

client.interceptors.request.use((config) => {
  const token = sessionStorage.getItem('admin-token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let isRefreshing = false
let refreshQueue: Array<(token: string) => void> = []

// 账号已被封禁 — 强制踢下线
function handleAccountDisabled(detail: string) {
  sessionStorage.removeItem('admin-token')
  sessionStorage.removeItem('admin-user')
  isRefreshing = false
  refreshQueue = []
  message.error(detail || '您的账号已被禁用，请联系运维人员')
  setTimeout(() => { window.location.href = '/login' }, 800)
}

client.interceptors.response.use(
  (response) => {
    const body = response.data as ApiResponse
    if (!body.success && body.error) {
      return Promise.reject(new Error(body.error.detail || 'Request failed'))
    }
    return response
  },
  async (error) => {
    const originalRequest = error.config
    const detail = error.response?.data?.error?.detail || ''

    // 账号封禁：不尝试刷新 token，直接踢下线
    if (error.response?.status === 401 && detail.includes('已被')) {
      handleAccountDisabled(detail)
      return Promise.reject(error)
    }

    if (error.response?.status === 401 && !originalRequest._retry) {
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
        const res = await axios.post<ApiResponse<LoginToken>>('/api/v1/current/refresh-token', {}, {
          headers: { Authorization: `Bearer ${sessionStorage.getItem('admin-token')}` },
        })
        const newToken = res.data.data.access_token
        sessionStorage.setItem('admin-token', newToken)

        refreshQueue.forEach((cb) => cb(newToken))
        refreshQueue = []

        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return client(originalRequest)
      } catch {
        sessionStorage.removeItem('admin-token')
        sessionStorage.removeItem('admin-user')
        refreshQueue = []
        window.location.href = '/login'
        return Promise.reject(error)
      } finally {
        isRefreshing = false
      }
    }

    if (error.response) {
      const msg = error.response.data?.error?.detail || error.message
      return Promise.reject(new Error(msg))
    }
    return Promise.reject(error)
  },
)

export default client

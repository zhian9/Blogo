import { useCallback } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { message } from 'antd'
import { useAuthStore } from '../store/authStore'
import { login as loginApi, register as registerApi, getCaptchaId, getCurrentUser, logout as logoutApi } from '../api/auth'
import type { LoginRequest, RegisterRequest } from '../types'

export function useAuth() {
  const { token, user, isAuthenticated, login: storeLogin, logout: storeLogout } = useAuthStore()
  const navigate = useNavigate()
  const location = useLocation()

  const login = useCallback(async (credentials: Omit<LoginRequest, 'captcha_id' | 'captcha_code'>) => {
    // Fetch captcha first (simplified: real captcha logic in Login page)
    const captchaRes = await getCaptchaId()
    const captchaId = captchaRes.data.captcha_id

    const res = await loginApi({
      ...credentials,
      captcha_id: captchaId,
      captcha_code: '', // This should come from user input — the Login page handles this
    })

    const { access_token } = res.data
    const userRes = await getCurrentUser()
    storeLogin(access_token, userRes.data)

    const redirect = new URLSearchParams(location.search).get('redirect') || '/'
    message.success('Welcome back!')
    navigate(redirect, { replace: true })

    return userRes.data
  }, [navigate, location.search, storeLogin])

  const register = useCallback(async (data: Omit<RegisterRequest, 'captcha_id' | 'captcha_code'>) => {
    const captchaRes = await getCaptchaId()
    const captchaId = captchaRes.data.captcha_id

    const res = await registerApi({
      ...data,
      captcha_id: captchaId,
      captcha_code: '',
    })

    const { access_token } = res.data
    const userRes = await getCurrentUser()
    storeLogin(access_token, userRes.data)

    const redirect = new URLSearchParams(location.search).get('redirect') || '/'
    message.success('Registration successful!')
    navigate(redirect, { replace: true })

    return userRes.data
  }, [navigate, location.search, storeLogin])

  const logout = useCallback(async () => {
    try { await logoutApi() } catch { /* ignore */ }
    storeLogout()
    message.info('You have been logged out')
  }, [storeLogout])

  const checkAuth = useCallback((action = 'perform this action') => {
    if (!isAuthenticated) {
      message.warning(`Please sign in to ${action}`)
      return false
    }
    return true
  }, [isAuthenticated])

  const requireAuth = useCallback((action = 'perform this action') => {
    if (!isAuthenticated) {
      message.warning(`Please sign in to ${action}`)
      const currentPath = location.pathname + location.search
      navigate(`/login?redirect=${encodeURIComponent(currentPath)}`)
      return false
    }
    return true
  }, [isAuthenticated, navigate, location.pathname, location.search])

  return {
    token,
    user,
    isAuthenticated,
    login,
    register,
    logout,
    checkAuth,
    requireAuth,
  }
}

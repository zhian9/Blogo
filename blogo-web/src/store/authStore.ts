import { create } from 'zustand'
import type { AuthUser } from '../types'

// sessionStorage isolates auth state per tab — two tabs can have different logins
const TOKEN_KEY = 'blog-token'
const USER_KEY = 'blog-user'
const storage = sessionStorage

interface AuthState {
  token: string | null
  user: AuthUser | null
  isAuthenticated: boolean
  login: (token: string, user: AuthUser) => void
  logout: () => void
  setUser: (user: AuthUser) => void
  initialize: () => void
}

function loadInitialToken(): string | null {
  return storage.getItem(TOKEN_KEY)
}

function loadInitialUser(): AuthUser | null {
  try {
    const raw = storage.getItem(USER_KEY)
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

function computeAuthenticated(): boolean {
  return !!(storage.getItem(TOKEN_KEY) && storage.getItem(USER_KEY))
}

export const useAuthStore = create<AuthState>((set) => ({
  token: loadInitialToken(),
  user: loadInitialUser(),
  isAuthenticated: computeAuthenticated(),

  login: (token, user) => {
    storage.setItem(TOKEN_KEY, token)
    storage.setItem(USER_KEY, JSON.stringify(user))
    set({ token, user, isAuthenticated: true })
  },

  logout: () => {
    storage.removeItem(TOKEN_KEY)
    storage.removeItem(USER_KEY)
    set({ token: null, user: null, isAuthenticated: false })
  },

  setUser: (user) => {
    storage.setItem(USER_KEY, JSON.stringify(user))
    set({ user })
  },

  initialize: () => {
    const token = loadInitialToken()
    const user = loadInitialUser()
    set({ token, user, isAuthenticated: !!(token && user) })
  },
}))

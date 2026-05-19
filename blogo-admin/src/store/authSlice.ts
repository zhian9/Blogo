import { createSlice, createAsyncThunk, type PayloadAction } from '@reduxjs/toolkit'
import client from '../api/client'
import type { LoginRequest, LoginToken, User, Menu } from '../types'

function roleNameToCode(name: string): string | undefined {
  const map: Record<string, string> = {
    '管理员': 'admin', '超级管理员': 'admin',
    '内容管理员': 'content_manager', '评论审核员': 'comment_moderator',
    '用户': 'user', '游客': 'guest',
  }
  return map[name]
}

interface AuthState {
  token: string | null
  user: User | null
  menus: Menu[]
  loading: boolean
}

const initialState: AuthState = {
  token: sessionStorage.getItem('admin-token'),
  user: JSON.parse(sessionStorage.getItem('admin-user') || 'null'),
  menus: [],
  loading: false,
}

export const loginAsync = createAsyncThunk(
  'auth/login',
  async (params: { credentials: LoginRequest; navigate: (path: string) => void }, { rejectWithValue }) => {
    try {
      const res = await client.post('/login', params.credentials)
      const token: LoginToken = res.data.data
      sessionStorage.setItem('admin-token', token.access_token)

      const [userRes, menusRes] = await Promise.all([
        client.get('/current/user'),
        client.get('/current/menus'),
      ])

      const user: User = userRes.data.data
      const menus: Menu[] = menusRes.data.data || []
      sessionStorage.setItem('admin-user', JSON.stringify(user))

      const firstRole = (user as any)?.roles?.[0]
      const roleCode = firstRole?.role_code || firstRole?.role?.code || roleNameToCode(firstRole?.role_name) || 'user'
      if (roleCode === 'comment_moderator') {
        params.navigate('/comments')
      } else {
        params.navigate('/')
      }
      return { token: token.access_token, user, menus }
    } catch (err: any) {
      return rejectWithValue(err.message || 'Login failed')
    }
  },
)

export const fetchUserInfo = createAsyncThunk('auth/fetchUserInfo', async () => {
  const [userRes, menusRes] = await Promise.all([
    client.get('/current/user'),
    client.get('/current/menus'),
  ])
  return { user: userRes.data.data as User, menus: menusRes.data.data as Menu[] || [] }
})

export const logoutAsync = createAsyncThunk('auth/logout', async (navigate: (path: string) => void) => {
  try {
    await client.post('/current/logout')
  } catch {}
  sessionStorage.removeItem('admin-token')
  sessionStorage.removeItem('admin-user')
  navigate('/login')
})

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    clearAuth(state) {
      state.token = null
      state.user = null
      state.menus = []
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loginAsync.pending, (state) => { state.loading = true })
      .addCase(loginAsync.fulfilled, (state, action) => {
        state.token = action.payload.token
        state.user = action.payload.user
        state.menus = action.payload.menus
        state.loading = false
      })
      .addCase(loginAsync.rejected, (state) => {
        state.loading = false
      })
      .addCase(fetchUserInfo.fulfilled, (state, action) => {
        state.user = action.payload.user
        state.menus = action.payload.menus
        sessionStorage.setItem('admin-user', JSON.stringify(action.payload.user))
      })
      .addCase(logoutAsync.fulfilled, (state) => {
        state.token = null
        state.user = null
        state.menus = []
      })
  },
})

export const { clearAuth } = authSlice.actions
export default authSlice.reducer

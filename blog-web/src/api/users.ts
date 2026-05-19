import client from './client'
import type { ApiResponse, User } from '../types'

export interface UserListParams {
  current?: number
  pageSize?: number
  username?: string
  name?: string
}

export async function searchUsers(params: UserListParams) {
  const res = await client.get<ApiResponse<User[]>>('/users', { params })
  return res.data
}

import client from './client'
import type { ApiResponse, ProfileDashboard } from '../types'

export async function getUserProfile(userId: string) {
  const res = await client.get<ApiResponse<ProfileDashboard>>(`/users/${userId}/profile`)
  return res.data
}

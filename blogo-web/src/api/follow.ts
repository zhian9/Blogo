import client from './client'
import type { ApiResponse } from '../types'

export async function followUser(userId: string) {
  const res = await client.post<ApiResponse>(`/users/${userId}/follow`)
  return res.data
}

export async function unfollowUser(userId: string) {
  const res = await client.delete<ApiResponse>(`/users/${userId}/follow`)
  return res.data
}

export async function checkFollow(userId: string) {
  const res = await client.get<ApiResponse<boolean>>(`/users/${userId}/follow`)
  return res.data
}

export async function listFollowers(userId: string, current = 1, pageSize = 20) {
  const res = await client.get<ApiResponse>(`/users/${userId}/followers?current=${current}&pageSize=${pageSize}`)
  return res.data
}

export async function listFollowing(userId: string, current = 1, pageSize = 20) {
  const res = await client.get<ApiResponse>(`/users/${userId}/following?current=${current}&pageSize=${pageSize}`)
  return res.data
}

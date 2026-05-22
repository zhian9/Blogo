import client from './client'
import type { ApiResponse, PublicStats, Statistics } from '../types'

export async function getLatestStatistics() {
  const res = await client.get<ApiResponse<Statistics>>('/statistics/latest')
  return res.data
}

export async function getPublicStats() {
  const res = await client.get<ApiResponse<PublicStats>>('/statistics/public')
  return res.data
}

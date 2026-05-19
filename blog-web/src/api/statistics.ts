import client from './client'
import type { ApiResponse, Statistics } from '../types'

export async function getLatestStatistics() {
  const res = await client.get<ApiResponse<Statistics>>('/statistics/latest')
  return res.data
}

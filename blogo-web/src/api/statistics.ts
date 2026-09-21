import client from './client'
import type { ApiResponse, PublicStats } from '../types'

export async function getPublicStats() {
  const res = await client.get<ApiResponse<PublicStats>>('/statistics/public')
  return res.data
}

/**
 * 上报一次页面访问（后台访问量统计的 PV/UV 数据来源）。
 * 统计属于 best-effort：失败时静默忽略，不能影响用户浏览。
 */
export async function recordVisit() {
  try {
    await client.post<ApiResponse<null>>('/statistics/visit')
  } catch {
    // 统计失败无需打扰用户
  }
}

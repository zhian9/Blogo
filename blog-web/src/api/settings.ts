import client from './client'
import type { ApiResponse, Setting } from '../types'

export async function getAllSettings() {
  const res = await client.get<ApiResponse<Setting[]>>('/settings/all')
  return res.data
}

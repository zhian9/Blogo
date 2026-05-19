import client from './client'
import type { ApiResponse, Tag } from '../types'

export async function getAllTags() {
  const res = await client.get<ApiResponse<Tag[]>>('/tags/all')
  return res.data
}

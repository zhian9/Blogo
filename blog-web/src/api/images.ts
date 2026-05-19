import client from './client'
import type { ApiResponse } from '../types'

export interface ImageItem {
  id: string
  url: string
  name: string
  size: number
  width: number
  height: number
  type: string
  category: string
  created_at: string
}

export async function getImages(params?: { category?: string; current?: number; pageSize?: number }) {
  const res = await client.get<ApiResponse<ImageItem[]>>('/images', { params })
  return res.data
}

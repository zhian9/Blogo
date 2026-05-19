import client from './client'
import type { ApiResponse } from '../types'

export interface LikeStatus {
  count: number
  liked: boolean
}

export async function getLikeStatus(articleId: string) {
  const res = await client.get<ApiResponse<LikeStatus>>(`/articles/${articleId}/like`)
  return res.data
}

export async function likeArticle(articleId: string) {
  const res = await client.post<ApiResponse>(`/articles/${articleId}/like`, {})
  return res.data
}

export async function unlikeArticle(articleId: string) {
  const res = await client.delete<ApiResponse>(`/articles/${articleId}/like`)
  return res.data
}

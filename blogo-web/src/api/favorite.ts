import client from './client'
import type { ApiResponse } from '../types'

export async function checkFavorite(articleId: string, userId: string) {
  const res = await client.get<ApiResponse<boolean>>(`/favorites/${articleId}`, { params: { user_id: userId } })
  return res.data
}

export async function addFavorite(articleId: string) {
  const res = await client.post<ApiResponse>('/favorites', { article_id: articleId })
  return res.data
}

export async function removeFavorite(articleId: string) {
  const res = await client.delete<ApiResponse>(`/favorites/${articleId}`)
  return res.data
}

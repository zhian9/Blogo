import client from './client'
import type { ApiResponse, Page } from '../types'

export async function getPageBySlug(slug: string) {
  const res = await client.get<ApiResponse<Page>>(`/pages/slug/${slug}`)
  return res.data
}

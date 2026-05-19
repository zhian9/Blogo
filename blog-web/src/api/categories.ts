import client from './client'
import type { ApiResponse, Category } from '../types'

export async function getAllCategories() {
  const res = await client.get<ApiResponse<Category[]>>('/categories/all')
  return res.data
}

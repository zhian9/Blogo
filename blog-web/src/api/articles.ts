import client from './client'
import type { ApiResponse, Article, ArchiveItem, ArticleForm } from '../types'

export interface ArticleListParams {
  current?: number
  pageSize?: number
  title?: string
  category_id?: string
  tag_name?: string
  status?: string
  is_top?: boolean
}

export async function getArticles(params: ArticleListParams) {
  const res = await client.get<ApiResponse<Article[]>>('/articles', { params })
  return res.data
}

export async function getArticleBySlug(slug: string) {
  const res = await client.get<ApiResponse<Article>>(`/articles/slug/${slug}`)
  return res.data
}

export async function getArticleById(id: string) {
  const res = await client.get<ApiResponse<Article>>(`/articles/${id}`)
  return res.data
}

export async function getArchives() {
  const res = await client.get<ApiResponse<ArchiveItem[]>>('/archives')
  return res.data
}

export async function incArticleViews(id: string) {
  const res = await client.post<ApiResponse>(`/articles/${id}/views`)
  return res.data
}

export async function createArticle(data: ArticleForm) {
  const res = await client.post<ApiResponse<Article>>('/articles', data)
  return res.data
}

export async function updateArticle(id: string, data: ArticleForm) {
  const res = await client.put<ApiResponse<Article>>(`/articles/${id}`, data)
  return res.data
}

export async function deleteArticle(id: string) {
  const res = await client.delete<ApiResponse<void>>(`/articles/${id}`)
  return res.data
}

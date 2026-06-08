import client from './client'
import type { ApiResponse, Comment, CommentForm } from '../types'

export async function getComments(articleId: string) {
  const res = await client.get<ApiResponse<Comment[]>>(`/articles/${articleId}/comments`)
  return res.data
}

export async function getProjectComments(projectId: string) {
  const res = await client.get<ApiResponse<Comment[]>>(`/projects/${projectId}/comments`)
  return res.data
}

export async function createComment(data: CommentForm) {
  const res = await client.post<ApiResponse<Comment>>('/comments', data)
  return res.data
}

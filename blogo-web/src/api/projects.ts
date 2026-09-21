import client from './client'
import type { ApiResponse, Project, ProjectForm } from '../types'

export interface ProjectListParams {
  current?: number
  pageSize?: number
  title?: string
  category_id?: string
  tag_name?: string
  status?: string
  project_state?: string
  visibility?: string
  is_top?: boolean
  is_featured?: boolean
  sort_by?: string
}

// Projects
export async function getProjects(params: ProjectListParams) {
  const res = await client.get<ApiResponse<Project[]>>('/projects', { params })
  return res.data
}

export async function getProjectBySlug(slug: string) {
  const res = await client.get<ApiResponse<Project>>(`/projects/slug/${slug}`)
  return res.data
}

export async function getFeaturedProjects() {
  const res = await client.get<ApiResponse<Project[]>>('/projects/featured')
  return res.data
}

export async function incProjectViews(id: string) {
  const res = await client.post<ApiResponse>(`/projects/${id}/views`)
  return res.data
}

export async function createProject(data: ProjectForm) {
  const res = await client.post<ApiResponse<Project>>('/projects', data)
  return res.data
}

export async function updateProject(id: string, data: ProjectForm) {
  const res = await client.put<ApiResponse<Project>>(`/projects/${id}`, data)
  return res.data
}

export async function deleteProject(id: string) {
  const res = await client.delete<ApiResponse<void>>(`/projects/${id}`)
  return res.data
}

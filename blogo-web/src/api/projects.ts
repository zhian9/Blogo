import client from './client'
import type { ApiResponse, Project, ProjectForm, ProjectTimeline, ProjectResource, ProjectLikeCountResult } from '../types'

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

export async function getProjectById(id: string) {
  const res = await client.get<ApiResponse<Project>>(`/projects/${id}`)
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

// Project Like
export async function getProjectLikeStatus(id: string) {
  const res = await client.get<ApiResponse<ProjectLikeCountResult>>(`/projects/${id}/like`)
  return res.data
}

export async function likeProject(id: string) {
  const res = await client.post<ApiResponse>(`/projects/${id}/like`)
  return res.data
}

export async function unlikeProject(id: string) {
  const res = await client.delete<ApiResponse>(`/projects/${id}/like`)
  return res.data
}

// Project Favorite
export async function getProjectFavoriteStatus(id: string) {
  const res = await client.get<ApiResponse<boolean>>(`/projects/${id}/favorite`)
  return res.data
}

export async function favoriteProject(id: string) {
  const res = await client.post<ApiResponse>(`/projects/${id}/favorite`)
  return res.data
}

export async function unfavoriteProject(id: string) {
  const res = await client.delete<ApiResponse>(`/projects/${id}/favorite`)
  return res.data
}

export async function getProjectFavoriteCount(id: string) {
  const res = await client.get<ApiResponse<number>>(`/projects/${id}/favorite/count`)
  return res.data
}

// Project Timeline
export async function getProjectTimeline(id: string) {
  const res = await client.get<ApiResponse<ProjectTimeline[]>>(`/projects/${id}/timeline`)
  return res.data
}

// Project Resources
export async function getProjectResources(id: string) {
  const res = await client.get<ApiResponse<ProjectResource[]>>(`/projects/${id}/resources`)
  return res.data
}

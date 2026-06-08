import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getProjects, getProjectBySlug, getFeaturedProjects, incProjectViews,
  createProject, updateProject, deleteProject,
  getProjectLikeStatus, likeProject, unlikeProject,
  getProjectFavoriteStatus, favoriteProject, unfavoriteProject,
  getProjectTimeline, getProjectResources,
} from '../api/projects'
import type { ProjectListParams } from '../api/projects'
import type { ProjectForm } from '../types'

// List
export function useProjects(params: ProjectListParams) {
  return useQuery({
    queryKey: ['projects', params],
    queryFn: () => getProjects(params),
    staleTime: 60_000,
  })
}

// Detail by slug
export function useProjectBySlug(slug: string) {
  return useQuery({
    queryKey: ['project', slug],
    queryFn: () => getProjectBySlug(slug),
    enabled: !!slug,
    staleTime: 120_000,
  })
}

// Featured
export function useFeaturedProjects() {
  return useQuery({
    queryKey: ['projects', 'featured'],
    queryFn: getFeaturedProjects,
    staleTime: 300_000,
  })
}

// Inc views
export function useIncProjectViews() {
  return useMutation({
    mutationFn: (id: string) => incProjectViews(id),
  })
}

// Create
export function useCreateProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: ProjectForm) => createProject(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      queryClient.invalidateQueries({ queryKey: ['profile'] })
    },
  })
}

// Update
export function useUpdateProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: ProjectForm }) => updateProject(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      queryClient.invalidateQueries({ queryKey: ['project'] })
      queryClient.invalidateQueries({ queryKey: ['profile'] })
    },
  })
}

// Delete
export function useDeleteProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteProject(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      queryClient.invalidateQueries({ queryKey: ['project'] })
      queryClient.invalidateQueries({ queryKey: ['profile'] })
    },
  })
}

// Like
export function useProjectLikeStatus(id: string) {
  return useQuery({
    queryKey: ['project-like', id],
    queryFn: () => getProjectLikeStatus(id),
    enabled: !!id,
    staleTime: 30_000,
  })
}

export function useLikeProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => likeProject(id),
    onSuccess: (_d, id) => {
      queryClient.invalidateQueries({ queryKey: ['project-like', id] })
    },
  })
}

export function useUnlikeProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => unlikeProject(id),
    onSuccess: (_d, id) => {
      queryClient.invalidateQueries({ queryKey: ['project-like', id] })
    },
  })
}

// Favorite
export function useProjectFavoriteStatus(id: string) {
  return useQuery({
    queryKey: ['project-fav', id],
    queryFn: () => getProjectFavoriteStatus(id),
    enabled: !!id,
    staleTime: 30_000,
  })
}

export function useFavoriteProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => favoriteProject(id),
    onSuccess: (_d, id) => {
      queryClient.invalidateQueries({ queryKey: ['project-fav', id] })
    },
  })
}

export function useUnfavoriteProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => unfavoriteProject(id),
    onSuccess: (_d, id) => {
      queryClient.invalidateQueries({ queryKey: ['project-fav', id] })
    },
  })
}

// Timeline
export function useProjectTimeline(id: string) {
  return useQuery({
    queryKey: ['project-timeline', id],
    queryFn: () => getProjectTimeline(id),
    enabled: !!id,
    staleTime: 120_000,
  })
}

// Resources
export function useProjectResources(id: string) {
  return useQuery({
    queryKey: ['project-resources', id],
    queryFn: () => getProjectResources(id),
    enabled: !!id,
    staleTime: 120_000,
  })
}

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getProjects, getProjectBySlug, getFeaturedProjects, incProjectViews,
  createProject, updateProject, deleteProject,
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

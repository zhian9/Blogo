import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getArticles, getArticleBySlug, getArchives, incArticleViews, createArticle, updateArticle, deleteArticle } from '../api/articles'
import type { ArticleListParams } from '../api/articles'
import type { ArticleForm } from '../types'

export function useArticles(params: ArticleListParams) {
  return useQuery({
    queryKey: ['articles', params],
    queryFn: () => getArticles(params),
    staleTime: 60_000,
  })
}

export function useArticleBySlug(slug: string) {
  return useQuery({
    queryKey: ['article', slug],
    queryFn: () => getArticleBySlug(slug),
    enabled: !!slug,
    staleTime: 120_000,
  })
}

export function useArchives() {
  return useQuery({
    queryKey: ['archives'],
    queryFn: getArchives,
    staleTime: 300_000,
  })
}

export function useIncViews() {
  return useMutation({
    mutationFn: (id: string) => incArticleViews(id),
  })
}

export function useCreateArticle() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: ArticleForm) => createArticle(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['articles'] })
      queryClient.invalidateQueries({ queryKey: ['profile'] })
    },
  })
}

export function useUpdateArticle() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: ArticleForm }) => updateArticle(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['articles'] })
      queryClient.invalidateQueries({ queryKey: ['article'] })
      queryClient.invalidateQueries({ queryKey: ['profile'] })
    },
  })
}

export function useDeleteArticle() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteArticle(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['articles'] })
      queryClient.invalidateQueries({ queryKey: ['article'] })
      queryClient.invalidateQueries({ queryKey: ['profile'] })
    },
  })
}

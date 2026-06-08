import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getComments, getProjectComments, createComment } from '../api/comments'
import type { CommentForm } from '../types'

export function useComments(articleId: string) {
  return useQuery({
    queryKey: ['comments', articleId],
    queryFn: () => getComments(articleId),
    enabled: !!articleId,
    staleTime: 60_000,
  })
}

export function useProjectComments(projectId: string) {
  return useQuery({
    queryKey: ['project-comments', projectId],
    queryFn: () => getProjectComments(projectId),
    enabled: !!projectId,
    staleTime: 60_000,
  })
}

export function useCreateComment() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CommentForm) => createComment(data),
    onSuccess: (_data, variables) => {
      if (variables.article_id) {
        queryClient.invalidateQueries({ queryKey: ['comments', variables.article_id] })
      }
      if (variables.project_id) {
        queryClient.invalidateQueries({ queryKey: ['project-comments', variables.project_id] })
      }
    },
  })
}

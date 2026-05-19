import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getComments, createComment } from '../api/comments'
import type { CommentForm } from '../types'

export function useComments(articleId: string) {
  return useQuery({
    queryKey: ['comments', articleId],
    queryFn: () => getComments(articleId),
    enabled: !!articleId,
    staleTime: 60_000,
  })
}

export function useCreateComment() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CommentForm) => createComment(data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['comments', variables.article_id] })
    },
  })
}

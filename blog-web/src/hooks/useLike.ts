import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getLikeStatus, likeArticle, unlikeArticle } from '../api/like'

export function useLikeStatus(articleId: string) {
  return useQuery({
    queryKey: ['like', articleId],
    queryFn: () => getLikeStatus(articleId),
    enabled: !!articleId,
    staleTime: 30_000,
  })
}

export function useLike(articleId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => likeArticle(articleId),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['like', articleId] }) },
  })
}

export function useUnLike(articleId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => unlikeArticle(articleId),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['like', articleId] }) },
  })
}

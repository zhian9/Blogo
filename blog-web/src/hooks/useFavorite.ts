import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { checkFavorite, addFavorite, removeFavorite } from '../api/favorite'
import { useAuthStore } from '../store/authStore'

export function useFavoriteStatus(articleId: string) {
  const userId = useAuthStore((s) => s.user?.id || '')
  return useQuery({
    queryKey: ['favorite', articleId, userId],
    queryFn: () => checkFavorite(articleId, userId),
    enabled: !!articleId && !!userId,
    staleTime: 30_000,
  })
}

export function useAddFavorite(articleId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => addFavorite(articleId),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['favorite', articleId] }) },
  })
}

export function useRemoveFavorite(articleId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => removeFavorite(articleId),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['favorite', articleId] }) },
  })
}

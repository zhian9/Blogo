import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { followUser, unfollowUser, checkFollow } from '../api/follow'

export function useFollowStatus(userId: string) {
  return useQuery({
    queryKey: ['follow', userId],
    queryFn: () => checkFollow(userId),
    enabled: !!userId,
    staleTime: 30_000,
  })
}

export function useFollow(userId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => followUser(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['follow', userId] })
      queryClient.invalidateQueries({ queryKey: ['profile', userId] })
    },
  })
}

export function useUnfollow(userId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => unfollowUser(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['follow', userId] })
      queryClient.invalidateQueries({ queryKey: ['profile', userId] })
    },
  })
}

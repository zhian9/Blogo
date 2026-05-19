import { useQuery } from '@tanstack/react-query'
import { getUserProfile } from '../api/profile'

export function useUserProfile(userId: string) {
  return useQuery({
    queryKey: ['profile', userId],
    queryFn: () => getUserProfile(userId),
    enabled: !!userId,
    staleTime: 30_000,
  })
}

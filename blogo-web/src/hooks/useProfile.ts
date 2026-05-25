import { useQuery } from '@tanstack/react-query'
import { getUserProfile } from '../api/profile'

export function useUserProfile(userId: string) {
  return useQuery({
    queryKey: ['profile', userId],
    queryFn: () => getUserProfile(userId),
    enabled: !!userId,
    staleTime: 0,           // 每次进入页面都重新拉取，确保贡献数据是新的
    refetchOnMount: true,
    refetchOnWindowFocus: true,
  })
}

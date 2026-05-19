import { useQuery } from '@tanstack/react-query'
import { searchUsers } from '../api/users'
import type { UserListParams } from '../api/users'

export function useUserSearch(params: UserListParams) {
  return useQuery({
    queryKey: ['users', params],
    queryFn: () => searchUsers(params),
    enabled: !!(params.username || params.name),
    staleTime: 30_000,
  })
}

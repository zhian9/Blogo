import { useQuery } from '@tanstack/react-query'
import { getAllTags } from '../api/tags'

export function useTags() {
  return useQuery({
    queryKey: ['tags'],
    queryFn: getAllTags,
    staleTime: 300_000,
  })
}

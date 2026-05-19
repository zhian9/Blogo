import { useQuery } from '@tanstack/react-query'
import { getImages } from '../api/images'

export function useImages(params?: { category?: string; current?: number; pageSize?: number }) {
  return useQuery({
    queryKey: ['images', params],
    queryFn: () => getImages(params),
    staleTime: 60_000,
  })
}

import { useQuery } from '@tanstack/react-query'
import { getAllCategories } from '../api/categories'

export function useCategories() {
  return useQuery({
    queryKey: ['categories'],
    queryFn: getAllCategories,
    staleTime: 300_000,
  })
}

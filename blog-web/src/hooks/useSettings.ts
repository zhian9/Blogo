import { useQuery } from '@tanstack/react-query'
import { getAllSettings } from '../api/settings'
import { getPageBySlug } from '../api/pages'
import { getLatestStatistics } from '../api/statistics'

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: getAllSettings,
    staleTime: 600_000,
  })
}

export function usePage(slug: string) {
  return useQuery({
    queryKey: ['page', slug],
    queryFn: () => getPageBySlug(slug),
    enabled: !!slug,
    staleTime: 300_000,
  })
}

export function useStatistics() {
  return useQuery({
    queryKey: ['statistics'],
    queryFn: getLatestStatistics,
    staleTime: 120_000,
  })
}

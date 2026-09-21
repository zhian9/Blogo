import { useQuery } from '@tanstack/react-query'
import { getAllSettings } from '../api/settings'
import { getPageBySlug } from '../api/pages'

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: getAllSettings,
    // 站点配置（首页文案、关于页文案等）属于后台随时会改的内容：
    // 不做长时间缓存，每次进入页面都重新校验，切回标签页也刷新，
    // 这样后台改完、前台刷新（或切回窗口）即可看到最新内容。
    staleTime: 0,
    refetchOnWindowFocus: true,
  })
}

export function usePage(slug: string) {
  return useQuery({
    queryKey: ['page', slug],
    queryFn: () => getPageBySlug(slug),
    enabled: !!slug,
    // 单页内容（关于、隐私政策等）同样由后台编辑，保持实时
    staleTime: 0,
    refetchOnWindowFocus: true,
  })
}

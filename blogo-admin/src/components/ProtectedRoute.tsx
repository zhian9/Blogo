import { Navigate, useLocation } from 'react-router-dom'
import { message } from 'antd'
import { useAppSelector } from '../store'
import { useEffect } from 'react'

// 角色 → 允许访问的路径前缀
const rolePathMap: Record<string, string[]> = {
  admin: [],
  content_manager: ['/dashboard', '/articles', '/categories', '/tags', '/comments', '/media', '/profile', '/403'],
  comment_moderator: ['/dashboard', '/comments', '/profile', '/403'],
  user: [],
  guest: [],
}

function roleNameToCode(name: string): string | undefined {
  const map: Record<string, string> = {
    '管理员': 'admin',
    '超级管理员': 'admin',
    '内容管理员': 'content_manager',
    '评论审核员': 'comment_moderator',
    '用户': 'user',
    '游客': 'guest',
  }
  return map[name]
}

// 获取当前用户的主角色码（root 视为 admin，兼容嵌套 role.code 和扁平 role_name）
export function getUserRoleCode(user: any): string {
  if (!user) return 'guest'
  if (user.id === 'root') return 'admin'
  const roles = user.roles || []
  for (const ur of roles) {
    const code = ur.role_code || ur.role?.code || roleNameToCode(ur.role_name)
    if (code === 'super_admin' || code === 'admin') return 'admin'
    if (code === 'content_manager') return 'content_manager'
    if (code === 'comment_moderator') return 'comment_moderator'
    if (code === 'user') return 'user'
  }
  return 'user'
}

// 判断当前路径是否允许访问
function isPathAllowed(pathname: string, roleCode: string): boolean {
  if (roleCode === 'admin') return true
  const allowed = rolePathMap[roleCode]
  if (!allowed || allowed.length === 0) return false
  // 仪表盘首页始终允许
  if (pathname === '/' || pathname === '/dashboard' || pathname.startsWith('/profile')) return true
  // 检查精确匹配或子路径前缀
  return allowed.some(p => pathname === p || pathname.startsWith(p + '/'))
}

export default function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const token = useAppSelector((s) => s.auth.token)
  const user = useAppSelector((s) => s.auth.user)
  const location = useLocation()
  const roleCode = getUserRoleCode(user)

  // 未登录 → 重定向到登录页
  if (!token) {
    return <Navigate to="/login" replace />
  }

  // user / guest 无权进入后台 → 清除登录状态并跳转到前台主页
  if (roleCode === 'user' || roleCode === 'guest') {
    sessionStorage.removeItem('admin-token')
    sessionStorage.removeItem('admin-user')
    window.location.href = '/'
    return null
  }

  // comment_moderator 访问非授权路径 → 提示并弹回 /comments
  if (roleCode === 'comment_moderator' && !isPathAllowed(location.pathname, roleCode)) {
    message.error('您无权访问该页面，已为您返回专属工作台')
    return <Navigate to="/comments" replace />
  }

  // content_manager 访问非授权路径 → 403
  if (roleCode === 'content_manager' && !isPathAllowed(location.pathname, roleCode)) {
    return <Navigate to="/403" replace />
  }

  return <>{children}</>
}

import { useMemo } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { Menu } from 'antd'
import {
  DashboardOutlined, FileTextOutlined, UserOutlined, TeamOutlined, MenuOutlined,
  AppstoreOutlined, TagOutlined, MessageOutlined, SettingOutlined,
  SafetyCertificateOutlined, HistoryOutlined, PictureOutlined, ToolOutlined,
} from '@ant-design/icons'
import { useAppSelector } from '../store'
import type { Menu as MenuType } from '../types'

// 静态图标映射（不支持动态 require，Vite ESM 环境需显式导入）
const iconRegistry: Record<string, React.ComponentType<any>> = {
  DashboardOutlined, FileTextOutlined, UserOutlined, TeamOutlined,
  MenuOutlined, AppstoreOutlined, TagOutlined, MessageOutlined,
  SettingOutlined, SafetyCertificateOutlined, HistoryOutlined,
  PictureOutlined, ToolOutlined,
}

function resolveIcon(iconName: string): React.ReactNode {
  if (!iconName) return undefined
  const Comp = iconRegistry[iconName]
  return Comp ? <Comp /> : undefined
}

function menuToAntdItem(m: MenuType): any {
  if (m.type === 'button') return null

  const hasChildren = m.children && m.children.length > 0
  let icon: React.ReactNode = undefined
  if (m.properties) {
    try { const p = JSON.parse(m.properties); if (p.icon) icon = resolveIcon(p.icon) } catch {}
  }

  const children = hasChildren
    ? m.children!.map(menuToAntdItem).filter(Boolean)
    : undefined

  const hasVisibleChildren = children && children.length > 0

  return {
    key: m.path || m.code,
    label: m.name,
    icon,
    children: hasVisibleChildren ? children : undefined,
  }
}

export default function Sidebar() {
  const navigate = useNavigate()
  const location = useLocation()
  const menus = useAppSelector((s) => s.auth.menus)

  const menuItems = useMemo(() => {
    if (!menus || menus.length === 0) return []
    return menus.map(menuToAntdItem).filter(Boolean)
  }, [menus])

  const pathname = location.pathname

  // 匹配当前路由到菜单项
  const { selectedKey, openKey } = useMemo(() => {
    let sel = '/dashboard'
    if (pathname !== '/') {
      sel = pathname.startsWith('/logs/') ? pathname : '/' + pathname.split('/')[1]
    }
    // 查找选中项的父级 key 用来自动展开
    let parentKey = ''
    for (const item of menuItems) {
      if (item.children) {
        for (const child of item.children) {
          if (child.key === sel || child.key === pathname) { parentKey = item.key; break }
        }
      }
      if (parentKey) break
    }
    return { selectedKey: sel, openKey: parentKey }
  }, [pathname, menuItems])

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <div style={{ height: 64, display: 'flex', alignItems: 'center', justifyContent: 'center', borderBottom: '1px solid rgba(0,0,0,0.06)' }}>
        <span style={{ fontSize: 20, fontWeight: 700, color: '#1890ff', whiteSpace: 'nowrap' }}>Blogo</span>
      </div>
      {menuItems.length > 0 ? (
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey]}
          defaultOpenKeys={openKey ? [openKey] : ['/dashboard']}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ flex: 1, borderRight: 0 }}
        />
      ) : (
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'rgba(255,255,255,0.2)', fontSize: 12 }}>
          菜单加载中...
        </div>
      )}
    </div>
  )
}

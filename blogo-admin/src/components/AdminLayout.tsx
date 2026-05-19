import { useState, useEffect } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import { Layout } from 'antd'
import Sidebar from './Sidebar'
import HeaderBar from './HeaderBar'
import { useAppDispatch, useAppSelector } from '../store'
import { fetchUserInfo } from '../store/authSlice'

const { Sider, Content } = Layout

export default function AdminLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const [themeMode, setThemeMode] = useState<'light' | 'dark'>(
    () => (localStorage.getItem('admin-theme') as 'light' | 'dark') || 'dark',
  )
  const location = useLocation()
  const [fadeKey, setFadeKey] = useState(0)
  const dispatch = useAppDispatch()
  const menus = useAppSelector((s) => s.auth.menus)

  // 页面刷新时恢复菜单
  useEffect(() => {
    if (!menus || menus.length === 0) {
      dispatch(fetchUserInfo())
    }
  }, [])

  const toggleTheme = () => {
    const next = themeMode === 'light' ? 'dark' : 'light'
    setThemeMode(next)
    localStorage.setItem('admin-theme', next)
    window.location.reload()
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider trigger={null} collapsible collapsed={collapsed}
        style={{ overflow: 'auto', height: '100vh', position: 'fixed', left: 0, top: 0, bottom: 0, zIndex: 10 }}>
        <Sidebar />
      </Sider>
      <Layout style={{ marginLeft: collapsed ? 80 : 200, transition: 'margin-left 0.2s' }}>
        <HeaderBar collapsed={collapsed} onToggle={() => setCollapsed(!collapsed)}
          themeMode={themeMode} onToggleTheme={toggleTheme} />
        <Content style={{ margin: 24, minHeight: 280 }}>
          <div key={fadeKey} className="page-fade">
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  )
}

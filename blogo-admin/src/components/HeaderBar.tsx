import { Layout, Breadcrumb, Button, Dropdown, Space, theme } from 'antd'
import {
  UserOutlined, LogoutOutlined, ProfileOutlined, BulbOutlined,
  MenuFoldOutlined, MenuUnfoldOutlined,
} from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../store'
import { logoutAsync } from '../store/authSlice'
import { getUserRoleCode } from './ProtectedRoute'

const { Header } = Layout

interface Props {
  collapsed: boolean
  onToggle: () => void
  themeMode: 'light' | 'dark'
  onToggleTheme: () => void
}

const pathLabels: Record<string, string> = {
  '/': '仪表盘',
  '/articles': '文章管理',
  '/articles/new': '新建文章',
  '/users': '用户管理',
  '/roles': '角色管理',
  '/menus': '菜单管理',
  '/categories': '分类管理',
  '/tags': '标签管理',
  '/comments': '评论管理',
  '/pages': '页面管理',
  '/settings': '系统设置',
  '/logs': '操作日志',
  '/profile': '个人设置',
}

export default function HeaderBar({ collapsed, onToggle, themeMode, onToggleTheme }: Props) {
  const navigate = useNavigate()
  const location = useLocation()
  const dispatch = useAppDispatch()
  const user = useAppSelector((s) => s.auth.user)
  const { token: themeToken } = theme.useToken()

  const roleCode = getUserRoleCode(user)
  const homePath = roleCode === 'comment_moderator' ? '/comments' : '/'

  const segments = location.pathname.split('/').filter(Boolean)
  const breadcrumbItems = [
    { title: '首页', path: homePath },
    ...segments.map((seg, i) => {
      const path = '/' + segments.slice(0, i + 1).join('/')
      return { title: pathLabels[path] || seg, path }
    }),
  ]

  return (
    <Header
      style={{
        padding: '0 24px',
        background: themeToken.colorBgContainer,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        borderBottom: `1px solid ${themeToken.colorBorderSecondary}`,
      }}
    >
      <Space>
        <Button
          type="text"
          icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          onClick={onToggle}
        />
        <Breadcrumb
          items={breadcrumbItems.map((item) => ({
            title: <a onClick={() => navigate(item.path)} style={{ color: 'inherit' }}>{item.title}</a>,
          }))}
        />
      </Space>

      <Space>
        <Button
          type="text"
          icon={<BulbOutlined />}
          onClick={onToggleTheme}
        />
        <Dropdown
          menu={{
            items: [
              { key: 'profile', icon: <ProfileOutlined />, label: '个人设置', onClick: () => navigate('/profile') },
              { type: 'divider' },
              {
                key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true,
                onClick: () => dispatch(logoutAsync(navigate)),
              },
            ],
          }}
        >
          <Space style={{ cursor: 'pointer' }}>
            <UserOutlined />
            {user?.name || user?.username || '管理员'}
          </Space>
        </Dropdown>
      </Space>
    </Header>
  )
}

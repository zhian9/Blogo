import { useNavigate } from 'react-router-dom'
import { Avatar, Dropdown, Button, Space, message } from 'antd'
import {
  UserOutlined,
  LoginOutlined,
  UserAddOutlined,
  LogoutOutlined,
  FileTextOutlined,
  SettingOutlined,
  HomeOutlined,
  EditOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '../store/authStore'
import { logout as logoutApi } from '../api/auth'

export default function AuthToggle() {
  const { token, user, isAuthenticated, logout: storeLogout } = useAuthStore()
  const navigate = useNavigate()

  const handleLogout = async () => {
    try { await logoutApi() } catch { /* ignore */ }
    storeLogout()
    message.info('已退出登录')
    navigate('/')
  }

  if (isAuthenticated && user) {
    const displayName = user.name || user.username || 'User'

    const menuItems = [
      {
        key: 'user-info',
        label: (
          <div style={{ padding: '4px 0' }}>
            <div style={{ fontWeight: 600, fontSize: 14 }}>{displayName}</div>
            <div style={{ fontSize: 12, color: '#999' }}>{user.email || user.username}</div>
          </div>
        ),
        disabled: true,
      },
      { type: 'divider' as const },
      {
        key: 'publish',
        icon: <EditOutlined />,
        label: '发布文章',
        onClick: () => navigate('/publish'),
      },
      {
        key: 'home',
        icon: <HomeOutlined />,
        label: '我的主页',
        onClick: () => navigate(`/user/${user.id}`),
      },
      {
        key: 'articles',
        icon: <FileTextOutlined />,
        label: '我的文章',
        onClick: () => navigate(`/user/${user.id}?tab=articles`),
      },
      {
        key: 'settings',
        icon: <SettingOutlined />,
        label: '编辑资料',
        onClick: () => navigate('/profile'),
      },
      { type: 'divider' as const },
      {
        key: 'logout',
        icon: <LogoutOutlined />,
        label: '退出登录',
        danger: true,
        onClick: handleLogout,
      },
    ]

    return (
      <Dropdown menu={{ items: menuItems }} placement="bottomRight" trigger={['click']}>
        <Avatar
          size={40}
          icon={!user.avatar ? <UserOutlined /> : undefined}
          src={user.avatar || undefined}
          style={{
            cursor: 'pointer',
            backgroundColor: user.avatar ? undefined : '#1890ff',
            flexShrink: 0,
          }}
        />
      </Dropdown>
    )
  }

  return (
    <Space>
      <Button
        ghost
        type="primary"
        icon={<LoginOutlined />}
        onClick={() => navigate('/login')}
      >
        登录
      </Button>
      <Button
        type="primary"
        icon={<UserAddOutlined />}
        onClick={() => navigate('/register')}
      >
        注册
      </Button>
    </Space>
  )
}

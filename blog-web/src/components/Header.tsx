import { useState } from 'react'
import { useNavigate, Link, useLocation } from 'react-router-dom'
import { Input, Button, Drawer } from 'antd'
import {
  MenuOutlined,
  SearchOutlined,
  HomeOutlined,
  FileTextOutlined,
  ClockCircleOutlined,
  AppstoreOutlined,
  CodeOutlined,
  InfoCircleOutlined,
  LoginOutlined,
  LogoutOutlined,
  UserAddOutlined,
} from '@ant-design/icons'
import { motion } from 'framer-motion'
import { useAppStore } from '../store/appStore'
import { useAuthStore } from '../store/authStore'
import { logout as logoutApi } from '../api/auth'
import AuthToggle from './AuthToggle'

const navItems = [
  { path: '/', label: '首页', icon: <HomeOutlined /> },
  { path: '/articles', label: '文章', icon: <FileTextOutlined /> },
  { path: '/archives', label: '归档', icon: <ClockCircleOutlined /> },
  { path: '/categories', label: '分类', icon: <AppstoreOutlined /> },
  { path: '/projects', label: '项目', icon: <CodeOutlined /> },
  { path: '/about', label: '关于', icon: <InfoCircleOutlined /> },
]

// ── accent colors per nav item ──
const itemAccents = [
  'rgba(255,255,255,0.12)',   // 首页
  'rgba(99,102,241,0.15)',    // 文章
  'rgba(16,185,129,0.12)',    // 归档
  'rgba(245,158,11,0.12)',    // 分类
  'rgba(139,92,246,0.15)',    // 项目
  'rgba(236,72,153,0.12)',    // 关于
]

export default function Header() {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const location = useLocation()
  const { searchKeyword, setSearchKeyword } = useAppStore()
  const { isAuthenticated, user, logout: storeLogout } = useAuthStore()
  const navigate = useNavigate()

  const isActive = (itemPath: string) => {
    if (itemPath === '/') return location.pathname === '/'
    return location.pathname.startsWith(itemPath)
  }

  const handleSearch = (value: string) => {
    setSearchKeyword(value)
    if (value.trim()) navigate(`/search?q=${encodeURIComponent(value.trim())}`)
  }

  const handleLogout = async () => {
    try { await logoutApi() } catch {}
    storeLogout()
    navigate('/')
  }

  return (
    <>
      {/* ========== NAVBAR ========== */}
      <nav
        style={{
          position: 'fixed', top: 0, left: 0, right: 0, zIndex: 100,
          height: 64,
          background: 'rgba(8, 8, 16, 0.72)',
          backdropFilter: 'blur(20px) saturate(180%)',
          WebkitBackdropFilter: 'blur(20px) saturate(180%)',
          borderBottom: '1px solid rgba(255,255,255,0.06)',
          boxShadow: '0 1px 0 rgba(255,255,255,0.03), 0 8px 32px rgba(0,0,0,0.4)',
        }}
      >
        <div style={{
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          height: '100%', padding: '0 48px', maxWidth: 1600, margin: '0 auto',
        }}>
          {/* ── Logo ── */}
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.2, ease: 'easeOut' }}
            whileHover={{ y: -1 }}
            whileTap={{ scale: 0.98 }}
          >
            <Link to="/" style={{
              display: 'flex', alignItems: 'center',
              textDecoration: 'none', flexShrink: 0,
              width: 100,
            }}>
              <span style={{
                fontFamily: "'Instrument Serif', 'Playfair Display', serif",
                fontStyle: 'italic',
                fontSize: 26,
                fontWeight: 400,
                color: '#ffffff',
                letterSpacing: '-0.03em',
                textShadow: '0 2px 10px rgba(0,0,0,0.5)',
                transition: 'text-shadow 0.3s ease, opacity 0.3s ease',
              }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.textShadow = '0 0 18px rgba(255,255,255,0.5), 0 2px 10px rgba(0,0,0,0.5)'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.textShadow = '0 2px 10px rgba(0,0,0,0.5)'
                }}
              >
                Blogo
              </span>
            </Link>
          </motion.div>

          {/* ── Nav links ── */}
          <div className="nav-desktop" style={{
            display: 'flex', alignItems: 'center', gap: 2,
          }}>
            {navItems.map((item, i) => {
              const active = isActive(item.path)
              const accent = itemAccents[i]
              return (
                <motion.div
                  key={item.path}
                  whileHover={{ scale: 1.03 }}
                  whileTap={{ scale: 0.97 }}
                >
                  <Link
                    to={item.path}
                    style={{
                      display: 'flex', alignItems: 'center', gap: 7,
                      padding: '8px 18px',
                      borderRadius: 20,
                      fontFamily: "'Noto Serif SC', 'Playfair Display', serif",
                      fontSize: 14,
                      fontWeight: active ? 500 : 400,
                      color: active ? '#ffffff' : 'rgba(255,255,255,0.6)',
                      letterSpacing: '0.03em',
                      textDecoration: 'none',
                      background: active ? accent : 'transparent',
                      border: active ? '1px solid rgba(255,255,255,0.12)' : '1px solid transparent',
                      boxShadow: active ? `0 0 24px ${accent}, 0 0 2px rgba(255,255,255,0.08)` : 'none',
                      transition: 'all 0.35s cubic-bezier(0.4, 0, 0.2, 1)',
                      position: 'relative',
                    }}
                    onMouseEnter={(e) => {
                      if (!active) {
                        e.currentTarget.style.background = 'rgba(255,255,255,0.08)'
                        e.currentTarget.style.color = '#ffffff'
                        e.currentTarget.style.boxShadow = '0 0 20px rgba(255,255,255,0.06)'
                        e.currentTarget.style.borderColor = 'rgba(255,255,255,0.08)'
                      }
                    }}
                    onMouseLeave={(e) => {
                      if (!active) {
                        e.currentTarget.style.background = 'transparent'
                        e.currentTarget.style.color = 'rgba(255,255,255,0.6)'
                        e.currentTarget.style.boxShadow = 'none'
                        e.currentTarget.style.borderColor = 'transparent'
                      }
                    }}
                  >
                    <span style={{ fontSize: 16, display: 'flex', alignItems: 'center' }}>
                      {item.icon}
                    </span>
                    {item.label}
                  </Link>
                </motion.div>
              )
            })}
          </div>

          {/* ── Right section ── */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexShrink: 0 }}>
            <Input.Search
              placeholder="搜索..."
              allowClear
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
              onSearch={handleSearch}
              prefix={<SearchOutlined style={{ color: 'rgba(255,255,255,0.3)', fontSize: 14 }} />}
              className="search-input"
              style={{ width: 180 }}
              styles={{
                input: {
                  background: 'rgba(255,255,255,0.04)',
                  borderColor: 'rgba(255,255,255,0.06)',
                  color: '#ffffff',
                  borderRadius: 20,
                  fontFamily: "'Noto Serif SC', serif",
                  fontSize: 13,
                },
              }}
            />

            <span className="nav-desktop">
              <AuthToggle />
            </span>

            <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
              <Button
                type="text"
                icon={<MenuOutlined style={{ color: 'rgba(255,255,255,0.55)', fontSize: 18 }} />}
                onClick={() => setDrawerOpen(true)}
                className="nav-mobile-btn"
                style={{
                  borderRadius: 20, width: 40, height: 40,
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                }}
              />
            </motion.div>
          </div>
        </div>
      </nav>

      {/* Spacer */}
      <div style={{ height: 64 }} />

      {/* ========== MOBILE DRAWER ========== */}
      <Drawer
        title={
          <span style={{
            fontFamily: "'Playfair Display', serif",
            fontStyle: 'italic',
            color: '#ffffff',
            fontSize: 22,
          }}>
            B<span style={{ fontFamily: "'Noto Serif SC', serif", fontStyle: 'normal', fontSize: 16, marginLeft: 6 }}>博客</span>
          </span>
        }
        placement="right"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        size="default"
        styles={{
          header: {
            background: 'rgba(8,8,16,0.98)',
            borderBottom: '1px solid rgba(255,255,255,0.06)',
          },
          body: {
            background: 'rgba(8,8,16,0.98)',
            padding: '16px 20px',
          },
          wrapper: { background: 'transparent' },
        }}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          {navItems.map((item, i) => {
            const active = isActive(item.path)
            return (
              <Link
                key={item.path}
                to={item.path}
                onClick={() => setDrawerOpen(false)}
                style={{
                  display: 'flex', alignItems: 'center', gap: 14,
                  fontSize: 15, padding: '12px 16px',
                  borderRadius: 14,
                  fontFamily: "'Noto Serif SC', serif",
                  fontWeight: active ? 500 : 400,
                  color: active ? '#ffffff' : 'rgba(255,255,255,0.55)',
                  background: active ? itemAccents[i] : 'transparent',
                  textDecoration: 'none',
                  letterSpacing: '0.03em',
                  transition: 'all 0.3s ease',
                }}
              >
                <span style={{ fontSize: 18, width: 24, textAlign: 'center' }}>{item.icon}</span>
                {item.label}
              </Link>
            )
          })}
          <div style={{ borderTop: '1px solid rgba(255,255,255,0.06)', paddingTop: 16, marginTop: 12 }}>
            {isAuthenticated ? (
              <Button block onClick={handleLogout} icon={<LogoutOutlined />} danger
                style={{ borderRadius: 12, height: 42, fontWeight: 500, fontFamily: "'Noto Serif SC', serif" }}>
                退出登录 ({user?.name || user?.username})
              </Button>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                <Button block type="primary" icon={<LoginOutlined />}
                  onClick={() => { setDrawerOpen(false); navigate('/login') }}
                  style={{ borderRadius: 12, height: 42, fontWeight: 500, fontFamily: "'Noto Serif SC', serif" }}>
                  登录
                </Button>
                <Button block icon={<UserAddOutlined />}
                  onClick={() => { setDrawerOpen(false); navigate('/register') }}
                  style={{ borderRadius: 12, height: 42, fontWeight: 500, fontFamily: "'Noto Serif SC', serif" }}>
                  注册
                </Button>
              </div>
            )}
          </div>
        </div>
      </Drawer>
    </>
  )
}

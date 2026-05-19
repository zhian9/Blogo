import { useState, useEffect, useMemo } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { Button, Typography, Spin, message, Breadcrumb, Input } from 'antd'
import {
  ArrowLeftOutlined, HomeOutlined, SearchOutlined, TeamOutlined,
  HeartOutlined,
} from '@ant-design/icons'
import { motion, AnimatePresence } from 'framer-motion'
import { useAppStore } from '../store/appStore'
import { listFollowers, listFollowing } from '../api/follow'
import UserCard from '../components/UserCard'
import type { AuthorInfo } from '../types'

const { Text } = Typography

interface Props {
  type: 'followers' | 'following'
}

export default function UserList({ type }: Props) {
  const { id } = useParams<{ id: string }>()
  const userId = id || ''
  const navigate = useNavigate()
  const theme = useAppStore((s) => s.theme)
  const isDark = theme === 'dark'

  const [users, setUsers] = useState<AuthorInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [searchQuery, setSearchQuery] = useState('')

  const fetchUsers = async (p: number) => {
    setLoading(true)
    try {
      const fetcher = type === 'followers' ? listFollowers : listFollowing
      const res = await fetcher(userId, p, 24)
      const list = (res.data || []) as AuthorInfo[]
      setUsers(p === 1 ? list : [...users, ...list])
      setTotal(res.total || list.length)
    } catch (err: any) {
      message.error(err.message || '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    setUsers([])
    setPage(1)
    setTotal(0)
    setSearchQuery('')
    fetchUsers(1)
  }, [userId, type])

  const title = type === 'followers' ? '粉丝' : '关注'
  const tabLabel = type === 'followers' ? '粉丝列表' : '关注列表'

  // Client-side search filter
  const filteredUsers = useMemo(() => {
    if (!searchQuery.trim()) return users
    const q = searchQuery.toLowerCase()
    return users.filter((u) => {
      const name = (u.name || '').toLowerCase()
      const username = (u.username || '').toLowerCase()
      const bio = (u.bio || '').toLowerCase()
      return name.includes(q) || username.includes(q) || bio.includes(q)
    })
  }, [users, searchQuery])

  const hasMore = users.length < total

  // ── Color palette ──
  const colors = {
    bg: isDark ? '#0f0f0f' : '#f5f7fa',
    surface: isDark ? '#1a1a2e' : '#ffffff',
    text: isDark ? '#e8e8f0' : '#1a1a2e',
    textSecondary: isDark ? '#8892b0' : '#6b7280',
    textMuted: isDark ? '#5a6080' : '#9ca3af',
    border: isDark ? '#2a2a4a' : '#e5e7eb',
    accent: '#4f6ef7',
    gradientFrom: isDark ? '#0f0c29' : '#e8f0fe',
    gradientVia: isDark ? '#1a1a4e' : '#dbeafe',
    gradientTo: isDark ? '#24243e' : '#eff6ff',
  }

  // ── Animation variants ──
  const containerVariants = {
    hidden: { opacity: 0 },
    visible: { opacity: 1, transition: { duration: 0.5, staggerChildren: 0.06 } },
  }
  const itemVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: { opacity: 1, y: 0, transition: { duration: 0.4, ease: 'easeOut' } },
  }

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={containerVariants}
      style={{ margin: -24, marginTop: -24 }}
    >
      {/* ==================== HERO HEADER ==================== */}
      <motion.div
        variants={itemVariants}
        style={{
          position: 'relative',
          width: '100%',
          background: `linear-gradient(160deg, ${colors.gradientFrom} 0%, ${colors.gradientVia} 40%, ${colors.gradientTo} 100%)`,
          overflow: 'hidden',
          padding: '40px 6% 32px',
        }}
      >
        {/* Decorative background elements */}
        <div style={{
          position: 'absolute', top: -60, right: -30,
          width: 240, height: 240, borderRadius: '50%',
          background: isDark ? 'rgba(79,110,247,0.07)' : 'rgba(79,110,247,0.05)',
          pointerEvents: 'none',
        }} />
        <div style={{
          position: 'absolute', bottom: -40, left: '5%',
          width: 180, height: 180, borderRadius: '50%',
          background: isDark ? 'rgba(139,92,246,0.05)' : 'rgba(139,92,246,0.04)',
          pointerEvents: 'none',
        }} />
        <div style={{
          position: 'absolute', inset: 0,
          backgroundImage: `radial-gradient(circle, ${isDark ? 'rgba(255,255,255,0.03)' : 'rgba(0,0,0,0.02)'} 1px, transparent 1px)`,
          backgroundSize: '20px 20px',
          pointerEvents: 'none',
        }} />

        {/* Back button */}
        <motion.div whileHover={{ x: -3 }} style={{ position: 'relative', zIndex: 2, marginBottom: 16 }}>
          <Button
            type="text"
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate(-1)}
            style={{
              color: isDark ? '#8892b0' : '#6b7280',
              fontSize: 14, padding: '4px 12px', borderRadius: 8,
              background: isDark ? 'rgba(255,255,255,0.05)' : 'rgba(255,255,255,0.6)',
              backdropFilter: 'blur(8px)', border: 'none',
            }}
          >
            返回
          </Button>
        </motion.div>

        {/* Breadcrumb */}
        <Breadcrumb
          style={{ marginBottom: 12, position: 'relative', zIndex: 2 }}
          items={[
            { title: <Link to="/" style={{ color: isDark ? '#8892b0' : '#6b7280' }}><HomeOutlined /> 首页</Link> },
            { title: <Link to={`/user/${userId}`} style={{ color: isDark ? '#8892b0' : '#6b7280' }}>用户主页</Link> },
            { title: <Text style={{ color: colors.accent, fontWeight: 500 }}>{title}</Text> },
          ]}
        />

        {/* Title with count */}
        <div style={{
          position: 'relative', zIndex: 2,
          display: 'flex', alignItems: 'baseline', gap: 12, flexWrap: 'wrap',
          marginBottom: 24,
        }}>
          <h1 style={{
            margin: 0, fontSize: 32, fontWeight: 800, color: colors.text,
            letterSpacing: '-0.5px',
          }}>
            {tabLabel}
          </h1>
          <span style={{
            fontSize: 18, fontWeight: 600, color: colors.accent,
            background: `${colors.accent}15`, padding: '2px 14px', borderRadius: 20,
          }}>
            {total} 人
          </span>
        </div>

        {/* Search + Tabs row */}
        <div style={{
          position: 'relative', zIndex: 2,
          display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap',
        }}>
          {/* Search input */}
          <div style={{ flex: 1, minWidth: 240, maxWidth: 400 }}>
            <Input
              prefix={<SearchOutlined style={{ color: colors.textMuted }} />}
              placeholder="搜索用户名或简介..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              allowClear
              style={{
                borderRadius: 12,
                height: 44,
                background: isDark ? 'rgba(255,255,255,0.06)' : 'rgba(255,255,255,0.7)',
                backdropFilter: 'blur(8px)',
                border: `1px solid ${isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)'}`,
                color: colors.text,
                fontSize: 14,
              }}
            />
          </div>

          {/* Tabs: followers | following */}
          <div style={{
            display: 'flex', gap: 4,
            background: isDark ? 'rgba(255,255,255,0.04)' : 'rgba(255,255,255,0.5)',
            borderRadius: 12, padding: 4,
            border: `1px solid ${isDark ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.06)'}`,
          }}>
            {(['followers', 'following'] as const).map((t) => {
              const isActive = type === t
              return (
                <Link key={t} to={`/user/${userId}/${t}`} style={{ textDecoration: 'none' }}>
                  <motion.div
                    whileHover={{ scale: 1.03 }}
                    whileTap={{ scale: 0.97 }}
                    style={{
                      display: 'flex', alignItems: 'center', gap: 6,
                      padding: '8px 18px', borderRadius: 10,
                      background: isActive
                        ? (isDark ? 'rgba(79,110,247,0.2)' : 'rgba(79,110,247,0.1)')
                        : 'transparent',
                      color: isActive ? colors.accent : colors.textMuted,
                      fontWeight: isActive ? 600 : 500,
                      fontSize: 14,
                      transition: 'all 0.25s ease',
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {t === 'followers' ? <TeamOutlined /> : <HeartOutlined />}
                    {t === 'followers' ? '粉丝' : '关注'}
                  </motion.div>
                </Link>
              )
            })}
          </div>
        </div>
      </motion.div>

      {/* ==================== CONTENT AREA ==================== */}
      <div style={{
        padding: '32px 6% 60px',
        background: colors.bg,
        minHeight: 400,
      }}>
        {/* Loading state */}
        {loading && users.length === 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            style={{
              display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
              padding: 80, gap: 16,
            }}
          >
            <Spin size="large" />
            <Text style={{ color: colors.textMuted, fontSize: 14 }}>加载中...</Text>
          </motion.div>
        ) : filteredUsers.length === 0 ? (
          /* Empty state */
          <AnimatePresence mode="wait">
            <motion.div
              key="empty"
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.95 }}
              transition={{ duration: 0.4 }}
              style={{
                display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
                padding: 80, gap: 20,
              }}
            >
              {/* Empty illustration */}
              <div style={{
                width: 120, height: 120, borderRadius: '50%',
                background: isDark ? 'rgba(79,110,247,0.08)' : 'rgba(79,110,247,0.06)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}>
                {type === 'followers' ? (
                  <TeamOutlined style={{ fontSize: 48, color: isDark ? '#4f6ef7' : '#6366f1', opacity: 0.6 }} />
                ) : (
                  <HeartOutlined style={{ fontSize: 48, color: isDark ? '#4f6ef7' : '#6366f1', opacity: 0.6 }} />
                )}
              </div>
              <Text style={{ color: colors.text, fontSize: 18, fontWeight: 600 }}>
                {searchQuery.trim() ? '未找到匹配的用户' : type === 'followers' ? '暂无粉丝' : '暂未关注任何人'}
              </Text>
              <Text style={{ color: colors.textMuted, fontSize: 14, textAlign: 'center', maxWidth: 360 }}>
                {searchQuery.trim()
                  ? '试试其他关键词搜索'
                  : type === 'followers'
                    ? '当有人关注时，他们会出现在这里'
                    : '去发现一些有趣的人并关注他们吧'}
              </Text>
            </motion.div>
          </AnimatePresence>
        ) : (
          /* User cards grid */
          <AnimatePresence mode="wait">
            <motion.div
              key="grid"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.3 }}
            >
              <motion.div
                variants={containerVariants}
                initial="hidden"
                animate="visible"
                style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(4, 1fr)',
                  gap: 24,
                }}
                className="user-list-grid"
              >
                {filteredUsers.map((u, i) => (
                  <UserCard key={u.id} user={u} index={i} />
                ))}
              </motion.div>

              {/* Load more */}
              {hasMore && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  transition={{ delay: 0.3 }}
                  style={{ textAlign: 'center', padding: '40px 0 20px' }}
                >
                  <motion.div whileHover={{ scale: 1.03 }} whileTap={{ scale: 0.97 }}>
                    <Button
                      loading={loading}
                      onClick={() => { const np = page + 1; setPage(np); fetchUsers(np) }}
                      size="large"
                      style={{
                        borderRadius: 12,
                        height: 48,
                        paddingInline: 40,
                        fontWeight: 600,
                        fontSize: 15,
                        background: isDark ? 'rgba(255,255,255,0.05)' : '#fff',
                        border: `1px solid ${colors.border}`,
                        color: colors.text,
                      }}
                    >
                      {loading ? '加载中...' : `加载更多用户（${total - users.length}）`}
                    </Button>
                  </motion.div>
                </motion.div>
              )}

              {/* End of list marker */}
              {!hasMore && users.length > 0 && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  transition={{ delay: 0.3 }}
                  style={{ textAlign: 'center', padding: '40px 0 20px' }}
                >
                  <div style={{
                    display: 'inline-flex', alignItems: 'center', gap: 8,
                    padding: '10px 20px', borderRadius: 20,
                    background: isDark ? 'rgba(255,255,255,0.03)' : 'rgba(0,0,0,0.02)',
                  }}>
                    <div style={{
                      width: 6, height: 6, borderRadius: '50%',
                      background: colors.accent, opacity: 0.5,
                    }} />
                    <Text style={{ color: colors.textMuted, fontSize: 13 }}>
                      已展示全部 {total} 位用户
                    </Text>
                  </div>
                </motion.div>
              )}
            </motion.div>
          </AnimatePresence>
        )}

        {/* Search result hint */}
        {searchQuery.trim() && filteredUsers.length > 0 && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            style={{ textAlign: 'center', paddingBottom: 20 }}
          >
            <Text style={{ color: colors.textMuted, fontSize: 13 }}>
              找到 {filteredUsers.length} 个匹配结果
            </Text>
          </motion.div>
        )}
      </div>

      {/* ==================== RESPONSIVE CSS ==================== */}
      <style>{`
        @media (max-width: 1400px) {
          .user-list-grid {
            grid-template-columns: repeat(3, 1fr) !important;
          }
        }
        @media (max-width: 1024px) {
          .user-list-grid {
            grid-template-columns: repeat(2, 1fr) !important;
            gap: 20px !important;
          }
        }
        @media (max-width: 640px) {
          .user-list-grid {
            grid-template-columns: 1fr !important;
            gap: 16px !important;
          }
        }
      `}</style>
    </motion.div>
  )
}

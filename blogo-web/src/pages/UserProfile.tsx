import { useState, useEffect, useRef, useMemo } from 'react'
import { useParams, useSearchParams, useNavigate, Link } from 'react-router-dom'
import { Typography, Tabs, Button, Avatar, Spin, Empty, message, Affix, Card } from 'antd'
import {
  EditOutlined, FileTextOutlined, HeartOutlined, EyeOutlined,
  UserOutlined, InfoCircleOutlined, StarOutlined,
  ArrowLeftOutlined, PlusOutlined,
  FireOutlined, ThunderboltOutlined,
} from '@ant-design/icons'
import { motion, AnimatePresence } from 'framer-motion'
import { useUserProfile } from '../hooks/useProfile'
import { useFollowStatus, useFollow, useUnfollow } from '../hooks/useFollow'
import { useAuthStore } from '../store/authStore'
import ArticleGridCard from '../components/ArticleGridCard'
import ContributionHeatmap from '../components/ContributionHeatmap'
import ActivityStats from '../components/ActivityStats'
import type { Article } from '../types'
import dayjs from '../utils/dayjs'

const { Title, Text, Paragraph } = Typography

// ── Dark theme tokens ──
const c = {
  bg: '#0a0a10',
  surface: 'rgba(255,255,255,0.03)',
  border: 'rgba(255,255,255,0.06)',
  text: 'rgba(255,255,255,0.85)',
  textMuted: 'rgba(255,255,255,0.35)',
  accent: '#4f6ef7',
  accentGlow: 'rgba(79,110,247,0.25)',
}

const tabMeta = [
  { key: 'articles', label: '文章', icon: <FileTextOutlined /> },
  { key: 'liked', label: '赞过', icon: <HeartOutlined /> },
  { key: 'favorites', label: '收藏', icon: <StarOutlined /> },
  { key: 'about', label: '关于', icon: <InfoCircleOutlined /> },
]

export default function UserProfile() {
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const userId = id || ''
  const defaultTab = searchParams.get('tab') || 'articles'
  const [activeTab, setActiveTab] = useState(defaultTab)
  const [uploading, setUploading] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const { data: profileData, isLoading } = useUserProfile(userId)
  const { data: followData } = useFollowStatus(userId)
  const followMutation = useFollow(userId)
  const unfollowMutation = useUnfollow(userId)

  const currentUser = useAuthStore((s) => s.user)
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const isSelf = currentUser ? currentUser.id === userId : false
  const setUser = useAuthStore((s) => s.setUser)

  const profile = profileData?.data
  const user = profile?.user
  const isFollowing = !!followData?.data

  const totalViews = useMemo(() => {
    if (!profile?.articles) return 0
    return profile.articles.reduce((sum: number, a: Article) => sum + (a.views || 0), 0)
  }, [profile])

  const contributionStats = profile?.contribution_stats

  // ── Follow handler ──
  const handleFollow = async () => {
    if (!isAuthenticated) { message.warning('请先登录'); return }
    try {
      if (isFollowing) {
        await unfollowMutation.mutateAsync()
        message.success('已取消关注')
      } else {
        await followMutation.mutateAsync()
        message.success('已关注')
      }
    } catch (err: any) { message.error(err.message || '操作失败') }
  }

  // ── Avatar upload ──
  const handleAvatarClick = () => fileInputRef.current?.click()
  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    if (!file.type.startsWith('image/')) { message.error('请选择图片文件'); return }
    if (file.size > 5 * 1024 * 1024) { message.error('文件不能超过 5MB'); return }
    setUploading(true)
    try {
      const formData = new FormData(); formData.append('file', file)
      const { default: client } = await import('../api/client')
      const uploadRes = await client.post('/images/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
      const imageUrl = uploadRes.data?.data?.url
      if (imageUrl) {
        const { updateCurrentUser } = await import('../api/auth')
        await updateCurrentUser({ avatar: imageUrl, name: user?.name || '', phone: user?.phone || '', email: user?.email || '', bio: user?.bio || '', remark: '' })
        if (isSelf && currentUser) setUser({ ...currentUser, avatar: imageUrl })
        message.success('头像已更新')
        window.location.reload()
      }
    } catch (err: any) { message.error(err.message || '上传失败') }
    finally { setUploading(false) }
  }

  useEffect(() => {
    const tab = searchParams.get('tab')
    if (tab) setActiveTab(tab)
  }, [searchParams])

  if (isLoading) return <div style={{ textAlign: 'center', padding: 120 }}><Spin size="large" /></div>
  if (!profile || !user) {
    return <div style={{ textAlign: 'center', padding: 120 }}><Empty description="用户不存在" /><Link to="/">返回首页</Link></div>
  }

  // ── Stats cards ──
  const statsCards = [
    { label: '文章', value: profile.articles?.length || 0, icon: <FileTextOutlined />, color: '#4f6ef7' },
    { label: '粉丝', value: user.follower_count || 0, icon: <HeartOutlined />, color: '#ec4899' },
    { label: '关注', value: user.following_count || 0, icon: <UserOutlined />, color: '#10b981' },
    { label: '阅读', value: totalViews, icon: <EyeOutlined />, color: '#f59e0b' },
    { label: '贡献', value: contributionStats?.total_contributions || 0, icon: <FireOutlined />, color: '#818cf8' },
    { label: '连续', value: `${contributionStats?.current_streak || 0}天`, icon: <ThunderboltOutlined />, color: '#c084fc' },
  ]

  // ── Tab content ──
  const renderTabContent = () => {
    switch (activeTab) {
      case 'articles': return renderArticleGrid(profile.articles || [], '还没有发布文章')
      case 'liked': return renderArticleGrid(profile.liked_articles || [], '还没有点赞文章')
      case 'favorites': return renderArticleGrid(profile.favorite_articles || [], '还没有收藏文章')
      case 'about': return renderAbout()
      default: return null
    }
  }

  const renderArticleGrid = (articles: Article[], emptyText: string) => {
    if (!articles || articles.length === 0) {
      return <div style={{ textAlign: 'center', padding: 60 }}><Empty description={emptyText} /></div>
    }
    return (
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: 20 }}>
        {articles.map((a, i) => (
          <motion.div key={a.id} initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, delay: i * 0.05 }}>
            <ArticleGridCard article={a} index={i} />
          </motion.div>
        ))}
      </div>
    )
  }

  const renderAbout = () => (
    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
      <div style={{ gridColumn: '1 / -1' }}>
        <Card bordered={false} style={{ background: c.surface, border: `1px solid ${c.border}`, borderRadius: 16 }}
          styles={{ body: { padding: '24px 28px' } }}>
          <Text strong style={{ color: c.text, fontSize: 15 }}>简介</Text>
          <Paragraph style={{ color: c.textMuted, margin: '12px 0 0', fontSize: 14, lineHeight: 1.8 }}>
            {user.bio || '这个人很懒，什么都没写...'}
          </Paragraph>
        </Card>
      </div>
      <Card bordered={false} style={{ background: c.surface, border: `1px solid ${c.border}`, borderRadius: 16 }}
        styles={{ body: { padding: '20px 24px' } }}>
        <Text strong style={{ color: c.text, fontSize: 14 }}>联系方式</Text>
        <div style={{ marginTop: 16, display: 'flex', flexDirection: 'column', gap: 12 }}>
          <InfoRow icon={<UserOutlined />} label="用户名" value={user.username} />
          <InfoRow icon={<UserOutlined />} label="邮箱" value={user.email || '未填写'} />
          <InfoRow icon={<UserOutlined />} label="手机" value={user.phone || '未填写'} />
        </div>
      </Card>
      <Card bordered={false} style={{ background: c.surface, border: `1px solid ${c.border}`, borderRadius: 16 }}
        styles={{ body: { padding: '20px 24px' } }}>
        <Text strong style={{ color: c.text, fontSize: 14 }}>账号信息</Text>
        <div style={{ marginTop: 16, display: 'flex', flexDirection: 'column', gap: 12 }}>
          <InfoRow icon={<UserOutlined />} label="加入时间" value={dayjs(user.created_at).format('YYYY-MM-DD')} />
          <InfoRow icon={<UserOutlined />} label="最近活跃" value={user.last_login_at ? dayjs(user.last_login_at).format('YYYY-MM-DD') : '未知'} />
          <InfoRow icon={<FileTextOutlined />} label="文章数" value={`${profile.articles?.length || 0} 篇`} />
          <InfoRow icon={<HeartOutlined />} label="获赞" value={`${totalViews} 次`} />
        </div>
      </Card>
    </div>
  )

  return (
    <div style={{ background: c.bg, minHeight: '100vh' }}>
      {/* ================================================================ */}
      {/*  HERO BANNER                                                     */}
      {/* ================================================================ */}
      <div style={{
        position: 'relative', overflow: 'hidden',
        background: 'linear-gradient(180deg, #0f0f2b 0%, #141430 30%, #0a0a10 100%)',
        padding: '48px 24px 40px',
      }}>
        {/* Starfield */}
        <div style={{ position: 'absolute', inset: 0, overflow: 'hidden' }}>
          {Array.from({ length: 40 }).map((_, i) => (
            <div key={i} style={{
              position: 'absolute',
              left: `${Math.random() * 100}%`, top: `${Math.random() * 100}%`,
              width: `${1 + Math.random() * 2}px`, height: `${1 + Math.random() * 2}px`,
              borderRadius: '50%', background: 'rgba(255,255,255,0.4)',
              animation: `twinkle ${2 + Math.random() * 3}s infinite ${Math.random() * 3}s`,
            }} />
          ))}
        </div>
        {/* Glow orbs */}
        <div style={{ position: 'absolute', top: -120, left: '20%', width: 300, height: 300, borderRadius: '50%', background: 'radial-gradient(circle, rgba(79,110,247,0.12) 0%, transparent 70%)' }} />
        <div style={{ position: 'absolute', bottom: -80, right: '10%', width: 240, height: 240, borderRadius: '50%', background: 'radial-gradient(circle, rgba(139,92,246,0.08) 0%, transparent 70%)' }} />

        <div style={{ position: 'relative', zIndex: 1, maxWidth: 720, margin: '0 auto', textAlign: 'center' }}>
          <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}
            style={{ position: 'absolute', left: 0, top: 0, color: c.textMuted }} />

          {/* Avatar */}
          <div style={{ position: 'relative', display: 'inline-block', marginBottom: 20 }}>
            <Avatar
              src={user.avatar ? (user.avatar.startsWith('http') ? user.avatar : `${import.meta.env.VITE_API_URL || ''}/api/v1/images/${user.avatar}/file`) : undefined}
              icon={!user.avatar ? <UserOutlined /> : undefined}
              size={96}
              style={{ border: '3px solid rgba(255,255,255,0.12)', boxShadow: `0 0 40px ${c.accentGlow}` }}
            />
            {isSelf && (
              <>
                <div onClick={handleAvatarClick} style={{
                  position: 'absolute', bottom: 2, right: 2,
                  width: 32, height: 32, borderRadius: '50%',
                  background: c.accent, border: '2px solid #0a0a10',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  cursor: 'pointer', color: '#fff', fontSize: 14,
                  transition: 'transform 0.2s ease',
                }}
                  onMouseEnter={(e) => e.currentTarget.style.transform = 'scale(1.1)'}
                  onMouseLeave={(e) => e.currentTarget.style.transform = 'scale(1)'}
                >
                  {uploading ? <Spin size="small" /> : <PlusOutlined />}
                </div>
                <input ref={fileInputRef} type="file" accept="image/*" style={{ display: 'none' }} onChange={handleFileChange} />
              </>
            )}
          </div>

          <Title level={2} style={{ color: '#ffffff', margin: '0 0 8px', fontWeight: 700, fontFamily: "'Barlow', sans-serif", letterSpacing: '-0.02em' }}>
            {user.name || user.username}
          </Title>
          <Text style={{ color: c.textMuted, fontSize: 15, fontFamily: "'Barlow', sans-serif" }}>
            @{user.username}{user.bio && <> · {user.bio.slice(0, 80)}</>}
          </Text>

          <div style={{ marginTop: 20 }}>
            {isSelf ? (
              <Link to="/profile">
                <Button type="default" icon={<EditOutlined />} style={{ borderRadius: 10, background: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.1)', color: c.text }}>
                  编辑资料
                </Button>
              </Link>
            ) : isAuthenticated ? (
              <Button
                type={isFollowing ? 'default' : 'primary'}
                onClick={handleFollow}
                loading={followMutation.isPending || unfollowMutation.isPending}
                style={isFollowing
                  ? { borderRadius: 10, background: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.1)', color: c.text }
                  : { borderRadius: 10, background: c.accent, border: 'none', boxShadow: `0 2px 16px ${c.accentGlow}` }
                }>
                {isFollowing ? '已关注' : '关注'}
              </Button>
            ) : null}
          </div>
        </div>
      </div>

      {/* ================================================================ */}
      {/*  STATS CARDS                                                     */}
      {/* ================================================================ */}
      <div style={{ maxWidth: 960, margin: '-24px auto 0', padding: '0 16px', position: 'relative', zIndex: 2 }}>
        <div style={{ display: 'grid', gridTemplateColumns: `repeat(${statsCards.length}, 1fr)`, gap: 12 }}>
          {statsCards.map((stat, i) => (
            <motion.div key={stat.label}
              initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.35, delay: i * 0.06, ease: 'easeOut' }}>
              <Link
                to={stat.label === '粉丝' ? `/user/${userId}/followers` : stat.label === '关注' ? `/user/${userId}/following` : '#'}
                style={{ textDecoration: 'none', pointerEvents: (stat.label === '粉丝' || stat.label === '关注') ? 'auto' : 'none' }}
                onClick={(e) => { if (stat.label !== '粉丝' && stat.label !== '关注') e.preventDefault() }}>
                <Card bordered={false}
                  style={{
                    background: 'rgba(20,20,40,0.8)', backdropFilter: 'blur(12px)',
                    border: '1px solid rgba(255,255,255,0.06)', borderRadius: 16,
                    textAlign: 'center', transition: 'all 0.3s ease',
                  }}
                  styles={{ body: { padding: '16px 12px' } }}
                  className="profile-stat-card"
                  onMouseEnter={(e) => {
                    e.currentTarget.style.background = 'rgba(30,30,55,0.85)'
                    e.currentTarget.style.borderColor = 'rgba(255,255,255,0.12)'
                    e.currentTarget.style.transform = 'translateY(-2px)'
                    e.currentTarget.style.boxShadow = '0 8px 24px rgba(0,0,0,0.3)'
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = 'rgba(20,20,40,0.8)'
                    e.currentTarget.style.borderColor = 'rgba(255,255,255,0.06)'
                    e.currentTarget.style.transform = 'translateY(0)'
                    e.currentTarget.style.boxShadow = 'none'
                  }}>
                  <div style={{ fontSize: 22, color: stat.color, marginBottom: 4 }}>{stat.icon}</div>
                  <div style={{ fontSize: 24, fontWeight: 800, color: '#ffffff', fontFamily: "'Barlow', sans-serif", lineHeight: 1.1 }}>
                    {typeof stat.value === 'number' && stat.value >= 1000 ? `${(stat.value / 1000).toFixed(1)}k` : stat.value}
                  </div>
                  <div style={{ fontSize: 11, color: c.textMuted, marginTop: 4, fontFamily: "'Barlow', sans-serif", letterSpacing: '0.04em' }}>
                    {stat.label}
                  </div>
                </Card>
              </Link>
            </motion.div>
          ))}
        </div>
      </div>

      {/* ================================================================ */}
      {/*  CONTRIBUTION ACTIVITY CENTER                                     */}
      {/* ================================================================ */}
      {profile.contributions && profile.contributions.length > 0 && (
        <motion.div
          initial={{ opacity: 0, y: 24 }} animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.3, ease: 'easeOut' }}
          style={{ maxWidth: 960, margin: '28px auto 0', padding: '0 16px' }}>
          <Card bordered={false}
            style={{
              background: 'linear-gradient(135deg, rgba(15,15,35,0.9) 0%, rgba(20,20,50,0.85) 100%)',
              backdropFilter: 'blur(20px)',
              border: '1px solid rgba(255,255,255,0.06)',
              borderRadius: 24,
              boxShadow: '0 4px 40px rgba(0,0,0,0.3), inset 0 1px 0 rgba(255,255,255,0.03)',
              overflow: 'hidden',
            }}
            styles={{ body: { padding: 'clamp(16px, 3vw, 32px)' } }}>
            {/* Header */}
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 20 }}>
              <div style={{
                width: 36, height: 36, borderRadius: 12, flexShrink: 0,
                background: 'rgba(129,140,248,0.15)', border: '1px solid rgba(129,140,248,0.3)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: 18, color: '#818cf8',
              }}>
                <FireOutlined />
              </div>
              <div style={{ minWidth: 0 }}>
                <Text strong style={{ color: '#ffffff', fontSize: 16, fontFamily: "'Barlow', sans-serif", letterSpacing: '-0.01em' }}>
                  Contribution Activity
                </Text>
                <Text style={{ color: 'rgba(255,255,255,0.3)', fontSize: 11, display: 'block', fontFamily: "'Barlow', sans-serif" }}>
                  过去 365 天活跃记录
                </Text>
              </div>
            </div>

            {/* Two-column layout: PC row, mobile column */}
            <div className="contribution-layout" style={{ display: 'flex', gap: 32, alignItems: 'stretch' }}>
              <div style={{ flex: 1, minWidth: 0, overflow: 'hidden' }}>
                <ContributionHeatmap data={profile.contributions} />
              </div>
              <div style={{ width: 260, minWidth: 220, flexShrink: 0 }} className="contribution-sidebar">
                <ActivityStats stats={profile.contribution_stats} />
              </div>
            </div>
          </Card>
        </motion.div>
      )}

      {/* ================================================================ */}
      {/*  TABS + CONTENT                                                   */}
      {/* ================================================================ */}
      <div style={{ maxWidth: 960, margin: '0 auto', padding: '28px 16px 0' }}>
        <Affix offsetTop={0}>
          <div style={{
            background: 'rgba(10,10,16,0.88)', backdropFilter: 'blur(16px)',
            borderBottom: '1px solid rgba(255,255,255,0.05)',
            borderRadius: '16px 16px 0 0', padding: '4px 16px',
          }}>
            <Tabs
              activeKey={activeTab}
              onChange={(key) => {
                setActiveTab(key)
                const url = new URL(window.location.href)
                url.searchParams.set('tab', key)
                window.history.replaceState({}, '', url.toString())
              }}
              style={{ marginBottom: 0 }}
              items={tabMeta.map((t) => {
                let count = 0
                if (t.key === 'articles') count = profile.articles?.length || 0
                else if (t.key === 'liked') count = profile.liked_articles?.length || 0
                else if (t.key === 'favorites') count = profile.favorite_articles?.length || 0
                return {
                  key: t.key,
                  label: (
                    <span style={{ display: 'flex', alignItems: 'center', gap: 6, fontFamily: "'Barlow', sans-serif", fontSize: 13 }}>
                      {t.icon} {t.label}
                      {t.key !== 'about' && (
                        <span style={{ fontSize: 11, color: c.textMuted, background: 'rgba(255,255,255,0.05)', padding: '0 6px', borderRadius: 6 }}>
                          {count}
                        </span>
                      )}
                    </span>
                  ),
                }
              })}
            />
          </div>
        </Affix>

        <div style={{ padding: '24px 0 48px' }}>
          <AnimatePresence mode="wait">
            <motion.div key={activeTab}
              initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -12 }}
              transition={{ duration: 0.25, ease: 'easeOut' }}>
              {renderTabContent()}
            </motion.div>
          </AnimatePresence>
        </div>
      </div>

      {/* ================================================================ */}
      {/*  STYLES                                                           */}
      {/* ================================================================ */}
      <style>{`
        @keyframes twinkle {
          0%, 100% { opacity: 0.2; transform: scale(1); }
          50% { opacity: 1; transform: scale(1.5); }
        }
        .profile-stat-card {
          transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
        }
        @media (max-width: 768px) {
          .contribution-layout {
            flex-direction: column !important;
          }
          .contribution-sidebar {
            width: 100% !important;
          }
        }
      `}</style>
    </div>
  )
}

// ── Helper ──
function InfoRow({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 10, fontSize: 13 }}>
      <span style={{ color: 'rgba(255,255,255,0.3)', fontSize: 13 }}>{icon}</span>
      <span style={{ color: 'rgba(255,255,255,0.4)', minWidth: 56, fontFamily: "'Barlow', sans-serif" }}>{label}</span>
      <span style={{ color: 'rgba(255,255,255,0.75)', fontFamily: "'Barlow', sans-serif" }}>{value || '-'}</span>
    </div>
  )
}

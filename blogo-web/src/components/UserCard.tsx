import { Link } from 'react-router-dom'
import { Avatar, Button, message } from 'antd'
import { UserOutlined, PlusOutlined, CheckOutlined, TeamOutlined, HeartOutlined } from '@ant-design/icons'
import { motion } from 'framer-motion'
import { useAuthStore } from '../store/authStore'
import { useAppStore } from '../store/appStore'
import { useFollowStatus, useFollow, useUnfollow } from '../hooks/useFollow'
import type { AuthorInfo } from '../types'

interface Props {
  user: AuthorInfo
  index: number
}

export default function UserCard({ user, index }: Props) {
  const currentUser = useAuthStore((s) => s.user)
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const isSelf = currentUser?.id === user.id
  const theme = useAppStore((s) => s.theme)
  const isDark = theme === 'dark'

  const { data: followData } = useFollowStatus(user.id)
  const followMutation = useFollow(user.id)
  const unfollowMutation = useUnfollow(user.id)
  const isFollowing = followData?.data || false
  const followLoading = followMutation.isPending || unfollowMutation.isPending

  const handleFollow = (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!isAuthenticated) { message.warning('请先登录'); return }
    if (isFollowing) {
      unfollowMutation.mutate(undefined, { onError: (err: any) => message.error(err.message) })
    } else {
      followMutation.mutate(undefined, { onError: (err: any) => message.error(err.message) })
    }
  }

  const cardVariants = {
    hidden: { opacity: 0, y: 24 },
    visible: {
      opacity: 1, y: 0,
      transition: { duration: 0.45, delay: index * 0.06, ease: 'easeOut' },
    },
  }

  return (
    <motion.div variants={cardVariants}>
      <Link to={`/user/${user.id}`} style={{ textDecoration: 'none' }}>
        <motion.div
          whileHover={{ y: -6 }}
          style={{
            background: isDark ? '#1a1a2e' : '#fff',
            borderRadius: 20,
            padding: '28px 24px 20px',
            border: `1px solid ${isDark ? '#2a2a4a' : '#e5e7eb'}`,
            display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 14,
            cursor: 'pointer',
            transition: 'all 0.3s ease',
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.boxShadow = isDark
              ? '0 12px 40px rgba(79,110,247,0.15)'
              : '0 12px 40px rgba(0,0,0,0.1)'
            e.currentTarget.style.borderColor = '#4f6ef7'
            const avatarEl = e.currentTarget.querySelector('.user-card-avatar') as HTMLElement
            if (avatarEl) avatarEl.style.transform = 'scale(1.08)'
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = 'none'
            e.currentTarget.style.borderColor = isDark ? '#2a2a4a' : '#e5e7eb'
            const avatarEl = e.currentTarget.querySelector('.user-card-avatar') as HTMLElement
            if (avatarEl) avatarEl.style.transform = 'scale(1)'
          }}
        >
          {/* Avatar with gradient ring */}
          <div className="user-card-avatar" style={{ transition: 'transform 0.35s ease' }}>
            <div style={{
              width: 88, height: 88, borderRadius: '50%',
              padding: 3,
              background: 'linear-gradient(135deg, #4f6ef7, #8b5cf6, #ec4899)',
              boxShadow: isDark ? '0 4px 20px rgba(79,110,247,0.3)' : '0 4px 20px rgba(79,110,247,0.2)',
            }}>
              <Avatar
                size={82}
                src={user.avatar || undefined}
                icon={!user.avatar ? <UserOutlined /> : undefined}
                style={{
                  backgroundColor: user.avatar ? undefined : '#4f6ef7',
                  border: '3px solid #fff',
                }}
              />
            </div>
          </div>

          {/* Name and username */}
          <div style={{ textAlign: 'center', width: '100%', minWidth: 0 }}>
            <h3 style={{
              margin: 0, fontSize: 16, fontWeight: 700,
              color: isDark ? '#e8e8f0' : '#1a1a2e',
              overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
            }}>
              {user.name || user.username || user.id}
            </h3>
            {(user.username && user.name) && (
              <p style={{
                margin: '2px 0 0', fontSize: 12,
                color: isDark ? '#5a6080' : '#9ca3af',
              }}>
                @{user.username}
              </p>
            )}
          </div>

          {/* Bio */}
          {user.bio ? (
            <p style={{
              margin: 0, fontSize: 13, color: isDark ? '#8892b0' : '#6b7280',
              lineHeight: 1.5, textAlign: 'center',
              overflow: 'hidden', textOverflow: 'ellipsis',
              display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical',
              minHeight: 20,
            }}>
              {user.bio}
            </p>
          ) : (
            <p style={{
              margin: 0, fontSize: 13, color: isDark ? '#5a6080' : '#d1d5db',
              textAlign: 'center', fontStyle: 'italic',
            }}>
              暂无简介
            </p>
          )}

          {/* Mini stats */}
          <div style={{
            display: 'flex', gap: 24, padding: '12px 0',
            borderTop: `1px solid ${isDark ? '#2a2a4a' : '#f3f4f6'}`,
            borderBottom: `1px solid ${isDark ? '#2a2a4a' : '#f3f4f6'}`,
            width: '100%', justifyContent: 'center',
          }}>
            <StatItem
              icon={<TeamOutlined />}
              value={user.follower_count || 0}
              label="粉丝"
              isDark={isDark}
            />
            <StatItem
              icon={<HeartOutlined />}
              value={user.following_count || 0}
              label="关注"
              isDark={isDark}
            />
          </div>

          {/* Follow button */}
          {!isSelf && (
            <motion.div
              whileTap={{ scale: 0.95 }}
              style={{ width: '100%' }}
              onClick={handleFollow}
            >
              <Button
                type={isFollowing ? 'default' : 'primary'}
                icon={isFollowing ? <CheckOutlined /> : <PlusOutlined />}
                loading={followLoading}
                block
                style={{
                  borderRadius: 10,
                  height: 38,
                  fontWeight: 600,
                  fontSize: 14,
                  border: isFollowing ? `1px solid ${isDark ? '#2a2a4a' : '#e5e7eb'}` : 'none',
                  background: isFollowing
                    ? (isDark ? 'rgba(255,255,255,0.06)' : '#f9fafb')
                    : 'linear-gradient(135deg, #4f6ef7, #6366f1)',
                  color: isFollowing ? (isDark ? '#e8e8f0' : '#374151') : '#fff',
                  boxShadow: isFollowing ? 'none' : '0 2px 12px rgba(79,110,247,0.3)',
                }}
              >
                {isFollowing ? '已关注' : '关注'}
              </Button>
            </motion.div>
          )}
        </motion.div>
      </Link>
    </motion.div>
  )
}

function StatItem({ icon, value, label, isDark }: {
  icon: React.ReactNode
  value: number
  label: string
  isDark: boolean
}) {
  return (
    <div style={{ textAlign: 'center' }}>
      <div style={{ fontSize: 13, color: isDark ? '#5a6080' : '#9ca3af', marginBottom: 2 }}>
        {icon}
      </div>
      <div style={{ fontSize: 16, fontWeight: 700, color: isDark ? '#e8e8f0' : '#1a1a2e', lineHeight: 1.2 }}>
        {value >= 1000 ? `${(value / 1000).toFixed(1)}k` : value}
      </div>
      <div style={{ fontSize: 11, color: isDark ? '#5a6080' : '#9ca3af' }}>{label}</div>
    </div>
  )
}

import { Link } from 'react-router-dom'
import { Avatar, Button, Typography, message } from 'antd'
import { UserOutlined, PlusOutlined, CheckOutlined } from '@ant-design/icons'
import { useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '../store/authStore'
import { useFollowStatus, useFollow, useUnfollow } from '../hooks/useFollow'

const { Text } = Typography

interface AuthorInfo {
  id: string
  name?: string
  username?: string
  avatar?: string
  bio?: string
  follower_count?: number
  following_count?: number
}

interface Props {
  author: AuthorInfo
  articleCount?: number
  articleSlug?: string
}

export default function AuthorCard({ author, articleCount = 0, articleSlug }: Props) {
  const queryClient = useQueryClient()
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const currentUser = useAuthStore((s) => s.user)
  const isSelf = currentUser?.id === author.id

  const { data: followData } = useFollowStatus(author.id)
  const followMutation = useFollow(author.id)
  const unfollowMutation = useUnfollow(author.id)
  const isFollowing = followData?.data || false

  const displayName = author.name || author.username || author.id
  const avatarSrc = author.avatar || undefined

  const invalidateArticle = () => {
    if (articleSlug) {
      queryClient.invalidateQueries({ queryKey: ['article', articleSlug] })
    }
  }

  const handleFollow = () => {
    if (!isAuthenticated) { message.warning('请先登录'); return }
    if (isFollowing) {
      unfollowMutation.mutate(undefined, {
        onSuccess: invalidateArticle,
        onError: (e: any) => message.error(e.message),
      })
    } else {
      followMutation.mutate(undefined, {
        onSuccess: invalidateArticle,
        onError: (e: any) => message.error(e.message),
      })
    }
  }

  return (
    <div style={{ textAlign: 'center', padding: '4px 0' }}>
      <Link to={`/user/${author.id}`}>
        <Avatar
          size={80}
          src={avatarSrc}
          icon={!avatarSrc ? <UserOutlined /> : undefined}
          style={{
            backgroundColor: avatarSrc ? undefined : '#1890ff',
            borderRadius: 16,
            marginBottom: 12,
          }}
        />
      </Link>

      <Link to={`/user/${author.id}`} style={{ textDecoration: 'none' }}>
        <div style={{ fontWeight: 700, fontSize: 16, color: '#1a1a1a', marginBottom: 4 }}>
          {displayName}
        </div>
      </Link>

      {author.bio && (
        <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 12, lineHeight: 1.5 }}>
          {author.bio}
        </Text>
      )}

      {!isSelf && (
        <Button
          type={isFollowing ? 'default' : 'primary'}
          icon={isFollowing ? <CheckOutlined /> : <PlusOutlined />}
          onClick={handleFollow}
          loading={followMutation.isPending || unfollowMutation.isPending}
          size="small"
          block
          style={{ borderRadius: 8, marginBottom: 16 }}
        >
          {isFollowing ? '已关注' : '关注'}
        </Button>
      )}

      <div style={{
        display: 'flex', justifyContent: 'center', gap: 20,
        padding: '12px 0', borderTop: '1px solid #f0f0f0', borderBottom: '1px solid #f0f0f0',
      }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontWeight: 700, fontSize: 15 }}>{articleCount}</div>
          <Text type="secondary" style={{ fontSize: 11 }}>文章</Text>
        </div>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontWeight: 700, fontSize: 15 }}>{author.follower_count || 0}</div>
          <Text type="secondary" style={{ fontSize: 11 }}>粉丝</Text>
        </div>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontWeight: 700, fontSize: 15 }}>{author.following_count || 0}</div>
          <Text type="secondary" style={{ fontSize: 11 }}>关注</Text>
        </div>
      </div>
    </div>
  )
}

import { Button, Space, Tooltip, message } from 'antd'
import { LikeOutlined, LikeFilled, StarOutlined, StarFilled, ShareAltOutlined, CommentOutlined } from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { useLikeStatus, useLike, useUnLike } from '../hooks/useLike'
import { useFavoriteStatus, useAddFavorite, useRemoveFavorite } from '../hooks/useFavorite'

interface Props {
  articleId: string
  onCommentClick?: () => void
  commentCount?: number
}

export default function PostActions({ articleId, onCommentClick, commentCount = 0 }: Props) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const navigate = useNavigate()
  const location = useLocation()

  const { data: likeData } = useLikeStatus(articleId)
  const likeMutation = useLike(articleId)
  const unlikeMutation = useUnLike(articleId)

  const { data: favData } = useFavoriteStatus(articleId)
  const addFavMutation = useAddFavorite(articleId)
  const removeFavMutation = useRemoveFavorite(articleId)

  const liked = likeData?.data?.liked || false
  const likeCount = likeData?.data?.count || 0
  const favorited = favData?.data || false

  const loginUrl = `/login?redirect=${encodeURIComponent(location.pathname + location.search)}`

  const handleLike = () => {
    if (!isAuthenticated) {
      message.warning('请登录后点赞')
      navigate(loginUrl)
      return
    }
    if (liked) {
      unlikeMutation.mutate(undefined, { onError: (e) => message.error(e.message) })
    } else {
      likeMutation.mutate(undefined, { onError: (e) => message.error(e.message) })
    }
  }

  const handleFavorite = () => {
    if (!isAuthenticated) {
      message.warning('请登录后收藏')
      navigate(loginUrl)
      return
    }
    if (favorited) {
      removeFavMutation.mutate(undefined, { onError: (e) => message.error(e.message) })
    } else {
      addFavMutation.mutate(undefined, { onError: (e) => message.error(e.message) })
    }
  }

  const handleShare = () => {
    const url = window.location.href
    if (navigator.share) {
      navigator.share({ url }).catch(() => {})
    } else {
      navigator.clipboard.writeText(url).then(
        () => message.success('链接已复制！'),
        () => message.error('复制失败')
      )
    }
  }

  const isLoading = likeMutation.isPending || unlikeMutation.isPending ||
    addFavMutation.isPending || removeFavMutation.isPending

  return (
    <Space size="middle" style={{ marginTop: 8, marginBottom: 8 }}>
      <Button
        type="text"
        icon={liked ? <LikeFilled style={{ color: '#1890ff' }} /> : <LikeOutlined />}
        onClick={handleLike}
        loading={likeMutation.isPending || unlikeMutation.isPending}
        disabled={isLoading && !liked}
      >
        {likeCount > 0 ? likeCount : '点赞'}
      </Button>

      <Button
        type="text"
        icon={favorited ? <StarFilled style={{ color: '#faad14' }} /> : <StarOutlined />}
        onClick={handleFavorite}
        loading={addFavMutation.isPending || removeFavMutation.isPending}
        disabled={isLoading && !favorited}
      >
        {favorited ? '已收藏' : '收藏'}
      </Button>

      {onCommentClick && (
        <Button type="text" icon={<CommentOutlined />} onClick={onCommentClick}>
          {commentCount > 0 ? commentCount : '评论'}
        </Button>
      )}

      <Tooltip title="分享">
        <Button type="text" icon={<ShareAltOutlined />} onClick={handleShare} />
      </Tooltip>
    </Space>
  )
}

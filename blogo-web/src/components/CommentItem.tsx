import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Avatar, Button, Tag, Typography, Input } from 'antd'
import {
  UserOutlined, ClockCircleOutlined, HeartOutlined,
  CommentOutlined, CaretDownOutlined, CaretUpOutlined,
  SendOutlined, CloseOutlined,
} from '@ant-design/icons'
import { motion, AnimatePresence } from 'framer-motion'
import { useAuthStore } from '../store/authStore'
import { useAppStore } from '../store/appStore'
import { useCreateComment } from '../hooks/useComments'
import dayjs from '../utils/dayjs'
import type { Comment, CommentForm } from '../types'

const { Text } = Typography

export type TreeNode = Comment & { children: TreeNode[] }

const DEFAULT_VISIBLE = 3
const EXPAND_STEP = 10

// ──────────────────────────────────────────
// Props
// ──────────────────────────────────────────
interface MainCommentProps {
  node: TreeNode
  articleId: string
  articleAuthorId?: string
  activeReplyId: string | null
  onReplyTarget: (id: string | null, name: string) => void
}

// ──────────────────────────────────────────
// MainComment component
// ──────────────────────────────────────────
export default function MainComment({ node, articleId, articleAuthorId, activeReplyId, onReplyTarget }: MainCommentProps) {
  const navigate = useNavigate()
  const theme = useAppStore((s) => s.theme)
  const isDark = theme === 'dark'

  const isAuthor = !!(articleAuthorId && node.user_id === articleAuthorId)
  const displayName = node.user?.name || node.user?.username || node.username || '匿名'
  const avatarSrc = node.user?.avatar || undefined
  const hasUser = !!node.user_id
  const isReplying = activeReplyId === node.id

  // Collapse state for replies
  const [visibleCount, setVisibleCount] = useState(DEFAULT_VISIBLE)
  const childCount = node.children.length
  const hasHidden = childCount > visibleCount
  const visibleChildren = node.children.slice(0, visibleCount)
  const hiddenCount = childCount - visibleCount
  const isFullyExpanded = !hasHidden && childCount > DEFAULT_VISIBLE

  const colors = {
    text: isDark ? '#e8e8f0' : '#1a1a2e',
    textSecondary: isDark ? '#8892b0' : '#6b7280',
    textMuted: isDark ? '#5a6080' : '#9ca3af',
    border: isDark ? '#2a2a4a' : '#e5e7eb',
    bgHover: isDark ? 'rgba(255,255,255,0.03)' : 'rgba(0,0,0,0.012)',
    accent: '#4f6ef7',
    replyBg: isDark ? 'rgba(255,255,255,0.015)' : '#f8f9fb',
    replyHover: isDark ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.02)',
  }

  const handleUserClick = () => {
    if (!hasUser) return
    navigate(`/user/${node.user_id}`)
  }

  const handleReplyClick = () => {
    if (isReplying) {
      onReplyTarget(null, '')
    } else {
      onReplyTarget(node.id, displayName)
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35, ease: 'easeOut' }}
      style={{ marginBottom: 8 }}
    >
      {/* ======== MAIN COMMENT CARD ======== */}
      <div
        style={{
          display: 'flex', gap: 14, padding: '16px 20px',
          borderRadius: 14,
          transition: 'background 0.2s ease',
          marginBottom: isReplying ? 0 : (childCount > 0 ? 8 : 16),
        }}
        onMouseEnter={(e) => { e.currentTarget.style.background = colors.bgHover }}
        onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
      >
        {/* Avatar */}
        <Avatar
          size={44}
          src={avatarSrc || undefined}
          icon={!avatarSrc ? <UserOutlined /> : undefined}
          onClick={handleUserClick}
          style={{
            flexShrink: 0,
            backgroundColor: avatarSrc ? undefined : '#4f6ef7',
            cursor: hasUser ? 'pointer' : 'default',
            borderRadius: 14,
            border: isAuthor ? '2.5px solid #f59e0b' : undefined,
            boxShadow: isAuthor ? '0 0 0 3px rgba(245,158,11,0.15)' : undefined,
          }}
        />

        <div style={{ flex: 1, minWidth: 0 }}>
          {/* Meta */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap', marginBottom: 4 }}>
            <Text
              strong
              onClick={handleUserClick}
              style={{
                fontSize: 14, cursor: hasUser ? 'pointer' : 'default',
                color: hasUser ? colors.accent : colors.text,
              }}
              onMouseEnter={(e) => { if (hasUser) e.currentTarget.style.opacity = '0.8' }}
              onMouseLeave={(e) => { if (hasUser) e.currentTarget.style.opacity = '1' }}
            >
              {displayName}
            </Text>
            {isAuthor && (
              <Tag color="orange" style={{ fontSize: 10, lineHeight: '18px', padding: '0 8px', borderRadius: 10, margin: 0, fontWeight: 600, border: 'none' }}>作者</Tag>
            )}
            {node.is_top && (
              <Tag color="red" style={{ fontSize: 10, lineHeight: '18px', padding: '0 8px', borderRadius: 10, margin: 0, fontWeight: 600, border: 'none' }}>置顶</Tag>
            )}
          </div>

          {/* Content */}
          <Text style={{ fontSize: 15, lineHeight: 1.7, color: colors.text, wordBreak: 'break-word', whiteSpace: 'pre-wrap', display: 'block' }}>
            {node.content}
          </Text>

          {/* Actions */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 18, marginTop: 10 }}>
            <Text style={{ fontSize: 12, color: colors.textMuted, display: 'flex', alignItems: 'center', gap: 4 }}>
              <ClockCircleOutlined style={{ fontSize: 11 }} />
              {dayjs(node.created_at).fromNow()}
            </Text>
            <ActionButton icon={<HeartOutlined />} label="0" hoverColor="#ef4444" isDark={isDark} />
            <ActionButton icon={<CommentOutlined />} label="回复" hoverColor={colors.accent} isDark={isDark} active={isReplying} onClick={handleReplyClick} />
          </div>
        </div>
      </div>

      {/* ======== INLINE REPLY FORM ======== */}
      <AnimatePresence>
        {isReplying && (
          <InlineReplyForm
            articleId={articleId}
            parentId={node.id}
            placeholder={`回复 @${displayName}...`}
            isDark={isDark}
            colors={colors}
            onSuccess={() => onReplyTarget(null, '')}
            onCancel={() => onReplyTarget(null, '')}
          />
        )}
      </AnimatePresence>

      {/* ======== FLATTENED REPLIES (SECOND LEVEL) ======== */}
      {childCount > 0 && (
        <div style={{ marginLeft: 56, marginBottom: 12 }}>
          <div style={{
            background: colors.replyBg,
            borderRadius: 14,
            padding: '4px 0',
          }}>
            <AnimatePresence>
              {visibleChildren.map((child) => (
                <ReplyCard
                  key={child.id}
                  node={child}
                  articleAuthorId={articleAuthorId}
                  isDark={isDark}
                  colors={colors}
                />
              ))}
            </AnimatePresence>

            {/* Expand / collapse */}
            {hasHidden && (
              <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.1 }} style={{ padding: '10px 16px 14px' }}>
                <Button
                  type="text" size="small"
                  icon={<CaretDownOutlined />}
                  onClick={() => setVisibleCount(Math.min(visibleCount + EXPAND_STEP, childCount))}
                  style={{ fontSize: 13, fontWeight: 500, color: colors.accent, padding: '4px 10px', borderRadius: 8 }}
                >
                  查看全部 {childCount} 条回复
                </Button>
                {hiddenCount > EXPAND_STEP && (
                  <Button
                    type="text" size="small"
                    onClick={() => setVisibleCount(childCount)}
                    style={{ fontSize: 13, fontWeight: 500, color: colors.textMuted, padding: '4px 10px', borderRadius: 8, marginLeft: 4 }}
                  >
                    展开剩余 {hiddenCount} 条
                  </Button>
                )}
              </motion.div>
            )}

            {isFullyExpanded && (
              <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} style={{ padding: '4px 16px 14px' }}>
                <Button
                  type="text" size="small"
                  icon={<CaretUpOutlined />}
                  onClick={() => setVisibleCount(DEFAULT_VISIBLE)}
                  style={{ fontSize: 13, fontWeight: 500, color: colors.textMuted, padding: '4px 10px', borderRadius: 8 }}
                >
                  收起回复
                </Button>
              </motion.div>
            )}
          </div>
        </div>
      )}
    </motion.div>
  )
}

// ──────────────────────────────────────────
// ReplyCard — second-level reply (no further nesting)
// ──────────────────────────────────────────
function ReplyCard({ node, articleAuthorId, isDark, colors }: {
  node: TreeNode
  articleAuthorId?: string
  isDark: boolean
  colors: Record<string, string>
}) {
  const navigate = useNavigate()
  const isAuthor = !!(articleAuthorId && node.user_id === articleAuthorId)
  const displayName = node.user?.name || node.user?.username || node.username || '匿名'
  const avatarSrc = node.user?.avatar || undefined
  const hasUser = !!node.user_id

  const handleUserClick = () => {
    if (!hasUser) return
    navigate(`/user/${node.user_id}`)
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.25, ease: 'easeOut' }}
      style={{ padding: '10px 16px' }}
      onMouseEnter={(e) => { e.currentTarget.style.background = colors.replyHover }}
      onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
    >
      <div style={{ display: 'flex', gap: 12 }}>
        <Avatar
          size={34}
          src={avatarSrc || undefined}
          icon={!avatarSrc ? <UserOutlined /> : undefined}
          onClick={handleUserClick}
          style={{
            flexShrink: 0,
            backgroundColor: avatarSrc ? undefined : '#4f6ef7',
            cursor: hasUser ? 'pointer' : 'default',
            borderRadius: 10,
            border: isAuthor ? '2px solid #f59e0b' : undefined,
          }}
        />
        <div style={{ flex: 1, minWidth: 0 }}>
          {/* Meta */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap', marginBottom: 3 }}>
            <Text
              strong
              onClick={handleUserClick}
              style={{
                fontSize: 13, cursor: hasUser ? 'pointer' : 'default',
                color: hasUser ? colors.accent : colors.text,
              }}
            >
              {displayName}
            </Text>
            {isAuthor && (
              <Tag color="orange" style={{ fontSize: 10, lineHeight: '16px', padding: '0 6px', borderRadius: 8, margin: 0, fontWeight: 600, border: 'none' }}>作者</Tag>
            )}
          </div>

          {/* Reply reference */}
          {node.parent && (
            <Text type="secondary" style={{ fontSize: 11, marginBottom: 2, display: 'block' }}>
              回复 <Text style={{ color: colors.accent, fontSize: 11 }}>@{node.parent.username || '匿名'}</Text>
            </Text>
          )}

          {/* Content */}
          <Text style={{ fontSize: 14, lineHeight: 1.65, color: colors.text, wordBreak: 'break-word', whiteSpace: 'pre-wrap', display: 'block' }}>
            {node.content}
          </Text>

          {/* Actions */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginTop: 6 }}>
            <Text style={{ fontSize: 11, color: colors.textMuted, display: 'flex', alignItems: 'center', gap: 3 }}>
              <ClockCircleOutlined style={{ fontSize: 10 }} />
              {dayjs(node.created_at).fromNow()}
            </Text>
            <ActionButton icon={<HeartOutlined />} label="0" hoverColor="#ef4444" isDark={isDark} size="small" />
          </div>
        </div>
      </div>
    </motion.div>
  )
}

// ──────────────────────────────────────────
// ActionButton helper
// ──────────────────────────────────────────
function ActionButton({ icon, label, hoverColor, isDark, active, onClick, size }: {
  icon: React.ReactNode
  label: string
  hoverColor: string
  isDark: boolean
  active?: boolean
  onClick?: () => void
  size?: 'small'
}) {
  const s = size === 'small'
  return (
    <Button
      type="text"
      size={s ? 'small' : undefined}
      icon={icon as any}
      onClick={onClick}
      style={{
        fontSize: s ? 11 : 12, color: active ? hoverColor : (isDark ? '#5a6080' : '#9ca3af'),
        padding: s ? '0 2px' : '0 4px', height: s ? 18 : 22,
        borderRadius: 6, fontWeight: active ? 600 : 400,
      }}
      onMouseEnter={(e) => { e.currentTarget.style.color = hoverColor }}
      onMouseLeave={(e) => { e.currentTarget.style.color = active ? hoverColor : (isDark ? '#5a6080' : '#9ca3af') }}
    >
      {label}
    </Button>
  )
}

// ──────────────────────────────────────────
// InlineReplyForm — appears below the comment being replied to
// ──────────────────────────────────────────
function InlineReplyForm({ articleId, parentId, placeholder, isDark, colors, onSuccess, onCancel }: {
  articleId: string
  parentId: string
  placeholder: string
  isDark: boolean
  colors: Record<string, string>
  onSuccess: () => void
  onCancel: () => void
}) {
  const [val, setVal] = useState('')
  const createMutation = useCreateComment()
  const { user, isAuthenticated } = useAuthStore()
  const inputBg = isDark ? 'rgba(255,255,255,0.05)' : '#f3f4f6'

  const handleSend = () => {
    const content = val.trim()
    if (!content) return
    const payload: CommentForm = {
      article_id: articleId,
      content,
      parent_id: parentId,
      username: (isAuthenticated && user) ? (user.name || user.username) : undefined,
    }
    createMutation.mutate(payload, {
      onSuccess: () => { setVal(''); onSuccess() },
      onError: (err: any) => { /* handled silently */ },
    })
  }

  return (
    <motion.div
      initial={{ opacity: 0, height: 0 }}
      animate={{ opacity: 1, height: 'auto' }}
      exit={{ opacity: 0, height: 0 }}
      style={{ overflow: 'hidden', marginLeft: 56, marginBottom: 8 }}
    >
      <div style={{
        background: colors.replyBg,
        borderRadius: 14,
        padding: '12px 16px',
        border: `1px solid ${isDark ? '#2a2a4a' : '#e5e7eb'}`,
      }}>
        <Input.TextArea
          value={val}
          onChange={(e) => setVal(e.target.value)}
          placeholder={placeholder}
          autoSize={{ minRows: 2, maxRows: 4 }}
          maxLength={2000}
          style={{
            borderRadius: 10, fontSize: 14, color: colors.text,
            background: inputBg, borderColor: isDark ? '#2a2a4a' : '#e5e7eb',
            resize: 'none', marginBottom: 10,
          }}
        />
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
          <Button size="small" icon={<CloseOutlined />} onClick={onCancel} style={{ borderRadius: 8, color: colors.textMuted, fontSize: 12 }}>
            取消
          </Button>
          <Button
            type="primary" size="small"
            icon={<SendOutlined />}
            onClick={handleSend}
            loading={createMutation.isPending}
            disabled={!val.trim()}
            style={{
              borderRadius: 8, fontWeight: 500, fontSize: 13,
              background: val.trim() ? 'linear-gradient(135deg, #4f6ef7, #6366f1)' : undefined,
              border: 'none', boxShadow: val.trim() ? '0 2px 10px rgba(79,110,247,0.3)' : 'none',
            }}
          >
            回复
          </Button>
        </div>
      </div>
    </motion.div>
  )
}

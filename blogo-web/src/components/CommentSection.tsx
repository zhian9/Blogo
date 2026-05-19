import { useState, useRef } from 'react'
import { Input, Avatar, Button, message, Spin, Empty, Typography } from 'antd'
import { SendOutlined, UserOutlined, MessageOutlined } from '@ant-design/icons'
import { motion, AnimatePresence } from 'framer-motion'
import { useComments, useCreateComment } from '../hooks/useComments'
import { useAuthStore } from '../store/authStore'
import { useAppStore } from '../store/appStore'
import MainComment, { type TreeNode } from './CommentItem'
import type { Comment, CommentForm } from '../types'

const { Text } = Typography

// ──────────────────────────────────────────
// Build tree from flat comments, then flatten beyond level 2
// ──────────────────────────────────────────
function buildFlat2LevelTree(comments: Comment[]): TreeNode[] {
  const map = new Map<string, TreeNode>()
  comments.forEach((c) => map.set(c.id, { ...c, children: [] }))

  const roots: TreeNode[] = []

  comments.forEach((c) => {
    const node = map.get(c.id)!
    if (c.parent_id && map.has(c.parent_id)) {
      map.get(c.parent_id)!.children.push(node)
    } else {
      roots.push(node)
    }
  })

  // Flatten: for each root, collect ALL descendants and put them directly under root
  function collectAllDescendants(nodes: TreeNode[]): TreeNode[] {
    const result: TreeNode[] = []
    for (const n of nodes) {
      result.push(n)
      if (n.children.length > 0) {
        result.push(...collectAllDescendants(n.children))
      }
    }
    return result
  }

  for (const root of roots) {
    root.children = collectAllDescendants(root.children)
  }

  return roots
}

// ──────────────────────────────────────────
// Props
// ──────────────────────────────────────────
interface Props {
  articleId: string
  visible: boolean
  articleAuthorId?: string
}

// ──────────────────────────────────────────
// Main component
// ──────────────────────────────────────────
export default function CommentSection({ articleId, visible, articleAuthorId }: Props) {
  const [activeReplyId, setActiveReplyId] = useState<string | null>(null)
  const [inputVal, setInputVal] = useState('')
  const inputRef = useRef<any>(null)

  const { data, isLoading } = useComments(articleId)
  const createMutation = useCreateComment()
  const { user, isAuthenticated } = useAuthStore()
  const theme = useAppStore((s) => s.theme)
  const isDark = theme === 'dark'

  const comments: Comment[] = data?.data || []
  const tree = buildFlat2LevelTree(comments)

  const handleReplyTarget = (id: string | null, _name: string) => {
    // Toggle: if clicking the same one, close it; otherwise open the new one
    setActiveReplyId((prev) => (prev === id ? null : id))
  }

  const handleSendTopLevel = () => {
    const content = inputVal.trim()
    if (!content) return

    const payload: CommentForm = {
      article_id: articleId,
      content,
      parent_id: undefined,
      username: (isAuthenticated && user) ? (user.name || user.username) : undefined,
    }

    createMutation.mutate(payload, {
      onSuccess: () => {
        message.success('评论成功')
        setInputVal('')
      },
      onError: (err: any) => {
        message.error(err.message || '评论失败')
      },
    })
  }

  const colors = {
    surface: isDark ? '#141418' : '#ffffff',
    text: isDark ? '#e8e8f0' : '#1a1a2e',
    textSecondary: isDark ? '#8892b0' : '#6b7280',
    textMuted: isDark ? '#5a6080' : '#9ca3af',
    border: isDark ? '#2a2a4a' : '#e5e7eb',
    accent: '#4f6ef7',
    inputBg: isDark ? 'rgba(255,255,255,0.05)' : '#f3f4f6',
    replyBar: isDark ? 'rgba(79,110,247,0.15)' : '#e0e7ff',
  }

  if (!visible) return null

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
      style={{
        display: 'flex', flexDirection: 'column',
        background: colors.surface,
        borderRadius: 20,
        border: `1px solid ${colors.border}`,
        overflow: 'hidden',
        boxShadow: isDark
          ? '0 4px 32px rgba(0,0,0,0.3)'
          : '0 4px 32px rgba(0,0,0,0.05)',
      }}
    >
      {/* ========== HEADER ========== */}
      <div style={{
        display: 'flex', alignItems: 'center', gap: 10,
        padding: '20px 24px 16px',
        borderBottom: `1px solid ${colors.border}`,
        flexShrink: 0,
      }}>
        <MessageOutlined style={{ fontSize: 18, color: colors.accent }} />
        <Text strong style={{ fontSize: 16, color: colors.text }}>评论</Text>
        {comments.length > 0 && (
          <span style={{
            background: `${colors.accent}15`, color: colors.accent,
            fontSize: 13, fontWeight: 600, padding: '2px 10px', borderRadius: 12,
          }}>
            {comments.length}
          </span>
        )}
      </div>

      {/* ========== COMMENT LIST ========== */}
      <div style={{
        flex: 1, overflowY: 'auto', padding: '16px 24px 8px',
        minHeight: 160, maxHeight: '60vh',
      }}>
        {isLoading ? (
          <div style={{
            display: 'flex', flexDirection: 'column',
            alignItems: 'center', justifyContent: 'center', padding: 60, gap: 12,
          }}>
            <Spin size="large" />
            <Text style={{ color: colors.textMuted, fontSize: 13 }}>加载评论中...</Text>
          </div>
        ) : tree.length === 0 ? (
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.4 }}
            style={{ padding: 50 }}
          >
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={
                <Text style={{ color: colors.textSecondary, fontSize: 14 }}>
                  暂无评论，来发表第一条精彩评论吧！
                </Text>
              }
            />
          </motion.div>
        ) : (
          tree.map((node) => (
            <MainComment
              key={node.id}
              node={node}
              articleId={articleId}
              articleAuthorId={articleAuthorId}
              activeReplyId={activeReplyId}
              onReplyTarget={handleReplyTarget}
            />
          ))
        )}
      </div>

      {/* ========== BOTTOM INPUT BAR ========== */}
      <div style={{
        flexShrink: 0,
        padding: '12px 24px 20px',
        borderTop: `1px solid ${colors.border}`,
        background: isDark ? 'rgba(15,15,15,0.85)' : 'rgba(255,255,255,0.85)',
        backdropFilter: 'blur(12px)',
      }}>
        {/* Active reply indicator */}
        <AnimatePresence>
          {activeReplyId && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              style={{ overflow: 'hidden', marginBottom: 10 }}
            >
              <div style={{
                display: 'flex', alignItems: 'center', gap: 8,
                padding: '6px 12px',
                background: colors.replyBar,
                borderRadius: 8,
                borderLeft: `3px solid ${colors.accent}`,
              }}>
                <Text style={{ fontSize: 13, color: colors.accent, fontWeight: 500 }}>
                  正在回复中...
                </Text>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Input */}
        <div
          style={{
            display: 'flex', alignItems: 'center', gap: 12,
            background: colors.inputBg,
            borderRadius: 14,
            padding: '6px 6px 6px 18px',
            border: `1px solid ${colors.border}`,
            transition: 'border-color 0.25s, box-shadow 0.25s',
          }}
          className="comment-input-wrapper"
        >
          <Avatar
            size={34}
            src={isAuthenticated && user?.avatar ? user.avatar : undefined}
            icon={!(isAuthenticated && user?.avatar) ? <UserOutlined /> : undefined}
            style={{
              flexShrink: 0,
              backgroundColor: (isAuthenticated && user?.avatar) ? undefined : '#4f6ef7',
              borderRadius: 10,
            }}
          />
          <Input
            ref={inputRef}
            value={inputVal}
            onChange={(e) => setInputVal(e.target.value)}
            onPressEnter={handleSendTopLevel}
            placeholder={isAuthenticated ? '写下你的想法...' : '登录后即可评论'}
            variant="borderless"
            maxLength={2000}
            style={{ flex: 1, fontSize: 14, color: colors.text, background: 'transparent' }}
          />
          <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
            <Button
              type="primary"
              shape="circle"
              icon={<SendOutlined />}
              onClick={handleSendTopLevel}
              loading={createMutation.isPending}
              disabled={!inputVal.trim()}
              style={{
                flexShrink: 0, width: 38, height: 38,
                background: inputVal.trim()
                  ? 'linear-gradient(135deg, #4f6ef7, #6366f1)'
                  : undefined,
                border: 'none',
                boxShadow: inputVal.trim() ? '0 2px 12px rgba(79,110,247,0.35)' : 'none',
              }}
            />
          </motion.div>
        </div>

        {inputVal.length > 1800 && (
          <div style={{ textAlign: 'right', marginTop: 4 }}>
            <Text style={{
              fontSize: 11,
              color: inputVal.length >= 2000 ? '#ef4444' : colors.textMuted,
            }}>
              {inputVal.length}/2000
            </Text>
          </div>
        )}
      </div>

      <style>{`
        .comment-input-wrapper:focus-within {
          border-color: #4f6ef7 !important;
          box-shadow: 0 0 0 3px rgba(79,110,247,0.12) !important;
        }
      `}</style>
    </motion.div>
  )
}

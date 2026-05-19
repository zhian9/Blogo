import { Link } from 'react-router-dom'
import { Tag, Space, Divider } from 'antd'
import {
  CalendarOutlined, EyeOutlined, ClockCircleOutlined,
  UserOutlined, LockOutlined, GlobalOutlined, TeamOutlined,
} from '@ant-design/icons'
import { motion } from 'framer-motion'
import type { Article } from '../types'
import dayjs from '../utils/dayjs'

interface Props {
  article: Article
  likeCount?: number
}

function VisibilityBadge({ visibility }: { visibility: string }) {
  switch (visibility) {
    case 'private':
      return (
        <span style={{
          display: 'inline-flex', alignItems: 'center', gap: 4,
          fontSize: 11, fontWeight: 600, padding: '2px 10px', borderRadius: 6,
          background: 'rgba(239,68,68,0.12)', border: '1px solid rgba(239,68,68,0.3)',
          color: '#f87171', backdropFilter: 'blur(4px)',
        }}>
          <LockOutlined style={{ fontSize: 10 }} /> 私密
        </span>
      )
    case 'partial_visible':
      return (
        <span style={{
          display: 'inline-flex', alignItems: 'center', gap: 4,
          fontSize: 11, fontWeight: 600, padding: '2px 10px', borderRadius: 6,
          background: 'rgba(168,85,247,0.12)', border: '1px solid rgba(168,85,247,0.3)',
          color: '#c084fc', backdropFilter: 'blur(4px)',
        }}>
          <TeamOutlined style={{ fontSize: 10 }} /> 部分可见
        </span>
      )
    case 'public':
    default:
      return (
        <span style={{
          display: 'inline-flex', alignItems: 'center', gap: 4,
          fontSize: 11, fontWeight: 600, padding: '2px 10px', borderRadius: 6,
          background: 'rgba(79,110,247,0.12)', border: '1px solid rgba(79,110,247,0.3)',
          color: '#818cf8', backdropFilter: 'blur(4px)',
        }}>
          <GlobalOutlined style={{ fontSize: 10 }} /> 公开
        </span>
      )
  }
}

export default function ArticleHeader({ article }: Props) {
  return (
    <motion.header
      initial={{ opacity: 0, filter: 'blur(8px)', y: 16 }}
      animate={{ opacity: 1, filter: 'blur(0px)', y: 0 }}
      transition={{ duration: 0.7, ease: [0.25, 0.46, 0.45, 0.94] }}
      style={{ marginBottom: 32 }}
    >
      {/* ── Tags row ── */}
      <Space size={6} style={{ marginBottom: 16 }}>
        {article.is_top && <Tag color="red" style={{ borderRadius: 6, fontSize: 11, fontWeight: 600 }}>置顶</Tag>}
        <VisibilityBadge visibility={article.visibility} />
        {article.category && (
          <Link to={`/categories?category_id=${article.category_id}`}>
            <Tag style={{
              borderRadius: 6, fontSize: 11, fontWeight: 500,
              background: 'rgba(79,110,247,0.15)',
              border: '1px solid rgba(79,110,247,0.3)',
              color: '#4f6ef7',
            }}>
              {article.category.name}
            </Tag>
          </Link>
        )}
      </Space>

      {/* ── Cinematic title ── */}
      <h1 style={{
        fontFamily: "'Instrument Serif', serif",
        fontStyle: 'italic',
        fontSize: 'clamp(28px, 5vw, 44px)',
        fontWeight: 400,
        color: '#ffffff',
        letterSpacing: '-0.03em',
        lineHeight: 1.2,
        margin: '0 0 20px',
      }}>
        {article.title}
      </h1>

      {/* ── Meta row ── */}
      <div style={{
        display: 'flex', alignItems: 'center', flexWrap: 'wrap',
        gap: '6px 20px', marginBottom: 0,
        fontFamily: "'Barlow', sans-serif",
      }}>
        <Link to={`/user/${article.author_id}`} style={{ textDecoration: 'none' }}>
          <span style={{
            fontSize: 13, color: '#4f6ef7', fontWeight: 500,
            display: 'flex', alignItems: 'center', gap: 4,
          }}>
            <UserOutlined style={{ fontSize: 12 }} />
            {article.author?.name || article.author?.username || article.author_id}
          </span>
        </Link>
        <span style={{ fontSize: 13, color: 'rgba(255,255,255,0.4)', display: 'flex', alignItems: 'center', gap: 4 }}>
          <CalendarOutlined style={{ fontSize: 12 }} />
          {dayjs(article.published_at).format('YYYY-MM-DD')}
        </span>
        <span style={{ fontSize: 13, color: 'rgba(255,255,255,0.4)', display: 'flex', alignItems: 'center', gap: 4 }}>
          <ClockCircleOutlined style={{ fontSize: 12 }} />
          {dayjs(article.published_at).fromNow()}
        </span>
        <span style={{ fontSize: 13, color: 'rgba(255,255,255,0.4)', display: 'flex', alignItems: 'center', gap: 4 }}>
          <EyeOutlined style={{ fontSize: 12 }} />
          {article.views || 0} 阅读
        </span>
      </div>

      {/* ── Tag pills ── */}
      {article.tags && article.tags.length > 0 && (
        <>
          <Divider style={{ margin: '18px 0', borderColor: 'rgba(255,255,255,0.06)' }} />
          <Space size={8}>
            {article.tags.map((tag) => (
              <Link key={tag.id} to={`/tags?tag_name=${encodeURIComponent(tag.name)}`}>
                <span style={{
                  display: 'inline-block',
                  fontFamily: "'Barlow', sans-serif",
                  fontSize: 12, fontWeight: 500,
                  color: 'rgba(255,255,255,0.55)',
                  background: 'rgba(255,255,255,0.05)',
                  padding: '4px 12px',
                  borderRadius: 8,
                  border: '1px solid rgba(255,255,255,0.08)',
                  transition: 'all 0.25s ease',
                }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.background = 'rgba(255,255,255,0.1)'
                    e.currentTarget.style.borderColor = 'rgba(255,255,255,0.2)'
                    e.currentTarget.style.color = '#ffffff'
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = 'rgba(255,255,255,0.05)'
                    e.currentTarget.style.borderColor = 'rgba(255,255,255,0.08)'
                    e.currentTarget.style.color = 'rgba(255,255,255,0.55)'
                  }}
                >
                  {tag.name}
                </span>
              </Link>
            ))}
          </Space>
        </>
      )}

      <Divider style={{ margin: '24px 0 0', borderColor: 'rgba(255,255,255,0.08)' }} />
    </motion.header>
  )
}

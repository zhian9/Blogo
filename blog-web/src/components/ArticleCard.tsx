import { useNavigate } from 'react-router-dom'
import { Tag, Typography, Button } from 'antd'
import {
  EyeOutlined, ClockCircleOutlined, EditOutlined, FileTextOutlined,
  GlobalOutlined, LockOutlined, TeamOutlined,
} from '@ant-design/icons'
import { motion } from 'framer-motion'
import { useAuthStore } from '../store/authStore'
import dayjs from '../utils/dayjs'
import type { Article } from '../types'

const { Text } = Typography

interface Props {
  article: Article
}

export default function ArticleCard({ article }: Props) {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const isAuthor = !!(user && article.author_id && user.id === article.author_id)
  const hasCover = !!(article.cover_image?.url)

  const handleEdit = (e: React.MouseEvent) => {
    e.stopPropagation()
    navigate(`/article/${article.slug}/edit`)
  }

  return (
    <motion.div
      initial={{ opacity: 0, filter: 'blur(6px)', y: 16 }}
      whileInView={{ opacity: 1, filter: 'blur(0px)', y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.5, ease: [0.25, 0.46, 0.45, 0.94] }}
      whileHover={{ y: -4 }}
      onClick={() => navigate(`/article/${article.slug}`)}
      className="liquid-glass-card article-card-h"
      style={{
        display: 'flex', flexDirection: 'row',
        cursor: 'pointer', overflow: 'hidden',
        marginBottom: 16, minHeight: 160, position: 'relative',
      }}
    >
      {/* ── Left cover ── */}
      <div style={{
        width: 200, minWidth: 200, position: 'relative', overflow: 'hidden',
        background: hasCover
          ? 'transparent'
          : 'linear-gradient(135deg, #4f6ef7 0%, #8b5cf6 100%)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
      }}>
        {hasCover && (
          <img
            src={article.cover_image!.url}
            alt={article.title}
            style={{
              position: 'absolute', top: 0, left: 0,
              width: '100%', height: '100%', objectFit: 'cover',
              transition: 'transform 0.5s cubic-bezier(0.25, 0.8, 0.25, 1.2), filter 0.5s ease',
            }}
            className="ac-cover-img"
          />
        )}
        {!hasCover && article.category && (
          <Text style={{ color: 'rgba(255,255,255,0.85)', fontSize: 18, fontWeight: 700, fontFamily: "'Instrument Serif', serif", fontStyle: 'italic' }}>
            #{article.category.name}
          </Text>
        )}
        {!hasCover && !article.category && (
          <FileTextOutlined style={{ fontSize: 48, color: 'rgba(255,255,255,0.4)' }} />
        )}
      </div>

      {/* ── Right content ── */}
      <div style={{ flex: 1, padding: '22px 24px', display: 'flex', flexDirection: 'column', justifyContent: 'space-between', minWidth: 0 }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            {article.is_top && (
              <Tag color="red" style={{ borderRadius: 4, margin: 0, flexShrink: 0, fontSize: 11 }}>置顶</Tag>
            )}
            {article.visibility === 'private' && (
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 3, fontSize: 10, fontWeight: 600, padding: '1px 8px', borderRadius: 4, background: 'rgba(239,68,68,0.12)', border: '1px solid rgba(239,68,68,0.25)', color: '#f87171', flexShrink: 0 }}><LockOutlined style={{ fontSize: 9 }} /> 私密</span>
            )}
            {article.visibility === 'partial_visible' && (
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 3, fontSize: 10, fontWeight: 600, padding: '1px 8px', borderRadius: 4, background: 'rgba(168,85,247,0.12)', border: '1px solid rgba(168,85,247,0.25)', color: '#c084fc', flexShrink: 0 }}><TeamOutlined style={{ fontSize: 9 }} /> 部分可见</span>
            )}
            <Text
              strong
              style={{
                fontFamily: "'Instrument Serif', serif",
                fontStyle: 'italic',
                fontSize: 20, fontWeight: 400, color: '#ffffff',
                letterSpacing: '-0.02em', lineHeight: 1.3,
                overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
              }}
            >
              {article.title}
            </Text>
          </div>

          <Text style={{
            fontSize: 14, color: 'rgba(255,255,255,0.45)', lineHeight: 1.7,
            fontFamily: "'Barlow', sans-serif",
            overflow: 'hidden', display: '-webkit-box',
            WebkitLineClamp: 2, WebkitBoxOrient: 'vertical',
            display: 'block',
          }}>
            {article.summary || article.content?.replace(/[#*`>\[\]!~-]/g, '').slice(0, 200)}
          </Text>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap', marginTop: 10 }}>
          {article.category && (
            <span style={{
              fontFamily: "'Barlow', sans-serif",
              fontSize: 11, fontWeight: 500, color: '#4f6ef7',
              background: 'rgba(79,110,247,0.12)',
              padding: '3px 10px', borderRadius: 6,
              letterSpacing: '0.04em',
              border: '1px solid rgba(79,110,247,0.2)',
            }}>
              {article.category.name}
            </span>
          )}
          <Text style={{ fontSize: 12, color: 'rgba(255,255,255,0.35)', fontFamily: "'Barlow', sans-serif", display: 'flex', alignItems: 'center', gap: 4 }}>
            <ClockCircleOutlined /> {dayjs(article.published_at).format('YYYY-MM-DD')}
          </Text>
          <Text style={{ fontSize: 12, color: 'rgba(255,255,255,0.35)', fontFamily: "'Barlow', sans-serif", display: 'flex', alignItems: 'center', gap: 4 }}>
            <EyeOutlined /> {article.views || 0} 阅读
          </Text>
        </div>
      </div>

      {/* Edit button */}
      {isAuthor && (
        <Button
          size="small" type="primary"
          icon={<EditOutlined />} onClick={handleEdit}
          className="ac-edit-btn"
          style={{
            position: 'absolute', bottom: 16, right: 16,
            opacity: 0, transition: 'opacity 0.3s ease',
            borderRadius: 8, zIndex: 2,
            background: '#4f6ef7', border: 'none',
          }}
        />
      )}

      <style>{`
        .article-card-h:hover .ac-cover-img {
          transform: scale(1.06);
          filter: brightness(1.15);
        }
        .article-card-h:hover .ac-edit-btn {
          opacity: 1 !important;
        }
        @media (max-width: 640px) {
          .article-card-h {
            flex-direction: column !important;
          }
          .article-card-h > div:first-child {
            width: 100% !important;
            min-width: 100% !important;
            height: 180px !important;
          }
        }
      `}</style>
    </motion.div>
  )
}

import { useNavigate } from 'react-router-dom'
import { Tag, Space, Typography } from 'antd'
import { EyeOutlined, ClockCircleOutlined, FileTextOutlined, LockOutlined, TeamOutlined } from '@ant-design/icons'
import { motion } from 'framer-motion'
import dayjs from '../utils/dayjs'
import type { Article } from '../types'

const { Text } = Typography

const estimateReadTime = (text: string) => {
  const minutes = Math.ceil(text.length / 250)
  return Math.max(1, minutes)
}

interface Props {
  article: Article
  index: number
}

export default function ArticleGridCard({ article, index }: Props) {
  const navigate = useNavigate()
  const hasCover = !!(article.cover_image?.url)
  const readTime = estimateReadTime(article.content || '')

  return (
    <motion.div
      initial={{ opacity: 0, filter: 'blur(8px)', y: 20 }}
      whileInView={{ opacity: 1, filter: 'blur(0px)', y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.6, delay: index * 0.08, ease: [0.25, 0.46, 0.45, 0.94] }}
      whileHover={{ y: -8 }}
      onClick={() => navigate(`/article/${article.slug}`)}
      className="liquid-glass-card"
      style={{
        cursor: 'pointer',
        overflow: 'hidden',
        display: 'flex', flexDirection: 'column',
        height: '100%',
      }}
    >
      {/* ── Cover image ── */}
      <div style={{
        width: '100%', height: 0, paddingBottom: '56.25%',
        position: 'relative', overflow: 'hidden',
        background: hasCover
          ? 'transparent'
          : `linear-gradient(135deg, #4f6ef7 0%, #8b5cf6 50%, #ec4899 100%)`,
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
            className="card-cover-img"
          />
        )}
        {/* Hover overlay */}
        <div style={{
          position: 'absolute', inset: 0,
          background: 'linear-gradient(to top, rgba(0,0,0,0.5) 0%, transparent 40%)',
          opacity: 0, transition: 'opacity 0.4s ease',
        }}
          className="card-cover-overlay"
        />
        {!hasCover && article.category && (
          <Text style={{
            position: 'absolute', inset: 0, display: 'flex',
            alignItems: 'center', justifyContent: 'center',
            color: 'rgba(255,255,255,0.9)', fontSize: 24, fontWeight: 700,
            fontFamily: "'Instrument Serif', serif", fontStyle: 'italic',
          }}>
            {article.category.name}
          </Text>
        )}
        {!hasCover && !article.category && (
          <FileTextOutlined style={{
            position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%,-50%)',
            fontSize: 56, color: 'rgba(255,255,255,0.4)',
          }} />
        )}
        {article.is_top && (
          <span style={{
            position: 'absolute', top: 12, right: 12, zIndex: 2,
            background: 'rgba(239,68,68,0.85)', color: '#fff',
            padding: '4px 12px', borderRadius: 6, fontSize: 11, fontWeight: 600,
            fontFamily: "'Barlow', sans-serif", letterSpacing: '0.04em',
          }}>
            置顶
          </span>
        )}
        {!article.is_top && article.visibility === 'private' && (
          <span style={{
            position: 'absolute', top: 12, right: 12, zIndex: 2,
            display: 'inline-flex', alignItems: 'center', gap: 4,
            background: 'rgba(239,68,68,0.7)', color: '#fff',
            padding: '4px 12px', borderRadius: 6, fontSize: 11, fontWeight: 600,
            fontFamily: "'Barlow', sans-serif", letterSpacing: '0.04em',
            backdropFilter: 'blur(4px)',
          }}>
            <LockOutlined style={{ fontSize: 10 }} /> 私密
          </span>
        )}
        {!article.is_top && article.visibility === 'partial_visible' && (
          <span style={{
            position: 'absolute', top: 12, right: 12, zIndex: 2,
            display: 'inline-flex', alignItems: 'center', gap: 4,
            background: 'rgba(168,85,247,0.7)', color: '#fff',
            padding: '4px 12px', borderRadius: 6, fontSize: 11, fontWeight: 600,
            fontFamily: "'Barlow', sans-serif", letterSpacing: '0.04em',
            backdropFilter: 'blur(4px)',
          }}>
            <TeamOutlined style={{ fontSize: 10 }} /> 部分可见
          </span>
        )}
      </div>

      {/* ── Content ── */}
      <div style={{ padding: '22px 20px 24px', display: 'flex', flexDirection: 'column', flex: 1 }}>
        {/* Title — Instrument Serif italic */}
        <h3 style={{
          fontFamily: "'Instrument Serif', serif",
          fontStyle: 'italic',
          fontSize: 20, fontWeight: 400, color: '#ffffff',
          letterSpacing: '-0.02em', lineHeight: 1.25, margin: '0 0 10px',
          overflow: 'hidden', display: '-webkit-box',
          WebkitLineClamp: 2, WebkitBoxOrient: 'vertical',
        }}>
          {article.title}
        </h3>

        {/* Summary */}
        <p style={{
          fontFamily: "'Barlow', sans-serif",
          fontSize: 14, fontWeight: 400, color: 'rgba(255,255,255,0.45)',
          lineHeight: 1.6, margin: '0 0 14px',
          overflow: 'hidden', display: '-webkit-box',
          WebkitLineClamp: 2, WebkitBoxOrient: 'vertical',
          flex: 1,
        }}>
          {article.summary || article.content?.replace(/[#*`>\[\]!~-]/g, '').slice(0, 140)}
        </p>

        {/* Category pill */}
        {article.category && (
          <div style={{ marginBottom: 14 }}>
            <span style={{
              display: 'inline-block',
              fontFamily: "'Barlow', sans-serif",
              fontSize: 11, fontWeight: 500, color: '#4f6ef7',
              background: 'rgba(79,110,247,0.12)',
              padding: '4px 12px', borderRadius: 6,
              letterSpacing: '0.04em',
              border: '1px solid rgba(79,110,247,0.2)',
            }}>
              {article.category.name}
            </span>
          </div>
        )}

        {/* Meta */}
        <Space size={16} style={{ fontSize: 12 }}>
          <Text style={{ fontSize: 12, color: 'rgba(255,255,255,0.35)', fontFamily: "'Barlow', sans-serif", display: 'flex', alignItems: 'center', gap: 4 }}>
            <ClockCircleOutlined /> {dayjs(article.published_at).format('MM-DD')}
          </Text>
          <Text style={{ fontSize: 12, color: 'rgba(255,255,255,0.35)', fontFamily: "'Barlow', sans-serif", display: 'flex', alignItems: 'center', gap: 4 }}>
            {readTime} 分钟
          </Text>
          <Text style={{ fontSize: 12, color: 'rgba(255,255,255,0.35)', fontFamily: "'Barlow', sans-serif", display: 'flex', alignItems: 'center', gap: 4 }}>
            <EyeOutlined /> {article.views || 0}
          </Text>
        </Space>
      </div>

      <style>{`
        .liquid-glass-card:hover .card-cover-img {
          transform: scale(1.06);
          filter: brightness(1.15);
        }
        .liquid-glass-card:hover .card-cover-overlay {
          opacity: 1;
        }
      `}</style>
    </motion.div>
  )
}

import { useNavigate } from 'react-router-dom'
import { Spin } from 'antd'
import { motion } from 'framer-motion'
import { useCategories } from '../hooks/useCategories'

const categoryAccents = [
  'rgba(79,110,247,0.2)',   // blue
  'rgba(16,185,129,0.2)',   // green
  'rgba(245,158,11,0.2)',   // amber
  'rgba(139,92,246,0.2)',   // violet
  'rgba(236,72,153,0.2)',   // pink
  'rgba(20,184,166,0.2)',   // teal
  'rgba(239,68,68,0.2)',    // red
  'rgba(250,140,22,0.2)',   // orange
]

const categoryBorders = [
  'rgba(79,110,247,0.35)',
  'rgba(16,185,129,0.35)',
  'rgba(245,158,11,0.35)',
  'rgba(139,92,246,0.35)',
  'rgba(236,72,153,0.35)',
  'rgba(20,184,166,0.35)',
  'rgba(239,68,68,0.35)',
  'rgba(250,140,22,0.35)',
]

export default function CategoriesSection() {
  const navigate = useNavigate()
  const { data, isLoading } = useCategories()
  const categories = data?.data || []

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: '60px 20px' }}>
        <Spin size="large" />
      </div>
    )
  }

  if (categories.length === 0) return null

  return (
    <section style={{ marginBottom: 80 }}>
      {/* ── Title ── */}
      <motion.div
        initial={{ opacity: 0, y: -16 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.7, ease: [0.25, 0.46, 0.45, 0.94] }}
        style={{ textAlign: 'center', marginBottom: 40 }}
      >
        <h2 style={{
          fontFamily: "'Instrument Serif', serif",
          fontStyle: 'italic',
          fontSize: 'clamp(28px, 5vw, 42px)',
          fontWeight: 400,
          color: '#ffffff',
          letterSpacing: '-0.03em',
          margin: '0 0 16px',
        }}>
          热门分类
        </h2>
        <div style={{
          width: 60, height: 2,
          background: 'linear-gradient(90deg, transparent, #4f6ef7, #8b5cf6, transparent)',
          margin: '0 auto', borderRadius: 1,
        }} />
      </motion.div>

      {/* ── Category pills grid ── */}
      <motion.div
        initial="hidden"
        whileInView="visible"
        viewport={{ once: true }}
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))',
          gap: 16, maxWidth: 1200, margin: '0 auto',
        }}
      >
        {categories.map((cat, i) => (
          <motion.div
            key={cat.id}
            initial={{ opacity: 0, filter: 'blur(6px)', y: 16 }}
            whileInView={{ opacity: 1, filter: 'blur(0px)', y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.45, delay: i * 0.06, ease: [0.25, 0.46, 0.45, 0.94] }}
            whileHover={{ scale: 1.05, y: -4 }}
            onClick={() => navigate(`/categories?category_id=${cat.id}`)}
            className="liquid-glass"
            style={{ cursor: 'pointer', padding: '28px 20px', textAlign: 'center' }}
          >
            <div style={{
              fontFamily: "'Instrument Serif', serif",
              fontStyle: 'italic',
              fontSize: 24, fontWeight: 400,
              color: '#ffffff',
              marginBottom: 6,
              letterSpacing: '-0.02em',
            }}>
              {cat.name}
            </div>
            <div style={{
              fontFamily: "'Barlow', sans-serif",
              fontSize: 13, fontWeight: 400,
              color: 'rgba(255,255,255,0.4)',
              letterSpacing: '0.04em',
            }}>
              {cat.article_count || 0} 篇文章
            </div>
          </motion.div>
        ))}
      </motion.div>
    </section>
  )
}

import { useNavigate } from 'react-router-dom'
import { Typography } from 'antd'
import { FileTextOutlined } from '@ant-design/icons'
import { motion } from 'framer-motion'
import { useArticles } from '../hooks/useArticles'
import ArticleGridCard from './ArticleGridCard'

const { Text } = Typography

export default function FeaturedArticles() {
  const navigate = useNavigate()
  const { data, isLoading } = useArticles({
    current: 1,
    pageSize: 4,
    status: 'published',
  })

  const articles = data?.data || []

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: '80px 20px' }}>
        <div style={{
          display: 'inline-block', width: 44, height: 44,
          borderRadius: '50%',
          border: '3px solid rgba(255,255,255,0.06)',
          borderTopColor: '#4f6ef7',
          animation: 'spin 0.8s linear infinite',
        }} />
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    )
  }

  if (articles.length === 0) {
    return (
      <div style={{ textAlign: 'center', padding: '80px 20px' }}>
        <FileTextOutlined style={{ fontSize: 48, color: 'rgba(255,255,255,0.15)', marginBottom: 16 }} />
        <p style={{ color: 'rgba(255,255,255,0.35)', fontFamily: "'Barlow', sans-serif" }}>暂无文章</p>
      </div>
    )
  }

  return (
    <section style={{ marginBottom: 80 }}>
      {/* ── Section title with decorative line ── */}
      <motion.div
        initial={{ opacity: 0, y: -16 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.7, ease: [0.25, 0.46, 0.45, 0.94] }}
        style={{ textAlign: 'center', marginBottom: 48 }}
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
          最新精选文章
        </h2>
        <div style={{
          width: 80, height: 2,
          background: 'linear-gradient(90deg, transparent, #4f6ef7, #8b5cf6, transparent)',
          margin: '0 auto',
          borderRadius: 1,
        }} />
      </motion.div>

      {/* ── Card grid ── */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
        gap: 28,
        maxWidth: 1400, margin: '0 auto',
      }}>
        {articles.map((article, idx) => (
          <ArticleGridCard key={article.id} article={article} index={idx} />
        ))}
      </div>

      {/* ── "View all" link ── */}
      <motion.div
        initial={{ opacity: 0 }}
        whileInView={{ opacity: 1 }}
        viewport={{ once: true }}
        transition={{ duration: 0.5, delay: 0.4 }}
        style={{ textAlign: 'center', marginTop: 48 }}
      >
        <motion.button
          type="button"
          whileHover={{ scale: 1.04, x: 6 }}
          whileTap={{ scale: 0.97 }}
          onClick={() => navigate('/articles')}
          className="liquid-glass"
          style={{
            display: 'inline-flex', alignItems: 'center', gap: 8,
            padding: '12px 28px',
            cursor: 'pointer',
            fontFamily: "'Barlow', sans-serif",
            fontSize: 15, fontWeight: 500,
            color: 'rgba(255,255,255,0.75)',
            letterSpacing: '0.03em',
            border: 'none', outline: 'none',
          }}
        >
          查看更多文章 →
        </motion.button>
      </motion.div>
    </section>
  )
}

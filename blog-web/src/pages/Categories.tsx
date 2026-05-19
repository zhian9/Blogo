import { useSearchParams, Link } from 'react-router-dom'
import { Spin } from 'antd'
import { AppstoreOutlined } from '@ant-design/icons'
import { motion } from 'framer-motion'
import ArticleList from '../components/ArticleList'
import { useCategories } from '../hooks/useCategories'

export default function Categories() {
  const [searchParams] = useSearchParams()
  const selectedCategoryId = searchParams.get('category_id') || ''
  const { data, isLoading } = useCategories()

  const categories = data?.data || []
  const selectedCategory = categories.find((c) => c.id === selectedCategoryId)

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 120 }}>
        <Spin size="large" />
      </div>
    )
  }

  if (categories.length === 0) {
    return (
      <div style={{
        display: 'flex', flexDirection: 'column', alignItems: 'center',
        padding: 100, gap: 20,
      }}>
        <div style={{
          width: 100, height: 100, borderRadius: '50%',
          background: 'rgba(79,110,247,0.08)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>
          <AppstoreOutlined style={{ fontSize: 44, color: '#4f6ef7', opacity: 0.5 }} />
        </div>
        <h2 style={{
          fontFamily: "'Instrument Serif', serif", fontStyle: 'italic',
          fontSize: 28, fontWeight: 400, color: '#ffffff', margin: 0,
        }}>
          暂无分类
        </h2>
        <p style={{
          color: 'rgba(255,255,255,0.4)', fontSize: 14,
          fontFamily: "'Barlow', sans-serif", margin: 0,
        }}>
          还没有任何分类
        </p>
      </div>
    )
  }

  return (
    <div style={{ maxWidth: 960, margin: '0 auto' }}>
      {/* ===== HEADER ===== */}
      <motion.div
        initial={{ opacity: 0, y: -16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease: [0.25, 0.46, 0.45, 0.94] }}
        style={{ textAlign: 'center', marginBottom: 36 }}
      >
        <h1 style={{
          fontFamily: "'Instrument Serif', serif", fontStyle: 'italic',
          fontSize: 'clamp(28px, 5vw, 44px)', fontWeight: 400,
          color: '#ffffff', letterSpacing: '-0.03em',
          margin: '0 0 8px',
        }}>
          <AppstoreOutlined style={{ marginRight: 12, color: '#4f6ef7' }} />
          {selectedCategory ? selectedCategory.name : '分类'}
        </h1>
        <p style={{
          fontFamily: "'Barlow', sans-serif",
          fontSize: 15, fontWeight: 300, color: 'rgba(255,255,255,0.45)',
          margin: 0, letterSpacing: '0.04em',
        }}>
          {selectedCategory
            ? `浏览 "${selectedCategory.name}" 分类下的所有文章`
            : '按分类浏览文章'}
        </p>
      </motion.div>

      {/* ===== CATEGORY PILLS ===== */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.15, duration: 0.5 }}
        style={{
          display: 'flex', flexWrap: 'wrap', gap: 10,
          justifyContent: 'center', marginBottom: 40,
        }}
      >
        {categories.map((cat, i) => {
          const isActive = selectedCategoryId === cat.id
          return (
            <motion.div
              key={cat.id}
              initial={{ opacity: 0, scale: 0.85 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.1 + i * 0.03, duration: 0.35 }}
              whileHover={{ scale: 1.06 }}
              whileTap={{ scale: 0.95 }}
            >
              <Link
                to={`/categories?category_id=${cat.id}`}
                style={{
                  display: 'inline-block',
                  padding: isActive ? '8px 20px' : '7px 18px',
                  borderRadius: 12,
                  fontFamily: "'Barlow', sans-serif",
                  fontSize: 14, fontWeight: isActive ? 600 : 400,
                  color: isActive ? '#ffffff' : 'rgba(255,255,255,0.55)',
                  background: isActive
                    ? 'rgba(79,110,247,0.2)'
                    : 'rgba(255,255,255,0.04)',
                  border: isActive
                    ? '1px solid rgba(79,110,247,0.35)'
                    : '1px solid rgba(255,255,255,0.07)',
                  textDecoration: 'none',
                  transition: 'all 0.25s ease',
                }}
                onMouseEnter={(e) => {
                  if (!isActive) {
                    e.currentTarget.style.background = 'rgba(255,255,255,0.08)'
                    e.currentTarget.style.borderColor = 'rgba(255,255,255,0.14)'
                    e.currentTarget.style.color = '#ffffff'
                  }
                }}
                onMouseLeave={(e) => {
                  if (!isActive) {
                    e.currentTarget.style.background = 'rgba(255,255,255,0.04)'
                    e.currentTarget.style.borderColor = 'rgba(255,255,255,0.07)'
                    e.currentTarget.style.color = 'rgba(255,255,255,0.55)'
                  }
                }}
              >
                {cat.name}
              </Link>
            </motion.div>
          )
        })}
      </motion.div>

      {/* ===== CONTENT ===== */}
      {selectedCategoryId ? (
        <motion.div
          key={selectedCategoryId}
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
        >
          <ArticleList
            params={{
              current: 1,
              pageSize: 10,
              category_id: selectedCategoryId,
              status: 'published',
            }}
            emptyDescription={`暂无 "${selectedCategory?.name || '...'}" 分类下的文章`}
          />
        </motion.div>
      ) : (
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.4 }}
          style={{
            display: 'flex', flexDirection: 'column', alignItems: 'center',
            padding: 60, gap: 16,
          }}
        >
          <div style={{
            width: 80, height: 80, borderRadius: '50%',
            background: 'rgba(79,110,247,0.06)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>
            <AppstoreOutlined style={{ fontSize: 36, color: '#4f6ef7', opacity: 0.5 }} />
          </div>
          <p style={{
            fontFamily: "'Barlow', sans-serif",
            fontSize: 16, color: 'rgba(255,255,255,0.4)',
            margin: 0, letterSpacing: '0.04em',
          }}>
            选择一个分类查看相关文章
          </p>
        </motion.div>
      )}
    </div>
  )
}

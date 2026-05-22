import { Space, Spin } from 'antd'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useCategories } from '../hooks/useCategories'
import { useTags } from '../hooks/useTags'
import { useArticles } from '../hooks/useArticles'

export default function Sidebar() {
  const { data: catData, isLoading: catLoading } = useCategories()
  const { data: tagData, isLoading: tagLoading } = useTags()
  const { data: hotData, isLoading: hotLoading } = useArticles({
    current: 1,
    pageSize: 5,
    status: 'published',
  })

  const categories = catData?.data || []
  const tags = tagData?.data || []
  const hotArticles = hotData?.data || []

  return (
    <aside>
      <Space vertical size={16} style={{ width: '100%' }}>
        {/* ── Categories ── */}
        <div className="liquid-glass-card" style={{ padding: '20px', overflow: 'hidden' }}>
          <h4 style={{
            fontFamily: "'Instrument Serif', serif", fontStyle: 'italic',
            fontSize: 20, fontWeight: 400, color: '#ffffff',
            letterSpacing: '-0.02em', margin: '0 0 16px',
          }}>
            分类
          </h4>
          {catLoading ? (
            <div style={{ textAlign: 'center', padding: 16 }}><Spin size="small" /></div>
          ) : (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
              {categories.map((cat) => (
                <Link key={cat.id} to={`/categories?category_id=${cat.id}`}>
                  <motion.span
                    whileHover={{ scale: 1.08 }}
                    style={{
                      display: 'inline-block',
                      fontFamily: "'Barlow', sans-serif",
                      fontSize: 12, fontWeight: 500,
                      color: 'rgba(255,255,255,0.65)',
                      background: 'rgba(255,255,255,0.05)',
                      padding: '4px 12px',
                      borderRadius: 8,
                      border: '1px solid rgba(255,255,255,0.08)',
                      transition: 'all 0.25s ease',
                      cursor: 'pointer',
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.background = 'rgba(79,110,247,0.15)'
                      e.currentTarget.style.borderColor = 'rgba(79,110,247,0.3)'
                      e.currentTarget.style.color = '#ffffff'
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.background = 'rgba(255,255,255,0.05)'
                      e.currentTarget.style.borderColor = 'rgba(255,255,255,0.08)'
                      e.currentTarget.style.color = 'rgba(255,255,255,0.65)'
                    }}
                  >
                    {cat.name}
                  </motion.span>
                </Link>
              ))}
            </div>
          )}
        </div>

        {/* ── Tags ── */}
        <div className="liquid-glass-card" style={{ padding: '20px', overflow: 'hidden' }}>
          <h4 style={{
            fontFamily: "'Instrument Serif', serif", fontStyle: 'italic',
            fontSize: 20, fontWeight: 400, color: '#ffffff',
            letterSpacing: '-0.02em', margin: '0 0 16px',
          }}>
            标签
          </h4>
          {tagLoading ? (
            <div style={{ textAlign: 'center', padding: 16 }}><Spin size="small" /></div>
          ) : (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
              {tags.map((tag) => (
                <Link key={tag.id} to={`/tags?tag_name=${encodeURIComponent(tag.name)}`}>
                  <motion.span
                    whileHover={{ scale: 1.08 }}
                    style={{
                      display: 'inline-block',
                      fontFamily: "'Barlow', sans-serif",
                      fontSize: 12, fontWeight: 500,
                      color: 'rgba(255,255,255,0.55)',
                      background: 'rgba(255,255,255,0.03)',
                      padding: '4px 10px',
                      borderRadius: 6,
                      border: '1px solid rgba(255,255,255,0.06)',
                      transition: 'all 0.25s ease',
                      cursor: 'pointer',
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.background = 'rgba(255,255,255,0.07)'
                      e.currentTarget.style.borderColor = 'rgba(255,255,255,0.14)'
                      e.currentTarget.style.color = '#ffffff'
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.background = 'rgba(255,255,255,0.03)'
                      e.currentTarget.style.borderColor = 'rgba(255,255,255,0.06)'
                      e.currentTarget.style.color = 'rgba(255,255,255,0.55)'
                    }}
                  >
                    {tag.name}
                  </motion.span>
                </Link>
              ))}
            </div>
          )}
        </div>

        {/* ── Hot Articles ── */}
        <div className="liquid-glass-card" style={{ padding: '20px' }}>
          <h4 style={{
            fontFamily: "'Instrument Serif', serif", fontStyle: 'italic',
            fontSize: 20, fontWeight: 400, color: '#ffffff',
            letterSpacing: '-0.02em', margin: '0 0 16px',
          }}>
            热门文章
          </h4>
          {hotLoading ? (
            <div style={{ textAlign: 'center', padding: 16 }}><Spin size="small" /></div>
          ) : (
            <Space vertical size={12} style={{ width: '100%' }}>
              {hotArticles
                .sort((a, b) => b.views - a.views)
                .slice(0, 5)
                .map((article) => (
                  <Link
                    key={article.id}
                    to={`/article/${article.slug}`}
                    style={{ display: 'block', textDecoration: 'none' }}
                  >
                    <div
                      style={{
                        fontFamily: "'Barlow', sans-serif",
                        fontSize: 14, fontWeight: 400,
                        color: 'rgba(255,255,255,0.55)',
                        lineHeight: 1.5,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                        padding: '6px 0',
                        transition: 'color 0.2s ease',
                      }}
                      onMouseEnter={(e) => { e.currentTarget.style.color = '#ffffff' }}
                      onMouseLeave={(e) => { e.currentTarget.style.color = 'rgba(255,255,255,0.55)' }}
                    >
                      {article.title}
                    </div>
                  </Link>
                ))}
            </Space>
          )}
        </div>
      </Space>
    </aside>
  )
}

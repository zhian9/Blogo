import { Spin } from 'antd'
import { motion } from 'framer-motion'
import type { Tag as TagType } from '../types'

interface Props {
  tags: TagType[]
  selectedTags: string[]
  onTagClick: (tagName: string) => void
  loading?: boolean
}

export default function TagCloud({ tags, selectedTags, onTagClick, loading }: Props) {
  if (loading) {
    return (
      <div className="liquid-glass-card" style={{ padding: '20px', marginBottom: 24 }}>
        <h4 style={{
          fontFamily: "'Instrument Serif', serif", fontStyle: 'italic',
          fontSize: 20, fontWeight: 400, color: '#ffffff',
          letterSpacing: '-0.02em', margin: '0 0 16px',
        }}>
          热门标签
        </h4>
        <div style={{ textAlign: 'center', padding: 20 }}><Spin size="small" /></div>
      </div>
    )
  }

  if (tags.length === 0) {
    return (
      <div className="liquid-glass-card" style={{ padding: '20px', marginBottom: 24 }}>
        <h4 style={{
          fontFamily: "'Instrument Serif', serif", fontStyle: 'italic',
          fontSize: 20, fontWeight: 400, color: '#ffffff',
          letterSpacing: '-0.02em', margin: '0 0 16px',
        }}>
          热门标签
        </h4>
        <p style={{ color: 'rgba(255,255,255,0.35)', fontSize: 13, fontFamily: "'Barlow', sans-serif" }}>
          暂无标签
        </p>
      </div>
    )
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.5 }}
      className="liquid-glass-card"
      style={{ padding: '20px', marginBottom: 24 }}
    >
      <h4 style={{
        fontFamily: "'Instrument Serif', serif", fontStyle: 'italic',
        fontSize: 20, fontWeight: 400, color: '#ffffff',
        letterSpacing: '-0.02em', margin: '0 0 16px',
      }}>
        热门标签
      </h4>

      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
        {tags.slice(0, 20).map((tag, i) => {
          const isSelected = selectedTags.includes(tag.name)
          return (
            <motion.span
              key={tag.id}
              initial={{ opacity: 0, scale: 0.8 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.35, delay: i * 0.04 }}
              whileHover={{ scale: 1.1 }}
              onClick={() => onTagClick(tag.name)}
              style={{
                display: 'inline-block',
                fontFamily: "'Barlow', sans-serif",
                fontSize: 12, fontWeight: isSelected ? 600 : 400,
                color: isSelected ? '#ffffff' : 'rgba(255,255,255,0.55)',
                background: isSelected ? 'rgba(79,110,247,0.2)' : 'rgba(255,255,255,0.04)',
                padding: '5px 12px',
                borderRadius: 8,
                border: isSelected
                  ? '1px solid rgba(79,110,247,0.4)'
                  : '1px solid rgba(255,255,255,0.07)',
                cursor: 'pointer',
                transition: 'all 0.25s ease',
              }}
              onMouseEnter={(e) => {
                if (!isSelected) {
                  e.currentTarget.style.background = 'rgba(255,255,255,0.08)'
                  e.currentTarget.style.borderColor = 'rgba(255,255,255,0.14)'
                  e.currentTarget.style.color = '#ffffff'
                }
              }}
              onMouseLeave={(e) => {
                if (!isSelected) {
                  e.currentTarget.style.background = 'rgba(255,255,255,0.04)'
                  e.currentTarget.style.borderColor = 'rgba(255,255,255,0.07)'
                  e.currentTarget.style.color = 'rgba(255,255,255,0.55)'
                }
              }}
            >
              {tag.name}
            </motion.span>
          )
        })}
      </div>
    </motion.div>
  )
}

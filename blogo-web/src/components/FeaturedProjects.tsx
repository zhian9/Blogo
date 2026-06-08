import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Tag, Empty, Spin } from 'antd'
import {
  StarOutlined,
  EyeOutlined,
  HeartOutlined,
  GithubOutlined,
  LinkOutlined,
} from '@ant-design/icons'
import { useFeaturedProjects } from '../hooks/useProjects'
import type { Project } from '../types'

const stateColors: Record<string, string> = {
  developing: '#f59e0b',
  completed: '#34d399',
  maintaining: '#60a5fa',
  paused: '#fbbf24',
  archived: '#9ca3af',
}

const stateLabels: Record<string, string> = {
  developing: '开发中',
  completed: '已完成',
  maintaining: '维护中',
  paused: '暂停开发',
  archived: '已归档',
}

function FeaturedProjectCard({ project, index }: { project: Project; index: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 24 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay: index * 0.1, ease: [0.25, 0.46, 0.45, 0.94] }}
      viewport={{ once: true }}
      style={{ position: 'relative' }}
    >
      <Link to={`/project/${project.slug}`} style={{ textDecoration: 'none' }}>
        <div
          style={{
            position: 'relative',
            borderRadius: 16,
            overflow: 'hidden',
            background: 'rgba(255,255,255,0.04)',
            border: '1px solid rgba(255,255,255,0.08)',
            backdropFilter: 'blur(12px)',
            transition: 'all 0.3s ease',
            cursor: 'pointer',
          }}
          className="liquid-glass-card"
        >
          {/* Cover */}
          <div style={{ position: 'relative', paddingBottom: '56.25%', overflow: 'hidden' }}>
            {project.cover_image?.url ? (
              <img
                src={project.cover_image.url}
                alt={project.title}
                style={{
                  position: 'absolute', inset: 0, width: '100%', height: '100%',
                  objectFit: 'cover', transition: 'transform 0.4s ease',
                }}
              />
            ) : (
              <div style={{
                position: 'absolute', inset: 0,
                background: 'linear-gradient(135deg, #4f6ef7 0%, #8b5cf6 50%, #ec4899 100%)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}>
                <span style={{
                  fontFamily: 'Instrument Serif, serif', fontStyle: 'italic',
                  fontSize: 32, color: 'rgba(255,255,255,0.6)',
                }}>
                  {project.title.charAt(0)}
                </span>
              </div>
            )}
            {/* State badge */}
            <div style={{ position: 'absolute', top: 12, right: 12 }}>
              <span style={{
                display: 'inline-flex', alignItems: 'center', gap: 4,
                padding: '2px 10px', borderRadius: 20,
                background: 'rgba(0,0,0,0.6)', backdropFilter: 'blur(8px)',
                fontSize: 12, color: stateColors[project.project_state] || '#9ca3af',
                border: `1px solid ${stateColors[project.project_state] || '#9ca3af'}33`,
              }}>
                <span style={{
                  width: 6, height: 6, borderRadius: '50%',
                  background: stateColors[project.project_state] || '#9ca3af',
                }} />
                {stateLabels[project.project_state] || project.project_state}
              </span>
            </div>
          </div>

          {/* Content */}
          <div style={{ padding: '16px 20px 20px' }}>
            <h3 style={{
              fontFamily: 'Instrument Serif, serif', fontStyle: 'italic',
              fontSize: 20, fontWeight: 500, color: '#fff', margin: 0,
              marginBottom: 6, lineHeight: 1.3,
              display: '-webkit-box', WebkitLineClamp: 1, WebkitBoxOrient: 'vertical', overflow: 'hidden',
            }}>
              {project.title}
            </h3>
            <p style={{
              fontSize: 14, color: 'rgba(255,255,255,0.55)', margin: 0, marginBottom: 12, lineHeight: 1.5,
              display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden',
            }}>
              {project.summary || project.content?.replace(/[#*`>\-\[\]()!]/g, '').slice(0, 100)}
            </p>

            {/* Tags */}
            {project.tags && project.tags.length > 0 && (
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, marginBottom: 12 }}>
                {project.tags.slice(0, 4).map(tag => (
                  <Tag key={tag.id} style={{
                    margin: 0, borderRadius: 12, fontSize: 11, padding: '1px 8px',
                    background: 'rgba(79,110,247,0.12)', border: '1px solid rgba(79,110,247,0.2)',
                    color: '#818cf8',
                  }}>
                    {tag.name}
                  </Tag>
                ))}
                {project.tags.length > 4 && (
                  <Tag style={{
                    margin: 0, borderRadius: 12, fontSize: 11, padding: '1px 8px',
                    background: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.1)',
                    color: 'rgba(255,255,255,0.4)',
                  }}>
                    +{project.tags.length - 4}
                  </Tag>
                )}
              </div>
            )}

            {/* Meta */}
            <div style={{ display: 'flex', alignItems: 'center', gap: 16, fontSize: 12, color: 'rgba(255,255,255,0.4)' }}>
              <span><EyeOutlined /> {project.views?.toLocaleString()}</span>
              <span><HeartOutlined /> {project.like_count}</span>
              {project.github_url && <GithubOutlined style={{ color: 'rgba(255,255,255,0.3)' }} />}
              {project.demo_url && <LinkOutlined style={{ color: 'rgba(255,255,255,0.3)' }} />}
            </div>
          </div>
        </div>
      </Link>
    </motion.div>
  )
}

export default function FeaturedProjects() {
  const { data, isLoading } = useFeaturedProjects()
  const projects = data?.data

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: '40px 0' }}>
        <Spin />
      </div>
    )
  }

  if (!projects || projects.length === 0) {
    return null
  }

  return (
    <section style={{ maxWidth: 1400, margin: '0 auto', padding: '0 24px' }}>
      {/* Section header */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
        viewport={{ once: true }}
        style={{ textAlign: 'center', marginBottom: 48 }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8, marginBottom: 8 }}>
          <StarOutlined style={{ color: '#f59e0b', fontSize: 20 }} />
          <h2 style={{
            fontFamily: 'Instrument Serif, serif', fontStyle: 'italic',
            fontSize: 32, fontWeight: 500, color: '#fff', margin: 0,
          }}>
            精选项目
          </h2>
        </div>
        <p style={{ color: 'rgba(255,255,255,0.45)', fontSize: 15, margin: 0 }}>
          Featured Projects
        </p>
      </motion.div>

      {/* Grid - PC 2x2, mobile horizontal scroll */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))',
        gap: 24,
      }}>
        {projects.slice(0, 4).map((project, i) => (
          <FeaturedProjectCard key={project.id} project={project} index={i} />
        ))}
      </div>

      {/* View all link */}
      <motion.div
        initial={{ opacity: 0 }}
        whileInView={{ opacity: 1 }}
        transition={{ duration: 0.5, delay: 0.4 }}
        viewport={{ once: true }}
        style={{ textAlign: 'center', marginTop: 40 }}
      >
        <Link to="/projects" style={{
          display: 'inline-flex', alignItems: 'center', gap: 6,
          color: '#818cf8', fontSize: 14, textDecoration: 'none',
          padding: '8px 20px', borderRadius: 20,
          border: '1px solid rgba(129,140,248,0.2)',
          background: 'rgba(129,140,248,0.06)',
          transition: 'all 0.3s ease',
        }}>
          查看全部项目 →
        </Link>
      </motion.div>
    </section>
  )
}

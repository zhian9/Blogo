import { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { Select, Input, Pagination, Spin, Empty, Tag, Button } from 'antd'
import {
  SearchOutlined, StarFilled, EyeOutlined, LikeOutlined,
  ClockCircleOutlined, PlusOutlined,
} from '@ant-design/icons'
import { motion } from 'framer-motion'
import { useProjects } from '../hooks/useProjects'
import { useCategories } from '../hooks/useCategories'
import { useTags } from '../hooks/useTags'
import { useAuthStore } from '../store/authStore'
import TagCloud from '../components/TagCloud'
import dayjs from '../utils/dayjs'

const PAGE_SIZE = 12

const PROJECT_STATES = [
  { value: '', label: '全部' },
  { value: 'developing', label: '开发中' },
  { value: 'completed', label: '已完成' },
  { value: 'maintaining', label: '维护中' },
  { value: 'paused', label: '暂停开发' },
  { value: 'archived', label: '已归档' },
]

const SORT_OPTIONS = [
  { value: 'latest', label: '最新发布' },
  { value: 'hot', label: '最多浏览' },
  { value: 'most_liked', label: '最多点赞' },
]

const STATE_COLOR_MAP: Record<string, string> = {
  developing: '#fa8c16',
  completed: '#52c41a',
  maintaining: '#1890ff',
  paused: '#fadb14',
  archived: '#8c8c8c',
}

const STATE_LABEL_MAP: Record<string, string> = {
  developing: '开发中',
  completed: '已完成',
  maintaining: '维护中',
  paused: '暂停开发',
  archived: '已归档',
}

const cardVariants = {
  hidden: { opacity: 0, y: 24, filter: 'blur(6px)' },
  visible: (i: number) => ({
    opacity: 1,
    y: 0,
    filter: 'blur(0px)',
    transition: {
      duration: 0.5,
      delay: i * 0.08,
      ease: [0.25, 0.46, 0.45, 0.94],
    },
  }),
}

export default function Projects() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)

  // ── Filter state ──
  const [search, setSearch] = useState('')
  const [projectState, setProjectState] = useState('')
  const [categoryId, setCategoryId] = useState<string | undefined>()
  const [sortBy, setSortBy] = useState<'latest' | 'hot' | 'most_liked'>('latest')
  const [selectedTags, setSelectedTags] = useState<string[]>([])
  const [currentPage, setCurrentPage] = useState(1)

  // ── API calls ──
  const { data: projectsData, isLoading: projectsLoading } = useProjects({
    current: currentPage,
    pageSize: PAGE_SIZE,
    title: search || undefined,
    project_state: projectState || undefined,
    category_id: categoryId,
    status: 'published',
    sort_by: sortBy,
  })

  const { data: categoriesData } = useCategories()
  const { data: tagsData } = useTags()

  const categories = categoriesData?.data || []
  const allTags = tagsData?.data || []

  const projects = projectsData?.data || []
  const total = projectsData?.total || 0

  // ── Local sorting / tag filtering ──
  const displayProjects = useMemo(() => {
    if (!Array.isArray(projects)) return []

    let list = [...projects]

    // Front-end tag filter
    if (selectedTags.length > 0) {
      list = list.filter((p) => {
        if (!Array.isArray(p.tags)) return true
        return selectedTags.some((tag) =>
          p.tags!.some((t) => t.name === tag),
        )
      })
    }

    // Front-end sort fallback
    if (sortBy === 'hot') {
      list.sort((a, b) => (b.views || 0) - (a.views || 0))
    } else if (sortBy === 'most_liked') {
      list.sort((a, b) => (b.like_count || 0) - (a.like_count || 0))
    }

    return list
  }, [projects, selectedTags, sortBy])

  // ── Page animation ──
  const pageVariants = {
    hidden: { opacity: 0 },
    visible: { opacity: 1, transition: { duration: 0.5 } },
  }

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={pageVariants}
      style={{ padding: 0, minHeight: '100vh' }}
    >
      {/* ═══════════════ Hero Section ═══════════════ */}
      <div
        className="liquid-glass"
        style={{
          maxWidth: 1400,
          margin: '40px auto 0',
          padding: '60px 24px 48px',
          textAlign: 'center',
          position: 'relative',
          background: 'rgba(79, 110, 247, 0.04)',
          border: '1px solid rgba(79, 110, 247, 0.1)',
          borderRadius: 24,
          overflow: 'hidden',
        }}
      >
        {/* Ambient glow */}
        <div
          style={{
            position: 'absolute',
            top: '-40%',
            left: '50%',
            transform: 'translateX(-50%)',
            width: 600,
            height: 400,
            borderRadius: '50%',
            background: 'radial-gradient(circle, rgba(79,110,247,0.12) 0%, transparent 70%)',
            pointerEvents: 'none',
          }}
        />

        <motion.h1
          className="cinematic-title"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.1 }}
          style={{
            fontSize: 'clamp(36px, 6vw, 56px)',
            color: '#ffffff',
            margin: '0 0 12px',
            position: 'relative',
          }}
        >
          项目库
        </motion.h1>

        <motion.p
          className="cinematic-subtitle"
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
          style={{
            fontSize: 16,
            color: 'rgba(255,255,255,0.5)',
            margin: '0 0 28px',
            position: 'relative',
          }}
        >
          记录我的产品、开源项目与技术实践
        </motion.p>

        {isAuthenticated && (
          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.35 }}
            style={{ position: 'relative' }}
          >
            <Link to="/project/new">
              <Button
                type="primary"
                icon={<PlusOutlined />}
                size="large"
                style={{
                  borderRadius: 12,
                  background: 'linear-gradient(135deg, #4f6ef7 0%, #8b5cf6 100%)',
                  border: 'none',
                  fontWeight: 600,
                  height: 44,
                  paddingInline: 28,
                  boxShadow: '0 4px 24px rgba(79,110,247,0.35)',
                }}
              >
                发布项目
              </Button>
            </Link>
          </motion.div>
        )}
      </div>

      {/* ═══════════════ Filter Bar ═══════════════ */}
      <div
        style={{
          maxWidth: 1400,
          margin: '32px auto 0',
          padding: '0 24px',
        }}
      >
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.25 }}
          className="liquid-glass-card"
          style={{
            padding: '20px 24px',
            display: 'flex',
            flexWrap: 'wrap',
            gap: 12,
            alignItems: 'center',
          }}
        >
          {/* Search */}
          <Input.Search
            placeholder="搜索项目..."
            allowClear
            prefix={<SearchOutlined style={{ color: 'rgba(255,255,255,0.35)' }} />}
            onSearch={(v) => { setSearch(v); setCurrentPage(1) }}
            onChange={(e) => { if (!e.target.value) { setSearch(''); setCurrentPage(1) } }}
            style={{ flex: '1 1 240px', maxWidth: 360 }}
            className="projects-search"
          />

          {/* 项目状态 */}
          <Select
            value={projectState}
            onChange={(v) => { setProjectState(v); setCurrentPage(1) }}
            options={PROJECT_STATES}
            style={{ width: 140 }}
            popupMatchSelectWidth={false}
          />

          {/* 分类 */}
          <Select
            value={categoryId}
            placeholder="分类"
            allowClear
            onChange={(v) => { setCategoryId(v); setCurrentPage(1) }}
            options={categories.map((c) => ({ value: c.id, label: c.name }))}
            style={{ width: 140 }}
            popupMatchSelectWidth={false}
          />

          {/* 排序 */}
          <Select
            value={sortBy}
            onChange={(v) => setSortBy(v)}
            options={SORT_OPTIONS}
            style={{ width: 140 }}
            popupMatchSelectWidth={false}
          />

          {/* 标签过滤 */}
          <Select
            mode="multiple"
            placeholder="标签"
            allowClear
            value={selectedTags}
            onChange={setSelectedTags}
            options={allTags.map((t) => ({ value: t.name, label: t.name }))}
            style={{ flex: '1 1 200px', maxWidth: 320 }}
            maxTagCount="responsive"
            popupMatchSelectWidth={false}
          />
        </motion.div>
      </div>

      {/* ═══════════════ Project Grid ═══════════════ */}
      <div
        style={{
          maxWidth: 1400,
          margin: '32px auto 0',
          padding: '0 24px',
          marginBottom: 60,
        }}
      >
        {projectsLoading ? (
          <div style={{ textAlign: 'center', padding: '80px 0' }}>
            <Spin size="large" />
          </div>
        ) : displayProjects.length === 0 ? (
          <Empty
            description={
              <span style={{ color: 'rgba(255,255,255,0.4)' }}>暂无项目</span>
            }
            style={{ padding: '80px 0' }}
          />
        ) : (
          <>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
                gap: 24,
              }}
            >
              {displayProjects.map((project, index) => (
                <motion.div
                  key={project.id}
                  custom={index}
                  initial="hidden"
                  whileInView="visible"
                  viewport={{ once: true }}
                  variants={cardVariants}
                  whileHover={{ y: -8 }}
                  style={{ cursor: 'pointer' }}
                >
                  <Link to={`/project/${project.slug}`} style={{ textDecoration: 'none' }}>
                    <div
                      className="liquid-glass-card project-card"
                      style={{
                        overflow: 'hidden',
                        height: '100%',
                        display: 'flex',
                        flexDirection: 'column',
                      }}
                    >
                      {/* ── Cover area (16:9) ── */}
                      <div
                        className="project-card-cover"
                        style={{
                          position: 'relative',
                          paddingTop: '56.25%',
                          overflow: 'hidden',
                          background: project.cover_image?.url
                            ? 'transparent'
                            : 'linear-gradient(135deg, rgba(79,110,247,0.3) 0%, rgba(139,92,246,0.3) 50%, rgba(16,185,129,0.2) 100%)',
                        }}
                      >
                        {project.cover_image?.url && (
                          <img
                            src={project.cover_image.url}
                            alt={project.title}
                            className="project-card-cover-img"
                            style={{
                              position: 'absolute',
                              top: 0,
                              left: 0,
                              width: '100%',
                              height: '100%',
                              objectFit: 'cover',
                              transition: 'transform 0.5s cubic-bezier(0.25, 0.8, 0.25, 1.2)',
                            }}
                          />
                        )}

                        {/* Featured badge (top-left) */}
                        {project.is_featured && (
                          <div
                            style={{
                              position: 'absolute',
                              top: 12,
                              left: 12,
                              display: 'flex',
                              alignItems: 'center',
                              gap: 4,
                              padding: '3px 10px',
                              borderRadius: 8,
                              background: 'rgba(0,0,0,0.55)',
                              backdropFilter: 'blur(10px)',
                              border: '1px solid rgba(250,219,20,0.3)',
                              zIndex: 2,
                            }}
                          >
                            <StarFilled style={{ fontSize: 11, color: '#fadb14' }} />
                            <span style={{ fontSize: 11, fontWeight: 600, color: '#fadb14' }}>
                              精选
                            </span>
                          </div>
                        )}

                        {/* State badge (top-right) */}
                        <div
                          style={{
                            position: 'absolute',
                            top: 12,
                            right: 12,
                            display: 'flex',
                            alignItems: 'center',
                            gap: 5,
                            padding: '3px 10px',
                            borderRadius: 8,
                            background: 'rgba(0,0,0,0.55)',
                            backdropFilter: 'blur(10px)',
                            border: `1px solid ${STATE_COLOR_MAP[project.project_state]}44`,
                            zIndex: 2,
                          }}
                        >
                          <span
                            style={{
                              width: 6,
                              height: 6,
                              borderRadius: '50%',
                              background: STATE_COLOR_MAP[project.project_state] || '#8c8c8c',
                              flexShrink: 0,
                            }}
                          />
                          <span style={{ fontSize: 11, fontWeight: 600, color: 'rgba(255,255,255,0.9)' }}>
                            {STATE_LABEL_MAP[project.project_state] || project.project_state}
                          </span>
                        </div>

                        {/* Gradient overlay at bottom of cover */}
                        <div
                          style={{
                            position: 'absolute',
                            bottom: 0,
                            left: 0,
                            right: 0,
                            height: 60,
                            background: 'linear-gradient(to top, rgba(0,0,0,0.5) 0%, transparent 100%)',
                            pointerEvents: 'none',
                          }}
                        />
                      </div>

                      {/* ── Content ── */}
                      <div style={{ padding: '18px 20px 20px', flex: 1, display: 'flex', flexDirection: 'column' }}>
                        {/* Title */}
                        <h3
                          style={{
                            fontFamily: "'Instrument Serif', serif",
                            fontStyle: 'italic',
                            fontSize: 20,
                            fontWeight: 400,
                            color: '#ffffff',
                            letterSpacing: '-0.02em',
                            lineHeight: 1.3,
                            margin: '0 0 8px',
                            overflow: 'hidden',
                            display: '-webkit-box',
                            WebkitLineClamp: 2,
                            WebkitBoxOrient: 'vertical',
                          }}
                        >
                          {project.title}
                        </h3>

                        {/* Summary */}
                        <p
                          style={{
                            fontSize: 13,
                            color: 'rgba(255,255,255,0.45)',
                            lineHeight: 1.7,
                            fontFamily: "'Barlow', sans-serif",
                            margin: '0 0 14px',
                            overflow: 'hidden',
                            display: '-webkit-box',
                            WebkitLineClamp: 2,
                            WebkitBoxOrient: 'vertical',
                            flex: 1,
                          }}
                        >
                          {project.summary}
                        </p>

                        {/* Tech stack tags */}
                        {project.tags && project.tags.length > 0 && (
                          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, marginBottom: 14 }}>
                            {project.tags.slice(0, 4).map((tag) => (
                              <Tag
                                key={tag.id}
                                style={{
                                  margin: 0,
                                  borderRadius: 6,
                                  fontSize: 11,
                                  padding: '1px 8px',
                                  background: 'rgba(79,110,247,0.1)',
                                  border: '1px solid rgba(79,110,247,0.2)',
                                  color: '#4f6ef7',
                                }}
                              >
                                {tag.name}
                              </Tag>
                            ))}
                            {project.tags.length > 4 && (
                              <Tag
                                style={{
                                  margin: 0,
                                  borderRadius: 6,
                                  fontSize: 11,
                                  padding: '1px 8px',
                                  background: 'rgba(255,255,255,0.04)',
                                  border: '1px solid rgba(255,255,255,0.08)',
                                  color: 'rgba(255,255,255,0.45)',
                                }}
                              >
                                +{project.tags.length - 4}
                              </Tag>
                            )}
                          </div>
                        )}

                        {/* Bottom meta */}
                        <div
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 14,
                            flexWrap: 'wrap',
                            borderTop: '1px solid rgba(255,255,255,0.06)',
                            paddingTop: 12,
                          }}
                        >
                          {/* State label */}
                          <span
                            style={{
                              fontSize: 11,
                              fontWeight: 600,
                              color: STATE_COLOR_MAP[project.project_state] || '#8c8c8c',
                              background: `${STATE_COLOR_MAP[project.project_state] || '#8c8c8c'}18`,
                              padding: '2px 8px',
                              borderRadius: 6,
                              border: `1px solid ${STATE_COLOR_MAP[project.project_state] || '#8c8c8c'}30`,
                            }}
                          >
                            {STATE_LABEL_MAP[project.project_state] || project.project_state}
                          </span>

                          <span
                            style={{
                              fontSize: 12,
                              color: 'rgba(255,255,255,0.35)',
                              fontFamily: "'Barlow', sans-serif",
                              display: 'flex',
                              alignItems: 'center',
                              gap: 4,
                            }}
                          >
                            <EyeOutlined /> {project.views || 0}
                          </span>

                          <span
                            style={{
                              fontSize: 12,
                              color: 'rgba(255,255,255,0.35)',
                              fontFamily: "'Barlow', sans-serif",
                              display: 'flex',
                              alignItems: 'center',
                              gap: 4,
                            }}
                          >
                            <LikeOutlined /> {project.like_count || 0}
                          </span>

                          <span
                            style={{
                              fontSize: 12,
                              color: 'rgba(255,255,255,0.35)',
                              fontFamily: "'Barlow', sans-serif",
                              display: 'flex',
                              alignItems: 'center',
                              gap: 4,
                              marginLeft: 'auto',
                            }}
                          >
                            <ClockCircleOutlined /> {dayjs(project.published_at).format('YYYY-MM-DD')}
                          </span>
                        </div>
                      </div>
                    </div>
                  </Link>
                </motion.div>
              ))}
            </div>

            {/* ═══════════════ Pagination ═══════════════ */}
            {total > PAGE_SIZE && (
              <motion.div
                initial={{ opacity: 0 }}
                whileInView={{ opacity: 1 }}
                transition={{ duration: 0.5, delay: 0.2 }}
                viewport={{ once: true }}
                style={{ textAlign: 'center', marginTop: 48 }}
              >
                <Pagination
                  current={currentPage}
                  pageSize={PAGE_SIZE}
                  total={total}
                  onChange={setCurrentPage}
                  showSizeChanger={false}
                  showQuickJumper
                  showTotal={(t) => `共 ${t} 个项目`}
                />
              </motion.div>
            )}
          </>
        )}
      </div>

      {/* ═══════════════ Tag Cloud (bottom section) ═══════════════ */}
      {allTags.length > 0 && (
        <div
          style={{
            maxWidth: 1400,
            margin: '0 auto 60px',
            padding: '0 24px',
          }}
        >
          <TagCloud
            tags={allTags}
            selectedTags={selectedTags}
            onTagClick={(tagName) => {
              const newTags = selectedTags.includes(tagName)
                ? selectedTags.filter((t) => t !== tagName)
                : [...selectedTags, tagName]
              setSelectedTags(newTags)
              setCurrentPage(1)
            }}
            loading={false}
          />
        </div>
      )}

      {/* ── Responsive & hover styles ── */}
      <style>{`
        .project-card:hover .project-card-cover-img {
          transform: scale(1.06);
        }

        .projects-search .ant-input-wrapper,
        .projects-search .ant-input-affix-wrapper {
          background: rgba(255,255,255,0.04) !important;
          border-color: rgba(255,255,255,0.08) !important;
          border-radius: 10px !important;
        }
        .projects-search .ant-input {
          background: transparent !important;
          color: #fff !important;
        }
        .projects-search .ant-input::placeholder {
          color: rgba(255,255,255,0.3) !important;
        }
        .projects-search .ant-input-search-button {
          background: rgba(79,110,247,0.15) !important;
          border-color: rgba(79,110,247,0.3) !important;
          border-radius: 0 10px 10px 0 !important;
          color: #4f6ef7 !important;
        }

        @media (max-width: 640px) {
          .project-card .project-card-cover {
            padding-top: 50% !important;
          }
        }
      `}</style>
    </motion.div>
  )
}

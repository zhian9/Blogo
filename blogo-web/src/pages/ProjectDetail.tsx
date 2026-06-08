import { useEffect, useMemo, useRef, useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { Typography, Spin, Button, Modal, message, Card, Affix, Drawer, Tag } from 'antd'
import {
  ArrowLeftOutlined, HomeOutlined, LockOutlined, EditOutlined,
  DeleteOutlined, ExclamationCircleOutlined, MenuOutlined,
  CalendarOutlined, EyeOutlined, ClockCircleOutlined, UserOutlined,
  DoubleLeftOutlined, DoubleRightOutlined, ColumnWidthOutlined,
  GithubOutlined, LinkOutlined, HeartOutlined, HeartFilled,
  StarOutlined, StarFilled, ShareAltOutlined, CommentOutlined,
} from '@ant-design/icons'
import {
  useProjectBySlug, useIncProjectViews, useDeleteProject,
  useProjectLikeStatus, useLikeProject, useUnlikeProject,
  useProjectFavoriteStatus, useFavoriteProject, useUnfavoriteProject,
  useProjectTimeline, useProjectResources,
} from '../hooks/useProjects'
import { useComments } from '../hooks/useComments'
import MarkdownRenderer from '../components/MarkdownRenderer'
import CommentSection from '../components/CommentSection'
import TableOfContents from '../components/TableOfContents'
import AuthorCard from '../components/AuthorCard'
import { motion } from 'framer-motion'
import { useAppStore } from '../store/appStore'
import { useAuthStore } from '../store/authStore'
import type { Project } from '../types'
import dayjs from '../utils/dayjs'
import '../styles/article-content.css'

const { Title, Text, Paragraph } = Typography

// Timeline type config
const TIMELINE_CONFIG: Record<string, { color: string; emoji: string }> = {
  launch:     { color: '#34d399', emoji: '🚀' },
  version:    { color: '#60a5fa', emoji: '🏷️' },
  feature:    { color: '#f59e0b', emoji: '✨' },
  milestone:  { color: '#818cf8', emoji: '🎯' },
  breaking:   { color: '#ef4444', emoji: '⚠️' },
  archived:   { color: '#9ca3af', emoji: '📦' },
}

const PROJECT_STATE_MAP: Record<string, { label: string; color: string }> = {
  developing:  { label: '开发中', color: '#3b82f6' },
  completed:   { label: '已完成', color: '#34d399' },
  maintaining: { label: '维护中', color: '#f59e0b' },
  paused:      { label: '已暂停', color: '#f97316' },
  archived:    { label: '已归档', color: '#9ca3af' },
}

function extractHeadings(project: Project | undefined) {
  const html = project?.html_content
  if (html) {
    const doc = new DOMParser().parseFromString(html, 'text/html')
    const headings = doc.querySelectorAll('h1, h2, h3')
    const items: { id: string; title: string; level: number }[] = []
    headings.forEach((h) => {
      const title = h.textContent?.trim() || ''
      if (!title) return
      const level = parseInt(h.tagName.charAt(1))
      const id = title.toLowerCase().replace(/\s+/g, '-').replace(/[^\w一-鿿-]/g, '')
      items.push({ id, title, level })
    })
    return items
  }

  const content = project?.content
  if (!content) return []
  const mdHeadingRegex = /^(#{1,3})\s+(.+)$/gm
  const items: { id: string; title: string; level: number }[] = []
  let match: RegExpExecArray | null
  while ((match = mdHeadingRegex.exec(content)) !== null) {
    let title = match[2].trim()
    title = title.replace(/`([^`]+)`/g, '$1')
    title = title.replace(/\*\*([^*]+)\*\*/g, '$1')
    title = title.replace(/__([^_]+)__/g, '$1')
    title = title.replace(/\*([^*]+)\*/g, '$1')
    title = title.replace(/_([^_]+)_/g, '$1')
    title = title.replace(/~~([^~]+)~~/g, '$1')
    title = title.replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    title = title.trim()
    if (!title) continue
    const id = title.toLowerCase().replace(/\s+/g, '-').replace(/[^\w一-鿿-]/g, '')
    items.push({ id, title, level: match[1].length })
  }
  return items
}

export default function ProjectDetail() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const { data, isLoading, error } = useProjectBySlug(slug || '')
  const incViews = useIncProjectViews()
  const deleteProject = useDeleteProject()
  const theme = useAppStore((s) => s.theme)
  const user = useAuthStore((s) => s.user)
  const token = useAuthStore((s) => s.token)

  const project = data?.data
  const headings = useMemo(() => extractHeadings(project), [project])
  const contentRef = useRef<HTMLDivElement>(null)

  // Like
  const { data: likeData } = useProjectLikeStatus(project?.id || '')
  const likeCount = likeData?.data?.count || 0
  const liked = likeData?.data?.liked || false
  const likeMutation = useLikeProject()
  const unlikeMutation = useUnlikeProject()

  // Favorite
  const { data: favData } = useProjectFavoriteStatus(project?.id || '')
  const favorited = favData?.data || false
  const favMutation = useFavoriteProject()
  const unfavMutation = useUnfavoriteProject()

  // Timeline
  const { data: timelineData } = useProjectTimeline(project?.id || '')
  const timeline = timelineData?.data || []

  // Resources
  const { data: resourcesData } = useProjectResources(project?.id || '')
  const resources = resourcesData?.data || []

  // Comments
  const [isCommentOpen, setIsCommentOpen] = useState(false)
  const { data: commentsData } = useComments(project?.id || '')
  const commentCount = (commentsData?.data || []).length

  const [tocVisible, setTocVisible] = useState(false)

  // Collapsible sidebars
  const [isLeftCollapsed, setIsLeftCollapsed] = useState(() => {
    return localStorage.getItem('project-left-collapsed') === '1'
  })
  const [isRightCollapsed, setIsRightCollapsed] = useState(() => {
    return localStorage.getItem('project-right-collapsed') === '1'
  })
  const toggleLeft = () => {
    const next = !isLeftCollapsed
    setIsLeftCollapsed(next)
    localStorage.setItem('project-left-collapsed', next ? '1' : '0')
  }
  const toggleRight = () => {
    const next = !isRightCollapsed
    setIsRightCollapsed(next)
    localStorage.setItem('project-right-collapsed', next ? '1' : '0')
  }
  const toggleImmersive = () => {
    setIsLeftCollapsed(true)
    setIsRightCollapsed(true)
    localStorage.setItem('project-left-collapsed', '1')
    localStorage.setItem('project-right-collapsed', '1')
  }

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'b') {
        e.preventDefault()
        toggleImmersive()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])

  // Reading progress
  const [progress, setProgress] = useState(0)
  useEffect(() => {
    const onScroll = () => {
      const scrollTop = window.scrollY
      const docHeight = document.documentElement.scrollHeight - window.innerHeight
      setProgress(docHeight > 0 ? Math.min((scrollTop / docHeight) * 100, 100) : 0)
    }
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  useEffect(() => {
    if (project?.id) incViews.mutate(project.id)
  }, [project?.id])

  useEffect(() => {
    if (project) document.title = `${project.seo_title || project.title} - Blogo`
  }, [project])

  const toggleComments = () => {
    const next = !isCommentOpen
    setIsCommentOpen(next)
    if (next) {
      setTimeout(() => {
        document.getElementById('comment-section')?.scrollIntoView({ behavior: 'smooth' })
      }, 100)
    }
  }

  const isAuthor = !!(token && project?.author_id && user?.id === project.author_id)

  const handleDelete = () => {
    if (!project) return
    Modal.confirm({
      title: '确认删除', icon: <ExclamationCircleOutlined />,
      content: `确定要删除「${project.title}」吗？`,
      okText: '删除', okType: 'danger', cancelText: '取消',
      onOk: async () => {
        try { await deleteProject.mutateAsync(project.id); message.success('已删除'); navigate('/') }
        catch (err: any) { message.error(err.message || '删除失败') }
      },
    })
  }

  const handleLike = () => {
    if (!token) {
      message.warning('请登录后点赞')
      navigate(`/login?redirect=${encodeURIComponent(location.pathname)}`)
      return
    }
    if (liked) {
      unlikeMutation.mutate(project!.id, { onError: (e) => message.error(e.message) })
    } else {
      likeMutation.mutate(project!.id, { onError: (e) => message.error(e.message) })
    }
  }

  const handleFavorite = () => {
    if (!token) {
      message.warning('请登录后收藏')
      navigate(`/login?redirect=${encodeURIComponent(location.pathname)}`)
      return
    }
    if (favorited) {
      unfavMutation.mutate(project!.id, { onError: (e) => message.error(e.message) })
    } else {
      favMutation.mutate(project!.id, { onError: (e) => message.error(e.message) })
    }
  }

  const handleShare = () => {
    const url = window.location.href
    if (navigator.share) {
      navigator.share({ url }).catch(() => {})
    } else {
      navigator.clipboard.writeText(url).then(
        () => message.success('链接已复制！'),
        () => message.error('复制失败')
      )
    }
  }

  if (isLoading) return <div style={{ textAlign: 'center', padding: 120 }}><Spin size="large" /></div>
  if (!project) {
    const errCode = (error as any)?.data?.error?.code || (error as any)?.status
    if (errCode === 403) {
      return <div style={{ textAlign: 'center', padding: 120 }}><LockOutlined style={{ fontSize: 48, color: '#bbb' }} /><Title level={3}>该项目无权限访问</Title><Text type="secondary">仅指定用户可查看该项目。</Text><br /><Link to="/">返回首页</Link></div>
    }
    return <div style={{ textAlign: 'center', padding: 120 }}><Title level={3}>项目不存在</Title><Link to="/">返回首页</Link></div>
  }
  if ((project.visibility === 'private' || project.visibility === 'partial_visible') && !isAuthor) {
    return <div style={{ textAlign: 'center', padding: 120 }}><LockOutlined style={{ fontSize: 48, color: '#bbb' }} /><Title level={3}>该项目无权限访问</Title><Text type="secondary">仅指定用户可查看该项目。</Text><br /><Link to="/">返回首页</Link></div>
  }

  const isDark = theme === 'dark'
  const bg = isDark ? '#0a0a0a' : '#f8f9fb'
  const cardBg = isDark ? '#141414' : '#fff'
  const border = isDark ? '#262626' : '#e8e8e8'
  const hasCover = !!project.cover_image?.url
  const stateInfo = PROJECT_STATE_MAP[project.project_state]

  return (
    <div style={{ background: bg, minHeight: '100vh' }}>
      {/* ============ Reading Progress Bar ============ */}
      <div style={{
        position: 'fixed', top: 0, left: 0, height: 3,
        width: `${progress}%`, background: 'linear-gradient(90deg, #3b82f6, #8b5cf6)',
        zIndex: 100, transition: 'width 0.1s linear',
      }} />

      {/* ============ Breadcrumb Bar ============ */}
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, delay: 0.15, ease: 'easeOut' }}
        style={{
          position: 'sticky', top: 0, zIndex: 9,
          background: 'rgba(8,8,16,0.78)',
          backdropFilter: 'blur(16px) saturate(180%)',
          WebkitBackdropFilter: 'blur(16px) saturate(180%)',
          borderBottom: '1px solid rgba(255,255,255,0.06)',
          boxShadow: '0 2px 16px rgba(0,0,0,0.4)',
          padding: '10px 32px',
          display: 'flex', alignItems: 'center', gap: 16,
        }}
      >
        {/* Left glow indicator */}
        <div style={{
          position: 'absolute', left: 0, top: '20%', bottom: '20%',
          width: 2, borderRadius: 1,
          background: 'linear-gradient(180deg, transparent, rgba(99,102,241,0.5), rgba(139,92,246,0.3), transparent)',
        }} />

        <motion.div whileHover={{ x: -3 }} whileTap={{ scale: 0.96 }}>
          <Button
            type="text" size="small"
            icon={<ArrowLeftOutlined style={{ fontSize: 14, transition: 'transform 0.3s ease' }} />}
            onClick={() => navigate(-1)}
            className="breadcrumb-back-btn"
            style={{
              display: 'flex', alignItems: 'center', gap: 6,
              padding: '6px 14px',
              borderRadius: 10,
              fontFamily: "'Noto Serif SC', 'Barlow', sans-serif",
              fontSize: 13, fontWeight: 500,
              color: 'rgba(255,255,255,0.65)',
              background: 'rgba(255,255,255,0.05)',
              border: '1px solid rgba(255,255,255,0.06)',
              transition: 'all 0.25s ease',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'rgba(255,255,255,0.12)'
              e.currentTarget.style.color = '#ffffff'
              e.currentTarget.style.borderColor = 'rgba(255,255,255,0.15)'
              e.currentTarget.style.boxShadow = '0 0 16px rgba(99,102,241,0.15)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'rgba(255,255,255,0.05)'
              e.currentTarget.style.color = 'rgba(255,255,255,0.65)'
              e.currentTarget.style.borderColor = 'rgba(255,255,255,0.06)'
              e.currentTarget.style.boxShadow = 'none'
            }}
          >
            <span className="breadcrumb-back-text">返回</span>
          </Button>
        </motion.div>

        <nav className="breadcrumb-hide-mobile" style={{ display: 'flex', alignItems: 'center', gap: 0 }}>
          <Link to="/" style={breadcrumbLinkStyle(false)} className="breadcrumb-link"
            onMouseEnter={breadcrumbHoverIn} onMouseLeave={breadcrumbHoverOut}>
            <HomeOutlined style={{ fontSize: 14 }} />
            <span>首页</span>
          </Link>
          <BreadcrumbSep />
          <Link to="/projects" style={breadcrumbLinkStyle(false)} className="breadcrumb-link"
            onMouseEnter={breadcrumbHoverIn} onMouseLeave={breadcrumbHoverOut}>
            项目库
          </Link>
          <BreadcrumbSep />
          <span style={breadcrumbLinkStyle(true)}>
            {project.title}
          </span>
        </nav>

        <div style={{ flex: 1 }} />

        <Button
          type="text" size="small"
          icon={<ColumnWidthOutlined />}
          onClick={toggleImmersive}
          title="沉浸模式 (Ctrl+B)"
          style={{
            fontSize: 13, color: 'rgba(255,255,255,0.3)',
            borderRadius: 8, padding: '4px 8px',
            transition: 'all 0.25s ease',
          }}
          onMouseEnter={(e) => { e.currentTarget.style.color = 'rgba(255,255,255,0.6)'; e.currentTarget.style.background = 'rgba(255,255,255,0.05)' }}
          onMouseLeave={(e) => { e.currentTarget.style.color = 'rgba(255,255,255,0.3)'; e.currentTarget.style.background = 'transparent' }}
        />

        {isAuthor && (
          <>
            <Button size="small" icon={<EditOutlined />} onClick={() => navigate(`/project/${project.slug}/edit`)}
              style={{
                borderRadius: 10, fontSize: 13, fontWeight: 500,
                background: 'rgba(99,102,241,0.12)', border: '1px solid rgba(99,102,241,0.2)',
                color: '#a5b4fc', fontFamily: "'Barlow', sans-serif",
                transition: 'all 0.25s ease',
              }}
              onMouseEnter={(e) => { e.currentTarget.style.background = 'rgba(99,102,241,0.2)'; e.currentTarget.style.color = '#ffffff' }}
              onMouseLeave={(e) => { e.currentTarget.style.background = 'rgba(99,102,241,0.12)'; e.currentTarget.style.color = '#a5b4fc' }}
            >
              <span className="breadcrumb-back-text">编辑</span>
            </Button>
            <Button size="small" danger icon={<DeleteOutlined />} onClick={handleDelete} loading={deleteProject.isPending}
              style={{
                borderRadius: 10, fontSize: 13, fontWeight: 500,
                background: 'rgba(239,68,68,0.08)', border: '1px solid rgba(239,68,68,0.15)',
                fontFamily: "'Barlow', sans-serif",
              }}
              onMouseEnter={(e) => { e.currentTarget.style.background = 'rgba(239,68,68,0.18)' }}
              onMouseLeave={(e) => { e.currentTarget.style.background = 'rgba(239,68,68,0.08)' }}
            >
              <span className="breadcrumb-back-text">删除</span>
            </Button>
          </>
        )}
        {headings.length > 0 && (
          <Button className="toc-mobile-btn" type="text" size="small" icon={<MenuOutlined />}
            onClick={() => setTocVisible(true)}
            style={{ borderRadius: 10, color: 'rgba(255,255,255,0.5)', fontFamily: "'Noto Serif SC', sans-serif" }}>
            目录
          </Button>
        )}
      </motion.div>

      {/* ============ Full-width Hero Section ============ */}
      <div style={{
        position: 'relative', width: '100%',
        height: 'clamp(300px, 40vw, 480px)',
        background: hasCover
          ? `url(${project.cover_image!.url}) center/cover no-repeat`
          : 'linear-gradient(135deg, #1e3a5f 0%, #3b82f6 40%, #8b5cf6 100%)',
        display: 'flex', alignItems: 'flex-end',
        justifyContent: 'center', overflow: 'hidden',
      }}>
        {/* Overlay gradient */}
        <div style={{
          position: 'absolute', inset: 0,
          background: hasCover
            ? 'linear-gradient(to top, rgba(0,0,0,0.75) 0%, rgba(0,0,0,0.1) 60%, rgba(0,0,0,0.3) 100%)'
            : 'linear-gradient(to top, rgba(0,0,0,0.6) 0%, transparent 60%)',
        }} />
        {/* Decorative circles (no-cover only) */}
        {!hasCover && (
          <>
            <div style={{ position: 'absolute', top: -60, right: -40, width: 200, height: 200, borderRadius: '50%', background: 'rgba(255,255,255,0.06)' }} />
            <div style={{ position: 'absolute', bottom: -40, left: '10%', width: 140, height: 140, borderRadius: '50%', background: 'rgba(255,255,255,0.04)' }} />
            <div style={{ position: 'absolute', top: '30%', left: '60%', width: 100, height: 100, borderRadius: '50%', background: 'rgba(255,255,255,0.03)' }} />
          </>
        )}
        {/* Hero content overlaid at bottom */}
        <div style={{
          position: 'relative', zIndex: 1, padding: '40px 32px',
          maxWidth: 900, width: '100%',
        }}>
          <Title level={1} style={{
            color: '#fff',
            fontSize: 'clamp(26px, 5vw, 44px)',
            fontWeight: 800, lineHeight: 1.25,
            marginBottom: 12,
            letterSpacing: '-0.5px',
            fontFamily: "'Instrument Serif', 'Noto Serif SC', serif",
            fontStyle: 'italic',
            textShadow: '0 2px 8px rgba(0,0,0,0.5)',
          }}>
            {project.title}
          </Title>

          {/* State badge */}
          {stateInfo && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
              <span style={{
                display: 'inline-flex', alignItems: 'center', gap: 6,
                padding: '4px 14px', borderRadius: 20,
                background: `${stateInfo.color}22`,
                border: `1px solid ${stateInfo.color}44`,
                fontSize: 13, fontWeight: 500,
                color: stateInfo.color,
              }}>
                <span style={{
                  width: 8, height: 8, borderRadius: '50%',
                  background: stateInfo.color,
                  boxShadow: `0 0 6px ${stateInfo.color}`,
                }} />
                {stateInfo.label}
              </span>
            </div>
          )}

          {/* Summary */}
          {project.summary && (
            <Paragraph style={{
              color: 'rgba(255,255,255,0.85)', fontSize: 15,
              marginBottom: 16, lineHeight: 1.6,
              maxWidth: 700,
            }}>
              {project.summary}
            </Paragraph>
          )}

          {/* Tech stack tags */}
          {project.tags && project.tags.length > 0 && (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 20 }}>
              {project.tags.map((tag) => (
                <Link key={tag.id} to={`/tags?tag_name=${encodeURIComponent(tag.name)}`}>
                  <span style={{
                    display: 'inline-block',
                    padding: '4px 16px', borderRadius: 20,
                    background: 'rgba(255,255,255,0.08)',
                    border: '1px solid rgba(255,255,255,0.15)',
                    backdropFilter: 'blur(8px)',
                    color: 'rgba(255,255,255,0.9)',
                    fontSize: 13, fontWeight: 500,
                    transition: 'all 0.25s ease',
                    cursor: 'pointer',
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.background = 'rgba(255,255,255,0.15)'
                    e.currentTarget.style.borderColor = 'rgba(255,255,255,0.25)'
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = 'rgba(255,255,255,0.08)'
                    e.currentTarget.style.borderColor = 'rgba(255,255,255,0.15)'
                  }}
                  >
                    {tag.name}
                  </span>
                </Link>
              ))}
            </div>
          )}

          {/* Action buttons row */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap', marginBottom: 16 }}>
            {project.github_url && (
              <a href={project.github_url} target="_blank" rel="noopener noreferrer"
                style={{
                  display: 'inline-flex', alignItems: 'center', gap: 8,
                  background: 'rgba(255,255,255,0.06)',
                  border: '1px solid rgba(255,255,255,0.12)',
                  backdropFilter: 'blur(12px)',
                  borderRadius: 12,
                  padding: '10px 24px',
                  color: '#fff', fontWeight: 500,
                  fontSize: 14, textDecoration: 'none',
                  transition: 'all 0.25s ease',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'rgba(255,255,255,0.1)'
                  e.currentTarget.style.borderColor = 'rgba(255,255,255,0.2)'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'rgba(255,255,255,0.06)'
                  e.currentTarget.style.borderColor = 'rgba(255,255,255,0.12)'
                }}
              >
                <GithubOutlined /> GitHub
              </a>
            )}
            {project.demo_url && (
              <a href={project.demo_url} target="_blank" rel="noopener noreferrer"
                style={{
                  display: 'inline-flex', alignItems: 'center', gap: 8,
                  background: 'rgba(255,255,255,0.06)',
                  border: '1px solid rgba(255,255,255,0.12)',
                  backdropFilter: 'blur(12px)',
                  borderRadius: 12,
                  padding: '10px 24px',
                  color: '#fff', fontWeight: 500,
                  fontSize: 14, textDecoration: 'none',
                  transition: 'all 0.25s ease',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'rgba(255,255,255,0.1)'
                  e.currentTarget.style.borderColor = 'rgba(255,255,255,0.2)'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'rgba(255,255,255,0.06)'
                  e.currentTarget.style.borderColor = 'rgba(255,255,255,0.12)'
                }}
              >
                <LinkOutlined /> Demo
              </a>
            )}
            <button
              onClick={handleLike}
              style={{
                display: 'inline-flex', alignItems: 'center', gap: 6,
                background: 'rgba(255,255,255,0.06)',
                border: '1px solid rgba(255,255,255,0.12)',
                backdropFilter: 'blur(12px)',
                borderRadius: 12,
                padding: '10px 20px',
                color: liked ? '#f43f5e' : '#fff',
                fontWeight: 500, fontSize: 14,
                cursor: 'pointer',
                transition: 'all 0.25s ease',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'rgba(255,255,255,0.1)'
                e.currentTarget.style.borderColor = 'rgba(255,255,255,0.2)'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'rgba(255,255,255,0.06)'
                e.currentTarget.style.borderColor = 'rgba(255,255,255,0.12)'
              }}
            >
              {liked ? <HeartFilled /> : <HeartOutlined />} {likeCount > 0 ? likeCount : ''}
            </button>
            <button
              onClick={handleFavorite}
              style={{
                display: 'inline-flex', alignItems: 'center', gap: 6,
                background: 'rgba(255,255,255,0.06)',
                border: '1px solid rgba(255,255,255,0.12)',
                backdropFilter: 'blur(12px)',
                borderRadius: 12,
                padding: '10px 20px',
                color: favorited ? '#f59e0b' : '#fff',
                fontWeight: 500, fontSize: 14,
                cursor: 'pointer',
                transition: 'all 0.25s ease',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'rgba(255,255,255,0.1)'
                e.currentTarget.style.borderColor = 'rgba(255,255,255,0.2)'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'rgba(255,255,255,0.06)'
                e.currentTarget.style.borderColor = 'rgba(255,255,255,0.12)'
              }}
            >
              {favorited ? <StarFilled /> : <StarOutlined />}
            </button>
            <button
              onClick={handleShare}
              style={{
                display: 'inline-flex', alignItems: 'center', gap: 6,
                background: 'rgba(255,255,255,0.06)',
                border: '1px solid rgba(255,255,255,0.12)',
                backdropFilter: 'blur(12px)',
                borderRadius: 12,
                padding: '10px 20px',
                color: '#fff',
                fontWeight: 500, fontSize: 14,
                cursor: 'pointer',
                transition: 'all 0.25s ease',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'rgba(255,255,255,0.1)'
                e.currentTarget.style.borderColor = 'rgba(255,255,255,0.2)'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'rgba(255,255,255,0.06)'
                e.currentTarget.style.borderColor = 'rgba(255,255,255,0.12)'
              }}
            >
              <ShareAltOutlined />
            </button>
          </div>

          {/* Stats row */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
            <Text style={{ color: 'rgba(255,255,255,0.75)', fontSize: 13 }}>
              <EyeOutlined /> {project.views || 0}
            </Text>
            <Text style={{ color: 'rgba(255,255,255,0.75)', fontSize: 13 }}>
              <HeartOutlined /> {likeCount}
            </Text>
            <Text style={{ color: 'rgba(255,255,255,0.75)', fontSize: 13 }}>
              <StarOutlined /> {project.favorite_count || 0}
            </Text>
            <Text style={{ color: 'rgba(255,255,255,0.75)', fontSize: 13 }}>
              <CommentOutlined /> {project.comment_count || 0}
            </Text>
          </div>
        </div>
      </div>

      {/* ============ Three Column Layout ============ */}
      <div className={`detail-grid ${isLeftCollapsed ? 'left-collapsed' : ''} ${isRightCollapsed ? 'right-collapsed' : ''} ${isLeftCollapsed && isRightCollapsed ? 'immersive' : ''}`}
        style={{ width: '100%', padding: isLeftCollapsed && isRightCollapsed ? '32px 0' : '32px 0', margin: 0 }}>
        {/* Left TOC */}
        <aside className={`detail-toc ${isLeftCollapsed ? 'collapsed' : ''}`}>
          {!isLeftCollapsed && (
            <Affix offsetTop={96}>
              <div style={{ position: 'relative' }}>
                <Button
                  type="text" size="small"
                  icon={<DoubleLeftOutlined />}
                  onClick={toggleLeft}
                  style={{ position: 'absolute', top: 0, right: 8, zIndex: 2, opacity: 0.4, fontSize: 12 }}
                  title="收起目录"
                />
                <TableOfContents key={project?.content?.slice(0, 100)} items={headings} contentRef={contentRef} />
              </div>
            </Affix>
          )}
        </aside>

        {/* Main Content */}
        <main className="detail-main" ref={contentRef}>
          <Card bordered={false} style={{
            borderRadius: isLeftCollapsed && isRightCollapsed ? 0 : 20,
            boxShadow: isLeftCollapsed && isRightCollapsed ? 'none' : '0 4px 24px rgba(0,0,0,0.06)',
            marginBottom: 32, overflow: 'visible',
          }} styles={{ body: { padding: isLeftCollapsed && isRightCollapsed
            ? '40px clamp(16px, 8vw, 120px)'
            : '40px clamp(24px, 5vw, 64px)' } }}>
            {/* Tags at top */}
            {project.tags && project.tags.length > 0 && (
              <div style={{ marginBottom: 24 }}>
                {project.tags.map((tag) => (
                  <Link key={tag.id} to={`/tags?tag_name=${encodeURIComponent(tag.name)}`}>
                    <Tag style={{ borderRadius: 8, padding: '2px 12px', fontSize: 12, marginBottom: 6 }}>{tag.name}</Tag>
                  </Link>
                ))}
              </div>
            )}
            {/* Markdown content */}
            <div className="article-body-enhanced">
              {project.html_content ? (
                <div dangerouslySetInnerHTML={{ __html: project.html_content }} />
              ) : (
                <MarkdownRenderer content={project.content} />
              )}
            </div>
          </Card>

          {/* Project Timeline */}
          {timeline.length > 0 && (
            <Card bordered={false} style={{
              borderRadius: 20,
              boxShadow: '0 4px 24px rgba(0,0,0,0.06)',
              marginBottom: 32, overflow: 'visible',
            }} styles={{ body: { padding: '32px clamp(24px, 5vw, 64px)' } }}>
              <Title level={4} style={{ marginBottom: 28, fontWeight: 700, letterSpacing: '-0.3px' }}>
                项目历程
              </Title>
              <div style={{ position: 'relative', paddingLeft: 32 }}>
                {/* Vertical line */}
                <div style={{
                  position: 'absolute', left: 7, top: 4, bottom: 4,
                  width: 2, borderRadius: 1,
                  background: 'linear-gradient(to bottom, #3b82f6, #8b5cf6, #3b82f6)',
                  opacity: 0.3,
                }} />
                {timeline.sort((a, b) => a.sort_order - b.sort_order).map((item) => {
                  const cfg = TIMELINE_CONFIG[item.type] || TIMELINE_CONFIG.milestone
                  return (
                    <div key={item.id} style={{ position: 'relative', marginBottom: 32 }}>
                      {/* Colored circle node */}
                      <div style={{
                        position: 'absolute', left: -32, top: 2,
                        width: 16, height: 16, borderRadius: '50%',
                        background: cfg.color,
                        boxShadow: `0 0 10px ${cfg.color}66`,
                        border: `2px solid ${isDark ? '#141414' : '#fff'}`,
                        zIndex: 1,
                      }} />
                      <div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 6 }}>
                          <span style={{ fontSize: 16 }}>{cfg.emoji}</span>
                          <Text style={{ fontSize: 13, color: 'rgba(255,255,255,0.5)', fontFamily: "'Barlow', monospace" }}>
                            {dayjs(item.event_date).format('YYYY-MM-DD')}
                          </Text>
                          {item.version && (
                            <Tag color={cfg.color} style={{ borderRadius: 10, fontSize: 11, padding: '0 10px', margin: 0 }}>
                              {item.version}
                            </Tag>
                          )}
                        </div>
                        <Text strong style={{ fontSize: 15, display: 'block', marginBottom: 4 }}>
                          {item.title}
                        </Text>
                        {item.description && (
                          <Text style={{ fontSize: 13, color: isDark ? '#999' : '#666', lineHeight: 1.7 }}>
                            {item.description}
                          </Text>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            </Card>
          )}

          {/* Project Resources */}
          {resources.length > 0 && (
            <Card bordered={false} style={{
              borderRadius: 20,
              boxShadow: '0 4px 24px rgba(0,0,0,0.06)',
              marginBottom: 32, overflow: 'visible',
            }} styles={{ body: { padding: '32px clamp(24px, 5vw, 64px)' } }}>
              <Title level={4} style={{ marginBottom: 20, fontWeight: 700, letterSpacing: '-0.3px' }}>
                相关资源
              </Title>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                {resources.sort((a, b) => a.sort_order - b.sort_order).map((res) => (
                  <a key={res.id} href={res.url} target="_blank" rel="noopener noreferrer"
                    style={{
                      display: 'flex', alignItems: 'center', gap: 12,
                      padding: '12px 16px', borderRadius: 12,
                      background: isDark ? 'rgba(255,255,255,0.03)' : 'rgba(0,0,0,0.02)',
                      border: `1px solid ${isDark ? 'rgba(255,255,255,0.06)' : 'rgba(0,0,0,0.06)'}`,
                      textDecoration: 'none',
                      transition: 'all 0.25s ease',
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.background = isDark ? 'rgba(255,255,255,0.06)' : 'rgba(0,0,0,0.04)'
                      e.currentTarget.style.transform = 'translateX(4px)'
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.background = isDark ? 'rgba(255,255,255,0.03)' : 'rgba(0,0,0,0.02)'
                      e.currentTarget.style.transform = 'translateX(0)'
                    }}
                  >
                    <LinkOutlined style={{ color: '#3b82f6', fontSize: 14 }} />
                    <div>
                      <Text style={{ fontSize: 14, fontWeight: 500, display: 'block' }}>{res.title}</Text>
                      <Text style={{ fontSize: 12, color: isDark ? '#666' : '#999' }}>{res.type}</Text>
                    </div>
                  </a>
                ))}
              </div>
            </Card>
          )}

          {/* PostActions (like, favorite, comment toggle) */}
          <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 32 }}>
            <PostActionsInline
              liked={liked}
              likeCount={likeCount}
              favorited={favorited}
              commentCount={commentCount}
              onLike={handleLike}
              onFavorite={handleFavorite}
              onComment={toggleComments}
              onShare={handleShare}
              isDark={isDark}
            />
          </div>

          {/* Comments */}
          <div id="comment-section">
            <CommentSection articleId={project.id} visible={isCommentOpen} />
          </div>
        </main>

        {/* Right Sidebar */}
        <aside className={`detail-sidebar ${isRightCollapsed ? 'collapsed' : ''}`}>
          {!isRightCollapsed && (
            <div style={{ position: 'relative' }}>
              <Button
                type="text" size="small"
                icon={<DoubleRightOutlined />}
                onClick={toggleRight}
                style={{ position: 'absolute', top: 0, left: 8, zIndex: 2, opacity: 0.4, fontSize: 12 }}
                title="收起侧栏"
              />
              <div style={{ position: 'sticky', top: 80, paddingTop: 0 }}>
                {/* Author Card */}
                {project.author && (
                  <div style={{ background: cardBg, borderRadius: 18, padding: '24px 18px', marginBottom: 20, boxShadow: '0 2px 12px rgba(0,0,0,0.06)', transition: 'all 0.3s ease' }} className="card-hover">
                    <AuthorCard author={project.author} articleCount={0} articleSlug={project.slug} />
                  </div>
                )}

                {/* Project Info Card */}
                <div style={{ background: cardBg, borderRadius: 14, padding: '16px 18px', marginBottom: 16, boxShadow: '0 1px 8px rgba(0,0,0,0.04)' }}>
                  <Text strong style={{ fontSize: 12, letterSpacing: 1.2, color: '#999', display: 'block', marginBottom: 12 }}>PROJECT INFO</Text>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 10, fontSize: 13 }}>
                    <Row label="浏览" value={`${project.views || 0}`} />
                    <Row label="点赞" value={`${likeCount}`} />
                    <Row label="收藏" value={`${project.favorite_count || 0}`} />
                    <Row label="评论" value={`${project.comment_count || 0}`} />
                    <Row label="创建" value={dayjs(project.created_at).format('YYYY-MM-DD')} />
                    <Row label="更新" value={dayjs(project.updated_at).format('YYYY-MM-DD')} />
                    <Row label="分类" value={project.category?.name || '-'} />
                    <Row label="状态" value={stateInfo?.label || '-'} />
                  </div>
                </div>

                {/* Related Resources (show resource list if any) */}
                {resources.length > 0 && (
                  <div style={{ background: cardBg, borderRadius: 14, padding: '16px 18px', boxShadow: '0 1px 8px rgba(0,0,0,0.04)' }}>
                    <Text strong style={{ fontSize: 12, letterSpacing: 1.2, color: '#999', display: 'block', marginBottom: 14 }}>RESOURCES</Text>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                      {resources.map((res) => (
                        <a key={res.id} href={res.url} target="_blank" rel="noopener noreferrer"
                          style={{ display: 'flex', alignItems: 'center', gap: 8, textDecoration: 'none', padding: '4px 0' }}>
                          <LinkOutlined style={{ fontSize: 12, color: '#3b82f6', flexShrink: 0 }} />
                          <Text style={{
                            fontSize: 13, lineHeight: 1.5,
                            color: isDark ? '#bbb' : '#444',
                            display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden',
                          } as React.CSSProperties}>{res.title}</Text>
                        </a>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </aside>
      </div>

      {/* ============ Fixed Floating Toggles (collapsed only) ============ */}
      {isLeftCollapsed && (
        <div
          onClick={toggleLeft}
          className="floating-toggle floating-toggle-left"
          style={{
            position: 'fixed', left: 0, top: '50%', transform: 'translateY(-50%)',
            width: 48, height: 180, zIndex: 50, cursor: 'pointer',
            background: isDark ? 'rgba(30,30,40,0.92)' : 'rgba(255,255,255,0.92)',
            backdropFilter: 'blur(12px)', borderRadius: '0 16px 16px 0',
            boxShadow: isDark ? '2px 0 20px rgba(0,0,0,0.5)' : '2px 0 20px rgba(0,0,0,0.1)',
            display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
            gap: 8, transition: 'width 0.25s ease, background 0.25s ease',
            border: `1px solid ${isDark ? '#333' : '#e8e8e8'}`,
            borderLeft: 'none',
          }}
          onMouseEnter={(e) => { e.currentTarget.style.width = '64px'; e.currentTarget.style.background = isDark ? 'rgba(40,40,55,0.95)' : 'rgba(255,255,255,0.98)' }}
          onMouseLeave={(e) => { e.currentTarget.style.width = '48px'; e.currentTarget.style.background = isDark ? 'rgba(30,30,40,0.92)' : 'rgba(255,255,255,0.92)' }}
        >
          <DoubleRightOutlined style={{ fontSize: 14, color: isDark ? '#aaa' : '#999' }} />
          <span style={{ writingMode: 'vertical-rl', textOrientation: 'mixed', fontSize: 11, color: isDark ? '#999' : '#888', letterSpacing: 3, userSelect: 'none' }}>目录</span>
        </div>
      )}
      {isRightCollapsed && (
        <div
          onClick={toggleRight}
          className="floating-toggle floating-toggle-right"
          style={{
            position: 'fixed', right: 0, top: '50%', transform: 'translateY(-50%)',
            width: 48, height: 180, zIndex: 50, cursor: 'pointer',
            background: isDark ? 'rgba(30,30,40,0.92)' : 'rgba(255,255,255,0.92)',
            backdropFilter: 'blur(12px)', borderRadius: '16px 0 0 16px',
            boxShadow: isDark ? '-2px 0 20px rgba(0,0,0,0.5)' : '-2px 0 20px rgba(0,0,0,0.1)',
            display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
            gap: 8, transition: 'width 0.25s ease, background 0.25s ease',
            border: `1px solid ${isDark ? '#333' : '#e8e8e8'}`,
            borderRight: 'none',
          }}
          onMouseEnter={(e) => { e.currentTarget.style.width = '64px'; e.currentTarget.style.background = isDark ? 'rgba(40,40,55,0.95)' : 'rgba(255,255,255,0.98)' }}
          onMouseLeave={(e) => { e.currentTarget.style.width = '48px'; e.currentTarget.style.background = isDark ? 'rgba(30,30,40,0.92)' : 'rgba(255,255,255,0.92)' }}
        >
          <DoubleLeftOutlined style={{ fontSize: 14, color: isDark ? '#aaa' : '#999' }} />
          <span style={{ writingMode: 'vertical-rl', textOrientation: 'mixed', fontSize: 11, color: isDark ? '#999' : '#888', letterSpacing: 3, userSelect: 'none' }}>侧栏</span>
        </div>
      )}

      {/* Mobile TOC Drawer */}
      <Drawer title="目录" placement="right" open={tocVisible} onClose={() => setTocVisible(false)} styles={{ body: { padding: 0 } }}>
        <TableOfContents items={headings} contentRef={contentRef} />
      </Drawer>

      {/* ============ Styles ============ */}
      <style>{`
        .detail-grid {
          display: flex; flex-direction: row; gap: 0;
          align-items: flex-start; width: 100%;
        }
        .detail-grid:not(.left-collapsed) .detail-toc {
          margin-right: 24px;
        }
        .detail-grid:not(.right-collapsed) .detail-sidebar {
          margin-left: 24px;
        }
        .detail-toc { width: 200px; flex-shrink: 0; transition: width 0.35s cubic-bezier(0.4,0,0.2,1); overflow: hidden; }
        .detail-toc.collapsed { width: 0px; }
        .detail-main { flex: 1; min-width: 0; transition: max-width 0.35s cubic-bezier(0.4,0,0.2,1), padding 0.35s cubic-bezier(0.4,0,0.2,1); }
        .detail-sidebar { width: 270px; flex-shrink: 0; transition: width 0.35s cubic-bezier(0.4,0,0.2,1); overflow: hidden; }
        .detail-sidebar.collapsed { width: 0px; }
        .toc-mobile-btn { display: none !important; }
        .card-hover:hover { transform: translateY(-2px); box-shadow: 0 6px 20px rgba(0,0,0,0.1) !important; }
        .collapse-btn:hover { opacity: 1 !important; }

        .detail-grid.immersive { gap: 0; }
        .detail-grid.immersive .detail-main { max-width: none !important; }

        .article-body-enhanced {
          font-size: 17px; line-height: 1.9; color: ${isDark ? '#d4d4d4' : '#333'}; word-break: break-word;
        }
        .article-body-enhanced h1 { font-size: 2em; font-weight: 800; margin: 1.8em 0 0.8em; line-height: 1.3; }
        .article-body-enhanced h2 { font-size: 1.5em; font-weight: 700; margin: 1.6em 0 0.7em; padding-left: 14px; border-left: 4px solid #3b82f6; line-height: 1.3; color: ${isDark ? '#e5e5e5' : '#1a1a1a'}; }
        .article-body-enhanced h3 { font-size: 1.2em; font-weight: 600; margin: 1.3em 0 0.5em; color: ${isDark ? '#ddd' : '#262626'}; }
        .article-body-enhanced p { margin: 0 0 1.1em; }
        .article-body-enhanced a { color: #3b82f6; text-decoration: none; border-bottom: 1px solid rgba(59,130,246,0.3); }
        .article-body-enhanced a:hover { border-bottom-color: #3b82f6; }
        .article-body-enhanced blockquote {
          margin: 1.2em 0; padding: 16px 22px; border-left: 4px solid #3b82f6;
          background: ${isDark ? '#1a1a2e' : '#f0f5ff'}; border-radius: 0 10px 10px 0; color: ${isDark ? '#bbb' : '#555'};
        }
        .article-body-enhanced code {
          background: ${isDark ? '#2d2d3f' : '#f0f0f0'};
          color: ${isDark ? '#f472b6' : '#d63384'}; padding: 2px 7px; border-radius: 5px; font-size: 0.88em;
          font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
        }
        .article-body-enhanced pre {
          margin: 1.3em 0; border-radius: 12px; overflow: hidden;
          box-shadow: 0 2px 12px rgba(0,0,0,0.08); position: relative;
        }
        .article-body-enhanced pre code {
          display: block; background: #1e1e2e; color: #cdd6f4; padding: 20px 24px;
          font-size: 14px; line-height: 1.7; border-radius: 12px; overflow-x: auto;
        }
        .article-body-enhanced table {
          width: 100%; margin: 1.2em 0; border-collapse: collapse; border-radius: 10px;
          overflow: hidden; box-shadow: 0 1px 6px rgba(0,0,0,0.05);
        }
        .article-body-enhanced thead { background: ${isDark ? '#1e1e2e' : '#f6f8fa'}; }
        .article-body-enhanced th { padding: 12px 16px; text-align: left; font-weight: 600; font-size: 14px; border-bottom: 2px solid ${isDark ? '#333' : '#e8e8e8'}; }
        .article-body-enhanced td { padding: 10px 16px; font-size: 14px; border-bottom: 1px solid ${isDark ? '#262626' : '#f0f0f0'}; }
        .article-body-enhanced tbody tr:hover { background: ${isDark ? '#1a1a24' : '#fafbfc'}; }
        .article-body-enhanced img { max-width: 100%; border-radius: 12px; box-shadow: 0 3px 12px rgba(0,0,0,0.1); margin: 1.2em auto; display: block; cursor: pointer; transition: transform 0.3s; }
        .article-body-enhanced img:hover { transform: scale(1.02); }
        .article-body-enhanced ul, .article-body-enhanced ol { padding-left: 1.8em; margin: 0 0 1em; }
        .article-body-enhanced li { margin-bottom: 0.4em; line-height: 1.75; }
        .article-body-enhanced hr { border: none; height: 1px; background: linear-gradient(to right, transparent, ${isDark ? '#333' : '#d9d9d9'}, transparent); margin: 2.2em 0; }

        @media (max-width: 1400px) { .detail-sidebar { display: none !important; } }
        @media (max-width: 1000px) {
          .detail-toc { display: none !important; }
          .toc-mobile-btn { display: inline-flex !important; }
          .detail-grid { flex-direction: column !important; }
          .detail-main { width: 100%; }
        }
        @media (max-width: 768px) {
          .detail-grid { padding: 16px 12px !important; }
          .ant-card-body { padding: 24px 18px !important; }
          .breadcrumb-hide-mobile { display: none; }
        }
      `}</style>
    </div>
  )
}

// ── Inline PostActions for project (uses project-specific hooks) ──
function PostActionsInline({ liked, likeCount, favorited, commentCount, onLike, onFavorite, onComment, onShare, isDark }: {
  liked: boolean; likeCount: number; favorited: boolean; commentCount: number;
  onLike: () => void; onFavorite: () => void; onComment: () => void; onShare: () => void; isDark: boolean;
}) {
  return (
    <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
      <Button
        type="text"
        icon={liked ? <HeartFilled style={{ color: '#f43f5e' }} /> : <HeartOutlined />}
        onClick={onLike}
      >
        {likeCount > 0 ? likeCount : '点赞'}
      </Button>
      <Button
        type="text"
        icon={favorited ? <StarFilled style={{ color: '#f59e0b' }} /> : <StarOutlined />}
        onClick={onFavorite}
      >
        {favorited ? '已收藏' : '收藏'}
      </Button>
      <Button type="text" icon={<CommentOutlined />} onClick={onComment}>
        {commentCount > 0 ? commentCount : '评论'}
      </Button>
      <Button type="text" icon={<ShareAltOutlined />} onClick={onShare} />
    </div>
  )
}

// ── Breadcrumb helpers ──
function BreadcrumbSep() {
  return (
    <span style={{
      color: 'rgba(255,255,255,0.18)', fontSize: 14, fontWeight: 300,
      margin: '0 10px', userSelect: 'none',
    }}>
      /
    </span>
  )
}

function breadcrumbLinkStyle(active: boolean): React.CSSProperties {
  return {
    display: 'flex', alignItems: 'center', gap: 6,
    padding: '4px 10px', borderRadius: 8,
    fontFamily: "'Noto Serif SC', 'Barlow', sans-serif",
    fontSize: 13, fontWeight: active ? 500 : 400,
    color: active ? 'rgba(255,255,255,0.9)' : 'rgba(255,255,255,0.5)',
    textDecoration: 'none', whiteSpace: 'nowrap',
    maxWidth: active ? 240 : undefined,
    overflow: 'hidden', textOverflow: 'ellipsis',
    background: active ? 'rgba(255,255,255,0.08)' : 'transparent',
    transition: 'all 0.25s ease',
  }
}

function breadcrumbHoverIn(e: React.MouseEvent<HTMLAnchorElement>) {
  e.currentTarget.style.background = 'rgba(255,255,255,0.1)'
  e.currentTarget.style.color = 'rgba(255,255,255,0.9)'
}

function breadcrumbHoverOut(e: React.MouseEvent<HTMLAnchorElement>) {
  e.currentTarget.style.background = 'transparent'
  e.currentTarget.style.color = 'rgba(255,255,255,0.5)'
}

// ── Metadata row ──
function Row({ label, value }: { label: string; value: string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
      <Text type="secondary" style={{ fontSize: 13 }}>{label}</Text>
      <Text style={{ fontSize: 13 }}>{value}</Text>
    </div>
  )
}

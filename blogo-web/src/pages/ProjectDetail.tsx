import { useEffect, useRef } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { Button, Spin, Tag, Typography } from 'antd'
import {
  ArrowLeftOutlined, BulbOutlined, EditOutlined, EyeOutlined,
  GithubOutlined, LinkOutlined,
} from '@ant-design/icons'
import { useIncProjectViews, useProjectBySlug } from '../hooks/useProjects'
import { useAuthStore } from '../store/authStore'
import dayjs from '../utils/dayjs'

const { Text, Title } = Typography

const c = {
  surface: 'rgba(255,255,255,0.03)',
  border: 'rgba(255,255,255,0.08)',
  text: 'rgba(255,255,255,0.75)',
  textMuted: 'rgba(255,255,255,0.45)',
  accent: '#4f6ef7',
}

const STATE_LABEL: Record<string, string> = {
  developing: '开发中', completed: '已完成', maintaining: '维护中', paused: '暂停', archived: '已归档',
}
const STATE_COLOR: Record<string, string> = {
  developing: '#4f6ef7', completed: '#22c55e', maintaining: '#f59e0b', paused: '#94a3b8', archived: '#64748b',
}

/**
 * 项目详情（展示型）。
 * 只展示：封面 + 名称 + 简介 + 技术栈 + 项目特点 + 仓库/演示链接。
 * 不含评论区、项目历程、项目资源、点赞与收藏（这些不属于展示定位）。
 */
export default function ProjectDetail() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const { data, isLoading } = useProjectBySlug(slug || '')
  const project = data?.data
  const user = useAuthStore((s) => s.user)
  const incViews = useIncProjectViews()
  const viewed = useRef(false)

  // 浏览量只上报一次
  useEffect(() => {
    if (project?.id && !viewed.current) {
      viewed.current = true
      incViews.mutate(project.id)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [project?.id])

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 120 }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!project) {
    return (
      <div style={{ textAlign: 'center', padding: 120 }}>
        <Title level={4} style={{ color: '#fff' }}>项目不存在或未公开</Title>
        <Link to="/projects"><Button type="link">返回项目列表</Button></Link>
      </div>
    )
  }

  const highlights = (project.highlights || '')
    .split('\n')
    .map((line) => line.replace(/^[-·•\s]+/, '').trim())
    .filter(Boolean)
  const tags = project.tags || []
  const isAuthor = !!user && user.id === project.author_id
  const stateLabel = STATE_LABEL[project.project_state] || project.project_state
  const stateColor = STATE_COLOR[project.project_state] || '#8c8c8c'

  return (
    <div style={{ maxWidth: 900, margin: '0 auto', paddingBottom: 80 }}>
      {/* ── 顶部操作 ── */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
        <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)} style={{ color: c.textMuted }}>
          返回
        </Button>
        {isAuthor && (
          <Button icon={<EditOutlined />} onClick={() => navigate(`/project/${project.slug}/edit`)}>
            编辑项目
          </Button>
        )}
      </div>

      {/* ── 封面 ── */}
      {project.cover_image?.url && (
        <div style={{
          borderRadius: 18, overflow: 'hidden', marginBottom: 28,
          border: `1px solid ${c.border}`, background: c.surface,
        }}>
          <img
            src={project.cover_image.url}
            alt={project.title}
            style={{ width: '100%', maxHeight: 380, objectFit: 'cover', display: 'block' }}
          />
        </div>
      )}

      {/* ── 标题区 ── */}
      <div style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap', marginBottom: 12 }}>
          <Title level={2} style={{ margin: 0, color: '#fff' }}>{project.title}</Title>
          <Tag style={{
            marginInlineEnd: 0, borderRadius: 999, fontSize: 12,
            color: stateColor, background: `${stateColor}18`, border: `1px solid ${stateColor}40`,
          }}>
            {stateLabel}
          </Tag>
        </div>

        {project.summary && (
          <Text style={{ display: 'block', color: c.text, fontSize: 16, lineHeight: 1.8, marginBottom: 14 }}>
            {project.summary}
          </Text>
        )}

        <div style={{ display: 'flex', alignItems: 'center', gap: 16, color: c.textMuted, fontSize: 13, flexWrap: 'wrap' }}>
          {project.author?.name && <span>{project.author.name}</span>}
          {project.published_at && <span>{dayjs(project.published_at).format('YYYY-MM-DD')}</span>}
          <span><EyeOutlined /> {project.views || 0}</span>
        </div>
      </div>

      {/* ── 技术栈 ── */}
      {tags.length > 0 && (
        <div style={{ marginBottom: 24 }}>
          <div style={{ color: c.textMuted, fontSize: 13, marginBottom: 10 }}>技术栈</div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {tags.map((tag) => (
              <Tag key={tag.id} style={{
                marginInlineEnd: 0, borderRadius: 999, fontSize: 12, padding: '3px 12px',
                color: c.accent, background: 'rgba(79,110,247,0.12)', border: '1px solid rgba(79,110,247,0.25)',
              }}>
                {tag.name}
              </Tag>
            ))}
          </div>
        </div>
      )}

      {/* ── 链接 ── */}
      {(project.github_url || project.demo_url) && (
        <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', marginBottom: 32 }}>
          {project.github_url && (
            <a
              href={project.github_url}
              target="_blank"
              rel="noopener noreferrer"
              className="liquid-glass-card"
              style={{
                display: 'inline-flex', alignItems: 'center', gap: 8, padding: '10px 20px',
                textDecoration: 'none', fontSize: 14, fontWeight: 500, color: c.text,
                border: `1px solid ${c.border}`, borderRadius: 10,
              }}
            >
              <GithubOutlined style={{ fontSize: 17 }} /> 仓库地址
            </a>
          )}
          {project.demo_url && (
            <a
              href={project.demo_url}
              target="_blank"
              rel="noopener noreferrer"
              className="liquid-glass-card"
              style={{
                display: 'inline-flex', alignItems: 'center', gap: 8, padding: '10px 20px',
                textDecoration: 'none', fontSize: 14, fontWeight: 500, color: c.text,
                border: `1px solid ${c.border}`, borderRadius: 10,
              }}
            >
              <LinkOutlined style={{ fontSize: 17 }} /> 在线演示
            </a>
          )}
        </div>
      )}

      {/* ── 项目特点 ── */}
      {highlights.length > 0 && (
        <div style={{ background: c.surface, border: `1px solid ${c.border}`, borderRadius: 16, padding: '24px 28px' }}>
          <div style={{ color: c.textMuted, fontSize: 13, marginBottom: 14 }}>
            <BulbOutlined /> 项目特点
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {highlights.map((item, index) => (
              <div key={index} style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
                <span style={{
                  flex: 'none', width: 6, height: 6, borderRadius: '50%',
                  background: c.accent, marginTop: 9,
                }} />
                <Text style={{ color: c.text, fontSize: 15, lineHeight: 1.8 }}>{item}</Text>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

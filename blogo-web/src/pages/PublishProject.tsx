import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Button, Card, Form, Input, Select, Space, Spin, Typography, message } from 'antd'
import {
  ArrowLeftOutlined, BulbOutlined, DeleteOutlined, GithubOutlined, LinkOutlined,
  PictureOutlined, PlusOutlined, SaveOutlined, SendOutlined, TagsOutlined,
} from '@ant-design/icons'
import { useCreateProject, useProjectBySlug, useUpdateProject } from '../hooks/useProjects'
import { useCategories } from '../hooks/useCategories'
import { useTags } from '../hooks/useTags'
import { useAuthStore } from '../store/authStore'
import CoverImageUpload from '../components/CoverImageUpload'
import type { ProjectForm } from '../types'

const { Text, Title } = Typography
const { TextArea } = Input

// ── 与项目其它页面一致的暗色 token ──
const c = {
  surface: 'rgba(255,255,255,0.03)',
  border: 'rgba(255,255,255,0.08)',
  text: 'rgba(255,255,255,0.85)',
  textMuted: 'rgba(255,255,255,0.45)',
  accent: '#4f6ef7',
}

const PROJECT_STATES = [
  { value: 'developing', label: '开发中' },
  { value: 'completed', label: '已完成' },
  { value: 'maintaining', label: '维护中' },
  { value: 'paused', label: '暂停' },
  { value: 'archived', label: '已归档' },
]

const cardStyle = { background: c.surface, border: `1px solid ${c.border}`, borderRadius: 14 }
const labelStyle = { color: c.textMuted, fontSize: 13 }

function slugify(text: string): string {
  const slug = text
    .toLowerCase()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_]+/g, '-')
    .replace(/-+/g, '-')
    .trim()
    .replace(/^-+|-+$/g, '')
  return slug || `project-${Date.now()}`
}

// 项目特点：一行一条，允许用户输入时带 · / - / • 前缀
function parseHighlights(text?: string): string[] {
  const lines = (text || '')
    .split('\n')
    .map((line) => line.replace(/^[-·•\s]+/, '').trim())
    .filter(Boolean)
  return lines.length ? lines : ['']
}

/**
 * 发布项目（展示型）。
 * 与文章发布不同：项目只展示「封面 + 简介 + 技术栈 + 项目特点 + 仓库/演示链接」，
 * 不需要长正文、可见性设置与 SEO 表单（SEO 由标题/摘要/技术栈自动生成）。
 */
export default function PublishProject() {
  const { slug } = useParams<{ slug?: string }>()
  const isEdit = !!slug
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [highlights, setHighlights] = useState<string[]>([''])
  const [submitting, setSubmitting] = useState(false)

  const token = useAuthStore((s) => s.token)
  const user = useAuthStore((s) => s.user)

  const createProject = useCreateProject()
  const updateProject = useUpdateProject()
  const { data: projectData, isLoading: loadingProject } = useProjectBySlug(slug || '')
  const { data: catData } = useCategories()
  const { data: tagData } = useTags()

  const project = projectData?.data
  const categories = catData?.data || []
  const tags = tagData?.data || []

  // 未登录跳登录页
  useEffect(() => {
    if (!token) {
      navigate(`/login?redirect=${encodeURIComponent(window.location.pathname)}`, { replace: true })
    }
  }, [token, navigate])

  // 编辑回显
  useEffect(() => {
    if (!project || !isEdit) return
    if (user && project.author_id !== user.id) {
      message.error('只能编辑自己的项目')
      navigate('/', { replace: true })
      return
    }
    form.setFieldsValue({
      title: project.title,
      slug: project.slug,
      summary: project.summary,
      category_id: project.category_id || undefined,
      project_state: project.project_state || 'developing',
      github_url: project.github_url || '',
      demo_url: project.demo_url || '',
      cover_image_id: project.cover_image_id || undefined,
      tag_ids: project.tags?.map((t: any) => t.id) || [],
    })
    setHighlights(parseHighlights(project.highlights))
  }, [project, isEdit, user, form, navigate])

  const updateHighlight = (index: number, value: string) => {
    setHighlights((prev) => prev.map((item, i) => (i === index ? value : item)))
  }

  const handleSubmit = async (status: 'draft' | 'published') => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)

      const tagNames = (values.tag_ids || []).map(
        (id: string) => tags.find((t: any) => t.id === id)?.name || '',
      ).filter(Boolean)

      const payload: ProjectForm = {
        title: values.title,
        slug: values.slug || slugify(values.title),
        summary: values.summary || '',
        highlights: highlights.map((h) => h.trim()).filter(Boolean).join('\n'),
        // 展示型项目不写长正文；编辑时保留原有内容，避免误清空
        content: project?.content || '',
        cover_image_id: values.cover_image_id || '',
        category_id: values.category_id || '',
        tag_ids: values.tag_ids || [],
        project_state: values.project_state || 'developing',
        github_url: values.github_url || '',
        demo_url: values.demo_url || '',
        status,
        visibility: 'public',
        is_top: false,
        is_featured: false,
        featured_order: 0,
        seo_title: values.title,
        seo_keywords: tagNames.join(','),
        seo_desc: values.summary || '',
      }

      if (isEdit && project) {
        await updateProject.mutateAsync({ id: project.id, data: payload })
        message.success(status === 'draft' ? '草稿已保存' : '项目已更新')
      } else {
        await createProject.mutateAsync(payload)
        message.success(status === 'draft' ? '草稿已保存' : '项目已发布')
      }
      navigate(`/project/${payload.slug}`)
    } catch (err: any) {
      if (err?.errorFields) return // 表单校验失败，antd 已在字段上提示
      message.error(err?.message || '保存失败，请稍后重试')
    } finally {
      setSubmitting(false)
    }
  }

  if (isEdit && loadingProject) {
    return (
      <div style={{ textAlign: 'center', padding: 120 }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div style={{ maxWidth: 860, margin: '0 auto', paddingBottom: 80 }}>
      {/* ── 顶部：标题 + 操作 ── */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 16, marginBottom: 24, flexWrap: 'wrap' }}>
        <Space size={12}>
          <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)} style={{ color: c.textMuted }} />
          <Title level={3} style={{ margin: 0, color: '#fff' }}>
            {isEdit ? '编辑项目' : '发布项目'}
          </Title>
          <Text style={labelStyle}>展示你的项目：特点 · 技术栈 · 链接</Text>
        </Space>
        <Space>
          <Button icon={<SaveOutlined />} loading={submitting} onClick={() => handleSubmit('draft')}>存草稿</Button>
          <Button type="primary" icon={<SendOutlined />} loading={submitting} onClick={() => handleSubmit('published')}>
            {isEdit ? '保存并发布' : '发布项目'}
          </Button>
        </Space>
      </div>

      <Form
        form={form}
        layout="vertical"
        initialValues={{ project_state: 'developing' }}
      >
        {/* ── 基本信息 ── */}
        <Card style={{ ...cardStyle, marginBottom: 16 }} styles={{ body: { padding: 24 } }}>
          <div style={{ marginBottom: 8, color: c.textMuted, fontSize: 13 }}>
            <PictureOutlined /> 封面图
          </div>
          <Form.Item name="cover_image_id" style={{ marginBottom: 20 }}>
            <CoverImageUpload coverUrl={project?.cover_image?.url} />
          </Form.Item>

          <Form.Item name="title" label={<span style={labelStyle}>项目名称</span>} rules={[{ required: true, message: '请输入项目名称' }]}>
            <Input placeholder="例如：Blogo 博客系统" size="large" />
          </Form.Item>

          <Form.Item
            name="slug"
            label={<span style={labelStyle}>访问路径（Slug）</span>}
            extra={<span style={{ color: c.textMuted, fontSize: 12 }}>留空则按项目名称自动生成</span>}
          >
            <Input addonBefore="/project/" placeholder="blogo" />
          </Form.Item>

          <Form.Item name="summary" label={<span style={labelStyle}>一句话简介</span>}>
            <TextArea rows={2} maxLength={200} showCount placeholder="一句话说清这个项目是做什么的" />
          </Form.Item>

          <Space size={16} style={{ display: 'flex', flexWrap: 'wrap' }} align="start">
            <Form.Item name="category_id" label={<span style={labelStyle}>分类</span>} style={{ marginBottom: 0, minWidth: 220 }}>
              <Select allowClear placeholder="选择分类" options={categories.map((item: any) => ({ value: item.id, label: item.name }))} />
            </Form.Item>
            <Form.Item name="project_state" label={<span style={labelStyle}>项目状态</span>} style={{ marginBottom: 0, minWidth: 220 }}>
              <Select options={PROJECT_STATES} />
            </Form.Item>
          </Space>
        </Card>

        {/* ── 技术栈 ── */}
        <Card style={{ ...cardStyle, marginBottom: 16 }} styles={{ body: { padding: 24 } }}>
          <Form.Item
            name="tag_ids"
            label={<span style={labelStyle}><TagsOutlined /> 技术栈</span>}
            extra={<span style={{ color: c.textMuted, fontSize: 12 }}>从已有标签里选，或直接输入新的技术名后回车</span>}
            style={{ marginBottom: 0 }}
          >
            <Select
              mode="tags"
              placeholder="例如：Go、Gin、Redis、React"
              options={tags.map((item: any) => ({ value: item.id, label: item.name }))}
              tokenSeparators={[',', '，']}
            />
          </Form.Item>
        </Card>

        {/* ── 项目特点 ── */}
        <Card style={{ ...cardStyle, marginBottom: 16 }} styles={{ body: { padding: 24 } }}>
          <div style={{ marginBottom: 12, color: c.textMuted, fontSize: 13 }}>
            <BulbOutlined /> 项目特点（一条一个要点，建议 3~5 条）
          </div>
          <Space direction="vertical" style={{ width: '100%' }} size={10}>
            {highlights.map((item, index) => (
              <div key={index} style={{ display: 'flex', gap: 8 }}>
                <Input
                  value={item}
                  onChange={(e) => updateHighlight(index, e.target.value)}
                  placeholder={index === 0 ? '例如：15 个微服务 + 网关，覆盖完整交易闭环' : '继续添加一条特点'}
                />
                <Button
                  type="text"
                  danger
                  icon={<DeleteOutlined />}
                  disabled={highlights.length === 1}
                  onClick={() => setHighlights((prev) => prev.filter((_, i) => i !== index))}
                />
              </div>
            ))}
          </Space>
          <Button
            type="dashed"
            icon={<PlusOutlined />}
            style={{ marginTop: 12, width: '100%' }}
            onClick={() => setHighlights((prev) => [...prev, ''])}
          >
            添加一条特点
          </Button>
        </Card>

        {/* ── 链接 ── */}
        <Card style={{ ...cardStyle, marginBottom: 24 }} styles={{ body: { padding: 24 } }}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Form.Item name="github_url" label={<span style={labelStyle}><GithubOutlined /> 仓库地址</span>} style={{ marginBottom: 0 }}>
              <Input placeholder="https://github.com/yourname/project" />
            </Form.Item>
            <Form.Item name="demo_url" label={<span style={labelStyle}><LinkOutlined /> 在线演示地址</span>} style={{ marginBottom: 0 }}>
              <Input placeholder="https://demo.example.com" />
            </Form.Item>
          </Space>
        </Card>

        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 12 }}>
          <Button icon={<SaveOutlined />} loading={submitting} onClick={() => handleSubmit('draft')}>存草稿</Button>
          <Button type="primary" icon={<SendOutlined />} loading={submitting} onClick={() => handleSubmit('published')}>
            {isEdit ? '保存并发布' : '发布项目'}
          </Button>
        </div>
      </Form>
    </div>
  )
}

import { useEffect, useState, useMemo, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Form, Input, Select, Switch, Button, Card, Space, message, Typography, Spin, Tag, Segmented, InputNumber, DatePicker, List, Collapse, Popconfirm } from 'antd'
import { TeamOutlined, GlobalOutlined, LockOutlined, SendOutlined, SaveOutlined, PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { useGetProjectQuery, useCreateProjectMutation, useUpdateProjectMutation } from '../../store/api'
import { useGetCategoriesQuery } from '../../store/api'
import { useGetTagsQuery } from '../../store/api'
import { useGetUsersQuery } from '../../store/api'
import { useGetProjectTimelineQuery, useCreateTimelineEntryMutation, useUpdateTimelineEntryMutation, useDeleteTimelineEntryMutation } from '../../store/api'
import MarkdownEditor from '../../components/MarkdownEditor'
import type { User, ProjectTimeline } from '../../types'
import dayjs from 'dayjs'

const { Title } = Typography
const { TextArea } = Input

function slugify(text: string): string {
  const slug = text.toLowerCase().replace(/[^\w\s-]/g, '').replace(/[\s_]+/g, '-').replace(/-+/g, '-').replace(/^-+|-+$/g, '')
  if (!slug) return `project-${Date.now()}`
  return slug
}

interface SelectedUser {
  id: string
  name: string
  username: string
}

const PROJECT_STATE_OPTIONS = [
  { label: '开发中', value: 'developing' },
  { label: '已完成', value: 'completed' },
  { label: '维护中', value: 'maintaining' },
  { label: '暂停开发', value: 'paused' },
  { label: '停止维护', value: 'archived' },
]

const TIMELINE_TYPE_OPTIONS = [
  { label: '🚀 首次发布', value: 'launch' },
  { label: '🏷️ 版本发布', value: 'version' },
  { label: '✨ 新功能', value: 'feature' },
  { label: '🎯 里程碑', value: 'milestone' },
  { label: '⚠️ 重大变更', value: 'breaking' },
  { label: '📦 归档', value: 'archived' },
]

export default function ProjectEdit() {
  const { id } = useParams<{ id: string }>()
  const isEdit = !!id
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [timelineForm] = Form.useForm()
  const [content, setContent] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [visibility, setVisibility] = useState<string>('public')
  const [selectedUsers, setSelectedUsers] = useState<SelectedUser[]>([])
  const [userSearch, setUserSearch] = useState('')
  const [showTimelineForm, setShowTimelineForm] = useState(false)
  const [editingTimeline, setEditingTimeline] = useState<ProjectTimeline | null>(null)

  const { data: projectData, isLoading } = useGetProjectQuery(id!, { skip: !isEdit })
  const { data: catData } = useGetCategoriesQuery({ current: 1, pageSize: 100 })
  const { data: tagData } = useGetTagsQuery({ current: 1, pageSize: 100 })
  const { data: userData } = useGetUsersQuery({ current: 1, pageSize: 1000 })
  const [createProject] = useCreateProjectMutation()
  const [updateProject] = useUpdateProjectMutation()

  const { data: timelineData } = useGetProjectTimelineQuery(id!, { skip: !isEdit })
  const [createTimelineEntry] = useCreateTimelineEntryMutation()
  const [updateTimelineEntry] = useUpdateTimelineEntryMutation()
  const [deleteTimelineEntry] = useDeleteTimelineEntryMutation()

  const project = projectData?.data
  const categories = catData?.data || []
  const tags = tagData?.data || []
  const users = (userData?.data || []) as User[]
  const timeline = timelineData?.data || []

  const isFeatured = Form.useWatch('is_featured', form)

  // ── 编辑回显 ──
  useEffect(() => {
    if (project && isEdit) {
      form.setFieldsValue({
        ...project,
        tag_ids: project.tags?.map((t: any) => t.id) || [],
      })
      setContent(project.content || '')
      setVisibility(project.visibility || 'public')
      if (project.visible_users) {
        const vuUsers: SelectedUser[] = []
        for (const vu of project.visible_users as any[]) {
          const u = users.find((usr) => usr.id === vu.user_id)
          vuUsers.push({ id: vu.user_id, name: u?.name || '', username: u?.username || '' })
        }
        setSelectedUsers(vuUsers)
        form.setFieldsValue({ visible_user_ids: vuUsers.map((u) => u.id) })
      }
    }
  }, [project, isEdit, form, users])

  // ── 去重可用用户 ──
  const filteredUsers = useMemo(() => {
    const selectedIds = new Set(selectedUsers.map((u) => u.id))
    let list = users.filter((u) => !selectedIds.has(u.id))
    if (userSearch) {
      const kw = userSearch.toLowerCase()
      list = list.filter((u) =>
        u.username?.toLowerCase().includes(kw) ||
        u.name?.toLowerCase().includes(kw)
      )
    }
    return list.slice(0, 50)
  }, [users, selectedUsers, userSearch])

  // ── Slug 自动生成 ──
  const handleTitleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const title = e.target.value
    const currentSlug = form.getFieldValue('slug')
    if (!currentSlug || currentSlug === slugify(form.getFieldValue('_prevTitle') || '')) {
      form.setFieldsValue({ slug: slugify(title), _prevTitle: title })
    } else {
      form.setFieldsValue({ _prevTitle: title })
    }
  }, [form])

  const handleContentChange = useCallback((markdown?: string) => {
    setContent(markdown || '')
    form.setFieldsValue({ content: markdown || '' })
  }, [form])

  // ── 切换可见性时清理脏数据 ──
  const handleVisibilityChange = useCallback((val: string | number) => {
    const v = String(val)
    setVisibility(v)
    if (v !== 'partial_visible') {
      setSelectedUsers([])
      setUserSearch('')
      form.setFieldsValue({ visible_user_ids: [] })
    }
  }, [form])

  // ── 用户选择 / 移除 ──
  const handleUserSelect = useCallback((userId: string) => {
    if (selectedUsers.find((u) => u.id === userId)) return
    const user = users.find((u) => u.id === userId)
    if (!user) return
    const next = [...selectedUsers, { id: user.id, name: user.name, username: user.username }]
    setSelectedUsers(next)
    setUserSearch('')
    form.setFieldsValue({ visible_user_ids: next.map((u) => u.id) })
  }, [selectedUsers, users, form])

  const handleRemoveUser = useCallback((userId: string) => {
    const next = selectedUsers.filter((u) => u.id !== userId)
    setSelectedUsers(next)
    form.setFieldsValue({ visible_user_ids: next.map((u) => u.id) })
  }, [selectedUsers, form])

  // ── 提交 ──
  const handleSubmit = useCallback((status: 'draft' | 'published') => {
    form.validateFields().then(async (values) => {
      setSubmitting(true)
      try {
        const payload = {
          ...values,
          content: content || values.content || '',
          status,
          visibility: values.visibility || 'public',
          visible_user_ids: visibility === 'partial_visible' ? (values.visible_user_ids || []) : [],
          is_featured: values.is_featured || false,
          featured_order: values.featured_order || 0,
        }
        delete (payload as any)._prevTitle
        if (isEdit) {
          await updateProject({ id: id!, body: payload }).unwrap()
          message.success('更新成功')
        } else {
          await createProject(payload).unwrap()
          message.success('创建成功')
        }
        navigate('/projects')
      } catch (err: any) {
        message.error(err.data?.error?.detail || err.message || '操作失败')
      } finally {
        setSubmitting(false)
      }
    })
  }, [form, content, visibility, isEdit, id, createProject, updateProject, navigate])

  // ── Timeline 操作 ──
  const handleAddTimeline = useCallback(() => {
    setEditingTimeline(null)
    timelineForm.resetFields()
    setShowTimelineForm(true)
  }, [timelineForm])

  const handleEditTimeline = useCallback((entry: ProjectTimeline) => {
    setEditingTimeline(entry)
    timelineForm.setFieldsValue({
      ...entry,
      event_date: entry.event_date ? dayjs(entry.event_date) : undefined,
    })
    setShowTimelineForm(true)
  }, [timelineForm])

  const handleCancelTimeline = useCallback(() => {
    setShowTimelineForm(false)
    setEditingTimeline(null)
    timelineForm.resetFields()
  }, [timelineForm])

  const handleSaveTimeline = useCallback(async () => {
    try {
      const values = await timelineForm.validateFields()
      const payload = {
        ...values,
        event_date: values.event_date ? values.event_date.format('YYYY-MM-DD') : '',
      }
      if (editingTimeline) {
        await updateTimelineEntry({ projectId: id!, tid: editingTimeline.id, body: payload }).unwrap()
        message.success('里程碑更新成功')
      } else {
        await createTimelineEntry({ projectId: id!, body: payload }).unwrap()
        message.success('里程碑添加成功')
      }
      setShowTimelineForm(false)
      setEditingTimeline(null)
      timelineForm.resetFields()
    } catch (err: any) {
      if (err.errorFields) return
      message.error(err.data?.error?.detail || err.message || '操作失败')
    }
  }, [timelineForm, editingTimeline, id, createTimelineEntry, updateTimelineEntry])

  const handleDeleteTimeline = useCallback(async (tid: string) => {
    try {
      await deleteTimelineEntry({ projectId: id!, tid }).unwrap()
      message.success('里程碑已删除')
    } catch (err: any) {
      message.error(err.data?.error?.detail || err.message || '删除失败')
    }
  }, [id, deleteTimelineEntry])

  if (isEdit && isLoading) return <div style={{ textAlign: 'center', padding: 80 }}><Spin size="large" /></div>

  return (
    <div>
      <Title level={4}>{isEdit ? '编辑项目' : '新建项目'}</Title>
      <Card>
        <Form form={form} layout="vertical" initialValues={{ status: 'draft', is_top: false, is_featured: false, visibility: 'public', project_state: 'developing', content: '' }}>
          <Form.Item name="title" label="标题" rules={[{ required: true, max: 255, message: '请输入标题' }]}>
            <Input placeholder="项目标题" onChange={handleTitleChange} style={{ fontSize: 18, fontWeight: 600 }} />
          </Form.Item>

          <Form.Item name="slug" label="Slug"
            rules={[{ required: true, pattern: /^[a-z0-9-]+$/, message: '仅支持小写字母、数字和连字符' }]}>
            <Input placeholder="项目-url-别名" addonBefore="/project/" />
          </Form.Item>

          <Form.Item name="summary" label="摘要">
            <TextArea rows={2} maxLength={1000} showCount placeholder="项目摘要（可选）" />
          </Form.Item>

          {/* ── 状态 + 可见性 + 置顶 ── */}
          <Space size="large" wrap style={{ marginBottom: 16, width: '100%' }}>
            <Form.Item name="status" label="状态" rules={[{ required: true }]} style={{ marginBottom: 0 }}>
              <Select style={{ width: 130 }}
                options={[{ label: '草稿', value: 'draft' }, { label: '已发布', value: 'published' }]} />
            </Form.Item>

            <div>
              <div style={{ marginBottom: 8, fontSize: 14, color: 'rgba(0,0,0,0.88)', fontWeight: 400 }}>可见性</div>
              <Form.Item name="visibility" style={{ marginBottom: 0 }}>
                <Segmented
                  onChange={(val) => handleVisibilityChange(val)}
                  options={[
                    { label: <span><GlobalOutlined /> 公开</span>, value: 'public' },
                    { label: <span><LockOutlined /> 私密</span>, value: 'private' },
                    { label: <span><TeamOutlined /> 部分可见</span>, value: 'partial_visible' },
                  ]}
                />
              </Form.Item>
            </div>

            <Form.Item name="is_top" label="置顶" valuePropName="checked" style={{ marginBottom: 0 }}>
              <Switch />
            </Form.Item>
          </Space>

          {/* ── 部分人可见 —— 用户选择器 ── */}
          {visibility === 'partial_visible' && (
            <div style={{
              marginBottom: 16, padding: '16px 20px',
              background: 'rgba(79,110,247,0.03)', borderRadius: 12,
              border: '1px solid rgba(79,110,247,0.12)',
            }}>
              <div style={{ marginBottom: 8, fontWeight: 500, color: '#4f6ef7', fontSize: 13 }}>
                <TeamOutlined /> 可见用户 · {selectedUsers.length} 人已选
              </div>

              {selectedUsers.length > 0 && (
                <div style={{ marginBottom: 10 }}>
                  {selectedUsers.map((u) => (
                    <Tag
                      key={u.id}
                      closable
                      onClose={() => handleRemoveUser(u.id)}
                      color="blue"
                      style={{ marginBottom: 4, borderRadius: 8, padding: '2px 10px' }}
                    >
                      {u.name || u.username}
                    </Tag>
                  ))}
                </div>
              )}

              <Select
                showSearch
                placeholder="搜索用户名或昵称..."
                style={{ width: 360 }}
                filterOption={false}
                onSearch={(val) => setUserSearch(val)}
                onSelect={(val) => val && handleUserSelect(val)}
                value={undefined}
                notFoundContent={userSearch ? '未找到相关用户' : '输入关键词搜索'}
                options={filteredUsers.map((u) => ({
                  label: `${u.name || ''}  @${u.username}`,
                  value: u.id,
                }))}
              />

              <Form.Item name="visible_user_ids" style={{ display: 'none' }}>
                <Input />
              </Form.Item>
            </div>
          )}

          {/* ── 项目特有字段 ── */}
          <Space size="large" wrap style={{ marginBottom: 16, width: '100%' }}>
            <Form.Item name="project_state" label="项目状态" style={{ marginBottom: 0 }}>
              <Select style={{ width: 160 }} options={PROJECT_STATE_OPTIONS} />
            </Form.Item>

            <Form.Item name="is_featured" label="精选项目" valuePropName="checked" style={{ marginBottom: 0 }}>
              <Switch />
            </Form.Item>

            {isFeatured && (
              <Form.Item name="featured_order" label="精选排序" style={{ marginBottom: 0 }}>
                <InputNumber min={0} style={{ width: 120 }} placeholder="0" />
              </Form.Item>
            )}
          </Space>

          <Form.Item name="github_url" label="GitHub 链接" style={{ marginBottom: 16 }}>
            <Input addonBefore="https://github.com/" placeholder="username/repo" />
          </Form.Item>

          <Form.Item name="demo_url" label="在线演示" style={{ marginBottom: 16 }}>
            <Input placeholder="https://demo.example.com" />
          </Form.Item>

          {/* ── 分类 + 标签 ── */}
          <Space size="large" wrap style={{ marginBottom: 16 }}>
            <Form.Item name="category_id" label="分类" rules={[{ required: true, message: '请选择分类' }]} style={{ marginBottom: 0 }}>
              <Select style={{ width: 200 }} placeholder="选择分类"
                options={categories.map((c) => ({ label: c.name, value: c.id }))}
                showSearch optionFilterProp="label" />
            </Form.Item>
            <Form.Item name="tag_ids" label="标签" style={{ marginBottom: 0 }}>
              <Select mode="tags" style={{ minWidth: 280 }} placeholder="输入标签名后按回车"
                options={tags.map((t) => ({ label: t.name, value: t.id }))} />
            </Form.Item>
          </Space>

          <Form.Item name="cover_image_id" label="封面图片ID" style={{ marginBottom: 16 }}>
            <Input placeholder="留空则不设置封面" />
          </Form.Item>

          <Form.Item name="content" label="内容" rules={[{ required: true, message: '请输入项目内容' }]} style={{ display: 'none' }}>
            <Input />
          </Form.Item>

          <div style={{ marginBottom: 24 }}>
            <MarkdownEditor value={content} onChange={handleContentChange} height={500} />
          </div>

          {/* ── SEO 字段 ── */}
          <Collapse
            ghost
            items={[{
              key: 'seo',
              label: <span style={{ fontWeight: 500 }}>SEO 设置</span>,
              children: (
                <>
                  <Form.Item name="seo_title" label="SEO 标题">
                    <Input placeholder="留空则使用项目标题" maxLength={255} />
                  </Form.Item>
                  <Form.Item name="seo_keywords" label="SEO 关键词">
                    <Input placeholder="多个关键词用逗号分隔" maxLength={500} />
                  </Form.Item>
                  <Form.Item name="seo_desc" label="SEO 描述">
                    <TextArea rows={2} maxLength={500} showCount placeholder="留空则使用摘要" />
                  </Form.Item>
                </>
              ),
            }]}
            style={{ marginBottom: 24 }}
          />

          {/* ── 项目历程 ── */}
          <div style={{ marginBottom: 24 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
              <Typography.Text strong style={{ fontSize: 16 }}>项目历程</Typography.Text>
              {isEdit && (
                <Button type="dashed" icon={<PlusOutlined />} onClick={handleAddTimeline}>
                  添加里程碑
                </Button>
              )}
            </div>

            {!isEdit && (
              <div style={{ color: 'rgba(0,0,0,0.45)', fontSize: 13 }}>
                请先保存项目后再添加里程碑
              </div>
            )}

            {isEdit && showTimelineForm && (
              <Card size="small" style={{ marginBottom: 16, background: 'rgba(79,110,247,0.03)', border: '1px solid rgba(79,110,247,0.12)' }}>
                <Form form={timelineForm} layout="vertical" size="small">
                  <Space size="middle" wrap style={{ width: '100%', marginBottom: 0 }}>
                    <Form.Item name="event_date" label="日期" rules={[{ required: true, message: '请选择日期' }]} style={{ marginBottom: 8 }}>
                      <DatePicker style={{ width: 150 }} />
                    </Form.Item>
                    <Form.Item name="type" label="类型" rules={[{ required: true, message: '请选择类型' }]} style={{ marginBottom: 8 }}>
                      <Select style={{ width: 160 }} options={TIMELINE_TYPE_OPTIONS} />
                    </Form.Item>
                    <Form.Item name="version" label="版本" style={{ marginBottom: 8 }}>
                      <Input placeholder="v1.0.0" style={{ width: 120 }} />
                    </Form.Item>
                  </Space>
                  <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]} style={{ marginBottom: 8 }}>
                    <Input placeholder="里程碑标题" />
                  </Form.Item>
                  <Form.Item name="description" label="描述" style={{ marginBottom: 8 }}>
                    <TextArea rows={2} placeholder="详细描述（可选）" />
                  </Form.Item>
                  <Form.Item name="link" label="链接" style={{ marginBottom: 8 }}>
                    <Input placeholder="相关链接（可选）" />
                  </Form.Item>
                  <Space>
                    <Button type="primary" size="small" onClick={handleSaveTimeline}>
                      {editingTimeline ? '更新' : '添加'}
                    </Button>
                    <Button size="small" onClick={handleCancelTimeline}>取消</Button>
                  </Space>
                </Form>
              </Card>
            )}

            {isEdit && timeline.length > 0 && (
              <List
                size="small"
                dataSource={timeline}
                renderItem={(entry: ProjectTimeline) => (
                  <List.Item
                    actions={[
                      <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEditTimeline(entry)}>编辑</Button>,
                      <Popconfirm title="确定删除此里程碑？" onConfirm={() => handleDeleteTimeline(entry.id)} okText="删除" cancelText="取消">
                        <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
                      </Popconfirm>,
                    ]}
                  >
                    <List.Item.Meta
                      title={
                        <span>
                          {TIMELINE_TYPE_OPTIONS.find((o) => o.value === entry.type)?.label || entry.type}
                          {' · '}
                          {entry.title}
                          {entry.version && <Tag style={{ marginLeft: 8 }}>{entry.version}</Tag>}
                        </span>
                      }
                      description={
                        <span>
                          {entry.event_date && <span style={{ marginRight: 12 }}>{entry.event_date}</span>}
                          {entry.description}
                        </span>
                      }
                    />
                  </List.Item>
                )}
              />
            )}

            {isEdit && timeline.length === 0 && !showTimelineForm && (
              <div style={{ color: 'rgba(0,0,0,0.45)', fontSize: 13, padding: '12px 0' }}>
                暂无里程碑记录
              </div>
            )}
          </div>

          <div style={{ marginTop: 24, display: 'flex', justifyContent: 'flex-end', gap: 12 }}>
            <Button onClick={() => navigate('/projects')}>取消</Button>
            <Button size="large" icon={<SaveOutlined />} onClick={() => handleSubmit('draft')} loading={submitting}>
              保存草稿
            </Button>
            <Button type="primary" size="large" icon={<SendOutlined />} onClick={() => handleSubmit('published')} loading={submitting} style={{ minWidth: 120 }}>
              发布项目
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  )
}

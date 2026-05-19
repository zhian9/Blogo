import { useEffect, useState, useMemo, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Form, Input, Select, Switch, Button, Card, Space, message, Typography, Spin, Tag, Segmented } from 'antd'
import { TeamOutlined, GlobalOutlined, LockOutlined, SendOutlined, SaveOutlined } from '@ant-design/icons'
import { useGetArticleQuery, useCreateArticleMutation, useUpdateArticleMutation } from '../../store/api'
import { useGetCategoriesQuery } from '../../store/api'
import { useGetTagsQuery } from '../../store/api'
import { useGetUsersQuery } from '../../store/api'
import MarkdownEditor from '../../components/MarkdownEditor'
import type { User } from '../../types'

const { Title } = Typography
const { TextArea } = Input

function slugify(text: string): string {
  const slug = text.toLowerCase().replace(/[^\w\s-]/g, '').replace(/[\s_]+/g, '-').replace(/-+/g, '-').replace(/^-+|-+$/g, '')
  if (!slug) return `post-${Date.now()}`
  return slug
}

interface SelectedUser {
  id: string
  name: string
  username: string
}

export default function ArticleEdit() {
  const { id } = useParams<{ id: string }>()
  const isEdit = !!id
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [content, setContent] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [visibility, setVisibility] = useState<string>('public')
  const [selectedUsers, setSelectedUsers] = useState<SelectedUser[]>([])
  const [userSearch, setUserSearch] = useState('')

  const { data: articleData, isLoading } = useGetArticleQuery(id!, { skip: !isEdit })
  const { data: catData } = useGetCategoriesQuery({ current: 1, pageSize: 100 })
  const { data: tagData } = useGetTagsQuery({ current: 1, pageSize: 100 })
  const { data: userData } = useGetUsersQuery({ current: 1, pageSize: 1000 })
  const [createArticle] = useCreateArticleMutation()
  const [updateArticle] = useUpdateArticleMutation()

  const article = articleData?.data
  const categories = catData?.data || []
  const tags = tagData?.data || []
  const users = (userData?.data || []) as User[]

  // ── 编辑回显 ──
  useEffect(() => {
    if (article && isEdit) {
      form.setFieldsValue({
        ...article,
        tag_ids: article.tags?.map((t: any) => t.id) || [],
      })
      setContent(article.content || '')
      setVisibility(article.visibility || 'public')
      if (article.visible_users) {
        const vuUsers: SelectedUser[] = []
        for (const vu of article.visible_users as any[]) {
          const u = users.find((usr) => usr.id === vu.user_id)
          vuUsers.push({ id: vu.user_id, name: u?.name || '', username: u?.username || '' })
        }
        setSelectedUsers(vuUsers)
        form.setFieldsValue({ visible_user_ids: vuUsers.map((u) => u.id) })
      }
    }
  }, [article, isEdit, form, users])

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
        }
        delete (payload as any)._prevTitle
        if (isEdit) {
          await updateArticle({ id: id!, body: payload }).unwrap()
          message.success('更新成功')
        } else {
          await createArticle(payload).unwrap()
          message.success('创建成功')
        }
        navigate('/articles')
      } catch (err: any) {
        message.error(err.data?.error?.detail || err.message || '操作失败')
      } finally {
        setSubmitting(false)
      }
    })
  }, [form, content, visibility, isEdit, id, createArticle, updateArticle, navigate])

  if (isEdit && isLoading) return <div style={{ textAlign: 'center', padding: 80 }}><Spin size="large" /></div>

  return (
    <div>
      <Title level={4}>{isEdit ? '编辑文章' : '新建文章'}</Title>
      <Card>
        <Form form={form} layout="vertical" initialValues={{ status: 'draft', is_top: false, visibility: 'public', content: '' }}>
          <Form.Item name="title" label="标题" rules={[{ required: true, max: 255, message: '请输入标题' }]}>
            <Input placeholder="文章标题" onChange={handleTitleChange} style={{ fontSize: 18, fontWeight: 600 }} />
          </Form.Item>

          <Form.Item name="slug" label="Slug"
            rules={[{ required: true, pattern: /^[a-z0-9-]+$/, message: '仅支持小写字母、数字和连字符' }]}>
            <Input placeholder="文章-url-别名" addonBefore="/article/" />
          </Form.Item>

          <Form.Item name="summary" label="摘要">
            <TextArea rows={2} maxLength={1000} showCount placeholder="文章摘要（可选）" />
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
                onSelect={(val: string) => handleUserSelect(val)}
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

          <Form.Item name="content" label="内容" rules={[{ required: true, message: '请输入文章内容' }]} style={{ display: 'none' }}>
            <Input />
          </Form.Item>

          <div style={{ marginBottom: 24 }}>
            <MarkdownEditor value={content} onChange={handleContentChange} height={500} />
          </div>

          <div style={{ marginTop: 24, display: 'flex', justifyContent: 'flex-end', gap: 12 }}>
            <Button onClick={() => navigate('/articles')}>取消</Button>
            <Button size="large" icon={<SaveOutlined />} onClick={() => handleSubmit('draft')} loading={submitting}>
              保存草稿
            </Button>
            <Button type="primary" size="large" icon={<SendOutlined />} onClick={() => handleSubmit('published')} loading={submitting} style={{ minWidth: 120 }}>
              发布
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  )
}

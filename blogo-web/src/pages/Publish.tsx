import { useState, useEffect, useMemo, useCallback, useRef } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Form, Input, Select, Button, Space, message, Typography, Collapse, Switch, Spin, Affix, Card, Segmented, Tag } from 'antd'
import {
  SendOutlined, SaveOutlined, EditOutlined, ArrowLeftOutlined,
  AppstoreOutlined, TagsOutlined, LinkOutlined, FileTextOutlined, PictureOutlined, SearchOutlined,
  GlobalOutlined, LockOutlined, TeamOutlined,
} from '@ant-design/icons'
import { useCreateArticle, useUpdateArticle, useArticleBySlug } from '../hooks/useArticles'
import { useCategories } from '../hooks/useCategories'
import { useTags } from '../hooks/useTags'
import { useUserSearch } from '../hooks/useUsers'
import { useAuthStore } from '../store/authStore'
import MarkdownEditor from '../components/MarkdownEditor'
import CoverImageUpload from '../components/CoverImageUpload'
import type { ArticleForm } from '../types'

const { Text } = Typography
const { TextArea } = Input

function roleNameToCode(name: string): string | undefined {
  const map: Record<string, string> = {
    '管理员': 'admin',
    '超级管理员': 'admin',
    '内容管理员': 'content_manager',
    '评论审核员': 'comment_moderator',
    '用户': 'user',
    '游客': 'guest',
  }
  return map[name]
}

function slugify(text: string): string {
  const slug = text
    .toLowerCase()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_]+/g, '-')
    .replace(/-+/g, '-')
    .trim()
    .replace(/^-+|-+$/g, '')
  if (!slug) return `post-${Date.now()}`
  return slug
}

// ── Dark theme tokens ──
const c = {
  bg: '#0a0a10',
  surface: 'rgba(255,255,255,0.03)',
  surfaceAlt: 'rgba(255,255,255,0.06)',
  border: 'rgba(255,255,255,0.08)',
  text: 'rgba(255,255,255,0.85)',
  textMuted: 'rgba(255,255,255,0.45)',
  accent: '#4f6ef7',
  accentBg: 'rgba(79,110,247,0.12)',
  accentBorder: 'rgba(79,110,247,0.25)',
}

interface SelectedUser {
  id: string
  username: string
  name: string
}

export default function Publish() {
  const { slug } = useParams<{ slug?: string }>()
  const isEdit = !!slug
  const [form] = Form.useForm()
  const navigate = useNavigate()
  const [content, setContent] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [visibility, setVisibility] = useState<string>('public')
  const [selectedUsers, setSelectedUsers] = useState<SelectedUser[]>([])
  const [userSearchText, setUserSearchText] = useState('')
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const token = useAuthStore((s) => s.token)
  const user = useAuthStore((s) => s.user)

  // 是否允许置顶（仅 admin / content_manager）
  // 兼容后端两种返回格式：嵌套 role.code 和扁平 role_name
  const canSetTop = user && (user.id === 'root' || (user.roles || []).some(ur => {
    const code = ur.role?.code || roleNameToCode(ur.role?.name || '')
    return code === 'admin' || code === 'content_manager'
  }))

  const createArticle = useCreateArticle()
  const updateArticle = useUpdateArticle()
  const { data: articleData, isLoading: loadingArticle } = useArticleBySlug(slug || '')
  const { data: catData } = useCategories()
  const { data: tagData } = useTags()

  // 用户搜索（带防抖）
  const [searchKeyword, setSearchKeyword] = useState('')
  const { data: userSearchData, isLoading: loadingUsers } = useUserSearch(
    searchKeyword ? { username: searchKeyword, current: 1, pageSize: 20 } : { current: 1, pageSize: 0 }
  )

  const categories = catData?.data || []
  const tags = tagData?.data || []
  const article = articleData?.data
  const searchResults = userSearchData?.data || []

  // ── 编辑回显 ──
  useEffect(() => {
    if (article && isEdit) {
      if (user && article.author_id !== user.id) {
        message.error('只能编辑自己的文章')
        navigate('/', { replace: true })
        return
      }
      form.setFieldsValue({
        ...article,
        tag_ids: article.tags?.map((t: any) => t.id) || [],
      })
      setContent(article.content || '')
      setVisibility(article.visibility || 'public')
      if (article.visible_users) {
        const vuUsers: SelectedUser[] = (article.visible_users as any[]).map((vu: any) => ({
          id: vu.user_id,
          username: '',
          name: '',
        }))
        setSelectedUsers(vuUsers)
        form.setFieldsValue({ visible_user_ids: vuUsers.map((u) => u.id) })
      }
    }
  }, [article, isEdit, user, form, navigate])

  useEffect(() => {
    if (!token) navigate(`/login?redirect=${encodeURIComponent(location.pathname)}`, { replace: true })
  }, [token, navigate])

  // ── 防抖搜索 ──
  const handleUserSearch = useCallback((val: string) => {
    setUserSearchText(val)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => setSearchKeyword(val), 300)
  }, [])

  // ── Slug ──
  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const title = e.target.value
    const currentSlug = form.getFieldValue('slug')
    if (!currentSlug || currentSlug === slugify(form.getFieldValue('_prevTitle') || '')) {
      form.setFieldsValue({ slug: slugify(title), _prevTitle: title })
    } else {
      form.setFieldsValue({ _prevTitle: title })
    }
  }

  const handleContentChange = (markdown?: string) => {
    setContent(markdown || '')
    form.setFieldsValue({ content: markdown || '' })
  }

  // ── 可见性切换 → 清理脏数据 ──
  const handleVisibilityChange = useCallback((val: string | number) => {
    const v = String(val)
    setVisibility(v)
    if (v !== 'partial_visible') {
      setSelectedUsers([])
      setUserSearchText('')
      setSearchKeyword('')
      form.setFieldsValue({ visible_user_ids: [] })
    }
  }, [form])

  // ── 用户选择 / 移除 ──
  const handleUserSelect = useCallback((userId: string) => {
    if (selectedUsers.find((u) => u.id === userId)) return
    const found = searchResults.find((u: any) => u.id === userId)
    const next = [...selectedUsers, {
      id: userId,
      username: found?.username || '',
      name: found?.name || '',
    }]
    setSelectedUsers(next)
    setUserSearchText('')
    setSearchKeyword('')
    form.setFieldsValue({ visible_user_ids: next.map((u) => u.id) })
  }, [selectedUsers, searchResults, form])

  const handleRemoveUser = useCallback((userId: string) => {
    const next = selectedUsers.filter((u) => u.id !== userId)
    setSelectedUsers(next)
    form.setFieldsValue({ visible_user_ids: next.map((u) => u.id) })
  }, [selectedUsers, form])

  // ── 可用用户列表（去重） ──
  const availableUsers = useMemo(() => {
    const selectedIds = new Set(selectedUsers.map((u) => u.id))
    return searchResults.filter((u: any) => !selectedIds.has(u.id))
  }, [searchResults, selectedUsers])

  // ── 提交 ──
  const handleSubmit = useCallback((status: 'draft' | 'published') => {
    form.validateFields().then(async (values) => {
      setSubmitting(true)
      const payload: ArticleForm = {
        title: values.title?.trim() || '',
        slug: values.slug?.trim() || slugify(values.title || ''),
        summary: values.summary || '',
        content: content || values.content || '',
        cover_image_id: values.cover_image_id || '',
        category_id: Array.isArray(values.category_id) ? (values.category_id[0] || '') : (values.category_id || ''),
        tag_ids: values.tag_ids || [],
        is_top: values.is_top || false,
        status,
        visibility: values.visibility || 'public',
        visible_user_ids: visibility === 'partial_visible' ? (values.visible_user_ids || []) : [],
        seo_title: values.seo_title || '',
        seo_keywords: values.seo_keywords || '',
        seo_desc: values.seo_desc || '',
      }

      try {
        if (isEdit && article) {
          await updateArticle.mutateAsync({ id: article.id, data: payload })
          message.success(status === 'published' ? '更新并发布成功！' : '更新并保存为草稿')
          setTimeout(() => navigate(`/article/${payload.slug || article.slug}`), 400)
        } else {
          const res = await createArticle.mutateAsync(payload)
          const slug = res?.data?.slug
          if (!slug) { message.error('发布失败：未获取到文章链接'); setSubmitting(false); return }
          message.success(status === 'published' ? '发布成功！' : '已保存为草稿')
          setTimeout(() => navigate(`/article/${slug}`), 400)
        }
      } catch (err: any) {
        message.error(err.message || '操作失败')
      } finally {
        setSubmitting(false)
      }
    })
  }, [form, content, visibility, isEdit, article, createArticle, updateArticle, navigate])

  const isPending = createArticle.isPending || updateArticle.isPending || submitting

  if (isEdit && loadingArticle) {
    return <div style={{ textAlign: 'center', padding: 120 }}><Spin size="large" /></div>
  }

  return (
    <div style={{ background: c.bg, minHeight: '100vh' }}>
      {/* ===== Top bar ===== */}
      <Affix offsetTop={0}>
        <div style={{
          background: 'rgba(10,10,16,0.92)',
          backdropFilter: 'blur(16px)',
          borderBottom: `1px solid ${c.border}`,
          padding: '0 24px', height: 56,
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          zIndex: 10,
        }}>
          <Space>
            <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}
              style={{ color: c.textMuted }}>
              返回
            </Button>
            <Text strong style={{ fontSize: 16, color: c.text }}>
              <EditOutlined /> {isEdit ? '编辑文章' : '发布文章'}
            </Text>
          </Space>
          <Space>
            <Button icon={<SaveOutlined />} onClick={() => handleSubmit('draft')} loading={isPending}
              style={{ borderRadius: 10, background: c.surfaceAlt, border: `1px solid ${c.border}`, color: c.text }}>
              保存草稿
            </Button>
            <Button type="primary" icon={<SendOutlined />} onClick={() => handleSubmit('published')} loading={isPending}
              style={{ borderRadius: 10, background: c.accent, border: 'none', boxShadow: '0 2px 12px rgba(79,110,247,0.3)' }}>
              发布
            </Button>
          </Space>
        </div>
      </Affix>

      <Form form={form} layout="vertical" initialValues={{ status: 'draft', is_top: false, visibility: 'public', content: '' }}>
        {/* ===== Title ===== */}
        <div style={{ padding: '24px 5vw 0' }}>
          <Form.Item name="title" rules={[{ required: true, message: '请输入标题' }]} style={{ marginBottom: 0 }}>
            <Input
              placeholder="输入文章标题..."
              onChange={handleTitleChange}
              maxLength={255}
              variant="borderless"
              style={{
                fontSize: 'clamp(22px, 4vw, 36px)', fontWeight: 700, padding: '16px 0', lineHeight: 1.3,
                background: 'transparent', color: '#ffffff',
              }}
            />
          </Form.Item>
        </div>

        {/* ===== Editor — CSDN 风格左右分栏 ===== */}
        <div>
          <Form.Item name="content" rules={[{ required: true, message: '请输入文章内容' }]} style={{ display: 'none' }}>
            <Input />
          </Form.Item>
          <MarkdownEditor value={content} onChange={handleContentChange} />
        </div>

        {/* ===== Settings — bottom of page ===== */}
        <div style={{ padding: '48px 5vw 48px' }}>
          <Card
            style={{ borderRadius: 16, background: c.surface, border: `1px solid ${c.border}`, maxWidth: 960, margin: '0 auto' }}
            styles={{ body: { padding: '20px 24px', background: 'transparent' } }}
          >
            <Text strong style={{ fontSize: 15, marginBottom: 16, display: 'block', color: c.text }}>
              <AppstoreOutlined /> 文章设置
            </Text>

            {/* ── 可见性 ── */}
            <div style={{ marginBottom: 16 }}>
              <div style={{ marginBottom: 8, color: c.textMuted, fontSize: 13 }}>可见范围</div>
              <Form.Item name="visibility" style={{ marginBottom: 0 }}>
                <Segmented
                  onChange={(val) => handleVisibilityChange(val)}
                  style={{
                    background: c.surfaceAlt,
                    border: `1px solid ${c.border}`,
                    borderRadius: 10,
                    padding: 3,
                  }}
                  options={[
                    {
                      label: <span style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '2px 6px' }}>
                        <GlobalOutlined /> 公开
                      </span>,
                      value: 'public',
                    },
                    {
                      label: <span style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '2px 6px' }}>
                        <LockOutlined /> 私密
                      </span>,
                      value: 'private',
                    },
                    {
                      label: <span style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '2px 6px' }}>
                        <TeamOutlined /> 部分可见
                      </span>,
                      value: 'partial_visible',
                    },
                  ]}
                />
              </Form.Item>
            </div>

            {/* ── 部分可见用户选择器 ── */}
            {visibility === 'partial_visible' && (
              <div style={{
                marginBottom: 16, padding: '16px 20px',
                background: c.accentBg, borderRadius: 12,
                border: `1px solid ${c.accentBorder}`,
              }}>
                <div style={{ marginBottom: 8, fontWeight: 500, color: c.accent, fontSize: 13 }}>
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
                        style={{
                          marginBottom: 4, borderRadius: 8,
                          padding: '2px 10px', fontSize: 12,
                        }}
                      >
                        {u.name || u.username}
                      </Tag>
                    ))}
                  </div>
                )}

                <Select
                  showSearch
                  placeholder="搜索用户..."
                  style={{ width: '100%', maxWidth: 400 }}
                  filterOption={false}
                  onSearch={handleUserSearch}
                  onSelect={(val) => val && handleUserSelect(val)}
                  value={undefined}
                  loading={loadingUsers}
                  notFoundContent={userSearchText && !loadingUsers ? '未找到相关用户' : '输入关键词搜索'}
                  options={availableUsers.map((u: any) => ({
                    label: `${u.name || ''}  @${u.username}`,
                    value: u.id,
                  }))}
                />

                <Form.Item name="visible_user_ids" style={{ display: 'none' }}>
                  <Input />
                </Form.Item>
              </div>
            )}

            <Space size="large" style={{ width: '100%', marginBottom: 8 }}>
              <Form.Item name="category_id" label={<span style={{ color: c.textMuted }}><AppstoreOutlined /> 分类</span>} style={{ marginBottom: 12, minWidth: 200 }}>
                <Select
                  mode="tags" maxCount={1}
                  placeholder="输入或选择分类"
                  options={categories.map((cat) => ({ label: cat.name, value: cat.id }))}
                  showSearch optionFilterProp="label"
                  style={{ borderRadius: 8 }}
                />
              </Form.Item>
              <Form.Item name="tag_ids" label={<span style={{ color: c.textMuted }}><TagsOutlined /> 标签</span>} style={{ marginBottom: 12, minWidth: 260 }}>
                <Select
                  mode="tags" placeholder="输入标签名后按回车"
                  options={tags.map((t) => ({ label: t.name, value: t.id }))}
                  style={{ borderRadius: 8 }}
                />
              </Form.Item>
              {canSetTop && (
                <Form.Item name="is_top" label={<span style={{ color: c.textMuted }}>置顶</span>} valuePropName="checked" style={{ marginBottom: 12 }}>
                  <Switch />
                </Form.Item>
              )}
            </Space>

            <Form.Item name="slug" label={<span style={{ color: c.textMuted }}><LinkOutlined /> Slug</span>}
              rules={[{ required: true, message: '请输入 Slug' }, { pattern: /^[a-z0-9-]+$/, message: '仅支持小写字母、数字和连字符' }]}
              style={{ marginBottom: 12 }}>
              <Input placeholder="文章 URL 别名" addonBefore="/article/" maxLength={255} style={{ borderRadius: 8 }} />
            </Form.Item>

            <Form.Item name="summary" label={<span style={{ color: c.textMuted }}><FileTextOutlined /> 摘要</span>} style={{ marginBottom: 12 }}>
              <TextArea rows={2} maxLength={1000} showCount placeholder="文章摘要，留空自动生成" style={{ borderRadius: 8 }} />
            </Form.Item>

            <div style={{ marginBottom: 12 }}>
              <div style={{ marginBottom: 8, color: c.textMuted, fontSize: 13 }}><PictureOutlined /> 封面图片</div>
              <Form.Item name="cover_image_id" style={{ marginBottom: 0 }}>
                <CoverImageUpload coverUrl={article?.cover_image?.url} />
              </Form.Item>
            </div>

            <Collapse ghost style={{ marginTop: 12 }}
              items={[{
                key: 'seo',
                label: <span style={{ color: c.textMuted }}><SearchOutlined /> SEO 设置（选填）</span>,
                children: (
                  <Space vertical style={{ width: '100%' }}>
                    <Form.Item name="seo_title" label={<span style={{ color: c.textMuted }}>SEO 标题</span>} style={{ marginBottom: 12 }}>
                      <Input placeholder="自定义页面标题" maxLength={255} style={{ borderRadius: 8 }} />
                    </Form.Item>
                    <Form.Item name="seo_keywords" label={<span style={{ color: c.textMuted }}>SEO 关键词</span>} style={{ marginBottom: 12 }}>
                      <Input placeholder="关键词1, 关键词2, 关键词3" maxLength={512} style={{ borderRadius: 8 }} />
                    </Form.Item>
                    <Form.Item name="seo_desc" label={<span style={{ color: c.textMuted }}>SEO 描述</span>} style={{ marginBottom: 0 }}>
                      <TextArea rows={2} placeholder="Meta 描述" maxLength={1000} style={{ borderRadius: 8 }} />
                    </Form.Item>
                  </Space>
                ),
              }]}
            />
          </Card>
        </div>

        {/* ===== Bottom bar ===== */}
        <Affix offsetBottom={0}>
          <div style={{
            background: 'rgba(10,10,16,0.92)',
            backdropFilter: 'blur(16px)',
            borderTop: `1px solid ${c.border}`,
            padding: '12px 24px',
            display: 'flex', justifyContent: 'center', gap: 12,
          }}>
            <Button size="large" onClick={() => navigate(-1)}
              style={{ borderRadius: 10, background: c.surfaceAlt, border: `1px solid ${c.border}`, color: c.text }}>
              取消
            </Button>
            <Button size="large" icon={<SaveOutlined />} onClick={() => handleSubmit('draft')} loading={isPending}
              style={{ borderRadius: 10, background: c.surfaceAlt, border: `1px solid ${c.border}`, color: c.text }}>
              保存草稿
            </Button>
            <Button type="primary" size="large" icon={<SendOutlined />} onClick={() => handleSubmit('published')} loading={isPending}
              style={{ minWidth: 140, borderRadius: 10, background: c.accent, border: 'none', boxShadow: '0 2px 12px rgba(79,110,247,0.3)' }}>
              发布
            </Button>
          </div>
        </Affix>
      </Form>
    </div>
  )
}

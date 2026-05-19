import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Table, Button, Space, Tag, Input, Select, message, Typography, Modal,
  Row, Col, Card, Statistic, Switch, Popconfirm, Popover, Image, Avatar,
} from 'antd'
import {
  PlusOutlined, DeleteOutlined, EditOutlined, EyeOutlined,
  ExclamationCircleOutlined, SearchOutlined, ReloadOutlined,
  FileTextOutlined, LockOutlined, GlobalOutlined, FireOutlined,
  PictureOutlined,
} from '@ant-design/icons'
import {
  useGetArticlesQuery, useDeleteArticleMutation, useBatchUpdateArticleStatusMutation,
  useToggleArticleTopMutation,
} from '../../store/api'
import { useAppSelector } from '../../store'
import dayjs from '../../utils/dayjs'
import type { Article } from '../../types'

const { Title, Text } = Typography

// ── Styles ──
const cardStyle = { borderRadius: 14, background: 'rgba(20,20,40,0.6)', border: '1px solid rgba(255,255,255,0.05)', backdropFilter: 'blur(8px)' }
const muted = { color: 'rgba(255,255,255,0.35)', fontSize: 12 }

export default function ArticleList() {
  const navigate = useNavigate()
  const user = useAppSelector((s) => s.auth.user)
  const [params, setParams] = useState({ current: 1, pageSize: 10 })
  const [searchTitle, setSearchTitle] = useState('')
  const [filterStatus, setFilterStatus] = useState<string>('')
  const [filterVisibility, setFilterVisibility] = useState<string>('')
  const [filterTop, setFilterTop] = useState<string>('')
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([])

  const queryParams = { ...params, title: searchTitle || undefined, status: filterStatus || undefined, visibility: filterVisibility || undefined, is_top: filterTop === '' ? undefined : filterTop === 'yes' }
  const { data, isLoading } = useGetArticlesQuery(queryParams)
  const [deleteArticle] = useDeleteArticleMutation()
  const [batchStatus] = useBatchUpdateArticleStatusMutation()
  const [toggleTop] = useToggleArticleTopMutation()

  const articles = (data?.data || []) as Article[]
  const total = data?.total || 0

  // ── Computed stats ──
  const stats = useMemo(() => ({
    total: total,
    published: articles.filter(a => a.status === 'published' && a.visibility === 'public').length,
    private: articles.filter(a => a.visibility === 'private').length,
    today: articles.filter(a => dayjs(a.published_at).isSame(dayjs(), 'day')).length,
  }), [articles, total])

  // ── Permission ──
  const canManage = (record: Article) => {
    if (!user) return false
    if (user.id === 'root' || user.roles?.some(r => r.role?.code === 'admin')) return true
    return record.author_id && record.author_id === user.id
  }

  // ── Delete ──
  const handleDelete = async (record: Article) => {
    try {
      await deleteArticle(record.id).unwrap()
      message.success(`已删除「${record.title}」`)
    } catch (err: any) { message.error(err.data?.error?.detail || '删除失败') }
  }

  // ── Toggle is_top ──
  const handleToggleTop = async (record: Article, checked: boolean) => {
    try {
      await toggleTop({ id: record.id, is_top: checked }).unwrap()
      message.success('置顶状态已更新')
    } catch (err: any) { message.error(err.data?.error?.detail || '更新失败') }
  }

  // ── Batch ──
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) { message.warning('请先选择文章'); return }
    Modal.confirm({
      title: '批量删除', icon: <ExclamationCircleOutlined />,
      content: `确定要删除选中的 ${selectedRowKeys.length} 篇文章吗？此操作不可逆。`,
      okText: '确认删除', okType: 'danger', cancelText: '取消',
      onOk: async () => {
        try {
          await Promise.all(selectedRowKeys.map(id => deleteArticle(id).unwrap()))
          message.success('批量删除完成')
          setSelectedRowKeys([])
        } catch (err: any) { message.error('删除失败') }
      },
    })
  }

  const handleBatchPublish = () => handleBatchChangeStatus('published')
  const handleBatchDraft = () => handleBatchChangeStatus('draft')

  const handleBatchChangeStatus = async (status: string) => {
    if (selectedRowKeys.length === 0) { message.warning('请先选择文章'); return }
    try {
      await batchStatus({ ids: selectedRowKeys, status }).unwrap()
      message.success(status === 'published' ? '已批量发布' : '已批量转草稿')
      setSelectedRowKeys([])
    } catch (err: any) { message.error(err.message) }
  }

  const handleReset = () => {
    setSearchTitle('')
    setFilterStatus('')
    setFilterVisibility('')
    setFilterTop('')
    setParams({ ...params, current: 1 })
  }

  // ── Columns ──
  const columns = [
    {
      title: '封面', dataIndex: 'cover_image', key: 'cover', width: 74,
      render: (_: any, r: Article) => {
        const url = (r as any).cover_image?.url
        const thumb = url ? (
          <img src={url} alt="" style={{ width: 60, height: 34, borderRadius: 6, objectFit: 'cover' }} />
        ) : (
          <div style={{ width: 60, height: 34, borderRadius: 6, background: 'rgba(255,255,255,0.04)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <PictureOutlined style={{ color: 'rgba(255,255,255,0.12)', fontSize: 14 }} />
          </div>
        )
        return url ? (
          <Popover content={<img src={url} alt="" style={{ maxWidth: 320, maxHeight: 200, borderRadius: 8 }} />} title="封面预览">
            {thumb}
          </Popover>
        ) : thumb
      },
    },
    {
      title: '标题 / Slug', dataIndex: 'title', key: 'title', ellipsis: true, sorter: (a: Article, b: Article) => a.title.localeCompare(b.title),
      render: (t: string, r: Article) => (
        <div>
          <div style={{ color: 'rgba(255,255,255,0.8)', fontSize: 13, fontWeight: 500, overflow: 'hidden', textOverflow: 'ellipsis' }}>{t}</div>
          <Text style={{ color: 'rgba(255,255,255,0.2)', fontSize: 11, fontFamily: "'JetBrains Mono', monospace" }}>/article/{r.slug}</Text>
        </div>
      ),
    },
    {
      title: '分类', dataIndex: 'category', key: 'category', width: 100,
      render: (c: any) => c?.name ? <Tag style={{ borderRadius: 6, fontSize: 11, background: 'rgba(79,110,247,0.1)', border: '1px solid rgba(79,110,247,0.2)', color: '#818cf8' }}>{c.name}</Tag> : <Text style={muted}>-</Text>,
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 80,
      render: (s: string) => <Tag color={s === 'published' ? 'green' : 'default'} style={{ borderRadius: 6, fontSize: 11 }}>{s === 'published' ? '已发布' : '草稿'}</Tag>,
    },
    {
      title: '可见', dataIndex: 'visibility', key: 'visibility', width: 90,
      render: (v: string) => {
        switch (v) {
          case 'private': return <Tag color="red" style={{ borderRadius: 6, fontSize: 11 }}><LockOutlined /> 私密</Tag>
          case 'partial_visible': return <Tag color="purple" style={{ borderRadius: 6, fontSize: 11 }}>部分可见</Tag>
          default: return <Tag color="blue" style={{ borderRadius: 6, fontSize: 11 }}><GlobalOutlined /> 公开</Tag>
        }
      },
    },
    {
      title: '置顶', dataIndex: 'is_top', key: 'is_top', width: 70, align: 'center' as const,
      render: (v: boolean, r: Article) => (
        <Switch size="small" checked={v} onChange={checked => handleToggleTop(r, checked)} disabled={!canManage(r)} />
      ),
    },
    {
      title: '浏览', dataIndex: 'views', key: 'views', width: 70, sorter: (a: Article, b: Article) => (a.views || 0) - (b.views || 0),
      render: (v: number) => <span style={{ fontFamily: "'Barlow', sans-serif", fontSize: 12, color: 'rgba(255,255,255,0.5)' }}>{v || 0}</span>,
    },
    {
      title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 150, sorter: (a: Article, b: Article) => new Date(a.updated_at).getTime() - new Date(b.updated_at).getTime(),
      render: (v: string) => <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>{dayjs(v).format('YYYY-MM-DD HH:mm')}</span>,
    },
    {
      title: '操作', key: 'actions', width: 150,
      render: (_: any, r: Article) => canManage(r) ? (
        <Space size={4}>
          <Button size="small" type="primary" ghost icon={<EditOutlined />} onClick={() => navigate(`/articles/${r.id}`)}>编辑</Button>
          <Popconfirm title="确定要删除这篇文章吗？此操作不可逆。" onConfirm={() => handleDelete(r)} okText="删除" okType="danger" cancelText="取消">
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ) : <Text style={{ fontSize: 11, color: 'rgba(255,255,255,0.25)' }}>无权限</Text>,
    },
  ]

  return (
    <div>
      {/* Header */}
      <div style={{ marginBottom: 20 }}>
        <Title level={4} style={{ fontWeight: 700, margin: 0 }}>文章管理</Title>
        <Text type="secondary" style={{ fontSize: 13 }}>内容运维中心 · 管理全站文章</Text>
      </div>

      {/* ── Stat Cards ── */}
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        {[
          { label: '文章总数', value: total, icon: <FileTextOutlined />, color: '#818cf8' },
          { label: '公开文章', value: stats.published, icon: <GlobalOutlined />, color: '#34d399' },
          { label: '私密/草稿', value: stats.private, icon: <LockOutlined />, color: '#f59e0b' },
          { label: '今日新增', value: stats.today, icon: <FireOutlined />, color: '#f472b6' },
        ].map((s, i) => (
          <Col xs={12} sm={6} key={s.label}>
            <Card bordered={false} style={cardStyle}
              styles={{ body: { padding: '16px 20px' } }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Statistic title={<span style={muted}>{s.label}</span>} value={s.value}
                  valueStyle={{ color: '#fff', fontSize: 24, fontWeight: 700, fontFamily: "'Barlow', sans-serif" }} />
                <div style={{ width: 36, height: 36, borderRadius: 10, background: `${s.color}15`, border: `1px solid ${s.color}20`, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18, color: s.color }}>{s.icon}</div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      {/* ── Filter Bar ── */}
      <div style={{ ...cardStyle, padding: '14px 20px', marginBottom: 14 }}>
        <Space wrap size={12}>
          <Input prefix={<SearchOutlined />} placeholder="标题 / Slug" value={searchTitle}
            onChange={e => setSearchTitle(e.target.value)} onPressEnter={() => setParams({ ...params, current: 1 })}
            style={{ width: 200 }} allowClear />
          <Select placeholder="状态" allowClear style={{ width: 120 }} value={filterStatus || undefined}
            onChange={v => { setFilterStatus(v || ''); setParams({ ...params, current: 1 }) }}
            options={[{ label: '已发布', value: 'published' }, { label: '草稿', value: 'draft' }]} />
          <Select placeholder="可见范围" allowClear style={{ width: 130 }} value={filterVisibility || undefined}
            onChange={v => { setFilterVisibility(v || ''); setParams({ ...params, current: 1 }) }}
            options={[{ label: '公开', value: 'public' }, { label: '私密', value: 'private' }, { label: '部分可见', value: 'partial_visible' }]} />
          <Select placeholder="置顶" allowClear style={{ width: 100 }} value={filterTop || undefined}
            onChange={v => { setFilterTop(v || ''); setParams({ ...params, current: 1 }) }}
            options={[{ label: '是', value: 'yes' }, { label: '否', value: 'no' }]} />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => setParams({ ...params, current: 1 })} style={{ borderRadius: 8 }}>查询</Button>
          <Button icon={<ReloadOutlined />} onClick={handleReset} style={{ borderRadius: 8 }}>重置</Button>
        </Space>
      </div>

      {/* ── Batch actions ── */}
      <div style={{ marginBottom: 12, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button icon={<PlusOutlined />} type="primary" onClick={() => navigate('/articles/new')} style={{ borderRadius: 8 }}>新建文章</Button>
          <Button onClick={handleBatchPublish} disabled={selectedRowKeys.length === 0} style={{ borderRadius: 8 }}>批量发布</Button>
          <Button onClick={handleBatchDraft} disabled={selectedRowKeys.length === 0} style={{ borderRadius: 8 }}>批量转草稿</Button>
          <Popconfirm title={`确定要删除 ${selectedRowKeys.length} 篇文章？`} onConfirm={handleBatchDelete} okText="删除" okType="danger" disabled={selectedRowKeys.length === 0}>
            <Button danger disabled={selectedRowKeys.length === 0} icon={<DeleteOutlined />} style={{ borderRadius: 8 }}>批量删除</Button>
          </Popconfirm>
        </Space>
        {selectedRowKeys.length > 0 && <Text style={{ color: 'rgba(255,255,255,0.3)', fontSize: 12 }}>已选 {selectedRowKeys.length} 篇</Text>}
      </div>

      {/* ── Table ── */}
      <Table<Article>
        rowKey="id" size="small" loading={isLoading} dataSource={articles}
        rowSelection={{ selectedRowKeys, onChange: keys => setSelectedRowKeys(keys as string[]) }}
        pagination={{ current: params.current, pageSize: params.pageSize, total, showSizeChanger: true, showTotal: t => <span style={{ color: 'rgba(255,255,255,0.3)', fontSize: 12 }}>共 {t} 篇</span> }}
        onChange={p => setParams({ current: p.current || 1, pageSize: p.pageSize || 10 })}
        columns={columns}
        locale={{ emptyText: <span style={{ color: 'rgba(255,255,255,0.2)' }}>暂无文章</span> }}
      />
    </div>
  )
}

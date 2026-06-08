import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Table, Button, Space, Tag, Input, Select, message, Modal,
  Row, Col, Card, Statistic, Switch, Popconfirm, Popover, Image,
} from 'antd'
import {
  PlusOutlined, DeleteOutlined, EditOutlined,
  SearchOutlined, ReloadOutlined,
  AppstoreOutlined, CheckCircleOutlined, CodeOutlined, FireOutlined,
  PictureOutlined, LockOutlined, GlobalOutlined,
} from '@ant-design/icons'
import {
  useGetProjectsQuery, useDeleteProjectMutation, useBatchUpdateProjectStatusMutation,
  useToggleProjectTopMutation, useToggleProjectFeaturedMutation,
} from '../../store/api'
import { useAppSelector } from '../../store'
import dayjs from '../../utils/dayjs'
import { getUserRoleCode } from '../../components/ProtectedRoute'
import type { Project, ApiResponse } from '../../types'

// ── Styles ──
const cardStyle = { borderRadius: 14, background: 'rgba(20,20,40,0.6)', border: '1px solid rgba(255,255,255,0.05)', backdropFilter: 'blur(8px)' }
const muted = { color: 'rgba(255,255,255,0.35)', fontSize: 12 }

// ── Project state color map ──
const projectStateColor: Record<string, string> = {
  developing: 'orange',
  completed: 'green',
  maintaining: 'blue',
  paused: 'yellow',
  archived: 'default',
}
const projectStateLabel: Record<string, string> = {
  developing: '开发中',
  completed: '已完成',
  maintaining: '维护中',
  paused: '暂停开发',
  archived: '停止维护',
}

export default function ProjectList() {
  const navigate = useNavigate()
  const user = useAppSelector((s) => s.auth.user)
  const [params, setParams] = useState({ current: 1, pageSize: 10 })
  const [searchTitle, setSearchTitle] = useState('')
  const [filterStatus, setFilterStatus] = useState<string>('')
  const [filterProjectState, setFilterProjectState] = useState<string>('')
  const [filterVisibility, setFilterVisibility] = useState<string>('')
  const [filterTop, setFilterTop] = useState<string>('')
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([])

  const queryParams = {
    ...params,
    title: searchTitle || undefined,
    status: filterStatus || undefined,
    project_state: filterProjectState || undefined,
    visibility: filterVisibility || undefined,
    is_top: filterTop === '' ? undefined : filterTop === 'yes',
  }
  const { data, isLoading } = useGetProjectsQuery(queryParams)
  const [deleteProject] = useDeleteProjectMutation()
  const [batchStatus] = useBatchUpdateProjectStatusMutation()
  const [toggleTop] = useToggleProjectTopMutation()
  const [toggleFeatured] = useToggleProjectFeaturedMutation()

  const projects = (data?.data || []) as Project[]
  const total = data?.total || 0

  // ── Computed stats ──
  const stats = useMemo(() => ({
    total: total,
    completed: projects.filter(p => p.project_state === 'completed').length,
    developing: projects.filter(p => p.project_state === 'developing').length,
    today: projects.filter(p => dayjs(p.created_at).isSame(dayjs(), 'day')).length,
  }), [projects, total])

  // ── Permission ──
  const canManage = (record: Project) => {
    if (!user) return false
    if (user.id === 'root' || user.roles?.some(r => r.role?.code === 'admin')) return true
    return record.author_id && record.author_id === user.id
  }

  // ── Delete ──
  const handleDelete = async (record: Project) => {
    try {
      await deleteProject(record.id).unwrap()
      message.success(`已删除「${record.title}」`)
    } catch (err: any) { message.error(err.data?.error?.detail || '删除失败') }
  }

  // ── Toggle is_top ──
  const handleToggleTop = async (record: Project, checked: boolean) => {
    try {
      await toggleTop({ id: record.id, is_top: checked }).unwrap()
      message.success('置顶状态已更新')
    } catch (err: any) { message.error(err.data?.error?.detail || '更新失败') }
  }

  // ── Toggle is_featured ──
  const handleToggleFeatured = async (record: Project, checked: boolean) => {
    try {
      await toggleFeatured({ id: record.id, is_featured: checked }).unwrap()
      message.success('精选状态已更新')
    } catch (err: any) { message.error(err.data?.error?.detail || '更新失败') }
  }

  // ── Batch ──
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) { message.warning('请先选择项目'); return }
    Modal.confirm({
      title: '确认删除',
      icon: <DeleteOutlined style={{ color: '#ff4d4f' }} />,
      content: (
        <div style={{ fontSize: 14, lineHeight: 1.6 }}>
          你选中了 <strong style={{ color: '#ff4d4f', fontSize: 16 }}>{selectedRowKeys.length}</strong> 个项目，
          删除后将<strong>无法恢复</strong>，包括项目资源、时间线等数据。
        </div>
      ),
      okText: '确认删除', okType: 'danger', cancelText: '取消',
      okButtonProps: { danger: true, type: 'primary', style: { borderRadius: 8 } },
      cancelButtonProps: { style: { borderRadius: 8 } },
      onOk: async () => {
        try {
          await Promise.all(selectedRowKeys.map(id => deleteProject(id).unwrap()))
          message.success('批量删除完成')
          setSelectedRowKeys([])
        } catch (err: any) { message.error('删除失败') }
      },
    })
  }

  const handleBatchPublish = () => handleBatchChangeStatus('published')
  const handleBatchDraft = () => handleBatchChangeStatus('draft')

  const handleBatchChangeStatus = async (status: string) => {
    if (selectedRowKeys.length === 0) { message.warning('请先选择项目'); return }
    try {
      await batchStatus({ ids: selectedRowKeys, status }).unwrap()
      message.success(status === 'published' ? '已批量发布' : '已批量转草稿')
      setSelectedRowKeys([])
    } catch (err: any) { message.error(err.message) }
  }

  const handleReset = () => {
    setSearchTitle('')
    setFilterStatus('')
    setFilterProjectState('')
    setFilterVisibility('')
    setFilterTop('')
    setParams({ ...params, current: 1 })
  }

  // ── Columns ──
  const columns = [
    {
      title: '封面', dataIndex: 'cover_image', key: 'cover', width: 74,
      render: (_: any, r: Project) => {
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
      title: '项目名 / Slug', dataIndex: 'title', key: 'title', ellipsis: true, sorter: (a: Project, b: Project) => a.title.localeCompare(b.title),
      render: (t: string, r: Project) => (
        <div>
          <div style={{ color: 'rgba(255,255,255,0.8)', fontSize: 13, fontWeight: 500, overflow: 'hidden', textOverflow: 'ellipsis' }}>{t}</div>
          <span style={{ color: 'rgba(255,255,255,0.2)', fontSize: 11, fontFamily: "'JetBrains Mono', monospace" }}>/project/{r.slug}</span>
        </div>
      ),
    },
    {
      title: '分类', dataIndex: 'category', key: 'category', width: 100,
      render: (c: any) => c?.name ? <Tag style={{ borderRadius: 6, fontSize: 11, background: 'rgba(79,110,247,0.1)', border: '1px solid rgba(79,110,247,0.2)', color: '#818cf8' }}>{c.name}</Tag> : <span style={muted}>-</span>,
    },
    {
      title: '技术栈', dataIndex: 'tags', key: 'tags', width: 180,
      render: (tags: any[]) => {
        if (!tags || tags.length === 0) return <span style={muted}>-</span>
        const maxShow = 3
        const shown = tags.slice(0, maxShow)
        const rest = tags.length - maxShow
        return (
          <Space size={4} wrap>
            {shown.map((tag: any) => (
              <Tag key={tag.id} style={{ borderRadius: 6, fontSize: 11, margin: 0 }}>{tag.name}</Tag>
            ))}
            {rest > 0 && <Tag style={{ borderRadius: 6, fontSize: 11, margin: 0, background: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.1)' }}>+{rest}</Tag>}
          </Space>
        )
      },
    },
    {
      title: '项目状态', dataIndex: 'project_state', key: 'project_state', width: 100,
      render: (s: string) => <Tag color={projectStateColor[s] || 'default'} style={{ borderRadius: 6, fontSize: 11 }}>{projectStateLabel[s] || s}</Tag>,
    },
    {
      title: '发布状态', dataIndex: 'status', key: 'status', width: 90,
      render: (s: string) => <Tag color={s === 'published' ? 'green' : 'default'} style={{ borderRadius: 6, fontSize: 11 }}>{s === 'published' ? '已发布' : '草稿'}</Tag>,
    },
    {
      title: '可见范围', dataIndex: 'visibility', key: 'visibility', width: 100,
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
      render: (v: boolean, r: Project) => (
        <Switch size="small" checked={v} onChange={checked => handleToggleTop(r, checked)} disabled={!canManage(r)} />
      ),
    },
    {
      title: '精选', dataIndex: 'is_featured', key: 'is_featured', width: 70, align: 'center' as const,
      render: (v: boolean, r: Project) => (
        <Switch size="small" checked={v} onChange={checked => handleToggleFeatured(r, checked)} disabled={!canManage(r)} />
      ),
    },
    {
      title: '浏览', dataIndex: 'views', key: 'views', width: 70, sorter: (a: Project, b: Project) => (a.views || 0) - (b.views || 0),
      render: (v: number) => <span style={{ fontFamily: "'Barlow', sans-serif", fontSize: 12, color: 'rgba(255,255,255,0.5)' }}>{v || 0}</span>,
    },
    {
      title: '操作', key: 'actions', width: 150,
      render: (_: any, r: Project) => canManage(r) ? (
        <Space size={4}>
          <Button size="small" type="primary" ghost icon={<EditOutlined />} onClick={() => navigate(`/projects/${r.id}`)}>编辑</Button>
          <Popconfirm
            title="确认删除"
            description={<span>确定要删除《<strong style={{ color: '#ff4d4f' }}>{r.title}</strong>》吗？此操作<strong>不可恢复</strong>。</span>}
            onConfirm={() => handleDelete(r)} okText="确认删除" okType="danger" cancelText="取消"
            okButtonProps={{ danger: true, type: 'primary' }}
          >
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ) : <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.25)' }}>无权限</span>,
    },
  ]

  return (
    <div>
      {/* Header */}
      <div style={{ marginBottom: 20 }}>
        <h4 style={{ fontWeight: 700, margin: 0, fontSize: 20, color: 'rgba(255,255,255,0.85)' }}>项目管理</h4>
        <span style={{ color: 'rgba(255,255,255,0.45)', fontSize: 13 }}>项目运维中心 · 管理全站项目</span>
      </div>

      {/* ── Stat Cards ── */}
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        {[
          { label: '项目总数', value: stats.total, icon: <AppstoreOutlined />, color: '#818cf8' },
          { label: '已完成项目', value: stats.completed, icon: <CheckCircleOutlined />, color: '#34d399' },
          { label: '开发中项目', value: stats.developing, icon: <CodeOutlined />, color: '#f59e0b' },
          { label: '今日新增', value: stats.today, icon: <FireOutlined />, color: '#f472b6' },
        ].map((s) => (
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
          <Select placeholder="发布状态" allowClear style={{ width: 120 }} value={filterStatus || undefined}
            onChange={v => { setFilterStatus(v || ''); setParams({ ...params, current: 1 }) }}
            options={[{ label: '已发布', value: 'published' }, { label: '草稿', value: 'draft' }]} />
          <Select placeholder="项目状态" allowClear style={{ width: 130 }} value={filterProjectState || undefined}
            onChange={v => { setFilterProjectState(v || ''); setParams({ ...params, current: 1 }) }}
            options={[
              { label: '开发中', value: 'developing' },
              { label: '已完成', value: 'completed' },
              { label: '维护中', value: 'maintaining' },
              { label: '暂停开发', value: 'paused' },
              { label: '停止维护', value: 'archived' },
            ]} />
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
          <Button icon={<PlusOutlined />} type="primary" onClick={() => navigate('/projects/new')} style={{ borderRadius: 8 }}>新增项目</Button>
          <Button onClick={handleBatchPublish} disabled={selectedRowKeys.length === 0} style={{ borderRadius: 8 }}>批量发布</Button>
          <Button onClick={handleBatchDraft} disabled={selectedRowKeys.length === 0} style={{ borderRadius: 8 }}>批量草稿</Button>
          <Popconfirm title={`确定要删除 ${selectedRowKeys.length} 个项目？`} onConfirm={handleBatchDelete} okText="删除" okType="danger" disabled={selectedRowKeys.length === 0}>
            <Button danger disabled={selectedRowKeys.length === 0} icon={<DeleteOutlined />} style={{ borderRadius: 8 }}>批量删除</Button>
          </Popconfirm>
        </Space>
        {selectedRowKeys.length > 0 && <span style={{ color: 'rgba(255,255,255,0.3)', fontSize: 12 }}>已选 {selectedRowKeys.length} 个</span>}
      </div>

      {/* ── Table ── */}
      <Table<Project>
        rowKey="id" size="small" loading={isLoading} dataSource={projects}
        rowSelection={{ selectedRowKeys, onChange: keys => setSelectedRowKeys(keys as string[]) }}
        pagination={{ current: params.current, pageSize: params.pageSize, total, showSizeChanger: true, showTotal: t => <span style={{ color: 'rgba(255,255,255,0.3)', fontSize: 12 }}>共 {t} 个</span> }}
        onChange={p => setParams({ current: p.current || 1, pageSize: p.pageSize || 10 })}
        columns={columns}
        locale={{ emptyText: <span style={{ color: 'rgba(255,255,255,0.2)' }}>暂无项目</span> }}
      />
    </div>
  )
}

import { useState, useEffect } from 'react'
import {
  Table, Button, Space, Tag, Popconfirm, message, Typography, Tabs,
  Input, Select, Row, Col, Card, Statistic, Avatar, Badge, Tooltip,
} from 'antd'
import {
  CheckOutlined, CloseOutlined, DeleteOutlined, SearchOutlined,
  ReloadOutlined, MessageOutlined, LinkOutlined, UserOutlined,
  MailOutlined, EnvironmentOutlined, ExclamationCircleOutlined,
  StopOutlined,
} from '@ant-design/icons'
import {
  useGetCommentsQuery, useApproveCommentMutation, useRejectCommentMutation,
  useDeleteCommentMutation, useGetCommentStatsQuery,
} from '../../store/api'
import dayjs from '../../utils/dayjs'

const { Title, Text, Paragraph } = Typography

const cardStyle = { borderRadius: 14, background: 'rgba(20,20,40,0.6)', border: '1px solid rgba(255,255,255,0.05)', backdropFilter: 'blur(8px)' }
const muted = { color: 'rgba(255,255,255,0.35)', fontSize: 12 }

const statusColors: Record<string, string> = { approved: 'green', pending: 'orange', rejected: 'red' }
const statusLabels: Record<string, string> = { approved: '已通过', pending: '待审核', rejected: '垃圾评论' }

export default function CommentManage() {
  const [params, setParams] = useState({ current: 1, pageSize: 10 })
  const [activeTab, setActiveTab] = useState('all')
  const [searchKeyword, setSearchKeyword] = useState('')
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([])
  const [replyTarget, setReplyTarget] = useState<string | null>(null)
  const [replyContent, setReplyContent] = useState('')

  const queryParams = { ...params, status: activeTab !== 'all' ? activeTab : undefined, keyword: searchKeyword || undefined }
  const { data, isLoading, refetch } = useGetCommentsQuery(queryParams)
  const { data: statsData } = useGetCommentStatsQuery()
  const [approve] = useApproveCommentMutation()
  const [reject] = useRejectCommentMutation()
  const [del] = useDeleteCommentMutation()

  const comments = data?.data || []
  const total = data?.total || 0
  const stats = statsData?.data || { total: 0, pending: 0, approved: 0, rejected: 0 }

  // Refetch on tab change
  useEffect(() => { refetch() }, [activeTab, refetch])

  const handleApprove = async (id: string) => {
    try { await approve(id); message.success('已通过审核'); refetch() } catch (err: any) { message.error(err.message) }
  }
  const handleReject = async (id: string) => {
    try { await reject(id); message.success('已标记为垃圾评论'); refetch() } catch (err: any) { message.error(err.message) }
  }
  const handleDelete = async (id: string) => {
    try { await del(id); message.success('已删除'); refetch() } catch (err: any) { message.error(err.message) }
  }

  const handleBatchApprove = async () => {
    if (selectedRowKeys.length === 0) { message.warning('请先选择评论'); return }
    await Promise.all(selectedRowKeys.map(id => approve(id).unwrap()))
    message.success(`已批量通过 ${selectedRowKeys.length} 条`)
    setSelectedRowKeys([])
    refetch()
  }
  const handleBatchReject = async () => {
    if (selectedRowKeys.length === 0) { message.warning('请先选择评论'); return }
    await Promise.all(selectedRowKeys.map(id => reject(id).unwrap()))
    message.success(`已批量标记 ${selectedRowKeys.length} 条`)
    setSelectedRowKeys([])
    refetch()
  }
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) { message.warning('请先选择评论'); return }
    await Promise.all(selectedRowKeys.map(id => del(id).unwrap()))
    message.success(`已批量删除 ${selectedRowKeys.length} 条`)
    setSelectedRowKeys([])
    refetch()
  }

  const handleReply = () => {
    if (!replyContent.trim()) { message.warning('请输入回复内容'); return }
    message.success('回复已成功发表')
    setReplyTarget(null)
    setReplyContent('')
  }

  const columns = [
    {
      title: '评论人', dataIndex: 'username', key: 'user', width: 180,
      render: (name: string, r: any) => (
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10 }}>
          <Avatar icon={<UserOutlined />} size={32} style={{ flexShrink: 0, background: r.status === 'approved' ? '#34d399' : r.status === 'rejected' ? '#6b7280' : '#f59e0b' }} />
          <div style={{ minWidth: 0 }}>
            <div style={{ color: 'rgba(255,255,255,0.8)', fontSize: 13, fontWeight: 500 }}>{name}</div>
            {r.email && <div style={{ display: 'flex', alignItems: 'center', gap: 4, marginTop: 1 }}><MailOutlined style={{ fontSize: 10, color: 'rgba(255,255,255,0.2)' }} /><Text style={{ fontSize: 10, color: 'rgba(255,255,255,0.25)' }}>{r.email}</Text></div>}
            {r.ip && <div style={{ display: 'flex', alignItems: 'center', gap: 4, marginTop: 1 }}><EnvironmentOutlined style={{ fontSize: 10, color: 'rgba(255,255,255,0.2)' }} /><Text code style={{ fontSize: 10, color: 'rgba(255,255,255,0.25)' }}>{r.ip}</Text></div>}
          </div>
        </div>
      ),
    },
    {
      title: '评论内容', dataIndex: 'content', key: 'content',
      render: (text: string, r: any) => (
        <div>
          <Paragraph ellipsis={{ rows: 2, expandable: true, symbol: '展开更多' }}
            style={{ color: 'rgba(255,255,255,0.7)', fontSize: 13, margin: 0, lineHeight: 1.7 }}>{text}</Paragraph>
          <div style={{ marginTop: 4, display: 'flex', alignItems: 'center', gap: 6 }}>
            <Tag color={statusColors[r.status]} style={{ fontSize: 10, borderRadius: 4, margin: 0 }}>{statusLabels[r.status]}</Tag>
          </div>
        </div>
      ),
    },
    { title: '时间', dataIndex: 'created_at', key: 'time', width: 160, render: (v: string) => <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>{dayjs(v).format('YYYY-MM-DD HH:mm')}</span> },
    {
      title: '操作', key: 'actions', width: 220,
      render: (_: any, r: any) => (
        <Space size={4}>
          {r.status === 'pending' && (
            <>
              <Button size="small" type="primary" icon={<CheckOutlined />} onClick={() => handleApprove(r.id)} style={{ borderRadius: 6, background: '#34d399', border: 'none' }}>通过</Button>
              <Button size="small" icon={<CloseOutlined />} onClick={() => handleReject(r.id)} style={{ borderRadius: 6, color: '#f59e0b', borderColor: '#f59e0b' }}>垃圾</Button>
            </>
          )}
          {r.status === 'approved' && (
            <Button size="small" icon={<MessageOutlined />} onClick={() => setReplyTarget(replyTarget === r.id ? null : r.id)} style={{ borderRadius: 6 }}>回复</Button>
          )}
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)} okText="删除" okType="danger">
            <Button size="small" danger icon={<DeleteOutlined />} style={{ borderRadius: 6 }} />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 20 }}>
        <Title level={4} style={{ fontWeight: 700, margin: 0 }}>评论管理</Title>
        <Text type="secondary" style={{ fontSize: 13 }}>全站互动审核中心 · 先审后发 · 垃圾过滤</Text>
      </div>

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        {[
          { label: '总评论', value: stats.total, icon: <MessageOutlined />, color: '#818cf8' },
          { label: '待审核', value: stats.pending, icon: <ExclamationCircleOutlined />, color: '#f59e0b' },
          { label: '已通过', value: stats.approved, icon: <CheckOutlined />, color: '#34d399' },
          { label: '垃圾评论', value: stats.rejected, icon: <StopOutlined />, color: '#6b7280' },
        ].map((s, i) => (
          <Col xs={12} sm={6} key={s.label}>
            <Card variant="borderless" style={cardStyle} styles={{ body: { padding: '16px 20px' } }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Statistic title={<span style={muted}>{s.label}</span>} value={s.value}
                  styles={{ content: { color: s.color === '#6b7280' ? 'rgba(255,255,255,0.4)' : '#fff', fontSize: 24, fontWeight: 700, fontFamily: "'Barlow', sans-serif" } }} />
                <div style={{ width: 36, height: 36, borderRadius: 10, background: `${s.color}15`, border: `1px solid ${s.color}20`, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18, color: s.color }}>{s.icon}</div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <Tabs activeKey={activeTab} onChange={k => { setActiveTab(k); setParams({ ...params, current: 1 }); setSelectedRowKeys([]) }}
        items={[
          { key: 'all', label: <span>全部<Badge count={stats.total} style={{ marginLeft: 6, backgroundColor: '#818cf8' }} /></span> },
          { key: 'pending', label: <span>待审核<Badge count={stats.pending} style={{ marginLeft: 6 }} /></span> },
          { key: 'approved', label: <span>已通过<Badge count={stats.approved} style={{ marginLeft: 6, backgroundColor: '#34d399' }} /></span> },
          { key: 'rejected', label: <span>垃圾评论<Badge count={stats.rejected} style={{ marginLeft: 6, backgroundColor: '#6b7280' }} /></span> },
        ]} />

      <div style={{ ...cardStyle, padding: '14px 20px', marginBottom: 14, marginTop: -16 }}>
        <Space wrap size={12}>
          <Input prefix={<SearchOutlined />} placeholder="评论内容 / 昵称" value={searchKeyword}
            onChange={e => setSearchKeyword(e.target.value)} style={{ width: 220 }} allowClear />
          <Button icon={<ReloadOutlined />} onClick={() => { setSearchKeyword(''); setParams({ ...params, current: 1 }); refetch() }} style={{ borderRadius: 8 }}>重置</Button>
        </Space>
      </div>

      {selectedRowKeys.length > 0 && (
        <div style={{ marginBottom: 12, display: 'flex', alignItems: 'center', gap: 8 }}>
          <Text style={{ color: 'rgba(255,255,255,0.3)', fontSize: 12 }}>已选 {selectedRowKeys.length} 条</Text>
          <Button size="small" type="primary" icon={<CheckOutlined />} onClick={handleBatchApprove} style={{ borderRadius: 6 }}>批量通过</Button>
          <Button size="small" icon={<StopOutlined />} onClick={handleBatchReject} style={{ borderRadius: 6 }}>批量标记垃圾</Button>
          <Popconfirm title="确定批量删除？" onConfirm={handleBatchDelete} okText="删除" okType="danger">
            <Button size="small" danger icon={<DeleteOutlined />} style={{ borderRadius: 6 }}>批量删除</Button>
          </Popconfirm>
        </div>
      )}

      <Table rowKey="id" size="small" loading={isLoading} dataSource={comments}
        rowSelection={{ selectedRowKeys, onChange: keys => setSelectedRowKeys(keys as string[]) }}
        pagination={{ current: params.current, pageSize: params.pageSize, total, showTotal: t => <span style={{ color: 'rgba(255,255,255,0.3)', fontSize: 12 }}>共 {t} 条</span> }}
        onChange={p => setParams({ current: p.current || 1, pageSize: p.pageSize || 10 })}
        columns={columns}
        expandable={{
          expandedRowRender: (r: any) => replyTarget === r.id ? (
            <div style={{ padding: '12px 16px', background: 'rgba(255,255,255,0.02)', borderRadius: 10, border: '1px solid rgba(255,255,255,0.05)' }}>
              <Input.TextArea rows={3} placeholder="输入回复内容..." value={replyContent} onChange={e => setReplyContent(e.target.value)} style={{ borderRadius: 8, marginBottom: 10 }} />
              <Space><Button type="primary" size="small" onClick={handleReply} style={{ borderRadius: 6 }}>提交回复</Button><Button size="small" onClick={() => { setReplyTarget(null); setReplyContent('') }} style={{ borderRadius: 6 }}>取消</Button></Space>
            </div>
          ) : null,
          expandedRowKeys: replyTarget ? [replyTarget] : [],
          showExpandColumn: false,
        }}
        locale={{ emptyText: <span style={{ color: 'rgba(255,255,255,0.2)' }}>暂无评论</span> }}
      />
    </div>
  )
}

import { useState, useMemo } from 'react'
import {
  Table, Button, Space, Tag, Popconfirm, Modal, Form, Input, Select,
  message, Typography, Row, Col, Card, Statistic, Switch, Avatar,
  Radio, Tree, Tooltip,
} from 'antd'
import {
  UserOutlined, CrownOutlined, EditOutlined, DeleteOutlined,
  LockOutlined, SearchOutlined, ReloadOutlined, CopyOutlined,
  SafetyCertificateOutlined, TeamOutlined, StopOutlined,
  FileTextOutlined, HeartOutlined, CommentOutlined, MailOutlined,
} from '@ant-design/icons'
import {
  useGetUsersQuery, useDeleteUserMutation,
  useChangeUserRoleMutation, useChangeUserStatusMutation,
} from '../../store/api'
import { useGetRolesQuery } from '../../store/api'
import dayjs from '../../utils/dayjs'
import type { User } from '../../types'

const { Title, Text } = Typography

const cardStyle = { borderRadius: 14, background: 'rgba(20,20,40,0.6)', border: '1px solid rgba(255,255,255,0.05)', backdropFilter: 'blur(8px)' }
const muted = { color: 'rgba(255,255,255,0.35)', fontSize: 12 }

const ROLE_DEFS: Record<string, { label: string; desc: string; color: string }> = {
  super_admin:       { label: '超级管理员', desc: '完整后台权限', color: '#722ed1' },
  content_manager:   { label: '内容管理员', desc: '管理文章/分类/标签/评论', color: '#177ddc' },
  comment_moderator: { label: '评论审核员', desc: '仅审核管理评论', color: '#fa8c16' },
  user:              { label: '普通用户', desc: '允许发布文章、评论、阅读', color: '#8c8c8c' },
  guest:             { label: '游客', desc: '只读访问', color: '#d9d9d9' },
}

const permissionTree = [
  { title: '仪表盘', key: 'dashboard', children: [] },
  { title: '内容管理', key: 'content', children: [
    { title: '文章管理', key: 'articles' },
    { title: '分类管理', key: 'categories' },
    { title: '标签管理', key: 'tags' },
    { title: '评论管理', key: 'comments' },
    { title: '媒体库', key: 'media' },
  ]},
  { title: '系统管理', key: 'system', children: [
    { title: '用户管理', key: 'users' },
    { title: '角色管理', key: 'roles' },
    { title: '菜单管理', key: 'menus' },
    { title: '系统设置', key: 'settings' },
  ]},
  { title: '安全审计', key: 'audit', children: [
    { title: '操作日志', key: 'logs-audit' },
  ]},
]

const rolePermissions: Record<string, string[]> = {
  super_admin:       ['dashboard', 'content', 'articles', 'categories', 'tags', 'comments', 'media', 'system', 'users', 'roles', 'menus', 'settings', 'audit', 'logs-audit'],
  content_manager:   ['dashboard', 'content', 'articles', 'categories', 'tags', 'comments', 'media'],
  comment_moderator: ['dashboard', 'content', 'comments'],
  user:              ['dashboard'],
  guest:             [],
}

export default function UserList() {
  const [params, setParams] = useState({ current: 1, pageSize: 10 })
  const [search, setSearch] = useState('')
  const [filterRole, setFilterRole] = useState('')
  const [filterStatus, setFilterStatus] = useState('')
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([])
  const [roleModalOpen, setRoleModalOpen] = useState(false)
  const [roleUser, setRoleUser] = useState<any>(null)
  const [selectedRole, setSelectedRole] = useState('')

  const queryParams = { ...params, likeUsername: search || undefined, likeName: search || undefined, status: filterStatus || undefined }
  const { data, isLoading } = useGetUsersQuery(queryParams)
  const { data: rolesData } = useGetRolesQuery({ current: 1, pageSize: 100, status: 'enabled' })
  const [deleteUser] = useDeleteUserMutation()
  const [changeRole] = useChangeUserRoleMutation()
  const [changeStatus] = useChangeUserStatusMutation()

  const users = (data?.data || []) as User[]
  const total = data?.total || 0
  const roles = rolesData?.data || []

  const stats = useMemo(() => ({
    total: total,
    superAdmin: users.filter(u => u.roles?.some(r => ((r.role?.code)) === 'super_admin' || ((r.role?.code)) === 'admin')).length,
    contentMgr: users.filter(u => u.roles?.some(r => ((r.role?.code)) === 'content_manager')).length,
    banned: users.filter(u => u.status === 'freezed').length,
  }), [users, total])

  const getUserRole = (u: User) => {
    if (u.id === 'root') return { code: 'super_admin', label: '超级管理员' }
    const codes = (u.roles || []).map(r => (r.role?.code) || '')
    if (codes.includes('super_admin') || codes.includes('admin')) return { code: 'super_admin', label: '超级管理员' }
    if (codes.includes('content_manager')) return { code: 'content_manager', label: '内容管理员' }
    if (codes.includes('comment_moderator')) return { code: 'comment_moderator', label: '评论审核员' }
    if (codes.includes('user')) return { code: 'user', label: '普通用户' }
    return { code: 'guest', label: '游客' }
  }

  const handleToggleStatus = async (u: User, checked: boolean) => {
    const newStatus = checked ? 'activated' : 'freezed'
    try {
      await changeStatus({ id: u.id, status: newStatus }).unwrap()
      message.success(checked ? '账号已启用' : '账号已禁用')
    } catch (err: any) { message.error(err?.data?.error?.detail || '状态变更失败') }
  }

  const openRoleModal = (u: User) => {
    setRoleUser(u)
    const r = getUserRole(u)
    setSelectedRole(r.code)
    setRoleModalOpen(true)
  }

  const handleRoleSave = async () => {
    if (!roleUser) return
    if (getUserRole(roleUser).code === 'super_admin' && selectedRole !== 'super_admin') {
      message.error('系统核心超级管理员角色不可被变更或降权')
      return
    }
    try {
      await changeRole({ id: roleUser.id, role_code: selectedRole }).unwrap()
      message.success('该用户角色权限已成功变更')
      setRoleModalOpen(false)
    } catch (err: any) { message.error(err?.data?.error?.detail || '变更失败') }
  }

  const handleDelete = async (id: string) => {
    try { await deleteUser(id).unwrap(); message.success('已删除') } catch (err: any) { message.error('删除失败') }
  }

  const handleReset = () => { setSearch(''); setFilterRole(''); setFilterStatus(''); setParams({ ...params, current: 1 }) }

  const handleCopyEmail = (email: string) => {
    navigator.clipboard.writeText(email).then(() => message.success('已复制'))
  }

  const columns = [
    {
      title: '用户信息', dataIndex: 'username', key: 'user', width: 200,
      render: (_: string, u: User) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <Avatar src={u.avatar || undefined} icon={<UserOutlined />} size={36} style={{ flexShrink: 0, background: '#818cf8' }} />
          <div style={{ minWidth: 0 }}>
            <div style={{ color: 'rgba(255,255,255,0.8)', fontSize: 13, fontWeight: 500, overflow: 'hidden', textOverflow: 'ellipsis' }}>{u.name || u.username}</div>
            <Text code style={{ fontSize: 11, color: 'rgba(255,255,255,0.3)' }}>{u.id?.substring(0, 12)}</Text>
          </div>
        </div>
      ),
    },
    {
      title: '邮箱', dataIndex: 'email', key: 'email', width: 200,
      render: (v: string) => v ? (
        <Space size={4}>
          <MailOutlined style={{ color: 'rgba(255,255,255,0.25)', fontSize: 12 }} />
          <span style={{ color: 'rgba(255,255,255,0.5)', fontSize: 12 }}>{v}</span>
          <Tooltip title="复制"><CopyOutlined onClick={() => handleCopyEmail(v)} style={{ color: 'rgba(255,255,255,0.2)', cursor: 'pointer', fontSize: 11 }} /></Tooltip>
        </Space>
      ) : <Text style={muted}>-</Text>,
    },
    {
      title: '角色', dataIndex: 'roles', key: 'role', width: 130,
      render: (_: any, u: User) => {
        const role = getUserRole(u)
        return <Tag color={ROLE_DEFS[role.code]?.color || 'default'} style={{ borderRadius: 6, fontSize: 11 }}>{role.label}</Tag>
      },
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 80, align: 'center' as const,
      render: (s: string, u: User) => (
        <Switch size="small" checked={s === 'activated'} onChange={checked => handleToggleStatus(u, checked)}
          checkedChildren={<span style={{ fontSize: 10 }}>正常</span>}
          unCheckedChildren={<span style={{ fontSize: 10 }}>冻结</span>} />
      ),
    },
    {
      title: '注册时间', dataIndex: 'created_at', key: 'created_at', width: 150, sorter: true,
      render: (v: string) => <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>{dayjs(v).format('YYYY-MM-DD HH:mm')}</span>,
    },
    {
      title: '操作', key: 'actions', width: 180,
      render: (_: any, u: User) => (
        <Space size={4}>
          <Button size="small" type="primary" ghost onClick={() => openRoleModal(u)}>修改权限</Button>
          <Popconfirm title="确定删除此用户？" onConfirm={() => handleDelete(u.id)} okText="删除" okType="danger">
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 20 }}>
        <Title level={4} style={{ fontWeight: 700, margin: 0 }}>用户管理</Title>
        <Text type="secondary" style={{ fontSize: 13 }}>全站用户监控 · 角色权限分发 · 账号运维</Text>
      </div>

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        {[
          { label: '总注册用户', value: stats.total, icon: <TeamOutlined />, color: '#818cf8' },
          { label: '超级管理员', value: stats.superAdmin, icon: <CrownOutlined />, color: '#722ed1' },
          { label: '内容管理员', value: stats.contentMgr, icon: <EditOutlined />, color: '#177ddc' },
          { label: '已冻结账号', value: stats.banned, icon: <StopOutlined />, color: '#ef4444' },
        ].map((s, i) => (
          <Col xs={12} sm={6} key={s.label}>
            <Card bordered={false} style={cardStyle} styles={{ body: { padding: '16px 20px' } }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Statistic title={<span style={muted}>{s.label}</span>} value={s.value}
                  valueStyle={{ color: s.color === '#ef4444' ? '#f87171' : '#fff', fontSize: 24, fontWeight: 700, fontFamily: "'Barlow', sans-serif" }} />
                <div style={{ width: 36, height: 36, borderRadius: 10, background: `${s.color}15`, border: `1px solid ${s.color}20`, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18, color: s.color }}>{s.icon}</div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <div style={{ ...cardStyle, padding: '14px 20px', marginBottom: 14 }}>
        <Space wrap size={12}>
          <Input prefix={<SearchOutlined />} placeholder="UID / 昵称 / Email" value={search}
            onChange={e => setSearch(e.target.value)} onPressEnter={() => setParams({ ...params, current: 1 })}
            style={{ width: 220 }} allowClear />
          <Select placeholder="角色" allowClear style={{ width: 140 }} value={filterRole || undefined}
            onChange={v => { setFilterRole(v || ''); setParams({ ...params, current: 1 }) }}
            options={Object.entries(ROLE_DEFS).map(([code, def]) => ({ label: def.label, value: code }))} />
          <Select placeholder="状态" allowClear style={{ width: 110 }} value={filterStatus || undefined}
            onChange={v => { setFilterStatus(v || ''); setParams({ ...params, current: 1 }) }}
            options={[{ label: '正常', value: 'activated' }, { label: '已冻结', value: 'freezed' }]} />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => setParams({ ...params, current: 1 })} style={{ borderRadius: 8 }}>查询</Button>
          <Button icon={<ReloadOutlined />} onClick={handleReset} style={{ borderRadius: 8 }}>重置</Button>
        </Space>
      </div>

      <Table<User>
        rowKey="id" size="small" loading={isLoading} dataSource={users}
        pagination={{ current: params.current, pageSize: params.pageSize, total, showSizeChanger: true, showTotal: t => <span style={{ color: 'rgba(255,255,255,0.3)', fontSize: 12 }}>共 {t} 人</span> }}
        onChange={p => setParams({ current: p.current || 1, pageSize: p.pageSize || 10 })}
        columns={columns}
        locale={{ emptyText: <span style={{ color: 'rgba(255,255,255,0.2)' }}>暂无用户</span> }}
      />

      <Modal open={roleModalOpen} onCancel={() => setRoleModalOpen(false)} onOk={handleRoleSave}
        title={<Space><SafetyCertificateOutlined /> 修改用户权限</Space>}
        okText="确认变更" cancelText="取消" width={560}>
        {roleUser && (
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 20, padding: '12px 16px', borderRadius: 12, background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.06)' }}>
              <Avatar src={roleUser.avatar || undefined} icon={<UserOutlined />} size={40} />
              <div>
                <div style={{ color: '#fff', fontWeight: 600, fontSize: 14 }}>{roleUser.name || roleUser.username}</div>
                <Text type="secondary" style={{ fontSize: 11 }}>UID: {roleUser.id}</Text>
              </div>
            </div>

            <div style={{ marginBottom: 16 }}>
              <Text strong style={{ color: 'rgba(255,255,255,0.7)', fontSize: 13, display: 'block', marginBottom: 10 }}>选择角色</Text>
              <Radio.Group value={selectedRole} onChange={e => setSelectedRole(e.target.value)}
                style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                {Object.entries(ROLE_DEFS).map(([code, def]) => (
                  <Radio.Button key={code} value={code}
                    disabled={code === 'guest'}
                    style={{ flex: '1 1 45%', height: 'auto', padding: '12px 10px', borderRadius: 10, textAlign: 'center', border: `1px solid ${selectedRole === code ? def.color + '40' : 'rgba(255,255,255,0.08)'}`, background: selectedRole === code ? def.color + '10' : 'transparent', transition: 'all 0.2s' }}>
                    <div style={{ color: def.color, fontWeight: 700, fontSize: 13 }}>{def.label}</div>
                    <div style={{ color: 'rgba(255,255,255,0.3)', fontSize: 10, marginTop: 2 }}>{def.desc}</div>
                  </Radio.Button>
                ))}
              </Radio.Group>
            </div>

            <div style={{ padding: '14px 16px', borderRadius: 10, background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.05)' }}>
              <Text strong style={{ color: 'rgba(255,255,255,0.5)', fontSize: 12, display: 'block', marginBottom: 8 }}>权限预览</Text>
              <Tree checkable defaultCheckedKeys={rolePermissions[selectedRole] || []}
                treeData={permissionTree}
                style={{ color: 'rgba(255,255,255,0.6)', fontSize: 12 }} />
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}

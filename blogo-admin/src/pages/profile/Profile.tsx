import { useState, useEffect, useRef } from 'react'
import { Card, Form, Input, Button, message, Typography, Avatar, Spin, Descriptions, Tag, Divider, Tooltip } from 'antd'
import {
  UserOutlined, CameraOutlined, MailOutlined, PhoneOutlined, EnvironmentOutlined,
  EditOutlined, LockOutlined, ClockCircleOutlined, CheckCircleOutlined,
  GlobalOutlined, DesktopOutlined, IdcardOutlined, SignatureOutlined,
} from '@ant-design/icons'
import { useAppSelector, useAppDispatch } from '../../store'
import { fetchUserInfo } from '../../store/authSlice'
import client from '../../api/client'
import dayjs from '../../utils/dayjs'

const { Title, Text, Paragraph } = Typography

// 玻璃卡片样式
const glassCard = (bg = 'rgba(255,255,255,0.03)') => ({
  background: bg,
  borderRadius: 16,
  border: '1px solid rgba(255,255,255,0.06)',
  backdropFilter: 'blur(12px)',
} as const)

function roleBadge(code?: string) {
  const map: Record<string, { label: string; color: string; bg: string }> = {
    super_admin: { label: '超级管理员', color: '#f43f5e', bg: 'rgba(244,63,94,0.12)' },
    admin: { label: '管理员', color: '#f43f5e', bg: 'rgba(244,63,94,0.12)' },
    content_manager: { label: '内容管理', color: '#8b5cf6', bg: 'rgba(139,92,246,0.12)' },
    comment_moderator: { label: '评论审核', color: '#f59e0b', bg: 'rgba(245,158,11,0.12)' },
    user: { label: '用户', color: '#3b82f6', bg: 'rgba(59,130,246,0.12)' },
  }
  const role = map[code || ''] || { label: code || '未知', color: '#6b7280', bg: 'rgba(107,114,128,0.12)' }
  return (
    <span style={{
      display: 'inline-block', fontSize: 11, fontWeight: 600,
      color: role.color, background: role.bg,
      padding: '2px 10px', borderRadius: 20, letterSpacing: '0.04em',
    }}>
      {role.label}
    </span>
  )
}

export default function Profile() {
  const user = useAppSelector((s) => s.auth.user)
  const dispatch = useAppDispatch()
  const [infoLoading, setInfoLoading] = useState(false)
  const [pwdLoading, setPwdLoading] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [avatarUrl, setAvatarUrl] = useState('')
  const [showPwdCard, setShowPwdCard] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [infoForm] = Form.useForm()
  const [pwdForm] = Form.useForm()

  useEffect(() => { dispatch(fetchUserInfo()) }, [dispatch])

  useEffect(() => {
    if (user) {
      setAvatarUrl(user.avatar || '')
      infoForm.setFieldsValue({
        name: user.name, phone: user.phone, email: user.email,
        avatar: user.avatar || '', bio: user.bio || '', remark: user.remark || '',
      })
    }
  }, [user, infoForm])

  if (!user) return <Spin size="large" style={{ display: 'block', marginTop: 120 }} />

  const roleCodes: string[] = (user as any)?.roles?.map((r: any) => r.role_code || r.role?.code || r.role_name) || []
  const roleNames: string[] = (user as any)?.roles?.map((r: any) => r.role_name || r.role?.name).filter(Boolean) || []
  const primaryRole = roleCodes[0] || 'user'

  const handleUpdateInfo = async (values: any) => {
    setInfoLoading(true)
    try {
      await client.put('/current/user', {
        name: values.name, phone: values.phone, email: values.email,
        avatar: values.avatar, bio: values.bio, remark: values.remark,
      })
      message.success('个人信息已更新')
      dispatch(fetchUserInfo())
    } catch (err: any) { message.error(err.message) }
    setInfoLoading(false)
  }

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const allowed = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
    if (!allowed.includes(file.type)) { message.error('仅支持 JPG、PNG、GIF、WebP 格式'); return }
    if (file.size > 5 * 1024 * 1024) { message.error('图片大小不能超过 5MB'); return }

    setUploading(true)
    try {
      const fd = new FormData(); fd.append('file', file); fd.append('category', 'avatar')
      const res = await client.post('/images/upload', fd)
      const url = res.data.data.url
      setAvatarUrl(url); infoForm.setFieldsValue({ avatar: url })
      await client.put('/current/user', {
        name: user.name, phone: user.phone || '', email: user.email || '',
        avatar: url, bio: user.bio || '', remark: user.remark || '',
      })
      dispatch(fetchUserInfo())
      message.success('头像已更新')
    } catch (err: any) { message.error(err.message || '上传失败') }
    finally { setUploading(false); if (fileInputRef.current) fileInputRef.current.value = '' }
  }

  const handleChangePwd = async (values: any) => {
    setPwdLoading(true)
    try {
      await client.put('/current/password', { old_password: values.old_password, new_password: values.new_password })
      message.success('密码已修改')
      pwdForm.resetFields(); setShowPwdCard(false)
    } catch (err: any) { message.error(err.message) }
    setPwdLoading(false)
  }

  return (
    <div style={{ maxWidth: 820, margin: '0 auto' }}>

      {/* ── 顶部横幅 ── */}
      <div style={{
        ...glassCard('linear-gradient(135deg, rgba(99,102,241,0.12), rgba(139,92,246,0.08), rgba(20,20,40,0.6))'),
        padding: '32px 36px', marginBottom: 24, display: 'flex', alignItems: 'center', gap: 28,
        position: 'relative', overflow: 'hidden',
      }}>
        {/* 装饰圆 */}
        <div style={{ position: 'absolute', top: -60, right: -40, width: 200, height: 200, borderRadius: '50%', background: 'rgba(99,102,241,0.06)', pointerEvents: 'none' }} />
        <div style={{ position: 'absolute', bottom: -30, left: '30%', width: 120, height: 120, borderRadius: '50%', background: 'rgba(139,92,246,0.05)', pointerEvents: 'none' }} />

        <div onClick={() => fileInputRef.current?.click()} style={{ cursor: 'pointer', position: 'relative', flexShrink: 0, zIndex: 1 }}>
          <Avatar
            size={90}
            src={avatarUrl || undefined}
            icon={!avatarUrl ? <UserOutlined /> : undefined}
            style={{
              backgroundColor: avatarUrl ? undefined : '#6366f1',
              borderRadius: 20,
              boxShadow: '0 8px 32px rgba(99,102,241,0.25)',
            }}
          />
          <div style={{
            position: 'absolute', bottom: -2, right: -2, width: 32, height: 32,
            borderRadius: 10, background: '#6366f1', display: 'flex',
            alignItems: 'center', justifyContent: 'center',
            border: '3px solid rgba(20,20,40,1)', boxShadow: '0 4px 12px rgba(99,102,241,0.4)',
          }}>
            {uploading ? <Spin size="small" /> : <CameraOutlined style={{ color: '#fff', fontSize: 14 }} />}
          </div>
        </div>
        <input ref={fileInputRef} type="file" accept="image/jpeg,image/png,image/gif,image/webp"
          style={{ display: 'none' }} onChange={handleFileChange} />

        <div style={{ position: 'relative', zIndex: 1, flex: 1 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 6 }}>
            <Title level={3} style={{ margin: 0, fontWeight: 700, letterSpacing: '-0.02em' }}>
              {user.name || user.username}
            </Title>
            {roleBadge(primaryRole)}
          </div>
          <Text style={{ color: 'rgba(255,255,255,0.45)', fontSize: 14 }}>
            @{user.username} · {user.email || '未设置邮箱'}
          </Text>
          {user.bio && (
            <Paragraph type="secondary" style={{ margin: '8px 0 0', fontSize: 13, maxWidth: 400, color: 'rgba(255,255,255,0.35)' }}
              ellipsis={{ rows: 2 }}>
              {user.bio}
            </Paragraph>
          )}
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20, marginBottom: 24 }}>

        {/* ── 账户信息卡 ── */}
        <div style={{ ...glassCard(), padding: '20px 24px' }}>
          <div style={{ fontSize: 13, fontWeight: 600, color: 'rgba(255,255,255,0.5)', marginBottom: 16, letterSpacing: '0.06em', textTransform: 'uppercase' }}>
            <IdcardOutlined style={{ marginRight: 8 }} />账户信息
          </div>
          <Descriptions column={1} size="small" colon={false}
            labelStyle={{ color: 'rgba(255,255,255,0.35)', fontSize: 12 }}
            contentStyle={{ color: 'rgba(255,255,255,0.75)', fontSize: 13 }}>
            <Descriptions.Item label="用户 ID">{user.id}</Descriptions.Item>
            <Descriptions.Item label="用户名">{user.username}</Descriptions.Item>
            <Descriptions.Item label="邮箱">{user.email || '—'}</Descriptions.Item>
            <Descriptions.Item label="手机">{user.phone || '—'}</Descriptions.Item>
            <Descriptions.Item label="角色">
              <span style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                {roleCodes.map((c: string) => roleBadge(c))}
                {roleCodes.length === 0 && <Text type="secondary">—</Text>}
              </span>
            </Descriptions.Item>
          </Descriptions>
        </div>

        {/* ── 活动信息卡 ── */}
        <div style={{ ...glassCard(), padding: '20px 24px' }}>
          <div style={{ fontSize: 13, fontWeight: 600, color: 'rgba(255,255,255,0.5)', marginBottom: 16, letterSpacing: '0.06em', textTransform: 'uppercase' }}>
            <ClockCircleOutlined style={{ marginRight: 8 }} />活动记录
          </div>
          <Descriptions column={1} size="small" colon={false}
            labelStyle={{ color: 'rgba(255,255,255,0.35)', fontSize: 12 }}
            contentStyle={{ color: 'rgba(255,255,255,0.75)', fontSize: 13 }}>
            <Descriptions.Item label="最后登录">
              {user.last_login_at ? dayjs(user.last_login_at).format('YYYY-MM-DD HH:mm:ss') : '—'}
            </Descriptions.Item>
            <Descriptions.Item label="登录 IP">
              <DesktopOutlined style={{ marginRight: 6, color: 'rgba(255,255,255,0.3)' }} />
              {user.last_login_ip || '—'}
            </Descriptions.Item>
            <Descriptions.Item label="注册时间">
              {(user as any).created_at ? dayjs((user as any).created_at).format('YYYY-MM-DD') : '—'}
            </Descriptions.Item>
            <Descriptions.Item label="粉丝">
              <span style={{ color: '#f59e0b', fontWeight: 600 }}>{user.follower_count || 0}</span>
              <span style={{ color: 'rgba(255,255,255,0.25)', margin: '0 8px' }}>|</span>
              <span style={{ color: '#3b82f6', fontWeight: 600 }}>{user.following_count || 0}</span>
              <span style={{ color: 'rgba(255,255,255,0.25)', marginLeft: 4, fontSize: 12 }}>关注</span>
            </Descriptions.Item>
          </Descriptions>
        </div>
      </div>

      {/* ── 编辑个人信息 ── */}
      <div style={{ ...glassCard(), padding: '24px 28px', marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20 }}>
          <div style={{ fontSize: 14, fontWeight: 600, color: 'rgba(255,255,255,0.7)', letterSpacing: '0.04em' }}>
            <EditOutlined style={{ marginRight: 8, color: '#6366f1' }} />编辑个人信息
          </div>
          <Tooltip title={showPwdCard ? '取消修改密码' : '修改密码'}>
            <Button
              type={showPwdCard ? 'default' : 'primary'}
              ghost={!showPwdCard}
              size="small"
              icon={<LockOutlined />}
              onClick={() => setShowPwdCard(!showPwdCard)}
              style={{ borderRadius: 8 }}
            >
              {showPwdCard ? '取消' : '改密'}
            </Button>
          </Tooltip>
        </div>

        <Form form={infoForm} layout="vertical" onFinish={handleUpdateInfo}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0 28px' }}>
            <Form.Item label={<span style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>用户名</span>}>
              <Input value={user.username} disabled style={{ borderRadius: 8 }} />
            </Form.Item>
            <Form.Item name="name" label={<span style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>显示名称</span>} rules={[{ required: true }]}>
              <Input placeholder="你的显示名称" style={{ borderRadius: 8 }} />
            </Form.Item>
            <Form.Item name="email" label={<span style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}><MailOutlined style={{ marginRight: 6 }} />邮箱</span>}>
              <Input placeholder="your@email.com" style={{ borderRadius: 8 }} />
            </Form.Item>
            <Form.Item name="phone" label={<span style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}><PhoneOutlined style={{ marginRight: 6 }} />手机</span>}>
              <Input placeholder="手机号码" style={{ borderRadius: 8 }} />
            </Form.Item>
          </div>
          <Form.Item name="avatar" label={<span style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}><GlobalOutlined style={{ marginRight: 6 }} />头像链接</span>}
            help={<span style={{ color: 'rgba(255,255,255,0.2)', fontSize: 11 }}>可直接粘贴图片 URL，也可点击上方头像上传</span>}>
            <Input placeholder="https://example.com/avatar.jpg" style={{ borderRadius: 8 }}
              onChange={(e) => setAvatarUrl(e.target.value)} />
          </Form.Item>
          <Form.Item name="bio" label={<span style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}><SignatureOutlined style={{ marginRight: 6 }} />个人简介</span>}>
            <Input.TextArea rows={3} maxLength={512} showCount style={{ borderRadius: 8 }}
              placeholder="简短介绍一下自己..." />
          </Form.Item>
          <Form.Item name="remark" label={<span style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>备注</span>}>
            <Input.TextArea rows={2} style={{ borderRadius: 8 }} placeholder="仅自己可见" />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit" loading={infoLoading}
              icon={<CheckCircleOutlined />} style={{ borderRadius: 8, height: 38, paddingInline: 28 }}>
              保存修改
            </Button>
          </Form.Item>
        </Form>
      </div>

      {/* ── 修改密码 ── */}
      {showPwdCard && (
        <div style={{ ...glassCard('rgba(99,102,241,0.04)'), padding: '24px 28px', marginBottom: 24, border: '1px solid rgba(99,102,241,0.12)' }}>
          <div style={{ fontSize: 14, fontWeight: 600, color: 'rgba(255,255,255,0.7)', marginBottom: 20, letterSpacing: '0.04em' }}>
            <LockOutlined style={{ marginRight: 8, color: '#6366f1' }} />修改密码
          </div>
          <Form form={pwdForm} layout="vertical" onFinish={handleChangePwd}>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '0 20px' }}>
              <Form.Item name="old_password" label={<span style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>当前密码</span>} rules={[{ required: true, message: '请输入当前密码' }]}>
                <Input.Password placeholder="输入当前密码" style={{ borderRadius: 8 }} />
              </Form.Item>
              <Form.Item name="new_password" label={<span style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>新密码</span>} rules={[{ required: true, message: '请输入新密码' }]}>
                <Input.Password placeholder="输入新密码" style={{ borderRadius: 8 }} />
              </Form.Item>
              <Form.Item name="confirm_password" label={<span style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>确认密码</span>}
                dependencies={['new_password']}
                rules={[
                  { required: true, message: '请确认新密码' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('new_password') === value) return Promise.resolve()
                      return Promise.reject(new Error('两次密码不一致'))
                    },
                  }),
                ]}>
                <Input.Password placeholder="再次输入新密码" style={{ borderRadius: 8 }} />
              </Form.Item>
            </div>
            <Form.Item style={{ marginBottom: 0 }}>
              <Button type="primary" htmlType="submit" loading={pwdLoading} danger
                icon={<LockOutlined />} style={{ borderRadius: 8, height: 38, paddingInline: 28 }}>
                更新密码
              </Button>
            </Form.Item>
          </Form>
        </div>
      )}

      <Divider style={{ borderColor: 'rgba(255,255,255,0.04)', margin: '8px 0 32px' }} />
    </div>
  )
}

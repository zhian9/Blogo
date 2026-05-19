import { useState, useRef, useCallback } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Form, Input, Button, Typography, message, Avatar, Spin } from 'antd'
import {
  ArrowLeftOutlined, UserOutlined, PhoneOutlined, MailOutlined,
  EditOutlined, SaveOutlined, CameraOutlined, UploadOutlined,
  CloseOutlined, CheckCircleFilled, LockOutlined, KeyOutlined,
} from '@ant-design/icons'
import { motion, AnimatePresence } from 'framer-motion'
import { useAuthStore } from '../store/authStore'
import { useAppStore } from '../store/appStore'
import { updateCurrentUser, uploadAvatar, changePassword } from '../api/auth'

const { Text } = Typography
const { TextArea } = Input

export default function EditProfile() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)
  const theme = useAppStore((s) => s.theme)
  const isDark = theme === 'dark'

  const [loading, setLoading] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [avatarUrl, setAvatarUrl] = useState(user?.avatar || '')
  const [dragOver, setDragOver] = useState(false)
  const [pwdLoading, setPwdLoading] = useState(false)
  const [showPwdForm, setShowPwdForm] = useState(false)
  const [pwdValues, setPwdValues] = useState({ old_password: '', new_password: '', confirm_password: '' })
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [form] = Form.useForm()

  // Reactive form values for live preview
  const watchedName = Form.useWatch('name', form)
  const watchedBio = Form.useWatch('bio', form)
  const watchedAvatar = Form.useWatch('avatar', form)

  const previewAvatar = watchedAvatar || avatarUrl
  const previewName = watchedName || user?.name || user?.username || ''
  const previewBio = (watchedBio ?? user?.bio ?? '')

  if (!user) {
    return (
      <div style={{
        display: 'flex', justifyContent: 'center', alignItems: 'center',
        minHeight: '60vh', flexDirection: 'column', gap: 16,
      }}>
        <h2 style={{ color: isDark ? '#e8e8f0' : '#1a1a2e' }}>请先登录</h2>
        <Link to="/login">
          <Button type="primary" size="large" style={{ borderRadius: 10 }}>前往登录</Button>
        </Link>
      </div>
    )
  }

  const handleAvatarClick = () => {
    fileInputRef.current?.click()
  }

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    await uploadFile(file)
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  const uploadFile = async (file: File) => {
    const allowed = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
    if (!allowed.includes(file.type)) {
      message.error('仅支持 JPG、PNG、GIF、WebP 格式')
      return
    }
    if (file.size > 5 * 1024 * 1024) {
      message.error('图片大小不能超过 5MB')
      return
    }
    setUploading(true)
    try {
      const res = await uploadAvatar(file)
      const url = res.data.url
      setAvatarUrl(url)
      form.setFieldsValue({ avatar: url })
      message.success('头像上传成功')
    } catch (err: any) {
      message.error(err.message || '上传失败')
    } finally {
      setUploading(false)
    }
  }

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setDragOver(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setDragOver(false)
  }, [])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setDragOver(false)
    const file = e.dataTransfer.files?.[0]
    if (file) uploadFile(file)
  }, [])

  const onFinish = async (values: any) => {
    setLoading(true)
    try {
      await updateCurrentUser({
        name: values.name?.trim() || '',
        phone: values.phone?.trim() || '',
        email: values.email?.trim() || '',
        avatar: values.avatar?.trim() || '',
        bio: values.bio?.trim() || '',
        remark: '',
      })
      setUser({ ...user, ...values })
      message.success('个人信息已更新')
      navigate(`/user/${user.id}`)
    } catch (err: any) {
      message.error(err.message || '更新失败')
    } finally {
      setLoading(false)
    }
  }

  const handleChangePassword = async () => {
    if (!pwdValues.old_password) { message.warning('请输入当前密码'); return }
    if (!pwdValues.new_password) { message.warning('请输入新密码'); return }
    if (pwdValues.new_password.length < 6) { message.warning('新密码至少 6 个字符'); return }
    if (pwdValues.new_password !== pwdValues.confirm_password) { message.warning('两次输入的新密码不一致'); return }

    setPwdLoading(true)
    try {
      await changePassword({ old_password: pwdValues.old_password, new_password: pwdValues.new_password })
      message.success('密码修改成功，即将跳转登录页')
      setPwdValues({ old_password: '', new_password: '', confirm_password: '' })
      setShowPwdForm(false)
      setTimeout(() => {
        useAuthStore.getState().logout()
        navigate('/login')
      }, 1500)
    } catch (err: any) {
      const msg = err.message || ''
      if (msg.includes('old password') || msg.includes('Incorrect')) {
        message.error('当前密码错误，请重试')
      } else {
        message.error(msg || '密码修改失败，请重试')
      }
    } finally {
      setPwdLoading(false)
    }
  }

  // ── Colors ──
  const colors = {
    bg: isDark ? '#0f0f0f' : '#f5f7fa',
    surface: isDark ? '#1a1a2e' : '#ffffff',
    surfaceAlt: isDark ? '#16213e' : '#f0f4ff',
    text: isDark ? '#e8e8f0' : '#1a1a2e',
    textSecondary: isDark ? '#8892b0' : '#6b7280',
    textMuted: isDark ? '#5a6080' : '#9ca3af',
    border: isDark ? '#2a2a4a' : '#e5e7eb',
    accent: '#4f6ef7',
    accentLight: isDark ? 'rgba(79,110,247,0.15)' : 'rgba(79,110,247,0.08)',
    gradientFrom: isDark ? '#0f0c29' : '#e8f0fe',
    gradientVia: isDark ? '#1a1a4e' : '#dbeafe',
    gradientTo: isDark ? '#24243e' : '#eff6ff',
    inputBg: isDark ? 'rgba(255,255,255,0.05)' : '#f9fafb',
    inputBorder: isDark ? '#2a2a4a' : '#e5e7eb',
    inputFocusBorder: '#4f6ef7',
    danger: '#ef4444',
    success: '#10b981',
  }

  // ── Animation variants ──
  const containerVariants = {
    hidden: { opacity: 0 },
    visible: { opacity: 1, transition: { duration: 0.5, staggerChildren: 0.08 } },
  }
  const fadeIn = {
    hidden: { opacity: 0, y: 20 },
    visible: { opacity: 1, y: 0, transition: { duration: 0.5, ease: 'easeOut' } },
  }
  const slideInLeft = {
    hidden: { opacity: 0, x: -30 },
    visible: { opacity: 1, x: 0, transition: { duration: 0.55, ease: 'easeOut' } },
  }
  const slideInRight = {
    hidden: { opacity: 0, x: 30 },
    visible: { opacity: 1, x: 0, transition: { duration: 0.55, ease: 'easeOut' } },
  }

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={containerVariants}
      style={{ margin: -24, marginTop: -24 }}
    >
      {/* ==================== HERO HEADER ==================== */}
      <motion.div
        variants={fadeIn}
        style={{
          position: 'relative', width: '100%',
          background: `linear-gradient(160deg, ${colors.gradientFrom} 0%, ${colors.gradientVia} 40%, ${colors.gradientTo} 100%)`,
          overflow: 'hidden',
          padding: '36px 6% 28px',
        }}
      >
        <div style={{
          position: 'absolute', top: -80, right: -40,
          width: 300, height: 300, borderRadius: '50%',
          background: isDark ? 'rgba(79,110,247,0.07)' : 'rgba(79,110,247,0.05)',
          pointerEvents: 'none',
        }} />
        <div style={{
          position: 'absolute', inset: 0,
          backgroundImage: `radial-gradient(circle, ${isDark ? 'rgba(255,255,255,0.03)' : 'rgba(0,0,0,0.02)'} 1px, transparent 1px)`,
          backgroundSize: '20px 20px', pointerEvents: 'none',
        }} />

        <motion.div whileHover={{ x: -3 }} style={{ position: 'relative', zIndex: 2, marginBottom: 12 }}>
          <Button
            type="text" icon={<ArrowLeftOutlined />}
            onClick={() => navigate(-1)}
            style={{
              color: isDark ? '#8892b0' : '#6b7280', fontSize: 14,
              padding: '4px 12px', borderRadius: 8,
              background: isDark ? 'rgba(255,255,255,0.05)' : 'rgba(255,255,255,0.6)',
              backdropFilter: 'blur(8px)', border: 'none',
            }}
          >
            返回
          </Button>
        </motion.div>

        <div style={{ position: 'relative', zIndex: 2 }}>
          <h1 style={{ margin: '0 0 4px', fontSize: 28, fontWeight: 800, color: colors.text, letterSpacing: '-0.5px' }}>
            编辑个人资料
          </h1>
          <Text style={{ color: colors.textMuted, fontSize: 14 }}>
            完善你的个人信息，让更多人认识你
          </Text>
        </div>
      </motion.div>

      {/* ==================== MAIN CONTENT: SPLIT LAYOUT ==================== */}
      <div style={{
        background: colors.bg,
        padding: '32px 6% 60px',
      }}>
        <div style={{
          display: 'flex', gap: 32,
          alignItems: 'flex-start',
        }}
          className="edit-profile-split"
        >
          {/* ========== LEFT: LIVE PREVIEW CARD (35-40%) ========== */}
          <motion.div
            variants={slideInLeft}
            style={{
              width: '38%', flexShrink: 0, position: 'sticky', top: 80,
            }}
            className="edit-profile-preview"
          >
            <Text style={{
              display: 'block', fontSize: 12, fontWeight: 600, color: colors.textMuted,
              textTransform: 'uppercase', letterSpacing: 1, marginBottom: 12,
            }}>
              预览效果
            </Text>

            <div style={{
              background: colors.surface,
              borderRadius: 24,
              overflow: 'hidden',
              border: `1px solid ${colors.border}`,
              boxShadow: isDark
                ? '0 4px 32px rgba(0,0,0,0.3)'
                : '0 4px 32px rgba(0,0,0,0.06)',
            }}>
              {/* Preview cover area */}
              <div style={{
                height: 120,
                background: `linear-gradient(135deg, #4f6ef7 0%, #8b5cf6 40%, #ec4899 100%)`,
                position: 'relative',
                overflow: 'hidden',
              }}>
                <div style={{
                  position: 'absolute', inset: 0,
                  backgroundImage: 'radial-gradient(circle, rgba(255,255,255,0.15) 1px, transparent 1px)',
                  backgroundSize: '16px 16px',
                }} />
              </div>

              {/* Avatar - overlaps the cover */}
              <div style={{
                display: 'flex', justifyContent: 'center', marginTop: -52,
                position: 'relative', zIndex: 2,
                cursor: 'pointer',
              }}
                onClick={handleAvatarClick}
              >
                <div style={{
                  width: 104, height: 104, borderRadius: '50%',
                  padding: 4,
                  background: colors.surface,
                  boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
                }}>
                  <AnimatePresence mode="wait">
                    <motion.div
                      key={previewAvatar || 'placeholder'}
                      initial={{ opacity: 0, scale: 0.8 }}
                      animate={{ opacity: 1, scale: 1 }}
                      transition={{ duration: 0.3 }}
                    >
                      <Avatar
                        size={96}
                        src={previewAvatar || undefined}
                        icon={!previewAvatar ? <UserOutlined /> : undefined}
                        style={{
                          backgroundColor: previewAvatar ? undefined : '#4f6ef7',
                          fontSize: 40,
                        }}
                      />
                    </motion.div>
                  </AnimatePresence>
                </div>
              </div>

              {/* Preview content */}
              <div style={{ padding: '16px 24px 28px', textAlign: 'center' }}>
                <AnimatePresence mode="wait">
                  <motion.h2
                    key={previewName}
                    initial={{ opacity: 0, y: -6 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.25 }}
                    style={{
                      margin: '0 0 4px', fontSize: 20, fontWeight: 700,
                      color: colors.text, wordBreak: 'break-word',
                    }}
                  >
                    {previewName || '你的昵称'}
                  </motion.h2>
                </AnimatePresence>

                <Text style={{ color: colors.textMuted, fontSize: 13 }}>
                  @{user.username}
                </Text>

                <AnimatePresence mode="wait">
                  <motion.p
                    key={previewBio}
                    initial={{ opacity: 0, y: -4 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.25 }}
                    style={{
                      margin: '14px 0 0', fontSize: 14, color: colors.textSecondary,
                      lineHeight: 1.7, wordBreak: 'break-word',
                      minHeight: 24,
                    }}
                  >
                    {previewBio || '你的个性签名将显示在这里'}
                  </motion.p>
                </AnimatePresence>

                {/* Mini stats */}
                <div style={{
                  display: 'flex', gap: 20, justifyContent: 'center',
                  marginTop: 18, paddingTop: 16,
                  borderTop: `1px solid ${colors.border}`,
                }}>
                  <div style={{ textAlign: 'center' }}>
                    <div style={{ fontSize: 16, fontWeight: 700, color: colors.text }}>
                      {user.follower_count || 0}
                    </div>
                    <div style={{ fontSize: 11, color: colors.textMuted }}>粉丝</div>
                  </div>
                  <div style={{ textAlign: 'center' }}>
                    <div style={{ fontSize: 16, fontWeight: 700, color: colors.text }}>
                      {user.following_count || 0}
                    </div>
                    <div style={{ fontSize: 11, color: colors.textMuted }}>关注</div>
                  </div>
                </div>

                {/* Real-time update indicator */}
                <div style={{
                  marginTop: 14, display: 'flex',
                  alignItems: 'center', justifyContent: 'center', gap: 6,
                }}>
                  <CheckCircleFilled style={{ color: colors.success, fontSize: 10 }} />
                  <Text style={{ fontSize: 11, color: colors.textMuted }}>
                    实时预览
                  </Text>
                </div>
              </div>
            </div>
          </motion.div>

          {/* ========== RIGHT: FORM SECTION (60-65%) ========== */}
          <motion.div
            variants={slideInRight}
            style={{ flex: 1, minWidth: 0 }}
          >
            <Form
              form={form}
              layout="vertical"
              initialValues={{
                name: user.name || '',
                phone: user.phone || '',
                email: user.email || '',
                avatar: user.avatar || '',
                bio: user.bio || '',
              }}
              onFinish={onFinish}
              size="large"
            >
              {/* Avatar upload card */}
              <motion.div
                variants={fadeIn}
                style={{
                  background: colors.surface,
                  borderRadius: 20,
                  padding: 28,
                  border: `1px solid ${colors.border}`,
                  boxShadow: isDark
                    ? '0 2px 16px rgba(0,0,0,0.2)'
                    : '0 2px 16px rgba(0,0,0,0.04)',
                  marginBottom: 24,
                }}
              >
                <h3 style={{
                  margin: '0 0 20px', fontSize: 16, fontWeight: 700,
                  color: colors.text, display: 'flex', alignItems: 'center', gap: 8,
                }}>
                  <CameraOutlined style={{ color: colors.accent }} /> 头像设置
                </h3>

                <div style={{ display: 'flex', gap: 24, alignItems: 'flex-start', flexWrap: 'wrap' }}>
                  {/* Current avatar */}
                  <div style={{ flexShrink: 0 }}>
                    <Avatar
                      size={80}
                      src={previewAvatar || undefined}
                      icon={!previewAvatar ? <UserOutlined /> : undefined}
                      style={{
                        backgroundColor: previewAvatar ? undefined : '#4f6ef7',
                        borderRadius: 16, fontSize: 36,
                      }}
                    />
                  </div>

                  {/* Upload area */}
                  <div style={{ flex: 1, minWidth: 260 }}>
                    {/* Drag & drop zone */}
                    <div
                      onDragOver={handleDragOver}
                      onDragLeave={handleDragLeave}
                      onDrop={handleDrop}
                      onClick={handleAvatarClick}
                      style={{
                        border: `2px dashed ${dragOver ? colors.accent : colors.inputBorder}`,
                        borderRadius: 14,
                        padding: '24px 20px',
                        textAlign: 'center',
                        cursor: 'pointer',
                        background: dragOver ? colors.accentLight : colors.inputBg,
                        transition: 'all 0.25s ease',
                      }}
                    >
                      {uploading ? (
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 10 }}>
                          <Spin size="small" />
                          <Text style={{ color: colors.textMuted, fontSize: 14 }}>上传中...</Text>
                        </div>
                      ) : (
                        <>
                          <UploadOutlined style={{
                            fontSize: 28, color: dragOver ? colors.accent : colors.textMuted,
                            marginBottom: 8, display: 'block',
                          }} />
                          <Text style={{ color: colors.textSecondary, fontSize: 14, fontWeight: 500 }}>
                            拖拽图片到此处，或点击上传
                          </Text>
                          <Text style={{ color: colors.textMuted, fontSize: 12, display: 'block', marginTop: 4 }}>
                            支持 JPG、PNG、GIF、WebP，最大 5MB
                          </Text>
                        </>
                      )}
                    </div>
                  </div>
                </div>

                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/jpeg,image/png,image/gif,image/webp"
                  style={{ display: 'none' }}
                  onChange={handleFileChange}
                />

                {/* Avatar URL field */}
                <div style={{ marginTop: 16 }}>
                  <Form.Item
                    name="avatar"
                    label={<Text style={{ color: colors.textSecondary, fontSize: 13, fontWeight: 500 }}>头像链接</Text>}
                    style={{ marginBottom: 0 }}
                    rules={[{ max: 512, message: '最多 512 个字符' }]}
                  >
                    <Input
                      placeholder="https://example.com/avatar.jpg"
                      style={{
                        borderRadius: 10, height: 42,
                        background: colors.inputBg,
                        borderColor: colors.inputBorder,
                        color: colors.text, fontSize: 14,
                      }}
                    />
                  </Form.Item>
                </div>
              </motion.div>

              {/* Basic info card */}
              <motion.div
                variants={fadeIn}
                style={{
                  background: colors.surface,
                  borderRadius: 20,
                  padding: 28,
                  border: `1px solid ${colors.border}`,
                  boxShadow: isDark
                    ? '0 2px 16px rgba(0,0,0,0.2)'
                    : '0 2px 16px rgba(0,0,0,0.04)',
                  marginBottom: 24,
                }}
              >
                <h3 style={{
                  margin: '0 0 20px', fontSize: 16, fontWeight: 700,
                  color: colors.text, display: 'flex', alignItems: 'center', gap: 8,
                }}>
                  <EditOutlined style={{ color: colors.accent }} /> 基本信息
                </h3>

                {/* Nickname */}
                <FormItemWrapper label="昵称" required colors={colors}>
                  <Form.Item
                    name="name"
                    rules={[{ required: true, message: '请输入昵称' }, { max: 64, message: '最多 64 个字符' }]}
                    style={{ marginBottom: 0 }}
                  >
                    <Input
                      prefix={<EditOutlined style={{ color: colors.textMuted }} />}
                      placeholder="你的昵称"
                      maxLength={64}
                      style={inputStyle(colors)}
                    />
                  </Form.Item>
                </FormItemWrapper>

                {/* Bio */}
                <FormItemWrapper label="个性签名" colors={colors}>
                  <Form.Item
                    name="bio"
                    rules={[{ max: 512, message: '最多 512 个字符' }]}
                    style={{ marginBottom: 0 }}
                  >
                    <TextArea
                      rows={4}
                      placeholder="介绍一下自己..."
                      maxLength={512}
                      showCount={{
                        formatter: ({ count, maxLength: max }) => (
                          <Text style={{ fontSize: 12, color: colors.textMuted }}>
                            {count}/{max}
                          </Text>
                        ),
                      }}
                      style={{
                        ...inputStyle(colors),
                        resize: 'vertical',
                        minHeight: 100,
                      }}
                    />
                  </Form.Item>
                </FormItemWrapper>

                {/* Phone */}
                <FormItemWrapper label="手机号" colors={colors}>
                  <Form.Item
                    name="phone"
                    rules={[{ max: 32, message: '最多 32 个字符' }]}
                    style={{ marginBottom: 0 }}
                  >
                    <Input
                      prefix={<PhoneOutlined style={{ color: colors.textMuted }} />}
                      placeholder="你的手机号"
                      maxLength={32}
                      style={inputStyle(colors)}
                    />
                  </Form.Item>
                </FormItemWrapper>

                {/* Email */}
                <FormItemWrapper label="邮箱" colors={colors}>
                  <Form.Item
                    name="email"
                    rules={[{ max: 128, message: '最多 128 个字符' }, { type: 'email', message: '请输入有效的邮箱地址' }]}
                    style={{ marginBottom: 0 }}
                  >
                    <Input
                      prefix={<MailOutlined style={{ color: colors.textMuted }} />}
                      placeholder="your@email.com"
                      maxLength={128}
                      style={inputStyle(colors)}
                    />
                  </Form.Item>
                </FormItemWrapper>
              </motion.div>

              {/* Password change card */}
              <motion.div
                variants={fadeIn}
                style={{
                  background: colors.surface,
                  borderRadius: 20,
                  padding: 28,
                  border: `1px solid ${colors.border}`,
                  boxShadow: isDark
                    ? '0 2px 16px rgba(0,0,0,0.2)'
                    : '0 2px 16px rgba(0,0,0,0.04)',
                  marginBottom: 24,
                }}
              >
                <div style={{
                  display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                  marginBottom: showPwdForm ? 20 : 0,
                }}>
                  <h3 style={{
                    margin: 0, fontSize: 16, fontWeight: 700,
                    color: colors.text, display: 'flex', alignItems: 'center', gap: 8,
                  }}>
                    <LockOutlined style={{ color: colors.accent }} /> 修改密码
                  </h3>
                  {!showPwdForm && (
                    <Button type="link" onClick={() => setShowPwdForm(true)} style={{ color: colors.accent, fontWeight: 500 }}>
                      修改
                    </Button>
                  )}
                </div>

                {showPwdForm && (
                  <motion.div
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: 'auto' }}
                    transition={{ duration: 0.3 }}
                    style={{ overflow: 'hidden' }}
                  >
                    <FormItemWrapper label="当前密码" required colors={colors}>
                      <Input.Password
                        prefix={<KeyOutlined style={{ color: colors.textMuted }} />}
                        placeholder="输入当前密码"
                        value={pwdValues.old_password}
                        onChange={(e) => setPwdValues({ ...pwdValues, old_password: e.target.value })}
                        style={inputStyle(colors)}
                      />
                    </FormItemWrapper>

                    <FormItemWrapper label="新密码" required colors={colors}>
                      <Input.Password
                        prefix={<LockOutlined style={{ color: colors.textMuted }} />}
                        placeholder="输入新密码（至少 6 位）"
                        value={pwdValues.new_password}
                        onChange={(e) => setPwdValues({ ...pwdValues, new_password: e.target.value })}
                        style={inputStyle(colors)}
                      />
                    </FormItemWrapper>

                    <FormItemWrapper label="确认新密码" required colors={colors}>
                      <Input.Password
                        prefix={<LockOutlined style={{ color: colors.textMuted }} />}
                        placeholder="再次输入新密码"
                        value={pwdValues.confirm_password}
                        onChange={(e) => setPwdValues({ ...pwdValues, confirm_password: e.target.value })}
                        onPressEnter={handleChangePassword}
                        style={inputStyle(colors)}
                      />
                    </FormItemWrapper>

                    <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 16 }}>
                      <Button
                        onClick={() => { setShowPwdForm(false); setPwdValues({ old_password: '', new_password: '', confirm_password: '' }) }}
                        style={{ borderRadius: 10, color: colors.textSecondary }}
                      >
                        取消
                      </Button>
                      <Button
                        type="primary"
                        onClick={handleChangePassword}
                        loading={pwdLoading}
                        style={{
                          borderRadius: 10, fontWeight: 500,
                          background: 'linear-gradient(135deg, #4f6ef7, #6366f1)',
                          border: 'none',
                        }}
                      >
                        修改密码
                      </Button>
                    </div>
                  </motion.div>
                )}
              </motion.div>

              {/* Action buttons */}
              <motion.div
                variants={fadeIn}
                style={{
                  display: 'flex', gap: 12, justifyContent: 'flex-end',
                }}
              >
                <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.97 }}>
                  <Button
                    size="large"
                    icon={<CloseOutlined />}
                    onClick={() => navigate(-1)}
                    style={{
                      borderRadius: 12, height: 48, paddingInline: 24,
                      fontWeight: 500, fontSize: 15,
                      background: 'transparent',
                      border: `1px solid ${colors.border}`,
                      color: colors.textSecondary,
                    }}
                  >
                    取消
                  </Button>
                </motion.div>
                <motion.div whileHover={{ scale: 1.03 }} whileTap={{ scale: 0.97 }}>
                  <Button
                    type="primary"
                    htmlType="submit"
                    size="large"
                    icon={<SaveOutlined />}
                    loading={loading}
                    style={{
                      borderRadius: 12, height: 48, paddingInline: 32,
                      fontWeight: 600, fontSize: 15,
                      background: 'linear-gradient(135deg, #4f6ef7, #6366f1)',
                      border: 'none',
                      boxShadow: '0 4px 16px rgba(79,110,247,0.35)',
                    }}
                  >
                    保存修改
                  </Button>
                </motion.div>
              </motion.div>
            </Form>
          </motion.div>
        </div>
      </div>

      {/* ==================== RESPONSIVE CSS ==================== */}
      <style>{`
        @media (max-width: 1024px) {
          .edit-profile-split {
            flex-direction: column !important;
          }
          .edit-profile-preview {
            width: 100% !important;
            position: static !important;
          }
        }
      `}</style>
    </motion.div>
  )
}

// ── Helper: Form item wrapper with custom label ──
function FormItemWrapper({ label, required, colors, children }: {
  label: string
  required?: boolean
  colors: Record<string, string>
  children: React.ReactNode
}) {
  return (
    <div style={{ marginBottom: 20 }}>
      <div style={{ marginBottom: 6, display: 'flex', alignItems: 'center', gap: 4 }}>
        <Text style={{ color: colors.textSecondary, fontSize: 13, fontWeight: 500 }}>
          {label}
        </Text>
        {required && (
          <Text style={{ color: '#ef4444', fontSize: 13 }}>*</Text>
        )}
      </div>
      {children}
    </div>
  )
}

// ── Shared input style ──
function inputStyle(colors: Record<string, string>) {
  return {
    borderRadius: 10,
    height: 44,
    background: colors.inputBg,
    borderColor: colors.inputBorder,
    color: colors.text,
    fontSize: 14,
  }
}

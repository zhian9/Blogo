import { useState, useEffect, useRef } from 'react'
import { useNavigate, Link, useSearchParams } from 'react-router-dom'
import { Form, Input, Button, Card, Typography, message, Space, Row, Col, Checkbox, Modal } from 'antd'
import { UserOutlined, LockOutlined, ArrowLeftOutlined, SafetyCertificateOutlined, MailOutlined } from '@ant-design/icons'
import { login as loginApi, getCaptchaId, getCurrentUser } from '../api/auth'
import { useAuthStore } from '../store/authStore'
import client from '../api/client'

const { Title, Text } = Typography

export default function Login() {
  const [captchaId, setCaptchaId] = useState('')
  const [captchaImg, setCaptchaImg] = useState('')
  const [captchaLoading, setCaptchaLoading] = useState(true)
  const [loading, setLoading] = useState(false)
  const [remember, setRemember] = useState(false)
  const [searchParams] = useSearchParams()
  const { isAuthenticated, login: storeLogin } = useAuthStore()
  const navigate = useNavigate()

  // ── Forgot password modal ──
  const [fpOpen, setFpOpen] = useState(false)
  const [fpStep, setFpStep] = useState(0)
  const [fpEmail, setFpEmail] = useState('')
  const [fpCode, setFpCode] = useState('')
  const [fpNewPwd, setFpNewPwd] = useState('')
  const [fpConfirmPwd, setFpConfirmPwd] = useState('')
  const [fpLoading, setFpLoading] = useState(false)
  const [fpCountdown, setFpCountdown] = useState(0)
  const fpTimer = useRef<ReturnType<typeof setInterval> | null>(null)

  const redirectTo = searchParams.get('redirect') || '/'

  useEffect(() => {
    if (isAuthenticated) navigate(redirectTo, { replace: true })
  }, [isAuthenticated, navigate, redirectTo])

  const fetchCaptcha = async () => {
    setCaptchaLoading(true)
    try {
      const res = await getCaptchaId()
      const id = res.data.captcha_id
      setCaptchaId(id)
      setCaptchaImg(`/api/v1/captcha/image?id=${id}&t=${Date.now()}`)
    } catch (err: any) {
      message.error(err.message || '验证码加载失败')
    } finally { setCaptchaLoading(false) }
  }

  useEffect(() => { fetchCaptcha() }, [])

  const onFinish = async (values: { username: string; password: string; captcha_code: string }) => {
    if (!captchaId) { message.warning('验证码加载中'); return }
    setLoading(true)
    try {
      const res = await loginApi({
        username: values.username, password: values.password,
        captcha_id: captchaId, captcha_code: values.captcha_code.trim(),
      })
      const { access_token } = res.data
      sessionStorage.setItem('blog-token', access_token)
      if (remember) localStorage.setItem('blog-remember', values.username)
      else localStorage.removeItem('blog-remember')
      const userRes = await getCurrentUser()
      storeLogin(access_token, userRes.data)
      message.success('欢迎回来！')
      navigate(redirectTo, { replace: true })
    } catch (err: any) {
      message.error(err.message || '登录失败')
      fetchCaptcha()
    } finally { setLoading(false) }
  }

  const handleForgotSend = async () => {
    if (!fpEmail) { message.warning('请输入邮箱'); return }
    setFpLoading(true)
    try {
      await client.post('/auth/forgot-password', { email: fpEmail })
      message.success('验证码已发送，请查收邮件')
      setFpStep(1)
      setFpCountdown(60)
      fpTimer.current = setInterval(() => setFpCountdown(c => { if (c <= 1) { clearInterval(fpTimer.current!); return 0 }; return c - 1 }), 1000)
    } catch (err: any) { message.error(err.message || '发送失败') }
    finally { setFpLoading(false) }
  }

  const handleForgotReset = async () => {
    if (!fpCode || !fpNewPwd || !fpConfirmPwd) { message.warning('请填写完整'); return }
    if (fpNewPwd.length < 6) { message.warning('新密码至少6位'); return }
    if (fpNewPwd !== fpConfirmPwd) { message.warning('两次密码不一致'); return }
    setFpLoading(true)
    try {
      await client.post('/auth/reset-password', { email: fpEmail, code: fpCode, new_password: fpNewPwd })
      message.success('密码重置成功，请使用新密码登录！')
      setFpOpen(false)
    } catch (err: any) { message.error(err.message || '重置失败') }
    finally { setFpLoading(false) }
  }

  const closeForgotModal = () => {
    if (fpTimer.current) clearInterval(fpTimer.current)
    setFpOpen(false)
  }

  useEffect(() => {
    const saved = localStorage.getItem('blog-remember')
    if (saved) setRemember(true)
  }, [])

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'linear-gradient(135deg, #0f0f1a 0%, #1a1a2e 50%, #16213e 100%)', padding: 24 }}>
      <Card style={{ width: 420, borderRadius: 12, boxShadow: '0 8px 24px rgba(0,0,0,0.12)' }}>
        <Link to="/" style={{ marginBottom: 16, display: 'inline-flex', alignItems: 'center', gap: 4, color: '#666' }}>
          <ArrowLeftOutlined /> 返回首页
        </Link>

        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div style={{ textAlign: 'center' }}>
            <Title level={2} style={{ marginBottom: 4 }}>登录</Title>
            <Text type="secondary">欢迎回到 Blogo 博客</Text>
          </div>

          <Form layout="vertical" onFinish={onFinish} size="large" initialValues={{ username: localStorage.getItem('blog-remember') || '' }}>
            <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
              <Input prefix={<UserOutlined />} placeholder="用户名" />
            </Form.Item>
            <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
              <Input.Password prefix={<LockOutlined />} placeholder="密码" />
            </Form.Item>

            <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
              <Col><Checkbox checked={remember} onChange={e => setRemember(e.target.checked)}>记住我</Checkbox></Col>
              <Col><a onClick={() => { setFpStep(0); setFpEmail(''); setFpCode(''); setFpNewPwd(''); setFpConfirmPwd(''); setFpCountdown(0); setFpOpen(true) }} style={{ fontSize: 13 }}>忘记密码？</a></Col>
            </Row>

            <Form.Item name="captcha_code" rules={[{ required: true, message: '请输入验证码' }]}>
              <Row gutter={8} align="stretch">
                <Col span={14}>
                  <Input prefix={<SafetyCertificateOutlined />} placeholder="验证码" style={{ height: 48 }} />
                </Col>
                <Col span={10}>
                  {captchaImg ? (
                    <img src={captchaImg} alt="captcha" onClick={fetchCaptcha}
                      style={{ cursor: 'pointer', width: '100%', height: 48, borderRadius: 6, objectFit: 'contain', border: '1px solid #d9d9d9', background: '#fff' }} />
                  ) : (
                    <div onClick={fetchCaptcha} style={{ cursor: 'pointer', width: '100%', height: 48, borderRadius: 6, border: '1px solid #d9d9d9', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 12, color: 'rgba(0,0,0,0.25)' }}>加载中</div>
                  )}
                </Col>
              </Row>
            </Form.Item>

            <Form.Item style={{ marginBottom: 0 }}>
              <Button type="primary" htmlType="submit" loading={loading} disabled={captchaLoading || !captchaId} block
                style={{ height: 48, borderRadius: 8, fontSize: 15, fontWeight: 600 }}>登 录</Button>
            </Form.Item>
          </Form>

          <Text style={{ textAlign: 'center', display: 'block' }}>
            还没有账号？{' '}
            <Link to={`/register${redirectTo !== '/' ? `?redirect=${encodeURIComponent(redirectTo)}` : ''}`} style={{ color: '#1890ff' }}>立即注册</Link>
          </Text>
        </Space>
      </Card>

      <Modal title="找回密码" open={fpOpen} onCancel={closeForgotModal} footer={null} width={420} destroyOnClose>
        {fpStep === 0 ? (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Text type="secondary">请输入注册时绑定的邮箱，我们将发送验证码</Text>
            <Input prefix={<MailOutlined />} placeholder="请输入邮箱" value={fpEmail} onChange={e => setFpEmail(e.target.value)} size="large" style={{ borderRadius: 8 }} />
            <Button type="primary" block loading={fpLoading} onClick={handleForgotSend} size="large" style={{ borderRadius: 8, height: 44 }}>发送验证码</Button>
          </Space>
        ) : (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <div style={{ background: '#f6ffed', padding: '8px 12px', borderRadius: 6, fontSize: 12, color: '#52c41a' }}>
              验证码已发送至 <b>{fpEmail}</b>，5分钟内有效
            </div>
            <Input placeholder="6位验证码" prefix={<SafetyCertificateOutlined />} value={fpCode} onChange={e => setFpCode(e.target.value)} maxLength={6} size="large" style={{ borderRadius: 8 }} />
            <Input.Password placeholder="新密码（至少6位）" prefix={<LockOutlined />} value={fpNewPwd} onChange={e => setFpNewPwd(e.target.value)} size="large" style={{ borderRadius: 8 }} />
            <Input.Password placeholder="确认新密码" prefix={<LockOutlined />} value={fpConfirmPwd} onChange={e => setFpConfirmPwd(e.target.value)} size="large" style={{ borderRadius: 8 }} />
            <Space style={{ width: '100%' }} direction="vertical" size={8}>
              <Button type="primary" block loading={fpLoading} onClick={handleForgotReset} size="large" style={{ borderRadius: 8, height: 44 }}>重置密码</Button>
              <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                <a onClick={() => { setFpStep(0); setFpCountdown(0); if (fpTimer.current) clearInterval(fpTimer.current) }} style={{ fontSize: 12 }}>← 返回上一步</a>
                <Button size="small" disabled={fpCountdown > 0} onClick={handleForgotSend} loading={fpLoading} style={{ borderRadius: 6 }}>
                  {fpCountdown > 0 ? `${fpCountdown}秒后重新获取` : '重新获取验证码'}
                </Button>
              </Space>
            </Space>
          </Space>
        )}
      </Modal>
    </div>
  )
}

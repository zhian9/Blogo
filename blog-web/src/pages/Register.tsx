import { useState, useEffect } from 'react'
import { useNavigate, Link, useSearchParams } from 'react-router-dom'
import { Form, Input, Button, Card, Typography, message, Space, Row, Col } from 'antd'
import { UserOutlined, LockOutlined, MailOutlined, PhoneOutlined, ArrowLeftOutlined, SafetyCertificateOutlined } from '@ant-design/icons'
import { register as registerApi, getCaptchaId, getCurrentUser } from '../api/auth'
import { useAuthStore } from '../store/authStore'
import client from '../api/client'

const { Title, Text } = Typography

export default function Register() {
  const [captchaId, setCaptchaId] = useState('')
  const [captchaImg, setCaptchaImg] = useState('')
  const [captchaLoading, setCaptchaLoading] = useState(true)
  const [loading, setLoading] = useState(false)
  const [emailAvailable, setEmailAvailable] = useState<boolean | null>(null)
  const [checkingEmail, setCheckingEmail] = useState(false)
  const [searchParams] = useSearchParams()
  const { isAuthenticated, login: storeLogin } = useAuthStore()
  const navigate = useNavigate()

  const redirectTo = searchParams.get('redirect') || '/'

  useEffect(() => {
    if (isAuthenticated) navigate(redirectTo, { replace: true })
  }, [isAuthenticated, navigate, redirectTo])

  const checkEmail = async (email: string) => {
    if (!email || !email.includes('@')) { setEmailAvailable(null); return }
    setCheckingEmail(true)
    try {
      const res = await client.get('/auth/check-email', { params: { email } })
      setEmailAvailable(res.data.data?.available !== false)
    } catch { setEmailAvailable(null) }
    finally { setCheckingEmail(false) }
  }

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

  const onFinish = async (values: any) => {
    if (values.password !== values.confirm_password) {
      message.error('两次密码不一致')
      return
    }
    if (!captchaId) { message.warning('验证码加载中，请稍候'); return }
    setLoading(true)
    try {
      const res = await registerApi({
        username: values.username, password: values.password,
        confirm_password: values.confirm_password, phone: values.phone,
        email: values.email, captcha_id: captchaId,
        captcha_code: values.captcha_code?.trim() || '',
      })
      const { access_token } = res.data
      sessionStorage.setItem('blog-token', access_token)
      const userRes = await getCurrentUser()
      storeLogin(access_token, userRes.data)
      message.success('注册成功！')
      navigate(redirectTo, { replace: true })
    } catch (err: any) {
      message.error(err.message || '注册失败')
      fetchCaptcha()
    } finally { setLoading(false) }
  }

  const inputStyle = { height: 48, borderRadius: 10 }

  const formItemStyle = { marginBottom: 14 }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'linear-gradient(135deg, #0f0f1a 0%, #1a1a2e 50%, #16213e 100%)', padding: 24 }}>
      <Card
        style={{ width: 480, borderRadius: 12, boxShadow: '0 8px 24px rgba(0,0,0,0.12)', transition: 'box-shadow 0.3s' }}
        styles={{ body: { padding: '22px 36px 20px' } }}
      >
        <Link to="/" style={{ marginBottom: 12, display: 'inline-flex', alignItems: 'center', gap: 4, color: '#999', fontSize: 13 }}>
          <ArrowLeftOutlined /> 返回首页
        </Link>

        <div style={{ textAlign: 'center', marginBottom: 20 }}>
          <Title level={2} style={{ marginBottom: 4 }}>创建账号</Title>
          <Text type="secondary">加入 Blogo 社区</Text>
        </div>

        <Form layout="vertical" onFinish={onFinish} size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]} style={formItemStyle}>
            <Input prefix={<UserOutlined />} placeholder="用户名" style={inputStyle} />
          </Form.Item>

          <Form.Item name="email" rules={[
            { required: true, message: '请输入邮箱' }, { type: 'email', message: '邮箱格式不正确' },
          ]} style={formItemStyle}
            hasFeedback validateStatus={emailAvailable === true ? 'success' : emailAvailable === false ? 'error' : checkingEmail ? 'validating' : undefined}
            help={emailAvailable === false ? '该邮箱已被注册，请直接登录或更换其他邮箱' : undefined}
          >
            <Input prefix={<MailOutlined />} placeholder="邮箱" style={inputStyle} onBlur={e => checkEmail(e.target.value)} />
          </Form.Item>

          <Form.Item name="password" rules={[
            { required: true, message: '请输入密码' }, { min: 6, message: '密码至少6位' },
          ]} style={formItemStyle}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" style={inputStyle} />
          </Form.Item>

          <Form.Item name="confirm_password" rules={[{ required: true, message: '请确认密码' }]} style={formItemStyle}>
            <Input.Password prefix={<LockOutlined />} placeholder="确认密码" style={inputStyle} />
          </Form.Item>

          <Form.Item name="phone" rules={[{ required: true, message: '请输入手机号' }]} style={formItemStyle}>
            <Input prefix={<PhoneOutlined />} placeholder="手机号" style={inputStyle} />
          </Form.Item>

          <Form.Item name="captcha_code" rules={[{ required: true, message: '请输入验证码' }]} style={{ marginBottom: 8 }}>
            <Row gutter={8} align="stretch">
              <Col span={14}>
                <Input prefix={<SafetyCertificateOutlined />} placeholder="验证码" style={{ height: 48 }} />
              </Col>
              <Col span={10}>
                {captchaImg ? (
                  <img src={captchaImg} alt="验证码" onClick={fetchCaptcha}
                    style={{ cursor: 'pointer', width: '100%', height: 48, borderRadius: 6, objectFit: 'contain', border: '1px solid #d9d9d9', background: '#fff' }} />
                ) : (
                  <div onClick={fetchCaptcha}
                    style={{ cursor: 'pointer', width: '100%', height: 48, borderRadius: 6, border: '1px solid #d9d9d9', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 12, color: 'rgba(0,0,0,0.25)' }}>
                    加载中
                  </div>
                )}
              </Col>
            </Row>
          </Form.Item>

          <Form.Item style={{ marginBottom: 12 }}>
            <Button type="primary" htmlType="submit" loading={loading} disabled={captchaLoading || !captchaId || emailAvailable === false} block
              style={{ height: 48, borderRadius: 8, fontSize: 15, fontWeight: 600, transition: 'all 0.25s' }}>
              创建账号
            </Button>
          </Form.Item>
        </Form>

        <Text style={{ textAlign: 'center', display: 'block', color: 'rgba(0,0,0,0.45)', fontSize: 13 }}>
          已有账号？{' '}
          <Link to={`/login${redirectTo !== '/' ? `?redirect=${encodeURIComponent(redirectTo)}` : ''}`}
            style={{ color: '#1890ff', fontWeight: 500 }}>
            去登录
          </Link>
        </Text>
      </Card>
    </div>
  )
}

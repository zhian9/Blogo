import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Form, Input, Button, Card, Typography, message, Space, Row, Col } from 'antd'
import { UserOutlined, LockOutlined, SafetyCertificateOutlined } from '@ant-design/icons'
import { useAppDispatch, useAppSelector } from '../../store'
import { loginAsync } from '../../store/authSlice'
import { getUserRoleCode } from '../../components/ProtectedRoute'
import client from '../../api/client'

const { Title, Text } = Typography

export default function LoginPage() {
  const [captchaId, setCaptchaId] = useState('')
  const [captchaImg, setCaptchaImg] = useState('')
  const [loading, setLoading] = useState(false)
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const authLoading = useAppSelector((s) => s.auth.loading)
  const token = useAppSelector((s) => s.auth.token)
  const user = useAppSelector((s) => s.auth.user)

  useEffect(() => {
    if (token && user) {
      const role = getUserRoleCode(user)
      navigate(role === 'comment_moderator' ? '/comments' : '/', { replace: true })
    }
  }, [token, user, navigate])

  const fetchCaptcha = async () => {
    try {
      const res = await client.get('/captcha/id')
      const id = res.data.data.captcha_id
      setCaptchaId(id)
      setCaptchaImg(`/api/v1/captcha/image?id=${id}&t=${Date.now()}`)
    } catch { message.error('验证码加载失败') }
  }

  useEffect(() => { fetchCaptcha() }, [])

  const onFinish = async (values: { username: string; password: string; captcha_code: string }) => {
    setLoading(true)
    dispatch(loginAsync({
      credentials: { username: values.username, password: values.password, captcha_id: captchaId, captcha_code: values.captcha_code },
      navigate,
    })).then((res: any) => {
      setLoading(false)
      if (res.error) { message.error(res.payload || '登录失败'); fetchCaptcha() }
    })
  }

  // 暗色主题 token
  const clr = {
    bg: '#0a0a14',
    cardBg: 'rgba(22,22,38,0.70)',
    cardBorder: 'rgba(255,255,255,0.06)',
    text: '#e5e7eb',
    textDim: 'rgba(255,255,255,0.40)',
    inputBg: 'rgba(255,255,255,0.04)',
    inputBorder: 'rgba(255,255,255,0.08)',
    primary: '#3b82f6',
  }

  return (
    <div style={{ position: 'relative', width: '100vw', height: '100vh', overflow: 'hidden', background: clr.bg }}>
      {/* 背景视频层 —— 将 src 替换为你的视频链接即可启用 */}
      <video
        className="login-bg-video"
        autoPlay muted loop playsInline
        poster=""
        style={{
          position: 'absolute', inset: 0, zIndex: 0,
          width: '100%', height: '100%', objectFit: 'cover',
          background: 'radial-gradient(ellipse 60% 50% at 50% 40%, #1e293b 0%, #0f172a 50%, #020617 100%)',
        }}
      >
        {/* <source src="https://your-cdn.com/bg.mp4" type="video/mp4" /> */}
      </video>
      {/* 动态光斑 */}
      <div style={{ position: 'absolute', top: '-20%', left: '-10%', width: '60%', height: '60%', background: 'radial-gradient(circle, rgba(59,130,246,0.08) 0%, transparent 70%)', zIndex: 0 }} />
      <div style={{ position: 'absolute', bottom: '-10%', right: '-5%', width: '50%', height: '50%', background: 'radial-gradient(circle, rgba(99,102,241,0.06) 0%, transparent 70%)', zIndex: 0 }} />

      {/* 登录卡片 */}
      <div style={{ position: 'relative', zIndex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', padding: 24 }}>
        <Card
          style={{
            width: 420,
            background: clr.cardBg,
            backdropFilter: 'blur(12px)',
            border: `1px solid ${clr.cardBorder}`,
            borderRadius: 16,
            boxShadow: '0 20px 40px rgba(0,0,0,0.5)',
          }}
          styles={{ body: { padding: 40 } }}
        >
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            <div style={{ textAlign: 'center' }}>
              <Title level={2} style={{ marginBottom: 4, color: '#f1f5f9', fontWeight: 700, letterSpacing: '-0.02em' }}>
                Blogo
              </Title>
              <Text style={{ color: clr.textDim, fontSize: 13 }}>后台管理系统</Text>
            </div>

            <Form layout="vertical" onFinish={onFinish} size="large">
              <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]} style={{ marginBottom: 16 }}>
                <Input
                  prefix={<UserOutlined style={{ color: clr.textDim }} />}
                  placeholder="用户名"
                  style={{
                    height: 48, borderRadius: 10,
                    background: clr.inputBg, borderColor: clr.inputBorder,
                    color: clr.text,
                  }}
                />
              </Form.Item>

              <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]} style={{ marginBottom: 16 }}>
                <Input.Password
                  prefix={<LockOutlined style={{ color: clr.textDim }} />}
                  placeholder="密码"
                  style={{
                    height: 48, borderRadius: 10,
                    background: clr.inputBg, borderColor: clr.inputBorder,
                    color: clr.text,
                  }}
                />
              </Form.Item>

              <Form.Item name="captcha_code" rules={[{ required: true, message: '请输入验证码' }]} style={{ marginBottom: 20 }}>
                <Row gutter={8} align="stretch">
                  <Col span={14}>
                    <Input
                      prefix={<SafetyCertificateOutlined style={{ color: clr.textDim }} />}
                      placeholder="验证码"
                      style={{
                        height: 48, borderRadius: 10,
                        background: clr.inputBg, borderColor: clr.inputBorder,
                        color: clr.text,
                      }}
                    />
                  </Col>
                  <Col span={10}>
                    {captchaImg ? (
                      <img src={captchaImg} alt="验证码" onClick={fetchCaptcha}
                        style={{
                          cursor: 'pointer', width: '100%', height: 48, borderRadius: 6,
                          objectFit: 'contain', border: `1px solid ${clr.inputBorder}`, background: '#fff',
                        }} />
                    ) : (
                      <div onClick={fetchCaptcha}
                        style={{
                          cursor: 'pointer', width: '100%', height: 48, borderRadius: 6,
                          border: `1px solid ${clr.inputBorder}`,
                          display: 'flex', alignItems: 'center', justifyContent: 'center',
                          fontSize: 11, color: clr.textDim,
                        }}>
                        点击加载
                      </div>
                    )}
                  </Col>
                </Row>
              </Form.Item>

              <Form.Item style={{ marginBottom: 0 }}>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={authLoading || loading}
                  block
                  style={{
                    height: 48, borderRadius: 10, fontSize: 15, fontWeight: 600,
                    background: `linear-gradient(135deg, ${clr.primary}, #6366f1)`,
                    border: 'none',
                    boxShadow: '0 4px 14px rgba(59,130,246,0.35)',
                    transition: 'all 0.25s',
                  }}
                >
                  登 录
                </Button>
              </Form.Item>
            </Form>
          </Space>
        </Card>
      </div>
    </div>
  )
}

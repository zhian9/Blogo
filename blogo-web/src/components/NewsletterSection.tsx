import { useState } from 'react'
import { Input } from 'antd'
import { MailOutlined, CheckCircleOutlined, CloseCircleOutlined, LoadingOutlined } from '@ant-design/icons'
import { AnimatePresence, motion } from 'framer-motion'
import client from '../api/client'

type Status = 'idle' | 'loading' | 'success' | 'error' | 'exists'

export default function NewsletterSection() {
  const [email, setEmail] = useState('')
  const [status, setStatus] = useState<Status>('idle')

  const isValidEmail = (v: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)

  const handleSubscribe = async () => {
    if (!email || !isValidEmail(email)) {
      setStatus('error')
      setTimeout(() => status === 'error' && setStatus('idle'), 2500)
      return
    }

    setStatus('loading')
    try {
      await client.post('/subscribe', { email: email.trim() })
      setStatus('success')
      setEmail('')
      setTimeout(() => setStatus('idle'), 3000)
    } catch (err: any) {
      const msg = err?.message || ''
      if (msg.includes('已订阅') || msg.includes('重复订阅')) {
        setStatus('exists')
      } else {
        setStatus('error')
      }
      setTimeout(() => setStatus('idle'), 3000)
    }
  }

  const isDisabled = status === 'loading' || status === 'success' || status === 'exists'

  return (
    <motion.section
      id="newsletter-section"
      initial={{ opacity: 0, y: 40 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.7, ease: [0.25, 0.46, 0.45, 0.94] }}
      style={{ marginBottom: 80 }}
    >
      <div
        className="liquid-glass-card"
        style={{ padding: '60px 40px', maxWidth: 720, margin: '0 auto', textAlign: 'center' }}
      >
        {/* Title */}
        <motion.div
          initial={{ opacity: 0, y: -16 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <h2 style={{
            fontFamily: "'Instrument Serif', serif",
            fontStyle: 'italic',
            fontSize: 'clamp(28px, 5vw, 42px)',
            fontWeight: 400, color: '#ffffff',
            letterSpacing: '-0.03em',
            margin: '0 0 12px',
          }}>
            <MailOutlined style={{ color: '#4f6ef7', marginRight: 12 }} />
            订阅更新
          </h2>

          <p style={{
            fontFamily: "'Barlow', sans-serif",
            fontSize: 16, fontWeight: 300,
            color: 'rgba(255,255,255,0.5)',
            lineHeight: 1.7, margin: '0 auto 36px',
            maxWidth: 480,
          }}>
            订阅我的新闻通讯，获取最新的技术文章、开发经验和行业洞察。每周精选一篇深度好文，直达你的邮箱。
          </p>
        </motion.div>

        {/* Input + Button */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.15 }}
          style={{
            display: 'flex', gap: 12, flexWrap: 'wrap',
            justifyContent: 'center', alignItems: 'center',
          }}
        >
          <Input
            type="email"
            placeholder="输入你的邮箱..."
            value={email}
            onChange={(e) => { setEmail(e.target.value); if (status === 'error') setStatus('idle') }}
            onPressEnter={handleSubscribe}
            disabled={isDisabled}
            style={{
              height: 50, borderRadius: 14, fontSize: 15, flex: '1 1 260px', maxWidth: 380,
              background: 'rgba(255,255,255,0.05)',
              border: `1px solid ${status === 'error' ? 'rgba(239,68,68,0.4)' : status === 'success' ? 'rgba(16,185,129,0.4)' : status === 'exists' ? 'rgba(245,158,11,0.4)' : 'rgba(255,255,255,0.1)'}`,
              color: '#ffffff',
              transition: 'border-color 0.3s ease, box-shadow 0.3s ease',
              boxShadow: status === 'error' ? '0 0 0 2px rgba(239,68,68,0.1)' : status === 'success' ? '0 0 0 2px rgba(16,185,129,0.1)' : 'none',
            }}
          />

          <motion.div whileHover={!isDisabled ? { scale: 1.04 } : {}} whileTap={!isDisabled ? { scale: 0.97 } : {}}>
            <motion.button
              type="button"
              onClick={handleSubscribe}
              disabled={isDisabled}
              className={status === 'success' ? 'liquid-glass-card' : 'liquid-glass-strong'}
              style={{
                display: 'inline-flex', alignItems: 'center', gap: 8,
                padding: '13px 28px',
                cursor: isDisabled ? 'default' : 'pointer',
                fontFamily: "'Barlow', sans-serif",
                fontSize: 15, fontWeight: 600, color: '#ffffff',
                letterSpacing: '0.03em',
                border: 'none', outline: 'none',
                opacity: status === 'loading' ? 0.7 : 1,
                borderColor: status === 'success' ? 'rgba(16,185,129,0.3)' : undefined,
              }}
            >
              {status === 'loading' ? (
                <><LoadingOutlined spin /> 订阅中...</>
              ) : status === 'success' ? (
                <><CheckCircleOutlined style={{ color: '#10b981' }} /> 已订阅</>
              ) : (
                <><MailOutlined /> 立即订阅</>
              )}
            </motion.button>
          </motion.div>
        </motion.div>

        {/* ---- Status feedback ---- */}
        <AnimatePresence mode="wait">
          {status === 'success' && (
            <motion.div
              key="success"
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.35 }}
              style={{
                marginTop: 20,
                display: 'inline-flex', alignItems: 'center', gap: 8,
                padding: '10px 20px',
                borderRadius: 12,
                background: 'rgba(16,185,129,0.08)',
                border: '1px solid rgba(16,185,129,0.2)',
              }}
            >
              <CheckCircleOutlined style={{ color: '#10b981', fontSize: 16 }} />
              <span style={{
                fontFamily: "'Barlow', sans-serif",
                fontSize: 14, fontWeight: 500,
                color: '#6ee7b7',
              }}>
                订阅成功！感谢你的关注。
              </span>
            </motion.div>
          )}

          {status === 'error' && (
            <motion.div
              key="error"
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.35 }}
              style={{
                marginTop: 20,
                display: 'inline-flex', alignItems: 'center', gap: 8,
                padding: '10px 20px',
                borderRadius: 12,
                background: 'rgba(239,68,68,0.08)',
                border: '1px solid rgba(239,68,68,0.2)',
              }}
            >
              <CloseCircleOutlined style={{ color: '#f87171', fontSize: 16 }} />
              <span style={{
                fontFamily: "'Barlow', sans-serif",
                fontSize: 14, fontWeight: 500,
                color: '#fca5a5',
              }}>
                服务开小差了，请稍后重试
              </span>
            </motion.div>
          )}

          {status === 'exists' && (
            <motion.div
              key="exists"
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.35 }}
              style={{
                marginTop: 20,
                display: 'inline-flex', alignItems: 'center', gap: 8,
                padding: '10px 20px',
                borderRadius: 12,
                background: 'rgba(245,158,11,0.08)',
                border: '1px solid rgba(245,158,11,0.2)',
              }}
            >
              <CheckCircleOutlined style={{ color: '#f59e0b', fontSize: 16 }} />
              <span style={{
                fontFamily: "'Barlow', sans-serif",
                fontSize: 14, fontWeight: 500,
                color: '#fcd34d',
              }}>
                该邮箱已订阅，无需重复订阅
              </span>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Trust badge */}
        <motion.p
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.3 }}
          style={{
            fontFamily: "'Barlow', sans-serif",
            fontSize: 13, color: 'rgba(255,255,255,0.3)',
            marginTop: 24, letterSpacing: '0.04em',
          }}
        >
          无垃圾邮件 · 随时取消订阅 · 隐私保护
        </motion.p>
      </div>
    </motion.section>
  )
}

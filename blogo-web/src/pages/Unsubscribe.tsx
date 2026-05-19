import { useState } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import { MailOutlined, CheckCircleOutlined, CloseCircleOutlined, LoadingOutlined } from '@ant-design/icons'
import { motion } from 'framer-motion'
import client from '../api/client'

export default function Unsubscribe() {
  const [searchParams] = useSearchParams()
  const email = searchParams.get('email') || ''
  const [status, setStatus] = useState<'idle' | 'loading' | 'success' | 'error'>('idle')

  const handleUnsubscribe = async () => {
    setStatus('loading')
    try {
      await client.get('/subscribe/unsubscribe', { params: { email } })
      setStatus('success')
    } catch {
      setStatus('error')
    }
  }

  return (
    <div style={{
      display: 'flex', justifyContent: 'center', alignItems: 'center',
      minHeight: '60vh', padding: 24,
    }}>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: 'easeOut' }}
        className="liquid-glass-card"
        style={{ maxWidth: 480, width: '100%', padding: '48px 40px', textAlign: 'center' }}
      >
        <h2 style={{
          fontFamily: "'Instrument Serif', serif", fontStyle: 'italic',
          fontSize: 28, fontWeight: 400, color: '#ffffff',
          letterSpacing: '-0.02em', margin: '0 0 16px',
        }}>
          <MailOutlined style={{ color: '#4f6ef7', marginRight: 10 }} />
          取消订阅
        </h2>

        {status === 'idle' && (
          <>
            <p style={{
              fontFamily: "'Barlow', sans-serif",
              fontSize: 15, color: 'rgba(255,255,255,0.5)',
              marginBottom: 28, lineHeight: 1.7,
            }}>
              确认取消 <strong style={{ color: 'rgba(255,255,255,0.85)' }}>{email}</strong> 的订阅？取消后将不再收到新文章通知。
            </p>
            <motion.button
              type="button"
              whileHover={{ scale: 1.04 }} whileTap={{ scale: 0.97 }}
              onClick={handleUnsubscribe}
              className="liquid-glass-strong"
              style={{
                display: 'inline-flex', alignItems: 'center', gap: 8,
                padding: '12px 32px', cursor: 'pointer',
                fontFamily: "'Barlow', sans-serif",
                fontSize: 15, fontWeight: 600, color: '#ffffff',
                border: 'none', outline: 'none',
                letterSpacing: '0.03em',
              }}
            >
              确认取消订阅
            </motion.button>
          </>
        )}

        {status === 'loading' && (
          <div style={{ padding: 24 }}>
            <LoadingOutlined spin style={{ fontSize: 32, color: '#4f6ef7', marginBottom: 12 }} />
            <p style={{ fontFamily: "'Barlow', sans-serif", color: 'rgba(255,255,255,0.5)', fontSize: 14 }}>
              处理中...
            </p>
          </div>
        )}

        {status === 'success' && (
          <div style={{ padding: 12 }}>
            <CheckCircleOutlined style={{ fontSize: 44, color: '#10b981', marginBottom: 12 }} />
            <p style={{
              fontFamily: "'Barlow', sans-serif",
              fontSize: 15, color: '#6ee7b7', fontWeight: 500, marginBottom: 20,
            }}>
              已成功取消订阅
            </p>
            <Link to="/" style={{
              fontFamily: "'Barlow', sans-serif",
              fontSize: 14, color: '#4f6ef7', textDecoration: 'none',
            }}>
              返回首页
            </Link>
          </div>
        )}

        {status === 'error' && (
          <div style={{ padding: 12 }}>
            <CloseCircleOutlined style={{ fontSize: 44, color: '#f87171', marginBottom: 12 }} />
            <p style={{
              fontFamily: "'Barlow', sans-serif",
              fontSize: 15, color: '#fca5a5', fontWeight: 500,
            }}>
              操作失败，请稍后重试
            </p>
          </div>
        )}
      </motion.div>
    </div>
  )
}

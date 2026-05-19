import { useState } from 'react'
import { Input, Button, Card, Typography, message } from 'antd'
import { MailOutlined, CheckCircleOutlined } from '@ant-design/icons'
import { motion } from 'framer-motion'
import { useAppStore } from '../store/appStore'

const { Title, Text } = Typography

export default function NewsletterMini() {
  const theme = useAppStore((s) => s.theme)
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [subscribed, setSubscribed] = useState(false)

  const handleSubscribe = async () => {
    if (!email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      message.error('请输入有效的邮箱地址')
      return
    }

    setLoading(true)
    try {
      await new Promise((resolve) => setTimeout(resolve, 600))
      message.success('订阅成功！')
      setSubscribed(true)
      setEmail('')
      setTimeout(() => setSubscribed(false), 2000)
    } catch (error) {
      message.error('订阅失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  const cardVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: 0.6, delay: 0.3 },
    },
  }

  return (
    <motion.div
      variants={cardVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true }}
    >
      <Card
        style={{
          borderRadius: 12,
          background: theme === 'dark'
            ? 'linear-gradient(135deg, #0f0f0f 0%, #1a1a1a 100%)'
            : 'linear-gradient(135deg, #e6f7ff 0%, #f0f5ff 100%)',
          border: `2px solid ${theme === 'dark' ? '#303030' : '#bae7ff'}`,
          overflow: 'hidden',
        }}
        styles={{ body: { padding: '24px 16px' } }}
      >
        <div style={{ textAlign: 'center' }}>
          {/* 图标 */}
          <motion.div
            animate={{ rotate: [0, 5, -5, 0] }}
            transition={{ duration: 2, repeat: Infinity }}
            style={{ fontSize: 32, marginBottom: 12 }}
          >
            <MailOutlined style={{ color: '#1890ff' }} />
          </motion.div>

          {/* 标题 */}
          <Title
            level={5}
            style={{
              margin: '0 0 8px',
              color: theme === 'dark' ? '#fff' : '#000',
              fontSize: 16,
            }}
          >
            订阅更新
          </Title>

          {/* 描述 */}
          <Text
            type="secondary"
            style={{
              fontSize: 12,
              display: 'block',
              marginBottom: 16,
              lineHeight: 1.5,
            }}
          >
            获取最新文章推送，每周精选。
          </Text>

          {/* 邮箱输入 */}
          <div style={{ marginBottom: 12 }}>
            <Input
              type="email"
              placeholder="你的邮箱"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              onPressEnter={handleSubscribe}
              disabled={loading || subscribed}
              style={{
                borderRadius: 6,
                fontSize: 13,
                height: 36,
                background: theme === 'dark' ? '#1a1a1a' : '#fff',
                border: `1px solid ${theme === 'dark' ? '#303030' : '#bae7ff'}`,
              }}
            />
          </div>

          {/* 订阅按钮 */}
          <motion.div whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }}>
            <Button
              type="primary"
              block
              size="small"
              loading={loading}
              onClick={handleSubscribe}
              disabled={subscribed}
              icon={subscribed ? <CheckCircleOutlined /> : <MailOutlined />}
              style={{
                borderRadius: 6,
                height: 36,
                fontWeight: 600,
                background: subscribed ? '#52c41a' : '#1890ff',
                border: 'none',
              }}
            >
              {subscribed ? '已订阅' : '订阅'}
            </Button>
          </motion.div>

          {/* 提示 */}
          <Text
            type="secondary"
            style={{
              fontSize: 11,
              display: 'block',
              marginTop: 12,
              color: theme === 'dark' ? '#666' : '#999',
            }}
          >
            无垃圾邮件 · 随时取消
          </Text>
        </div>
      </Card>
    </motion.div>
  )
}

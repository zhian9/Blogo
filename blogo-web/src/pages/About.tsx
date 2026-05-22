import { Typography, Spin, Space } from 'antd'
import {
  GithubOutlined, MailOutlined, CodeOutlined,
  InfoCircleOutlined, LaptopOutlined, CloudServerOutlined,
} from '@ant-design/icons'
import { motion } from 'framer-motion'
import MarkdownRenderer from '../components/MarkdownRenderer'
import { usePage } from '../hooks/useSettings'

const { Title, Text } = Typography

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.12, delayChildren: 0.1 } },
}

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.6, ease: [0.25, 0.46, 0.45, 0.94] } },
}

const techStack = [
  { icon: <CodeOutlined />, label: 'Go · Gin · GORM', desc: '高性能后端框架' },
  { icon: <LaptopOutlined />, label: 'React · TypeScript', desc: '现代前端工程化' },
  { icon: <CloudServerOutlined />, label: 'Docker · Linux', desc: 'DevOps 与部署' },
]

export default function About() {
  const { data, isLoading } = usePage('about')
  const page = data?.data

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!page) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        style={{ textAlign: 'center', padding: '120px 24px' }}
      >
        <InfoCircleOutlined style={{ fontSize: 56, color: 'rgba(255,255,255,0.2)', marginBottom: 20 }} />
        <Title level={3} style={{ color: '#ffffff', marginBottom: 8 }}>页面未找到</Title>
        <Text style={{ color: 'rgba(255,255,255,0.4)' }}>关于页面尚未创建，请先在后管添加 slug 为 "about" 的页面。</Text>
      </motion.div>
    )
  }

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={containerVariants}
      style={{ maxWidth: 960, margin: '0 auto', paddingBottom: 80 }}
    >
      {/* ── Header ── */}
      <motion.div variants={itemVariants} style={{ textAlign: 'center', marginBottom: 56, marginTop: 24 }}>
        <h1 style={{
          fontFamily: "'Instrument Serif', serif",
          fontStyle: 'italic',
          fontSize: 'clamp(32px, 5vw, 56px)',
          fontWeight: 400,
          color: '#ffffff',
          letterSpacing: '-0.03em',
          margin: '0 0 12px',
        }}>
          {page.title}
        </h1>
        <p style={{
          fontFamily: "'Barlow', sans-serif",
          fontSize: 16,
          color: 'rgba(255,255,255,0.4)',
          letterSpacing: '0.04em',
          margin: 0,
        }}>
          全栈开发者 · 开源爱好者
        </p>
      </motion.div>

      {/* ── Tech Stack Cards ── */}
      <motion.div
        variants={itemVariants}
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
          gap: 16,
          marginBottom: 48,
        }}
      >
        {techStack.map((t) => (
          <div
            key={t.label}
            className="liquid-glass-card"
            style={{ padding: '24px', textAlign: 'center' }}
          >
            <div style={{
              fontSize: 28, color: '#4f6ef7', marginBottom: 12,
              display: 'flex', justifyContent: 'center',
            }}>
              {t.icon}
            </div>
            <div style={{
              fontFamily: "'Barlow', sans-serif",
              fontSize: 14, fontWeight: 600,
              color: '#ffffff', marginBottom: 6,
            }}>
              {t.label}
            </div>
            <div style={{
              fontFamily: "'Barlow', sans-serif",
              fontSize: 12, color: 'rgba(255,255,255,0.4)',
            }}>
              {t.desc}
            </div>
          </div>
        ))}
      </motion.div>

      {/* ── Social Links ── */}
      <motion.div
        variants={itemVariants}
        style={{ display: 'flex', justifyContent: 'center', gap: 20, marginBottom: 48 }}
      >
        <a
          href="https://github.com/zhian9"
          target="_blank"
          rel="noopener noreferrer"
          className="liquid-glass-card"
          style={{
            display: 'inline-flex', alignItems: 'center', gap: 8,
            padding: '10px 24px', textDecoration: 'none',
            fontFamily: "'Barlow', sans-serif",
            fontSize: 14, fontWeight: 500, color: 'rgba(255,255,255,0.55)',
            transition: 'all 0.25s ease',
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.color = '#ffffff'
            e.currentTarget.style.borderColor = 'rgba(255,255,255,0.2)'
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.color = 'rgba(255,255,255,0.55)'
            e.currentTarget.style.borderColor = 'rgba(255,255,255,0.07)'
          }}
        >
          <GithubOutlined style={{ fontSize: 18 }} />
          GitHub
        </a>
        <a
          href="mailto:admin@blogo.dev"
          className="liquid-glass-card"
          style={{
            display: 'inline-flex', alignItems: 'center', gap: 8,
            padding: '10px 24px', textDecoration: 'none',
            fontFamily: "'Barlow', sans-serif",
            fontSize: 14, fontWeight: 500, color: 'rgba(255,255,255,0.55)',
            transition: 'all 0.25s ease',
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.color = '#ffffff'
            e.currentTarget.style.borderColor = 'rgba(255,255,255,0.2)'
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.color = 'rgba(255,255,255,0.55)'
            e.currentTarget.style.borderColor = 'rgba(255,255,255,0.07)'
          }}
        >
          <MailOutlined style={{ fontSize: 18 }} />
          Email
        </a>
      </motion.div>

      {/* ── Divider ── */}
      <motion.div
        variants={itemVariants}
        style={{
          width: 60, height: 2,
          background: 'linear-gradient(90deg, transparent, rgba(79,110,247,0.5), transparent)',
          margin: '0 auto 48px',
        }}
      />

      {/* ── Markdown Content ── */}
      <motion.div
        variants={itemVariants}
        className="liquid-glass-card"
        style={{ padding: '36px 44px' }}
      >
        <MarkdownRenderer content={page.content} />
      </motion.div>
    </motion.div>
  )
}

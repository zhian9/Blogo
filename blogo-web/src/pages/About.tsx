import { Typography, Spin, Space } from 'antd'
import {
  GithubOutlined, MailOutlined, CodeOutlined,
  InfoCircleOutlined, LaptopOutlined, CloudServerOutlined,
} from '@ant-design/icons'
import { motion } from 'framer-motion'
import MarkdownRenderer from '../components/MarkdownRenderer'
import { usePage, useSettings } from '../hooks/useSettings'

const { Title, Text } = Typography

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.12, delayChildren: 0.1 } },
}

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.6, ease: [0.25, 0.46, 0.45, 0.94] } },
}

// 技术栈卡片的图标按顺序取用（文案由后台「系统设置 → about_tech_stack」配置）
const TECH_STACK_ICONS = [<CodeOutlined />, <LaptopOutlined />, <CloudServerOutlined />]

const DEFAULT_TECH_STACK = 'Go · Gin · GORM|高性能后端框架;;React · TypeScript|现代前端工程化;;Docker · Linux|DevOps 与部署'

export default function About() {
  const { data, isLoading } = usePage('about')
  const page = data?.data
  // 副标题 / 技术栈 / GitHub 链接都来自后台「系统设置」，未配置时用默认值
  const { data: settingsData } = useSettings()
  const settings = (settingsData?.data || []).reduce<Record<string, string>>((acc, item) => {
    acc[item.key] = item.value
    return acc
  }, {})
  const aboutSubtitle = settings.about_subtitle || '全栈开发者 · 开源爱好者'
  const githubUrl = settings.github_url || 'https://github.com/zhian9'
  // 未配置邮箱时不显示 Email 按钮（避免出现假地址）
  const contactEmail = settings.contact_email || ''
  // 关于页正文：优先用「系统设置 → about_content」，留空时回退到「页面管理 → about」的正文
  const aboutContent = (settings.about_content || page?.content || '').trim()
  const techStack = (settings.about_tech_stack || DEFAULT_TECH_STACK)
    .split(';;')
    .map((chunk, index) => {
      const [label, desc] = chunk.split('|')
      return {
        icon: TECH_STACK_ICONS[index % TECH_STACK_ICONS.length],
        label: (label || '').trim(),
        desc: (desc || '').trim(),
      }
    })
    .filter((item) => item.label)
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
      <motion.div variants={itemVariants} style={{ textAlign: 'center', marginBottom: 36, marginTop: 24 }}>
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
          {aboutSubtitle}
        </p>

        {/* ── Social Links：跟在标题区里，避免单独一行时显得空荡 ── */}
        <div style={{ display: 'flex', justifyContent: 'center', gap: 12, marginTop: 20 }}>
          <a
            href={githubUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="liquid-glass-card"
            style={{
              display: 'inline-flex', alignItems: 'center', gap: 6,
              padding: '8px 18px', textDecoration: 'none',
              fontFamily: "'Barlow', sans-serif",
              fontSize: 13, fontWeight: 500, color: 'rgba(255,255,255,0.55)',
              transition: 'all 0.25s ease',
            }}
            onMouseEnter={(e) => { e.currentTarget.style.color = '#ffffff' }}
            onMouseLeave={(e) => { e.currentTarget.style.color = 'rgba(255,255,255,0.55)' }}
          >
            <GithubOutlined style={{ fontSize: 15 }} />
            GitHub
          </a>
          {contactEmail ? (
            <a
              href={`mailto:${contactEmail}`}
              className="liquid-glass-card"
              style={{
                display: 'inline-flex', alignItems: 'center', gap: 6,
                padding: '8px 18px', textDecoration: 'none',
                fontFamily: "'Barlow', sans-serif",
                fontSize: 13, fontWeight: 500, color: 'rgba(255,255,255,0.55)',
                transition: 'all 0.25s ease',
              }}
              onMouseEnter={(e) => { e.currentTarget.style.color = '#ffffff' }}
              onMouseLeave={(e) => { e.currentTarget.style.color = 'rgba(255,255,255,0.55)' }}
            >
              <MailOutlined style={{ fontSize: 15 }} />
              Email
            </a>
          ) : null}
        </div>
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

      {/* ── Markdown Content（内容为空时不渲染空卡片） ── */}
      {aboutContent ? (
        <motion.div
          variants={itemVariants}
          className="liquid-glass-card"
          style={{ padding: '36px 44px' }}
        >
          <MarkdownRenderer content={aboutContent} />
        </motion.div>
      ) : null}
    </motion.div>
  )
}

import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowRightOutlined, UserOutlined, MailOutlined } from '@ant-design/icons'
import { motion } from 'framer-motion'
import VideoBackground from './VideoBackground'
import { getPublicStats } from '../api/statistics'

const HERO_VIDEO = 'https://cmxxx.dpdns.org/blogo.mp4'

function formatStat(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1).replace(/\.0$/, '')}K+`
  return `${n}+`
}

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.18, delayChildren: 0.6 } },
}

const blurInVariants = {
  hidden: { opacity: 0, filter: 'blur(16px)', y: 20 },
  visible: { opacity: 1, filter: 'blur(0px)', y: 0, transition: { duration: 1, ease: [0.25, 0.46, 0.45, 0.94] } },
}

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.7, ease: [0.25, 0.46, 0.45, 0.94] } },
}

const statVariants = {
  hidden: { opacity: 0, y: 24, scale: 0.95 },
  visible: (i: number) => ({
    opacity: 1, y: 0, scale: 1,
    transition: { duration: 0.6, delay: 1.0 + i * 0.15, ease: 'easeOut' },
  }),
}

export default function HeroSection() {
  const navigate = useNavigate()
  const [stats, setStats] = useState([
    { value: '0+', label: '篇文章' },
    { value: '0+', label: '个分类' },
    { value: '0+', label: '位读者' },
  ])

  useEffect(() => {
    getPublicStats()
      .then((res) => {
        setStats([
          { value: formatStat(res.data.article_count), label: '篇文章' },
          { value: formatStat(res.data.category_count), label: '个分类' },
          { value: formatStat(res.data.user_count), label: '位读者' },
        ])
      })
      .catch(() => {
        // keep defaults on error
      })
  }, [])

  return (
    <section style={{
      position: 'relative',
      width: '100%',
      height: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      overflow: 'hidden',
    }}>
      {/* ===== video (first in DOM, below overlay) ===== */}
      <VideoBackground src={HERO_VIDEO} />

      {/* ===== overlay (second in DOM, sits on top of video via DOM order) ===== */}
      <div
        aria-hidden="true"
        style={{
          position: 'absolute',
          inset: 0,
          background: `
            linear-gradient(to bottom,
              rgba(0,0,0,0.3) 0%,
              rgba(0,0,0,0.2) 50%,
              rgba(0,0,0,0.5) 100%)
          `,
          pointerEvents: 'none',
        }}
      />

      {/* ===== content (z-index lifts it above both) ===== */}
      <motion.div
        initial="hidden"
        animate="visible"
        variants={containerVariants}
        style={{
          position: 'relative',
          zIndex: 10,
          width: '100%',
          maxWidth: 1200,
          margin: '0 auto',
          padding: '120px 24px 80px',
          display: 'flex', flexDirection: 'column', alignItems: 'center',
          textAlign: 'center',
        }}
      >
        {/* ── Title ── */}
        <motion.h1
          variants={blurInVariants}
          style={{
            fontFamily: "'Instrument Serif', serif",
            fontStyle: 'italic',
            fontSize: 'clamp(52px, 10vw, 120px)',
            fontWeight: 400,
            color: '#ffffff',
            letterSpacing: '-0.04em',
            lineHeight: 0.9,
            marginBottom: 24,
            maxWidth: 900,
            textShadow: '0 0 120px rgba(79,110,247,0.3), 0 0 200px rgba(139,92,246,0.15)',
          }}
        >
          Blogo
          <br />
          <span style={{ fontSize: '0.7em', letterSpacing: '-0.02em' }}>博客</span>
        </motion.h1>

        {/* ── Subtitle ── */}
        <motion.p
          variants={blurInVariants}
          style={{
            fontFamily: "'Barlow', sans-serif",
            fontSize: 'clamp(16px, 2vw, 20px)',
            fontWeight: 300,
            color: 'rgba(255,255,255,0.65)',
            letterSpacing: '0.06em',
            lineHeight: 1.7,
            maxWidth: 580,
            marginBottom: 48,
          }}
        >
          基于 React + Go 构建的技术博客。
          <br />
          分享后端开发、系统设计与开源相关的思考。
        </motion.p>

        {/* ── CTA Buttons ── */}
        <motion.div
          variants={itemVariants}
          style={{
            display: 'flex', flexWrap: 'wrap', gap: 16,
            justifyContent: 'center', marginBottom: 72,
          }}
        >
          <motion.div whileHover={{ scale: 1.04 }} whileTap={{ scale: 0.97 }}>
            <button
              type="button"
              onClick={() => navigate('/articles')}
              className="liquid-glass-strong"
              style={{
                display: 'inline-flex', alignItems: 'center', gap: 10,
                padding: '14px 32px', cursor: 'pointer',
                fontFamily: "'Barlow', sans-serif",
                fontSize: 16, fontWeight: 600, color: '#ffffff',
                letterSpacing: '0.03em', border: 'none', outline: 'none',
              }}
            >
              <ArrowRightOutlined /> 阅读最新文章
            </button>
          </motion.div>

          <motion.div whileHover={{ scale: 1.04 }} whileTap={{ scale: 0.97 }}>
            <button
              type="button"
              onClick={() => navigate('/about')}
              className="liquid-glass"
              style={{
                display: 'inline-flex', alignItems: 'center', gap: 10,
                padding: '14px 32px', cursor: 'pointer',
                fontFamily: "'Barlow', sans-serif",
                fontSize: 16, fontWeight: 500, color: 'rgba(255,255,255,0.85)',
                letterSpacing: '0.03em', border: 'none', outline: 'none',
              }}
            >
              <UserOutlined /> 关于我
            </button>
          </motion.div>

          <motion.div whileHover={{ scale: 1.04 }} whileTap={{ scale: 0.97 }}>
            <button
              type="button"
              onClick={() => {
                document.getElementById('newsletter-section')?.scrollIntoView({ behavior: 'smooth' })
              }}
              className="liquid-glass"
              style={{
                display: 'inline-flex', alignItems: 'center', gap: 10,
                padding: '14px 32px', cursor: 'pointer',
                fontFamily: "'Barlow', sans-serif",
                fontSize: 16, fontWeight: 500, color: 'rgba(255,255,255,0.85)',
                letterSpacing: '0.03em', border: 'none', outline: 'none',
              }}
            >
              <MailOutlined /> 订阅更新
            </button>
          </motion.div>
        </motion.div>

        {/* ── Stats ── */}
        <motion.div
          initial="hidden"
          animate="visible"
          style={{ display: 'flex', flexWrap: 'wrap', gap: 16, justifyContent: 'center' }}
        >
          {stats.map((stat, i) => (
            <motion.div
              key={stat.label}
              custom={i}
              variants={statVariants}
              className="liquid-glass-card"
              whileHover={{ y: -6, scale: 1.03 }}
              style={{ padding: '20px 36px', textAlign: 'center', minWidth: 150 }}
            >
              <div style={{
                fontFamily: "'Instrument Serif', serif", fontStyle: 'italic',
                fontSize: 32, fontWeight: 400, color: '#ffffff',
                letterSpacing: '-0.02em', lineHeight: 1.1, marginBottom: 4,
              }}>
                {stat.value}
              </div>
              <div style={{
                fontFamily: "'Barlow', sans-serif",
                fontSize: 13, fontWeight: 400,
                color: 'rgba(255,255,255,0.45)',
                letterSpacing: '0.06em',
              }}>
                {stat.label}
              </div>
            </motion.div>
          ))}
        </motion.div>
      </motion.div>
    </section>
  )
}

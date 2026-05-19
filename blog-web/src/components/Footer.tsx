import { Space } from 'antd'
import { GithubOutlined, CopyrightOutlined } from '@ant-design/icons'
import { useSettings } from '../hooks/useSettings'

export default function Footer() {
  const { data } = useSettings()
  const settings = data?.data
  const siteTitle = 'Blogo'
  const icp = settings?.find((s) => s.key === 'icp_license')?.value || ''

  return (
    <footer className="liquid-glass-nav" style={{ position: 'relative', marginTop: 0 }}>
      <div style={{
        textAlign: 'center', padding: '32px 16px', maxWidth: 800, margin: '0 auto',
      }}>
        <Space vertical size={10}>
          {/* Site name */}
          <span style={{
            fontFamily: "'Instrument Serif', serif",
            fontStyle: 'italic',
            fontSize: 22,
            fontWeight: 400,
            color: '#ffffff',
            letterSpacing: '-0.02em',
          }}>
            {siteTitle}
          </span>

          {/* Copyright */}
          <span style={{
            fontFamily: "'Barlow', sans-serif",
            fontSize: 13, fontWeight: 400,
            color: 'rgba(255,255,255,0.35)',
            letterSpacing: '0.04em',
          }}>
            <CopyrightOutlined style={{ marginRight: 4 }} />
            {new Date().getFullYear()} {siteTitle}. All rights reserved.
          </span>

          {/* ICP */}
          {icp && (
            <span style={{
              fontFamily: "'Barlow', sans-serif",
              fontSize: 12, color: 'rgba(255,255,255,0.25)',
            }}>
              {icp}
            </span>
          )}

          {/* GitHub */}
          <a
            href="https://github.com/zhian9"
            target="_blank"
            rel="noopener noreferrer"
            aria-label="GitHub"
            style={{
              display: 'inline-flex',
              color: 'rgba(255,255,255,0.35)',
              fontSize: 22,
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.color = '#ffffff'
              e.currentTarget.style.transform = 'scale(1.15)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.color = 'rgba(255,255,255,0.35)'
              e.currentTarget.style.transform = 'scale(1)'
            }}
          >
            <GithubOutlined />
          </a>
        </Space>
      </div>
    </footer>
  )
}

import { CopyrightOutlined } from '@ant-design/icons'

export default function Footer() {
  const currentYear = new Date().getFullYear()

  return (
    <footer
      style={{
        position: 'relative',
        marginTop: 0,
        borderTop: '1px solid rgba(255,255,255,0.05)',
        background: 'linear-gradient(180deg, rgba(15,15,28,0.6) 0%, rgba(10,10,20,0.85) 100%)',
        backdropFilter: 'blur(12px)',
        WebkitBackdropFilter: 'blur(12px)',
      }}
    >
      <div
        style={{
          maxWidth: 960,
          margin: '0 auto',
          padding: 'clamp(24px, 4vw, 40px) clamp(12px, 3vw, 24px)',
          textAlign: 'center',
        }}
      >
        {/* ── Logo ── */}
        <div
          style={{
            fontFamily: "'Georgia', 'Times New Roman', 'Instrument Serif', serif",
            fontStyle: 'italic',
            fontSize: 'clamp(20px, 2.5vw, 26px)',
            fontWeight: 400,
            color: '#ffffff',
            letterSpacing: '-0.02em',
            marginBottom: 8,
          }}
        >
          Blogo
        </div>

        {/* ── Tech stack ── */}
        <p
          style={{
            margin: '0 0 18px',
            fontFamily: "'Barlow', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
            fontSize: 'clamp(11px, 1.2vw, 13px)',
            fontWeight: 400,
            color: 'rgba(255,255,255,0.3)',
            letterSpacing: '0.06em',
          }}
        >
          Built with Go · React · MySQL · Redis
        </p>

        {/* ── Divider ── */}
        <div
          style={{
            width: 60,
            height: 1,
            margin: '0 auto 16px',
            background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.08), transparent)',
          }}
        />

        {/* ── Copyright ── */}
        <p
          style={{
            margin: '0 0 10px',
            fontFamily: "'Barlow', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
            fontSize: 'clamp(11px, 1.1vw, 12px)',
            fontWeight: 400,
            color: 'rgba(255,255,255,0.3)',
            letterSpacing: '0.03em',
          }}
        >
          <CopyrightOutlined style={{ marginRight: 4 }} />
          {currentYear} Chen Zhian
        </p>

        {/* ── ICP License ── */}
        <a
          href="https://beian.miit.gov.cn/"
          target="_blank"
          rel="noopener noreferrer"
          aria-label="ICP备案信息"
          style={{
            display: 'inline-block',
            fontFamily: "'Barlow', 'PingFang SC', 'Microsoft YaHei', sans-serif",
            fontSize: 'clamp(11px, 1.1vw, 12px)',
            fontWeight: 400,
            color: 'rgba(255,255,255,0.22)',
            textDecoration: 'none',
            letterSpacing: '0.04em',
            padding: '4px 12px',
            borderRadius: 6,
            border: '1px solid transparent',
            transition: 'all 0.3s ease',
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.color = 'rgba(255,255,255,0.55)'
            e.currentTarget.style.borderColor = 'rgba(255,255,255,0.1)'
            e.currentTarget.style.background = 'rgba(255,255,255,0.03)'
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.color = 'rgba(255,255,255,0.22)'
            e.currentTarget.style.borderColor = 'transparent'
            e.currentTarget.style.background = 'transparent'
          }}
        >
          辽ICP备2026010701号-1
        </a>
      </div>
    </footer>
  )
}

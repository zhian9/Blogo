import { motion } from 'framer-motion'
import { FireOutlined, ThunderboltOutlined, TrophyOutlined, RiseOutlined } from '@ant-design/icons'
import type { ContributionStats } from '../types'

function pct(total: number): string {
  if (total >= 1000) return 'Top 1%'
  if (total >= 500) return 'Top 5%'
  if (total >= 200) return 'Top 15%'
  if (total >= 50) return 'Top 30%'
  return 'Keep going'
}

interface Props { stats: ContributionStats }

export default function ActivityStats({ stats }: Props) {
  const rows = [
    { icon: <FireOutlined />, label: '年度贡献', value: stats.total_contributions.toLocaleString(), sub: 'Contributions', color: '#3fb950' },
    { icon: <ThunderboltOutlined />, label: '当前连续', value: `${stats.current_streak}`, sub: 'day streak', color: '#d29922' },
    { icon: <TrophyOutlined />, label: '最长连续', value: `${stats.longest_streak}`, sub: 'day record', color: '#f778ba' },
    { icon: <RiseOutlined />, label: stats.active_level, value: pct(stats.total_contributions), sub: '', color: '#79c0ff' },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 10, justifyContent: 'center', height: '100%' }}>
      {rows.map((r, i) => (
        <motion.div
          key={r.label}
          initial={{ opacity: 0, x: 16 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.35, delay: 0.5 + i * 0.08, ease: 'easeOut' }}
          style={{
            display: 'flex', alignItems: 'center', gap: 12,
            padding: '12px 16px',
            borderRadius: 12,
            border: '1px solid rgba(255,255,255,0.04)',
            background: 'rgba(255,255,255,0.015)',
            transition: 'background 0.2s ease, border-color 0.2s ease',
            cursor: 'default',
          }}
          onMouseEnter={e => { e.currentTarget.style.background='rgba(255,255,255,0.04)'; e.currentTarget.style.borderColor='rgba(255,255,255,0.08)' }}
          onMouseLeave={e => { e.currentTarget.style.background='rgba(255,255,255,0.015)'; e.currentTarget.style.borderColor='rgba(255,255,255,0.04)' }}
        >
          <div style={{
            width: 36, height: 36, borderRadius: 10,
            background: `${r.color}15`, border: `1px solid ${r.color}30`,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 16, color: r.color, flexShrink: 0,
          }}>{r.icon}</div>
          <div style={{ minWidth: 0 }}>
            <div style={{ fontSize: 10, color: 'rgba(255,255,255,0.3)', fontFamily: "'Barlow', sans-serif", letterSpacing: '0.05em', marginBottom: 1 }}>{r.label}</div>
            <div style={{ fontSize: 20, fontWeight: 700, color: '#fff', fontFamily: "'Barlow', sans-serif", lineHeight: 1.15 }}>
              {r.value}
              {r.sub ? <span style={{ fontSize: 10, fontWeight: 400, color: 'rgba(255,255,255,0.2)', marginLeft: 4 }}>{r.sub}</span> : null}
            </div>
          </div>
        </motion.div>
      ))}
    </div>
  )
}

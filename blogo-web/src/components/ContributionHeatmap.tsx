import { useMemo, useState, useCallback, useRef, useEffect } from 'react'
import { motion } from 'framer-motion'
import type { ContributionDay } from '../types'

// ── GitHub green scale ──
const CELL_COLORS = ['rgba(255,255,255,0.04)', '#1a5e30', '#0e8c3a', '#2db84d', '#39d353']
function cellColor(n: number) { if (n===0) return CELL_COLORS[0]; if (n<=2) return CELL_COLORS[1]; if (n<=5) return CELL_COLORS[2]; if (n<=9) return CELL_COLORS[3]; return CELL_COLORS[4] }

const toDateStr = (d: Date) => `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`
const DAY_NAMES = ['Sun','Mon','Tue','Wed','Thu','Fri','Sat']
const MONTH_NAMES = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec']
const CELL = 13
const GAP = 3
const LABEL_W = 30
const ROWS = 7

// ── Calendar generation ──
interface Cell {
  date: string
  count: number
  publishCount: number
  editCount: number
  dayOfWeek: number  // 0=Sun
  weekIndex: number
  month: number
}

function buildCalendar(data: ContributionDay[]): { cells: (Cell | null)[][]; monthMarkers: { week: number; label: string }[] } {
  const map = new Map<string, ContributionDay>()
  for (const d of data) map.set(d.date, d)

  // Date range: 364 days ago → today
  const today = new Date()
  today.setHours(0,0,0,0)
  const start = new Date(today)
  start.setDate(start.getDate() - 364)

  // Generate all days
  const all: Cell[] = []
  for (let i = 0; i < 365; i++) {
    const d = new Date(start)
    d.setDate(d.getDate() + i)
    const ds = toDateStr(d)
    const cd = map.get(ds)
    all.push({
      date: ds,
      count: cd?.count ?? 0,
      publishCount: cd?.publish_count ?? 0,
      editCount: cd?.edit_count ?? 0,
      dayOfWeek: d.getDay(),
      weekIndex: 0,
      month: d.getMonth(),
    })
  }

  // Pad leading days so the grid starts on Sunday
  const firstDow = all[0].dayOfWeek
  const leadingNulls: null[] = Array(firstDow).fill(null)

  // Combine and split into columns (7 rows each)
  const combined = [...leadingNulls, ...all]
  const columns: (Cell | null)[][] = []
  for (let i = 0; i < combined.length; i += ROWS) {
    columns.push(combined.slice(i, i + ROWS))
  }

  // Build month markers: record which column each month starts at
  const monthMarkers: { week: number; label: string }[] = []
  let lastMonth = -1
  for (let col = 0; col < columns.length; col++) {
    const firstCell = columns[col].find((c): c is Cell => c !== null)
    if (firstCell && firstCell.month !== lastMonth) {
      monthMarkers.push({ week: col, label: MONTH_NAMES[firstCell.month] })
      lastMonth = firstCell.month
    }
  }

  return { cells: columns, monthMarkers }
}

interface Props { data: ContributionDay[] }

export default function ContributionHeatmap({ data }: Props) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const [tooltip, setTooltip] = useState<{ cell: Cell; x: number; y: number } | null>(null)
  const [canScrollL, setCanScrollL] = useState(false)
  const [canScrollR, setCanScrollR] = useState(true)

  const { cells, monthMarkers } = useMemo(() => {
    if (!data || data.length === 0) return { cells: [] as (Cell | null)[][], monthMarkers: [] as { week: number; label: string }[] }
    return buildCalendar(data)
  }, [data])

  // Scroll shadow
  const updateFade = useCallback(() => {
    const el = scrollRef.current; if (!el) return
    setCanScrollL(el.scrollLeft > 2)
    setCanScrollR(el.scrollLeft < el.scrollWidth - el.clientWidth - 2)
  }, [])
  useEffect(() => {
    const el = scrollRef.current; if (!el) return
    updateFade()
    el.addEventListener('scroll', updateFade, { passive: true })
    return () => el.removeEventListener('scroll', updateFade)
  }, [updateFade])

  // Tooltip
  const showTip = useCallback((cell: Cell, e: React.MouseEvent) => {
    const rect = scrollRef.current?.getBoundingClientRect()
    if (!rect) return
    setTooltip({ cell, x: e.clientX - rect.left, y: e.clientY - rect.top })
  }, [])

  if (!data || data.length === 0) {
    return <div style={{ color: 'rgba(255,255,255,0.25)', fontSize: 13, padding: '32px 0', textAlign: 'center' }}>暂无贡献数据</div>
  }

  const gridW = cells.length * (CELL + GAP)

  return (
    <div style={{ overflow: 'hidden' }}>
      {/* ── Scrollable area: month labels + grid together ── */}
      <div
        ref={scrollRef}
        style={{ overflowX: 'auto', overflowY: 'hidden', scrollBehavior: 'smooth', position: 'relative' }}
        className="heatmap-scroll"
      >
        {/* Fade edges */}
        {canScrollL && <div style={{ position:'absolute',left:0,top:0,bottom:0,width:28,zIndex:2,background:'linear-gradient(to right, rgba(10,10,16,0.9), transparent)',pointerEvents:'none' }} />}
        {canScrollR && <div style={{ position:'absolute',right:0,top:0,bottom:0,width:28,zIndex:2,background:'linear-gradient(to left, rgba(10,10,16,0.9), transparent)',pointerEvents:'none' }} />}

        <div style={{ display:'flex', flexDirection:'column', width: gridW + LABEL_W, minWidth: 'max-content' }}>
          {/* Month labels row */}
          <div style={{ height: 18, marginLeft: LABEL_W, position: 'relative' }}>
            {monthMarkers.map((m, i) => (
              <span key={i} style={{
                position: 'absolute',
                left: m.week * (CELL + GAP),
                fontSize: 10, fontWeight: 500,
                color: 'rgba(255,255,255,0.3)',
                fontFamily: "'Barlow', sans-serif",
                lineHeight: '18px',
              }}>{m.label}</span>
            ))}
          </div>

          {/* Days + grid row */}
          <div style={{ display: 'flex' }}>
            {/* Day-of-week labels */}
            <div style={{ display:'flex', flexDirection:'column', gap: GAP, width: LABEL_W, flexShrink: 0, paddingTop: 1 }}>
              {DAY_NAMES.map((name, i) => (
                <div key={i} style={{ height: CELL, display:'flex', alignItems:'center' }}>
                  <span style={{
                    fontSize: 9, color: i%2===0 ? 'rgba(255,255,255,0.2)' : 'rgba(255,255,255,0.12)',
                    fontFamily: "'Barlow', sans-serif",
                  }}>{name}</span>
                </div>
              ))}
            </div>

            {/* Cell grid — columns = weeks */}
            <div style={{ display:'flex', gap: GAP, flexShrink: 0 }}>
              {cells.map((col, ci) => (
                <div key={ci} style={{ display:'flex', flexDirection:'column', gap: GAP }}>
                  {col.map((cell, ri) => {
                    if (!cell) return <div key={ri} style={{ width:CELL, height:CELL, flexShrink:0 }} />
                    // Highlight today's column with subtle border
                    const isToday = cell.date === toDateStr(new Date())
                    return (
                      <motion.div
                        key={ri}
                        initial={{ opacity:0, scale:0 }}
                        animate={{ opacity:1, scale:1 }}
                        transition={{ duration:0.2, delay: (ci*ROWS+ri)*0.0015, ease:'easeOut' }}
                        onMouseEnter={(e) => showTip(cell, e)}
                        onMouseLeave={() => setTooltip(null)}
                        style={{
                          width: CELL, height: CELL, borderRadius: 3,
                          background: cellColor(cell.count),
                          cursor: 'pointer', flexShrink: 0,
                          outline: isToday ? '1px solid rgba(255,255,255,0.25)' : undefined,
                          outlineOffset: 1,
                          transition: 'transform 0.1s ease, box-shadow 0.1s ease',
                        }}
                        whileHover={cell.count>0
                          ? { scale:1.6, boxShadow:`0 0 12px ${cellColor(cell.count)}`, zIndex:3, transition:{duration:0.1} }
                          : { scale:1.3, boxShadow:'0 0 0 1px rgba(255,255,255,0.12)', zIndex:3, transition:{duration:0.1} }
                        }
                      />
                    )
                  })}
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* ── Tooltip (inside scroll area) ── */}
        {tooltip && (
          <div style={{
            position: 'absolute',
            left: Math.min(tooltip.x + 14, gridW + LABEL_W - 195),
            top: Math.max(tooltip.y - 70, 20),
            zIndex: 10,
            background: 'rgba(12,12,24,0.97)',
            backdropFilter: 'blur(16px)',
            border: '1px solid rgba(255,255,255,0.1)',
            borderRadius: 10,
            padding: '10px 14px',
            boxShadow: '0 6px 24px rgba(0,0,0,0.5)',
            pointerEvents: 'none',
            minWidth: 170,
            fontFamily: "'Barlow', sans-serif",
          }}>
            <div style={{ fontSize:13, fontWeight:600, color:'#fff', marginBottom:6 }}>
              {tooltip.cell.date}
            </div>
            {tooltip.cell.publishCount > 0 && <div style={{ fontSize:11, color:'#7ee787' }}>{tooltip.cell.publishCount} 篇发布</div>}
            {tooltip.cell.editCount > 0 && <div style={{ fontSize:11, color:'#7ee787' }}>{tooltip.cell.editCount} 次编辑</div>}
            {tooltip.cell.count === 0 && <div style={{ fontSize:11, color:'rgba(255,255,255,0.3)' }}>无贡献</div>}
            <div style={{ marginTop:5, paddingTop:4, borderTop:'1px solid rgba(255,255,255,0.06)', fontSize:12, fontWeight:600, color:'#39d353' }}>
              {tooltip.cell.count} 次贡献
            </div>
          </div>
        )}
      </div>

      {/* ── Legend ── */}
      <div style={{ display:'flex', alignItems:'center', justifyContent:'flex-end', gap:4, marginTop:8 }}>
        <span style={{ fontSize:10, color:'rgba(255,255,255,0.18)', fontFamily:"'Barlow', sans-serif", marginRight:2 }}>Less</span>
        {CELL_COLORS.map((c,i)=><div key={i} style={{ width:CELL, height:CELL, borderRadius:3, background:c }} />)}
        <span style={{ fontSize:10, color:'rgba(255,255,255,0.18)', fontFamily:"'Barlow', sans-serif", marginLeft:2 }}>More</span>
      </div>

      <style>{`
        .heatmap-scroll::-webkit-scrollbar { height:3px; }
        .heatmap-scroll::-webkit-scrollbar-track { background:transparent; }
        .heatmap-scroll::-webkit-scrollbar-thumb { background:rgba(255,255,255,0.05); border-radius:3px; }
      `}</style>
    </div>
  )
}

import { useEffect, useState, useRef, useCallback } from 'react'

interface TocItem {
  id: string
  title: string
  level: number
}

interface Props {
  items: TocItem[]
  contentRef: React.RefObject<HTMLDivElement | null>
}

export default function TableOfContents({ items, contentRef }: Props) {
  const [activeId, setActiveId] = useState('')
  const observerRef = useRef<IntersectionObserver | null>(null)

  useEffect(() => {
    if (!contentRef.current || items.length === 0) return
    if (observerRef.current) observerRef.current.disconnect()

    // Delay to ensure React DOM has rendered content (dangerouslySetInnerHTML / react-markdown)
    const timer = setTimeout(() => {
      const container = contentRef.current
      if (!container) return

      const headingEls = container.querySelectorAll('h1, h2, h3')
      const elMap = new Map<string, HTMLElement>()

      // 按顺序匹配：TOC 第 n 项对应文档中第 n 个标题
      items.forEach((item, idx) => {
        if (idx < headingEls.length) {
          const el = headingEls[idx] as HTMLElement
          el.id = item.id
          elMap.set(item.id, el)
        }
      })

      // 如果按位置匹配数不够，回退到文本匹配
      if (elMap.size < items.length) {
        const usedEls = new Set(elMap.values())
        items.forEach((item) => {
          if (elMap.has(item.id)) return
          for (let i = 0; i < headingEls.length; i++) {
            const el = headingEls[i] as HTMLElement
            if (!usedEls.has(el) && el.textContent?.trim() === item.title) {
              el.id = item.id
              elMap.set(item.id, el)
              usedEls.add(el)
              break
            }
          }
        })
      }

      if (elMap.size === 0) return

      const observer = new IntersectionObserver(
        (entries) => {
          const visible = entries.filter((e) => e.isIntersecting)
          if (visible.length > 0) {
            setActiveId(visible[0].target.id)
          }
        },
        { rootMargin: '-80px 0px -70% 0px', threshold: 0 }
      )
      observerRef.current = observer
      elMap.forEach((el) => observer.observe(el))
    }, 150)

    return () => {
      clearTimeout(timer)
      if (observerRef.current) observerRef.current.disconnect()
    }
  }, [items, contentRef])

  const scrollTo = useCallback((id: string) => {
    // 先按 id 查找（由 useEffect 预设）
    let el: HTMLElement | null = document.getElementById(id)
    // 回退：按文本内容匹配（确保首次点击也能跳转）
    if (!el && contentRef.current) {
      const item = items.find((i) => i.id === id)
      if (item) {
        const headings = contentRef.current.querySelectorAll('h1, h2, h3')
        for (const h of headings) {
          if ((h as HTMLElement).textContent?.trim() === item.title) {
            el = h as HTMLElement
            el.id = id // 补设 ID，下次直接用
            break
          }
        }
      }
    }
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'start' })
      setActiveId(id)
    }
  }, [items, contentRef])

  if (items.length === 0) return null

  return (
    <nav style={{ padding: '8px 0' }}>
      <div style={{
        fontSize: 12, fontWeight: 600, textTransform: 'uppercase',
        letterSpacing: 1.5, color: '#8c8c8c', padding: '0 0 12px 16px',
      }}>
        目录
      </div>
      {items.map((item) => (
        <button
          type="button"
          key={item.id}
          onClick={() => scrollTo(item.id)}
          style={{
            display: 'block', width: '100%', textAlign: 'left',
            padding: `5px 16px 5px ${12 + item.level * 12}px`,
            fontSize: 13, lineHeight: 1.5,
            color: activeId === item.id ? '#1890ff' : '#595959',
            fontWeight: activeId === item.id ? 600 : 400,
            background: activeId === item.id ? 'rgba(24,144,255,0.06)' : 'transparent',
            border: 'none',
            borderLeft: activeId === item.id ? '3px solid #1890ff' : '3px solid transparent',
            cursor: 'pointer', transition: 'all 0.2s ease',
            overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
          }}
          onMouseEnter={(e) => {
            if (activeId !== item.id) {
              e.currentTarget.style.color = '#1890ff'
              e.currentTarget.style.background = 'rgba(24,144,255,0.03)'
            }
          }}
          onMouseLeave={(e) => {
            if (activeId !== item.id) {
              e.currentTarget.style.color = '#595959'
              e.currentTarget.style.background = 'transparent'
            }
          }}
        >
          {item.title}
        </button>
      ))}
    </nav>
  )
}

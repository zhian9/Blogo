import { useRef, useEffect, useCallback, useState, useMemo } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'

// ── Slash command definitions ──
interface SlashCmd { id: string; icon: string; label: string; shortcut: string; preview: string }
const slashCommands: SlashCmd[] = [
  { id: 'h1', icon: 'H1', label: '一级标题', shortcut: '#', preview: '# 标题' },
  { id: 'h2', icon: 'H2', label: '二级标题', shortcut: '##', preview: '## 标题' },
  { id: 'h3', icon: 'H3', label: '三级标题', shortcut: '###', preview: '### 标题' },
  { id: 'code', icon: '</>', label: '代码块', shortcut: '```', preview: '```\n\n```' },
  { id: 'table', icon: '⊞', label: '表格', shortcut: '|', preview: '| 列1 | 列2 | 列3 |\n| --- | --- | --- |\n| | | |' },
  { id: 'quote', icon: '"', label: '引用块', shortcut: '>', preview: '> 引用内容' },
  { id: 'ul', icon: '•', label: '无序列表', shortcut: '*', preview: '* 列表项' },
  { id: 'ol', icon: '1.', label: '有序列表', shortcut: '1.', preview: '1. 列表项' },
  { id: 'img', icon: '🖼', label: '插入图片', shortcut: '![]()', preview: '![描述](url)' },
  { id: 'link', icon: '🔗', label: '插入链接', shortcut: '[]()', preview: '[文本](url)' },
  { id: 'hr', icon: '—', label: '分割线', shortcut: '---', preview: '---' },
]

// ── Formatting toolbar actions ──
interface FormatAction { id: string; label: string; wrap: [string, string] }
const formatActions: FormatAction[] = [
  { id: 'bold', label: '加粗', wrap: ['**', '**'] },
  { id: 'italic', label: '斜体', wrap: ['_', '_'] },
  { id: 'code', label: '行内代码', wrap: ['`', '`'] },
  { id: 'link', label: '链接', wrap: ['[', '](url)'] },
  { id: 'strike', label: '删除线', wrap: ['~~', '~~'] },
]

interface Props { value: string; onChange: (v: string) => void }

export default function MarkdownEditor({ value, onChange }: Props) {
  const sourceRef = useRef<HTMLTextAreaElement>(null)
  const previewRef = useRef<HTMLDivElement>(null)
  const syncing = useRef(false)
  const wrapRef = useRef<HTMLDivElement>(null)

  // ── Slash menu state ──
  const [slashOpen, setSlashOpen] = useState(false)
  const [slashIdx, setSlashIdx] = useState(0)
  const [slashPos, setSlashPos] = useState({ top: 0, left: 0 })
  const [slashAnchor, setSlashAnchor] = useState(0) // cursor position when / was typed

  // ── Floating toolbar state ──
  const [floatOpen, setFloatOpen] = useState(false)
  const [floatPos, setFloatPos] = useState({ top: 0, left: 0 })

  // ── Helpers ──
  const getLineStart = useCallback((text: string, cursor: number) => {
    const before = text.lastIndexOf('\n', cursor - 1)
    return before === -1 ? 0 : before + 1
  }, [])

  const insertAtCursor = useCallback((text: string, insert: string, cursorPos: number, selectFrom?: number) => {
    const ta = sourceRef.current
    onChange(text)
    if (!ta) return
    requestAnimationFrame(() => {
      if (selectFrom !== undefined) {
        ta.selectionStart = selectFrom; ta.selectionEnd = cursorPos
      } else {
        ta.selectionStart = ta.selectionEnd = cursorPos
      }
      ta.focus()
    })
  }, [onChange])

  // ── Sync scroll ──
  const handleSourceScroll = useCallback(() => {
    if (syncing.current) return
    const src = sourceRef.current; const prv = previewRef.current
    if (!src || !prv) return
    const r = src.scrollTop / Math.max(src.scrollHeight - src.clientHeight, 1)
    syncing.current = true; prv.scrollTop = r * Math.max(prv.scrollHeight - prv.clientHeight, 1)
    requestAnimationFrame(() => { syncing.current = false })
  }, [])

  const handlePreviewScroll = useCallback(() => {
    if (syncing.current) return
    const src = sourceRef.current; const prv = previewRef.current
    if (!src || !prv) return
    const r = prv.scrollTop / Math.max(prv.scrollHeight - prv.clientHeight, 1)
    syncing.current = true; src.scrollTop = r * Math.max(src.scrollHeight - src.clientHeight, 1)
    requestAnimationFrame(() => { syncing.current = false })
  }, [])

  // ── Show/hide float toolbar on selection ──
  const handleSelect = useCallback(() => {
    const ta = sourceRef.current; if (!ta) return
    const start = ta.selectionStart; const end = ta.selectionEnd
    if (start === end) { setFloatOpen(false); return }
    // Position above selection
    const rect = ta.getBoundingClientRect()
    // Approximate position: count newlines before selectionStart
    const textBefore = value.substring(0, start)
    const lineCount = (textBefore.match(/\n/g) || []).length
    const lastNewline = textBefore.lastIndexOf('\n')
    const col = start - (lastNewline === -1 ? 0 : lastNewline + 1)
    const lineH = 15 * 1.8 // fontSize * lineHeight
    setFloatPos({
      top: 24 + lineCount * lineH + 8, // relative to textarea
      left: col * 9.2 + 28, // approximate char width
    })
    setFloatOpen(true)
  }, [value])

  // ── Execute slash command ──
  const executeSlash = useCallback((cmd: SlashCmd) => {
    const ta = sourceRef.current; if (!ta) return
    const lineStart = getLineStart(value, slashAnchor)
    const before = value.substring(0, lineStart)
    const after = value.substring(slashAnchor)
    let insert = ''; let cursor = 0; let selFrom: number | undefined

    switch (cmd.id) {
      case 'h1': insert = '# '; cursor = lineStart + 2; break
      case 'h2': insert = '## '; cursor = lineStart + 3; break
      case 'h3': insert = '### '; cursor = lineStart + 4; break
      case 'code': insert = '```\n\n```'; cursor = lineStart + 4; break
      case 'table': insert = '| 列1 | 列2 | 列3 |\n| --- | --- | --- |\n|  |  |  |'; cursor = lineStart + 11; selFrom = cursor; break
      case 'quote': insert = '> '; cursor = lineStart + 2; break
      case 'ul': insert = '* '; cursor = lineStart + 2; break
      case 'ol': insert = '1. '; cursor = lineStart + 3; break
      case 'img': insert = '![描述](url)'; cursor = lineStart + 2; selFrom = lineStart + 4; break
      case 'link': insert = '[文本](url)'; cursor = lineStart + 1; selFrom = lineStart + 3; break
      case 'hr': insert = '---\n'; cursor = lineStart + 4; break
      default: insert = ''; cursor = lineStart
    }
    const newVal = before + insert + after
    insertAtCursor(newVal, '', cursor, selFrom)
    setSlashOpen(false)
  }, [value, slashAnchor, getLineStart, insertAtCursor])

  // ── Keyboard handler ──
  const handleKeyDown = useCallback((e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    const ta = e.currentTarget

    // ── Slash menu open: intercept keys ──
    if (slashOpen) {
      if (e.key === 'ArrowDown') { e.preventDefault(); setSlashIdx(i => Math.min(i + 1, slashCommands.length - 1)); return }
      if (e.key === 'ArrowUp') { e.preventDefault(); setSlashIdx(i => Math.max(i - 1, 0)); return }
      if (e.key === 'Enter') { e.preventDefault(); executeSlash(slashCommands[slashIdx]); return }
      if (e.key === 'Escape') { e.preventDefault(); setSlashOpen(false); return }
      setSlashOpen(false)
    }

    // ── Tab → spaces ──
    if (e.key === 'Tab') {
      e.preventDefault()
      const start = ta.selectionStart; const end = ta.selectionEnd
      onChange(value.substring(0, start) + '  ' + value.substring(end))
      requestAnimationFrame(() => { ta.selectionStart = ta.selectionEnd = start + 2 })
      return
    }

    // ── Ctrl shortcuts ──
    if (e.ctrlKey || e.metaKey) {
      const s = ta.selectionStart; const ed = ta.selectionEnd
      const sel = value.substring(s, ed)
      const applyWrap = (w: [string, string]) => {
        e.preventDefault()
        onChange(value.substring(0, s) + w[0] + sel + w[1] + value.substring(ed))
        requestAnimationFrame(() => { ta.selectionStart = s + w[0].length; ta.selectionEnd = ed + w[0].length })
      }
      switch (e.key) {
        case 'b': applyWrap(['**', '**']); return
        case 'i': applyWrap(['_', '_']); return
        case 'k': applyWrap(['[', '](url)']); return
        default: break
      }
    }
  }, [slashOpen, slashIdx, executeSlash, onChange, value])

  // ── Input handler: detect / at line start ──
  const handleInput = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newVal = e.target.value; onChange(newVal)
    const ta = e.target; const cursor = ta.selectionStart
    const lineStart = getLineStart(newVal, cursor)
    const lineText = newVal.substring(lineStart, cursor)

    // Detect: at line start AND the last character typed is /
    if (lineText === '/' || (/^\s+\/$/).test(lineText)) {
      const lines = newVal.substring(0, lineStart).split('\n').length - 1
      const col = lineText.length
      const lineH = 15 * 1.8; const charW = 9.2
      setSlashPos({ top: 24 + lines * lineH + lineH + 4, left: 28 + col * charW })
      setSlashAnchor(cursor)
      setSlashIdx(0)
      setSlashOpen(true)
    } else if (slashOpen && !lineText.startsWith('/') && !(/^\s+\/$/).test(lineText)) {
      setSlashOpen(false)
    }
  }, [onChange, getLineStart, slashOpen])

  // ── Close slash on blur ──
  useEffect(() => {
    if (!slashOpen) return
    const close = () => setSlashOpen(false)
    const id = setTimeout(close, 5000) // auto-close after 5s
    return () => clearTimeout(id)
  }, [slashOpen])

  // ── Apply float format ──
  const applyFormat = useCallback((action: FormatAction) => {
    const ta = sourceRef.current; if (!ta) return
    const s = ta.selectionStart; const ed = ta.selectionEnd
    if (s === ed) return
    const sel = value.substring(s, ed)
    const [l, r] = action.wrap
    onChange(value.substring(0, s) + l + sel + r + value.substring(ed))
    requestAnimationFrame(() => { ta.selectionStart = s + l.length; ta.selectionEnd = ed + l.length; ta.focus() })
    setFloatOpen(false)
  }, [value, onChange])

  const isEmpty = !value || value.trim().length === 0

  return (
    <div className="blogo-editor-wrapper" style={{ display: 'flex', height: 'calc(100vh - 280px)', minHeight: 500, borderTop: '1px solid rgba(255,255,255,0.06)' }} ref={wrapRef}>
      {/* ── Left: Markdown source ── */}
      <div style={{ flex: 1, minWidth: 0, borderRight: '1px solid rgba(255,255,255,0.06)', display: 'flex', flexDirection: 'column', position: 'relative' }}>
        <div style={{ padding: '8px 20px', borderBottom: '1px solid rgba(255,255,255,0.04)', display: 'flex', alignItems: 'center', gap: 8, flexShrink: 0 }}>
          <span style={{ width: 8, height: 8, borderRadius: '50%', background: '#818cf8' }} />
          <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.3)', fontFamily: `"Barlow", sans-serif`, letterSpacing: '0.05em' }}>MARKDOWN</span>
        </div>
        <div style={{ flex: 1, position: 'relative' }}>
          <textarea ref={sourceRef} value={value}
            onChange={handleInput} onScroll={handleSourceScroll}
            onKeyDown={handleKeyDown} onSelect={handleSelect}
            onBlur={() => { setFloatOpen(false); /* slash auto-closes via timeout */ }}
            placeholder="开始写作..."
            autoFocus spellCheck={false}
            style={{ position: 'absolute', inset: 0, width: '100%', height: '100%', border: 'none', outline: 'none', resize: 'none', background: 'transparent', color: 'rgba(255,255,255,0.82)', fontFamily: "'JetBrains Mono','Fira Code','Consolas',monospace", fontSize: 15, lineHeight: '1.8', padding: '24px 28px', tabSize: 2 }} />

          {/* ── Empty-state hint ── */}
          {isEmpty && (
            <div style={{ position: 'absolute', top: 56, left: 28, right: 28, pointerEvents: 'none' }}>
              <span style={{ fontSize: 12, color: 'rgba(255,255,255,0.1)', lineHeight: 2, fontFamily: `"Barlow", sans-serif` }}>
                输入 / 唤起快捷格式菜单 · Ctrl+B 加粗 · Ctrl+I 斜体 · Ctrl+K 链接
              </span>
            </div>
          )}

          {/* ── Slash command menu ── */}
          {slashOpen && (
            <div style={{ position: 'absolute', top: slashPos.top, left: slashPos.left, zIndex: 100, background: 'rgba(16,16,32,0.97)', backdropFilter: 'blur(20px)', border: '1px solid rgba(255,255,255,0.1)', borderRadius: 12, padding: '6px 0', minWidth: 260, boxShadow: '0 12px 40px rgba(0,0,0,0.5)' }}>
              <div style={{ padding: '6px 14px 4px', fontSize: 10, color: 'rgba(255,255,255,0.25)', fontFamily: `"Barlow", sans-serif`, letterSpacing: '0.06em' }}>格式菜单</div>
              {slashCommands.map((cmd, i) => (
                <div key={cmd.id} onClick={() => executeSlash(cmd)}
                  onMouseEnter={() => setSlashIdx(i)}
                  style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '7px 14px', cursor: 'pointer', background: i === slashIdx ? 'rgba(79,110,247,0.15)' : 'transparent', borderLeft: i === slashIdx ? '2px solid #818cf8' : '2px solid transparent', transition: 'background 0.1s' }}>
                  <span style={{ width: 24, height: 24, borderRadius: 6, background: 'rgba(255,255,255,0.05)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 12, color: 'rgba(255,255,255,0.6)', flexShrink: 0, fontFamily: `"Barlow", sans-serif`, fontWeight: 700 }}>{cmd.icon}</span>
                  <span style={{ flex: 1, color: 'rgba(255,255,255,0.8)', fontSize: 13 }}>{cmd.label}</span>
                  <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.2)', fontFamily: "'JetBrains Mono', monospace" }}>{cmd.shortcut}</span>
                </div>
              ))}
            </div>
          )}

          {/* ── Floating selection toolbar ── */}
          {floatOpen && (
            <div style={{ position: 'absolute', top: Math.max(floatPos.top - 36, 4), left: floatPos.left, zIndex: 99, display: 'flex', gap: 2, background: 'rgba(16,16,32,0.96)', backdropFilter: 'blur(12px)', border: '1px solid rgba(255,255,255,0.1)', borderRadius: 10, padding: '4px 6px', boxShadow: '0 8px 24px rgba(0,0,0,0.4)' }}
              onMouseDown={e => e.preventDefault()}>
              {formatActions.map(a => (
                <button key={a.id} onClick={() => applyFormat(a)}
                  style={{ border: 'none', background: 'transparent', color: 'rgba(255,255,255,0.6)', padding: '6px 10px', borderRadius: 6, cursor: 'pointer', fontSize: 12, fontFamily: `"Barlow", sans-serif`, fontWeight: 500, transition: 'all 0.12s', lineHeight: 1 }}
                  onMouseEnter={e => { e.currentTarget.style.background = 'rgba(255,255,255,0.06)'; e.currentTarget.style.color = '#fff' }}
                  onMouseLeave={e => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'rgba(255,255,255,0.6)' }}>
                  {a.label}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* ── Right: Preview ── */}
      <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column' }}>
        <div style={{ padding: '8px 20px', borderBottom: '1px solid rgba(255,255,255,0.04)', display: 'flex', alignItems: 'center', gap: 8, flexShrink: 0 }}>
          <span style={{ width: 8, height: 8, borderRadius: '50%', background: '#34d399' }} />
          <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.3)', fontFamily: `"Barlow", sans-serif`, letterSpacing: '0.05em' }}>PREVIEW</span>
        </div>
        <div ref={previewRef} onScroll={handlePreviewScroll} style={{ flex: 1, overflow: 'auto', padding: '24px 28px' }}>
          <div className="wmde-markdown" style={{ color: 'rgba(255,255,255,0.82)', fontSize: 16, lineHeight: 1.9 }}>
            <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeHighlight]}>
              {value || '*暂无内容*'}
            </ReactMarkdown>
          </div>
        </div>
      </div>

      {/* ── Styles ── */}
      <style>{`
        .blogo-editor-wrapper,.blogo-editor-wrapper *,.blogo-editor-wrapper *::before,.blogo-editor-wrapper *::after{box-sizing:border-box}
        .blogo-editor-wrapper textarea::selection,.blogo-editor-wrapper textarea ::selection{background:rgba(79,110,247,0.35)!important;color:#fff!important}
        .blogo-editor-wrapper ::selection{background:rgba(79,110,247,0.3);color:#fff}
        .blogo-editor-wrapper .wmde-markdown h1{font-size:2em;font-weight:800;color:#fff;margin:1.4em 0 0.5em;border:none;padding:0}
        .blogo-editor-wrapper .wmde-markdown h2{font-size:1.5em;font-weight:700;color:#fff;margin:1.3em 0 0.4em;border:none;padding:0}
        .blogo-editor-wrapper .wmde-markdown h3{font-size:1.2em;font-weight:600;color:#e5e5e5;margin:1em 0 0.3em}
        .blogo-editor-wrapper .wmde-markdown p{margin:0 0 0.8em;word-break:break-word}
        .blogo-editor-wrapper .wmde-markdown code{background:rgba(255,255,255,0.06);color:#f472b6;padding:2px 6px;border-radius:4px;font-size:.88em}
        .blogo-editor-wrapper .wmde-markdown pre{background:rgba(0,0,0,0.4);border-radius:10px;padding:16px 20px;overflow-x:auto}
        .blogo-editor-wrapper .wmde-markdown pre code{background:transparent;padding:0;color:rgba(255,255,255,0.8)}
        .blogo-editor-wrapper .wmde-markdown blockquote{border-left:3px solid #818cf8;padding:4px 16px;margin:1em 0;color:rgba(255,255,255,0.5)}
        .blogo-editor-wrapper .wmde-markdown a{color:#818cf8}
        .blogo-editor-wrapper .wmde-markdown ul,.blogo-editor-wrapper .wmde-markdown ol{padding-left:1.6em;margin:.4em 0}
        .blogo-editor-wrapper .wmde-markdown li{margin:.2em 0}
        .blogo-editor-wrapper .wmde-markdown table{border-collapse:collapse;width:100%;margin:1em 0}
        .blogo-editor-wrapper .wmde-markdown th,.blogo-editor-wrapper .wmde-markdown td{border:1px solid rgba(255,255,255,0.1);padding:8px 12px;text-align:left}
        .blogo-editor-wrapper .wmde-markdown th{background:rgba(255,255,255,0.03)}
        .blogo-editor-wrapper .wmde-markdown img{max-width:100%;border-radius:8px}
        .blogo-editor-wrapper .wmde-markdown hr{border:none;height:1px;background:rgba(255,255,255,0.08);margin:2em 0}
      `}</style>
    </div>
  )
}

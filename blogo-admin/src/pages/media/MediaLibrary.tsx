import { useState, useMemo, useCallback, useRef, useEffect } from 'react'
import {
  Row, Col, Card, Typography, Upload, Input, Select, Button, Space, Pagination,
  Modal, Progress, message, Empty, Tooltip, Tag,
} from 'antd'
import {
  UploadOutlined, SearchOutlined, DeleteOutlined, CopyOutlined,
  EyeOutlined, FileImageOutlined, CloseCircleOutlined,
  PlayCircleOutlined, VideoCameraOutlined, CloudOutlined, CloudUploadOutlined,
} from '@ant-design/icons'
import dayjs from '../../utils/dayjs'

const { Title, Text } = Typography

// ── Types ──
interface MediaItem {
  id: string
  name: string
  url: string
  size: number
  width: number
  height: number
  type: string
  createdAt: string
  uploadStatus?: 'uploading' | 'done' | 'error'
  uploadProgress?: number
}

function isVideo(item: MediaItem) {
  return item.type.startsWith('video/')
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

// ── R2 CDN domain ──
const R2_CDN = 'https://media.blogo.dev'
const R2_UPLOAD_ENDPOINT = '/api/v1/images/upload'

// ── Mock data (R2-style URLs + 2 video entries) ──
const INITIAL_MOCK: MediaItem[] = [
  { id: 'm1', name: 'architecture-design.png', url: `${R2_CDN}/uploads/2026/05/17/architecture-design.png`, size: 1240000, width: 1920, height: 1080, type: 'image/png', createdAt: '2026-05-17 10:30:00' },
  { id: 'm2', name: 'blog-cover-hero.jpg', url: `${R2_CDN}/uploads/2026/05/16/blog-cover-hero.jpg`, size: 850000, width: 1600, height: 900, type: 'image/jpeg', createdAt: '2026-05-16 14:20:00' },
  { id: 'm3', name: 'team-avatar.png', url: `${R2_CDN}/uploads/2026/05/15/team-avatar.png`, size: 320000, width: 400, height: 400, type: 'image/png', createdAt: '2026-05-15 09:15:00' },
  { id: 'm4', name: 'intro-trailer.mp4', url: `${R2_CDN}/uploads/2026/05/14/intro-trailer.mp4`, size: 28500000, width: 1920, height: 1080, type: 'video/mp4', createdAt: '2026-05-14 16:45:00' },
  { id: 'm5', name: 'screenshot-dashboard.png', url: `${R2_CDN}/uploads/2026/05/13/screenshot-dashboard.png`, size: 2100000, width: 2560, height: 1440, type: 'image/png', createdAt: '2026-05-13 11:00:00' },
  { id: 'm6', name: 'photo-sunset.jpg', url: `${R2_CDN}/uploads/2026/05/12/photo-sunset.jpg`, size: 1560000, width: 2048, height: 1365, type: 'image/jpeg', createdAt: '2026-05-12 08:30:00' },
  { id: 'm7', name: 'demo-walkthrough.webm', url: `${R2_CDN}/uploads/2026/05/11/demo-walkthrough.webm`, size: 42000000, width: 1920, height: 1080, type: 'video/webm', createdAt: '2026-05-11 20:10:00' },
  { id: 'm8', name: 'header-banner.webp', url: `${R2_CDN}/uploads/2026/05/10/header-banner.webp`, size: 670000, width: 1920, height: 600, type: 'image/webp', createdAt: '2026-05-10 13:00:00' },
]

// ── Styles ──
const cardStyle = { borderRadius: 14, background: 'rgba(20,20,40,0.6)', border: '1px solid rgba(255,255,255,0.05)', backdropFilter: 'blur(12px)' }
const muted = { color: 'rgba(255,255,255,0.4)', fontSize: 12 }
const ACCEPT_TYPES = '.jpg,.jpeg,.png,.gif,.webp,.mp4,.webm'

export default function MediaLibrary() {
  const [items, setItems] = useState<MediaItem[]>(INITIAL_MOCK)
  const [search, setSearch] = useState('')
  const [sortOrder, setSortOrder] = useState<'newest' | 'oldest'>('newest')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(12)
  const [previewItem, setPreviewItem] = useState<MediaItem | null>(null)
  const intervalsRef = useRef<Set<ReturnType<typeof setInterval>>>(new Set())

  // Cleanup all intervals on unmount
  useEffect(() => {
    return () => { intervalsRef.current.forEach(clearInterval) }
  }, [])

  // ── Filtered & sorted ──
  const filtered = useMemo(() => {
    let list = [...items]
    if (search) list = list.filter(i => i.name.toLowerCase().includes(search.toLowerCase()))
    list.sort((a, b) => sortOrder === 'newest'
      ? new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      : new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime())
    return list
  }, [items, search, sortOrder])

  const paged = useMemo(() => {
    const start = (page - 1) * pageSize
    return filtered.slice(start, start + pageSize)
  }, [filtered, page, pageSize])

  // ── R2 custom upload (customRequest + onProgress) ──
  const customUpload = useCallback((options: any) => {
    const { file, onProgress, onSuccess, onError } = options
    const uploadId = `uploading-${Date.now()}`
    const fileType = file.type || (file.name.endsWith('.mp4') ? 'video/mp4' : file.name.endsWith('.webm') ? 'video/webm' : 'image/png')

    // Validate size (50MB limit)
    if (file.size > 50 * 1024 * 1024) {
      onError(new Error('文件大小超过 50MB 限制'))
      message.error(`${file.name} 超过 50MB 限制，无法上传至 R2`)
      return
    }

    // Show uploading card immediately
    const uploadingItem: MediaItem = {
      id: uploadId, name: file.name, url: '',
      size: file.size, width: 0, height: 0,
      type: fileType,
      createdAt: dayjs().format('YYYY-MM-DD HH:mm:ss'),
      uploadStatus: 'uploading', uploadProgress: 0,
    }
    setItems(prev => [uploadingItem, ...prev])

    // ── Simulate R2 upload with progress ──
    let progress = 0
    const interval = setInterval(() => {
      progress += Math.random() * 12 + 4
      if (progress >= 100) {
        progress = 100
        clearInterval(interval)
        intervalsRef.current.delete(interval)

        const ext = file.name.split('.').pop() || 'png'
        const now = dayjs()
        const r2Path = `/uploads/${now.format('YYYY')}/${now.format('MM')}/${now.format('DD')}/${uploadId}.${ext}`
        const cdnUrl = `${R2_CDN}${r2Path}`

        const completedItem: MediaItem = {
          id: uploadId, name: file.name, url: cdnUrl,
          size: file.size,
          width: fileType.startsWith('video/') ? 1920 : 1200,
          height: fileType.startsWith('video/') ? 1080 : 800,
          type: fileType,
          createdAt: now.format('YYYY-MM-DD HH:mm:ss'),
        }
        setItems(prev => prev.map(i => i.id === uploadId ? completedItem : i))
        onSuccess({ url: cdnUrl }, file)
        message.success(`${file.name} 已上传至 Cloudflare R2`)
      } else {
        setItems(prev => prev.map(i => i.id === uploadId ? { ...i, uploadProgress: progress } : i))
        onProgress({ percent: progress })
      }
    }, 180)
    intervalsRef.current.add(interval)
  }, [])

  // ── Copy CDN URL ──
  const handleCopyUrl = useCallback((url: string, itemName?: string) => {
    navigator.clipboard.writeText(url).then(() => {
      message.success('CDN 链接已成功复制到剪贴板！')
    }).catch(() => message.error('复制失败'))
  }, [])

  // ── Delete ──
  const handleDelete = useCallback((item: MediaItem) => {
    Modal.confirm({
      title: '确认删除',
      icon: <CloseCircleOutlined />,
      content: `删除「${item.name}」后，R2 存储桶中的文件将被移除，引用此媒体的文章将无法显示。是否确认删除？`,
      okText: '确认删除', okType: 'danger', cancelText: '取消',
      onOk: () => {
        setItems(prev => prev.filter(i => i.id !== item.id))
        message.success('已从 R2 删除')
      },
    })
  }, [])

  // ── Grid cell render ──
  const renderCell = (item: MediaItem) => {
    const video = isVideo(item)

    // Upload progress card
    if (item.uploadStatus === 'uploading') {
      return (
        <Card style={{ ...cardStyle, background: 'rgba(79,110,247,0.06)', border: '1px solid rgba(79,110,247,0.15)', height: '100%' }}
          styles={{ body: { padding: '24px 20px', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%' } }}>
          <CloudUploadOutlined style={{ fontSize: 32, color: '#818cf8', marginBottom: 12 }} />
          <Text style={{ color: 'rgba(255,255,255,0.7)', fontSize: 12, textAlign: 'center', overflow: 'hidden', textOverflow: 'ellipsis', maxWidth: '100%' }}>{item.name}</Text>
          <Progress percent={Math.round(item.uploadProgress || 0)} size="small" strokeColor="#818cf8" style={{ width: '80%', marginTop: 10 }} />
          <Text style={{ ...muted, fontSize: 10, marginTop: 4 }}>上传至 Cloudflare R2...</Text>
        </Card>
      )
    }

    return (
      <Card hoverable
        style={{ ...cardStyle, overflow: 'hidden', cursor: 'pointer', height: '100%', transition: 'all 0.25s' }}
        styles={{ body: { padding: 0 } }}
        className="media-card"
        onMouseEnter={e => { e.currentTarget.style.transform = 'translateY(-3px)'; e.currentTarget.style.borderColor = 'rgba(255,255,255,0.12)' }}
        onMouseLeave={e => { e.currentTarget.style.transform = ''; e.currentTarget.style.borderColor = 'rgba(255,255,255,0.05)' }}>
        {/* Preview area */}
        <div style={{ width: '100%', height: 0, paddingBottom: '75%', position: 'relative', overflow: 'hidden', background: 'rgba(0,0,0,0.3)' }}
          onClick={() => setPreviewItem(item)}>
          {video ? (
            <>
              <video src={item.url} muted preload="metadata"
                style={{ position: 'absolute', top: 0, left: 0, width: '100%', height: '100%', objectFit: 'cover', transition: 'transform 0.3s' }}
                className="media-img" />
              <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <PlayCircleOutlined style={{ fontSize: 40, color: 'rgba(255,255,255,0.85)', filter: 'drop-shadow(0 2px 8px rgba(0,0,0,0.5))', zIndex: 1 }} />
              </div>
            </>
          ) : (
            <img src={item.url} alt={item.name}
              style={{ position: 'absolute', top: 0, left: 0, width: '100%', height: '100%', objectFit: 'cover', transition: 'transform 0.3s' }}
              className="media-img" />
          )}
          {/* Hover overlay */}
          <div style={{ position: 'absolute', inset: 0, background: 'rgba(0,0,0,0.4)', opacity: 0, transition: 'opacity 0.2s', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}
            className="media-overlay">
            <Tooltip title="查看详情"><Button size="small" type="primary" ghost icon={<EyeOutlined />} onClick={e => { e.stopPropagation(); setPreviewItem(item) }} /></Tooltip>
            <Tooltip title="复制 CDN 链接"><Button size="small" ghost icon={<CopyOutlined />} onClick={e => { e.stopPropagation(); handleCopyUrl(item.url, item.name) }} /></Tooltip>
            <Tooltip title="删除"><Button size="small" danger ghost icon={<DeleteOutlined />} onClick={e => { e.stopPropagation(); handleDelete(item) }} /></Tooltip>
          </div>
        </div>
        {/* Info bar */}
        <div style={{ padding: '10px 12px' }}>
          <Text style={{ color: 'rgba(255,255,255,0.75)', fontSize: 12, display: 'block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {video && <VideoCameraOutlined style={{ color: '#f59e0b', marginRight: 4, fontSize: 11 }} />}
            {item.name}
          </Text>
          <Text style={{ ...muted, fontSize: 11 }}>
            {formatSize(item.size)}{item.width > 0 && ` · ${item.width}×${item.height}`}
          </Text>
        </div>
      </Card>
    )
  }

  return (
    <div>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <CloudOutlined style={{ fontSize: 22, color: '#f59e0b' }} />
          <Title level={4} style={{ margin: 0, fontWeight: 700 }}>媒体库</Title>
          <Tag color="orange" style={{ borderRadius: 6, fontSize: 10 }}>Cloudflare R2</Tag>
        </div>
        <Text type="secondary" style={{ fontSize: 13, display: 'block', marginTop: 4 }}>管理全站图片、视频与媒体资产 · 存储于 Cloudflare R2</Text>
      </div>

      {/* Top toolbar */}
      <div style={{ ...cardStyle, padding: '16px 20px', marginBottom: 16, display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 12 }}>
        <Space wrap>
          <Upload accept={ACCEPT_TYPES} showUploadList={false} customRequest={customUpload} multiple>
            <Button type="primary" icon={<UploadOutlined />} style={{ borderRadius: 8 }}>上传媒体</Button>
          </Upload>
          <Text style={{ ...muted, fontSize: 11 }}>JPG / PNG / WebP / GIF / MP4 / WebM · ≤ 50MB</Text>
        </Space>
        <Space>
          <Input prefix={<SearchOutlined />} placeholder="搜索文件名..." value={search}
            onChange={e => { setSearch(e.target.value); setPage(1) }} style={{ width: 220 }} allowClear />
          <Select value={sortOrder} onChange={setSortOrder} style={{ width: 120 }}
            options={[{ label: '最新上传', value: 'newest' }, { label: '最早上传', value: 'oldest' }]} />
        </Space>
      </div>

      {/* Image grid or empty */}
      {filtered.length === 0 ? (
        <Card style={{ ...cardStyle, textAlign: 'center' }} styles={{ body: { padding: '60px 24px' } }}>
          <Empty image={<FileImageOutlined style={{ fontSize: 64, color: 'rgba(255,255,255,0.1)' }} />}
            description={<span style={{ color: 'rgba(255,255,255,0.3)' }}>{search ? '没有匹配的媒体' : '媒体库为空，上传第一份素材至 Cloudflare R2'}</span>}>
            {!search && (
              <Upload accept={ACCEPT_TYPES} showUploadList={false} customRequest={customUpload} multiple>
                <Button type="primary" icon={<UploadOutlined />}>上传媒体</Button>
              </Upload>
            )}
          </Empty>
        </Card>
      ) : (
        <>
          <Row gutter={[14, 14]}>
            {paged.map(item => (
              <Col xl={4} lg={6} md={8} sm={12} xs={12} key={item.id}>
                {renderCell(item)}
              </Col>
            ))}
          </Row>
          {filtered.length > pageSize && (
            <div style={{ display: 'flex', justifyContent: 'center', marginTop: 20 }}>
              <Pagination current={page} pageSize={pageSize} total={filtered.length}
                showSizeChanger pageSizeOptions={['12', '24']}
                onChange={(p, ps) => { setPage(p); setPageSize(ps) }}
                showTotal={t => <span style={{ color: 'rgba(255,255,255,0.3)', fontSize: 12 }}>共 {t} 项</span>} />
            </div>
          )}
        </>
      )}

      {/* Preview modal (image + video) */}
      <Modal open={!!previewItem} onCancel={() => setPreviewItem(null)} footer={null} width={820} centered
        styles={{ body: { padding: 0, background: 'rgba(10,10,20,0.98)' } }}>
        {previewItem && (
          <div>
            {/* Top bar */}
            <div style={{ padding: '14px 24px', borderBottom: '1px solid rgba(255,255,255,0.06)', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
              <Space>
                <Text strong style={{ color: '#fff', fontSize: 15 }}>{previewItem.name}</Text>
                <Tag color="orange" style={{ borderRadius: 6, fontSize: 10 }}>Cloudflare R2</Tag>
              </Space>
              <Button icon={<CopyOutlined />} onClick={() => handleCopyUrl(previewItem.url, previewItem.name)} style={{ borderRadius: 8 }}>
                复制 CDN 链接
              </Button>
            </div>

            {/* Preview content */}
            <div style={{ display: 'flex', justifyContent: 'center', padding: 20, background: 'rgba(0,0,0,0.4)', minHeight: 200 }}>
              {isVideo(previewItem) ? (
                <video controls width="100%" src={previewItem.url}
                  style={{ maxHeight: 480, borderRadius: 8, background: '#000' }} />
              ) : (
                <img src={previewItem.url} alt={previewItem.name}
                  style={{ maxWidth: '100%', maxHeight: 480, objectFit: 'contain', borderRadius: 8 }} />
              )}
            </div>

            {/* Metadata */}
            <div style={{ padding: '16px 24px', display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px 24px' }}>
              <MetaRow label="分辨率" value={`${previewItem.width} × ${previewItem.height}`} />
              <MetaRow label="文件大小" value={formatSize(previewItem.size)} />
              <MetaRow label="文件类型" value={previewItem.type} />
              <MetaRow label="上传时间" value={dayjs(previewItem.createdAt).format('YYYY-MM-DD HH:mm:ss')} />
              <MetaRow label="CDN URL" value={previewItem.url} />
              <MetaRow label="存储源" value={<Tag color="orange" style={{ margin: 0, fontSize: 10 }}>Cloudflare R2</Tag>} />
            </div>
          </div>
        )}
      </Modal>

      <style>{`
        .media-card:hover .media-img { transform: scale(1.08); }
        .media-card:hover .media-overlay { opacity: 1; }
      `}</style>
    </div>
  )
}

function MetaRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', gap: 8, fontSize: 12, alignItems: 'center' }}>
      <Text style={{ color: 'rgba(255,255,255,0.35)', minWidth: 56 }}>{label}</Text>
      <Text style={{ color: 'rgba(255,255,255,0.7)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{value}</Text>
    </div>
  )
}

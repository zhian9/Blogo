import { useState, useRef } from 'react'
import { Modal, Button, Spin, message, Row, Col, Card, Typography, Space, Upload, Empty } from 'antd'
import {
  PlusOutlined, DeleteOutlined, PictureOutlined,
  CloudUploadOutlined, FolderOpenOutlined,
} from '@ant-design/icons'
import { useImages } from '../hooks/useImages'
import client from '../api/client'
import type { ImageItem } from '../api/images'

const { Text } = Typography

interface Props {
  value?: string       // image ID (form controlled)
  onChange?: (id: string) => void
  coverUrl?: string    // existing cover URL (edit mode)
}

export default function CoverImageUpload({ value, onChange, coverUrl }: Props) {
  const [uploading, setUploading] = useState(false)
  const [previewUrl, setPreviewUrl] = useState(coverUrl || '')
  const [modalOpen, setModalOpen] = useState(false)
  const [mode, setMode] = useState<'upload' | 'library' | null>(null)
  const [pickerPage, setPickerPage] = useState(1)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const hasImage = !!value
  const { data: imageData } = useImages({ pageSize: 50 })
  const images = imageData?.data || []

  // ── R2 upload ──
  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    if (!file.type.startsWith('image/')) { message.error('仅支持图片格式'); return }
    if (file.size > 5 * 1024 * 1024) { message.error('图片大小不能超过 5MB'); return }

    const localUrl = URL.createObjectURL(file)
    setPreviewUrl(localUrl)
    setUploading(true)
    try {
      const formData = new FormData(); formData.append('file', file); formData.append('category', 'article_cover')
      const res = await client.post('/images/upload', formData)
      const imageId = res.data.data.id
      const imageUrl = res.data.data.url
      URL.revokeObjectURL(localUrl)
      setPreviewUrl(imageUrl)
      onChange?.(imageId)
      setModalOpen(false)
      setMode(null)
      message.success('封面上传成功')
    } catch (err: any) {
      URL.revokeObjectURL(localUrl)
      setPreviewUrl('')
      message.error(err.message || '上传失败')
    } finally { setUploading(false) }
  }

  // ── Pick from library ──
  const handlePickFromLibrary = (img: ImageItem) => {
    setPreviewUrl(img.url)
    onChange?.(img.id)
    setModalOpen(false)
    setMode(null)
    message.success('已选择封面')
  }

  // ── Remove ──
  const handleRemove = (e: React.MouseEvent) => {
    e.stopPropagation()
    setPreviewUrl('')
    onChange?.('')
  }

  return (
    <div>
      <input ref={fileInputRef} type="file" accept="image/*" style={{ display: 'none' }} onChange={handleFileChange} />

      {/* ── 16:9 preview card ── */}
      <div
        onClick={() => !uploading && setModalOpen(true)}
        style={{
          width: '100%', maxWidth: 400, aspectRatio: '16/9',
          borderRadius: 14, overflow: 'hidden', cursor: 'pointer',
          border: hasImage ? '1px solid rgba(255,255,255,0.08)' : '2px dashed rgba(255,255,255,0.12)',
          background: hasImage ? 'transparent' : 'rgba(255,255,255,0.02)',
          position: 'relative', transition: 'border-color 0.2s',
        }}
        onMouseEnter={e => e.currentTarget.style.borderColor = 'rgba(255,255,255,0.2)'}
        onMouseLeave={e => e.currentTarget.style.borderColor = hasImage ? 'rgba(255,255,255,0.08)' : 'rgba(255,255,255,0.12)'}
      >
        {hasImage ? (
          <>
            <img src={previewUrl} alt="cover" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
            <div onClick={handleRemove} style={{
              position: 'absolute', top: 8, right: 8, width: 30, height: 30,
              borderRadius: 8, background: 'rgba(0,0,0,0.5)', display: 'flex',
              alignItems: 'center', justifyContent: 'center', cursor: 'pointer',
              opacity: 0, transition: 'opacity 0.2s',
            }} className="cover-delete-btn">
              <DeleteOutlined style={{ color: '#fff', fontSize: 14 }} />
            </div>
            <style>{`.cover-delete-btn:hover, .cover-preview:hover .cover-delete-btn { opacity: 1 !important; }`}</style>
          </>
        ) : uploading ? (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', gap: 12 }}>
            <Spin size="small" />
            <Text style={{ color: 'rgba(255,255,255,0.3)', fontSize: 12 }}>上传中...</Text>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', gap: 8 }}>
            <PlusOutlined style={{ fontSize: 28, color: 'rgba(255,255,255,0.2)' }} />
            <Text style={{ color: 'rgba(255,255,255,0.3)', fontSize: 13 }}>选择封面</Text>
          </div>
        )}
      </div>

      {/* ── Selection modal ── */}
      <Modal open={modalOpen} onCancel={() => { setModalOpen(false); setMode(null) }} footer={null}
        width={mode === 'library' ? 800 : 420} centered
        styles={{ body: { padding: 0, background: 'rgba(12,12,24,0.98)' } }}>
        {!mode ? (
          <div style={{ padding: '32px 24px', display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Text strong style={{ color: '#fff', fontSize: 16, textAlign: 'center', marginBottom: 8 }}>选择封面来源</Text>
            <Button size="large" block icon={<CloudUploadOutlined />} onClick={() => { setMode('upload'); fileInputRef.current?.click() }}
              style={{ height: 56, borderRadius: 12, background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.08)', color: '#fff', fontSize: 15 }}>
              本地上传
            </Button>
            <Button size="large" block icon={<FolderOpenOutlined />} onClick={() => setMode('library')}
              style={{ height: 56, borderRadius: 12, background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.08)', color: '#fff', fontSize: 15 }}>
              从媒体库选择
            </Button>
          </div>
        ) : mode === 'library' ? (
          <div style={{ padding: '20px 24px', maxHeight: '70vh', overflow: 'auto' }}>
            <Text strong style={{ color: '#fff', fontSize: 15, display: 'block', marginBottom: 16 }}>从媒体库选择封面</Text>
            {images.length === 0 ? (
              <Empty description={<span style={{ color: 'rgba(255,255,255,0.25)' }}>媒体库为空</span>} />
            ) : (
              <Row gutter={[10, 10]}>
                {images.map((img: ImageItem) => (
                  <Col span={8} key={img.id}>
                    <Card hoverable size="small" onClick={() => handlePickFromLibrary(img)}
                      style={{ borderRadius: 10, background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.06)', overflow: 'hidden' }}
                      styles={{ body: { padding: 0 } }}>
                      <div style={{ width: '100%', paddingBottom: '56.25%', position: 'relative', overflow: 'hidden', background: 'rgba(0,0,0,0.3)' }}>
                        <img src={img.url} alt={img.name} style={{ position: 'absolute', inset: 0, width: '100%', height: '100%', objectFit: 'cover' }} />
                      </div>
                      <div style={{ padding: '8px 10px' }}>
                        <Text style={{ color: 'rgba(255,255,255,0.6)', fontSize: 11, display: 'block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{img.name}</Text>
                      </div>
                    </Card>
                  </Col>
                ))}
              </Row>
            )}
          </div>
        ) : null}
      </Modal>
    </div>
  )
}

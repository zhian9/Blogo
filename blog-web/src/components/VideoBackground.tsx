import { useEffect, useRef } from 'react'

interface Props {
  src: string
}

export default function VideoBackground({ src }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null)

  useEffect(() => {
    const v = videoRef.current
    if (!v) return
    console.log('🎬 VideoBackground:', src)
    v.play().then(() => console.log('✅ playing')).catch((e) => console.error('❌', e.name))
  }, [src])

  return (
    <video
      ref={videoRef}
      key={src}
      src={src}
      autoPlay
      muted
      loop
      playsInline
      preload="auto"
      style={{
        position: 'absolute',
        inset: 0,
        width: '100%',
        height: '100%',
        objectFit: 'cover',
        objectPosition: 'center center',
        // no z-index — DOM order controls layering
        filter: 'brightness(0.85) saturate(1.2) contrast(1.1)',
        background: '#020208',
      }}
    />
  )
}

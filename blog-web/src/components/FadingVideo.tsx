import { useRef, useEffect, useState, useCallback } from 'react'

interface Props {
  src: string
  fadeDuration?: number
  className?: string
}

/**
 * FadingVideo — smooth crossfade loop video background.
 *
 * Two overlapping <video> elements. The active video plays to near-end,
 * then the standby fades in and starts from 0. This avoids visible
 * seams or black flashes on loop.
 */
export default function FadingVideo({ src, fadeDuration = 1500, className }: Props) {
  const [activeIndex, setActiveIndex] = useState(0)
  const [opacities, setOpacities] = useState<[number, number]>([1, 0])
  const [loaded, setLoaded] = useState(false)

  const video0Ref = useRef<HTMLVideoElement>(null)
  const video1Ref = useRef<HTMLVideoElement>(null)
  const videoRefs = [video0Ref, video1Ref]

  const fadingRef = useRef(false)
  const activeRef = useRef(0) // always current, avoids stale closures

  // Keep ref in sync
  useEffect(() => {
    activeRef.current = activeIndex
  }, [activeIndex])

  // ── Crossfade trigger ──
  const doCrossfade = useCallback(() => {
    if (fadingRef.current) return
    fadingRef.current = true

    const current = activeRef.current
    const next = current === 0 ? 1 : 0

    // Swap opacities: fade out current, fade in next
    setOpacities(current === 0 ? [0, 1] : [1, 0])
    setActiveIndex(next)

    // Start the next video from beginning
    const nextVideo = videoRefs[next].current
    if (nextVideo) {
      nextVideo.currentTime = 0
      nextVideo.play().catch(() => {})
    }

    // After fade completes, pause the old video
    setTimeout(() => {
      fadingRef.current = false
      const oldVideo = videoRefs[current].current
      if (oldVideo && activeRef.current !== current) {
        oldVideo.pause()
      }
    }, fadeDuration)
  }, [fadeDuration])

  // ── Bootstrap: preload + autoplay both, then start video 0 ──
  useEffect(() => {
    const v0 = video0Ref.current
    const v1 = video1Ref.current

    if (!v0 || !v1) return

    // Preload metadata for both
    v0.load()
    v1.load()

    let mounted = true

    const startPlayback = async () => {
      try {
        // Ensure video 1 is preloaded and paused at start
        v1.pause()
        v1.currentTime = 0

        // Play video 0
        await v0.play()
        if (mounted) setLoaded(true)
      } catch {
        // Autoplay blocked — show a one-click play button
        if (mounted) setLoaded(true) // still show overlay
      }
    }

    // Wait for video 0 to have enough data
    if (v0.readyState >= 2) {
      startPlayback()
    } else {
      const onCanPlay = () => {
        v0.removeEventListener('canplay', onCanPlay)
        startPlayback()
      }
      v0.addEventListener('canplay', onCanPlay)
    }

    return () => { mounted = false }
  }, [src])

  // ── Listen for near-end on active video ──
  useEffect(() => {
    const activeVideo = videoRefs[activeIndex].current
    if (!activeVideo) return

    const onTimeUpdate = () => {
      if (!activeVideo.duration || fadingRef.current) return
      // Trigger crossfade 1.5s before the end
      if (activeVideo.currentTime >= activeVideo.duration - 1.5) {
        doCrossfade()
      }
    }

    activeVideo.addEventListener('timeupdate', onTimeUpdate)
    return () => activeVideo.removeEventListener('timeupdate', onTimeUpdate)
  }, [activeIndex, doCrossfade])

  return (
    <div
      className={className}
      style={{
        position: 'absolute',
        inset: 0,
        overflow: 'hidden',
        zIndex: 0,
        background: '#000',
      }}
    >
      {/* Fallback gradient while loading */}
      {!loaded && (
        <div style={{
          position: 'absolute', inset: 0, zIndex: 1,
          background: 'linear-gradient(160deg, #050510 0%, #0a0a20 40%, #020208 100%)',
        }} />
      )}

      {[0, 1].map((i) => (
        <video
          key={i}
          ref={i === 0 ? video0Ref : video1Ref}
          src={src}
          muted
          autoPlay
          playsInline
          preload="auto"
          disablePictureInPicture
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            objectFit: 'cover',
            objectPosition: 'center top',
            opacity: opacities[i],
            transition: `opacity ${fadeDuration}ms cubic-bezier(0.4, 0, 0.2, 1)`,
            willChange: 'opacity',
          }}
        />
      ))}
    </div>
  )
}

import { motion } from 'framer-motion'

export default function HeroIllustration() {
  const floatingVariants = {
    animate: {
      y: [0, -20, 0],
      transition: {
        duration: 3,
        repeat: Infinity,
        ease: 'easeInOut',
      },
    },
  }

  return (
    <motion.div
      className="relative w-full h-full flex items-center justify-center"
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.8, delay: 0.3 }}
    >
      <svg
        viewBox="0 0 400 400"
        className="w-full max-w-md"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        {/* 背景渐变圆 */}
        <defs>
          <linearGradient id="bgGradient" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" style={{ stopColor: '#1890ff', stopOpacity: 0.1 }} />
            <stop offset="100%" style={{ stopColor: '#722ed1', stopOpacity: 0.1 }} />
          </linearGradient>
          <linearGradient id="heroGradient" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" style={{ stopColor: '#1890ff', stopOpacity: 1 }} />
            <stop offset="100%" style={{ stopColor: '#722ed1', stopOpacity: 1 }} />
          </linearGradient>
        </defs>

        {/* 背景气泡 */}
        <circle cx="100" cy="80" r="40" fill="url(#bgGradient)" />
        <circle cx="320" cy="150" r="50" fill="url(#bgGradient)" />
        <circle cx="80" cy="280" r="35" fill="url(#bgGradient)" />

        {/* 程序员身体 */}
        <g>
          {/* 头 */}
          <circle cx="200" cy="120" r="30" fill="url(#heroGradient)" />

          {/* 脸部细节 */}
          <circle cx="190" cy="115" r="6" fill="#fff" />
          <circle cx="210" cy="115" r="6" fill="#fff" />
          <path d="M 190 130 Q 200 135 210 130" stroke="#fff" strokeWidth="2" fill="none" />

          {/* 身体 */}
          <rect x="175" y="155" width="50" height="60" rx="8" fill="url(#heroGradient)" />

          {/* 左臂（拿着键盘）*/}
          <g>
            <line x1="175" y1="165" x2="140" y2="185" stroke="url(#heroGradient)" strokeWidth="8" strokeLinecap="round" />
            {/* 键盘 */}
            <rect x="125" y="175" width="35" height="20" rx="4" fill="#1f1f1f" stroke="#1890ff" strokeWidth="2" />
            <g fill="#1890ff" opacity="0.7">
              <rect x="129" y="179" width="4" height="4" />
              <rect x="136" y="179" width="4" height="4" />
              <rect x="143" y="179" width="4" height="4" />
              <rect x="129" y="186" width="4" height="4" />
              <rect x="136" y="186" width="4" height="4" />
              <rect x="143" y="186" width="4" height="4" />
            </g>
          </g>

          {/* 右臂 */}
          <line x1="225" y1="165" x2="260" y2="180" stroke="url(#heroGradient)" strokeWidth="8" strokeLinecap="round" />

          {/* 左腿 */}
          <line x1="185" y1="215" x2="180" y2="270" stroke="url(#heroGradient)" strokeWidth="8" strokeLinecap="round" />
          {/* 左鞋 */}
          <ellipse cx="180" cy="275" rx="8" ry="6" fill="#1f1f1f" />

          {/* 右腿 */}
          <line x1="215" y1="215" x2="220" y2="270" stroke="url(#heroGradient)" strokeWidth="8" strokeLinecap="round" />
          {/* 右鞋 */}
          <ellipse cx="220" cy="275" rx="8" ry="6" fill="#1f1f1f" />
        </g>

        {/* 浮动代码块 */}
        <motion.g variants={floatingVariants} animate="animate">
          <rect x="260" y="50" width="120" height="80" rx="8" fill="#1f1f1f" stroke="#1890ff" strokeWidth="2" opacity="0.9" />
          <text x="270" y="70" fontSize="10" fill="#1890ff" fontFamily="monospace">
            {'const hero = () =>'}
          </text>
          <text x="270" y="85" fontSize="10" fill="#52c41a" fontFamily="monospace">
            {'  return <Hero />'}
          </text>
          <text x="270" y="100" fontSize="10" fill="#ff7875" fontFamily="monospace">
            {'}'}
          </text>
          <text x="270" y="115" fontSize="9" fill="#8c8c8c" fontFamily="monospace">
            {'export default'}
          </text>
        </motion.g>

        {/* 浮动代码片段 */}
        <motion.g
          animate={{ y: [20, -20, 20], rotate: [0, 5, 0] }}
          transition={{ duration: 4, repeat: Infinity, ease: 'easeInOut', delay: 0.5 }}
        >
          <rect x="50" y="180" width="90" height="60" rx="6" fill="#fff" stroke="#722ed1" strokeWidth="2" opacity="0.85" />
          <text x="60" y="200" fontSize="9" fill="#722ed1" fontFamily="monospace">
            {'React'}
          </text>
          <text x="60" y="215" fontSize="9" fill="#1890ff" fontFamily="monospace">
            {'+ Go'}
          </text>
          <text x="60" y="230" fontSize="8" fill="#666" fontFamily="monospace">
            {'Blog'}
          </text>
        </motion.g>

        {/* 顶部装饰星星 */}
        <motion.g
          animate={{ opacity: [0.3, 1, 0.3] }}
          transition={{ duration: 2, repeat: Infinity, ease: 'easeInOut' }}
        >
          <circle cx="320" cy="40" r="3" fill="#1890ff" />
          <circle cx="360" cy="80" r="2" fill="#722ed1" />
          <circle cx="50" cy="50" r="2.5" fill="#52c41a" />
        </motion.g>
      </svg>
    </motion.div>
  )
}

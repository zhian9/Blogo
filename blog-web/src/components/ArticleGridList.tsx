import { Empty, Spin } from 'antd'
import { motion } from 'framer-motion'
import ArticleGridCard from './ArticleGridCard'
import { useAppStore } from '../store/appStore'
import type { Article } from '../types'

interface Props {
  articles: Article[]
  loading: boolean
}

export default function ArticleGridList({ articles, loading }: Props) {
  const theme = useAppStore((s) => s.theme)

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.08,
        delayChildren: 0.1,
      },
    },
  }

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '80px 20px' }}>
        <Spin size="large" />
      </div>
    )
  }

  if (articles.length === 0) {
    return (
      <Empty
        description="暂无文章"
        style={{
          padding: '80px 20px',
          color: theme === 'dark' ? '#666' : '#999',
        }}
      />
    )
  }

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
        gap: '24px',
        width: '100%',
      }}
    >
      {articles.map((article, idx) => (
        <ArticleGridCard
          key={article.id}
          article={article}
          index={idx}
        />
      ))}
    </motion.div>
  )
}

import { useState, useMemo } from 'react'
import { Row, Col, Pagination, Space } from 'antd'
import { motion } from 'framer-motion'
import ArticleListHeader from '../components/ArticleListHeader'
import ArticleGridList from '../components/ArticleGridList'
import TagCloud from '../components/TagCloud'
import NewsletterMini from '../components/NewsletterMini'
import Sidebar from '../components/Sidebar'
import { useArticles } from '../hooks/useArticles'
import { useCategories } from '../hooks/useCategories'
import { useTags } from '../hooks/useTags'
import { useAppStore } from '../store/appStore'

const PAGE_SIZE = 12

export default function Articles() {
  const theme = useAppStore((s) => s.theme)

  // 状态管理
  const [search, setSearch] = useState('')
  const [categoryId, setCategoryId] = useState<string | undefined>()
  const [sortBy, setSortBy] = useState<'latest' | 'hot'>('latest')
  const [selectedTags, setSelectedTags] = useState<string[]>([])
  const [currentPage, setCurrentPage] = useState(1)

  // API 调用
  const { data: articlesData, isLoading: articlesLoading } = useArticles({
    current: currentPage,
    pageSize: PAGE_SIZE,
    title: search || undefined,
    category_id: categoryId,
    status: 'published',
  })

  const { data: categoriesData } = useCategories()
  const { data: tagsData } = useTags()

  const categories = categoriesData?.data || []
  const allTags = tagsData?.data || []

  // 获取数据
  const articles = articlesData?.data || []
  const total = articlesData?.total || 0

  // 本地排序（如果后端不支持，可在此处理）
  const sortedArticles = useMemo(() => {
    if (!Array.isArray(articles)) return []

    let sorted = [...articles]

    // 应用标签过滤（前端过滤）
    if (selectedTags.length > 0) {
      sorted = sorted.filter((article) => {
        // 假设 article 中有 tags 字段，如果没有则跳过过滤
        if (!Array.isArray((article as any).tags)) return true
        return selectedTags.some((tag) =>
          (article as any).tags.some((t: any) => t.name === tag || t === tag),
        )
      })
    }

    // 排序
    if (sortBy === 'hot') {
      sorted.sort((a, b) => (b.views || 0) - (a.views || 0))
    } else {
      // latest - 已由后端排序，这里保持原序
    }

    return sorted
  }, [articles, selectedTags, sortBy])

  // 分页
  const paginatedArticles = sortedArticles.slice(0, PAGE_SIZE)
  const displayTotal = total // 注意：如果前端过滤了标签，总数可能不准确

  const pageVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: { duration: 0.5 },
    },
  }

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={pageVariants}
      style={{
        padding: '0',
      }}
    >
      {/* 头部搜索和筛选 */}
      <div
        style={{
          padding: '40px 24px',
          borderBottom: `1px solid ${theme === 'dark' ? '#303030' : '#f0f0f0'}`,
          marginBottom: 40,
        }}
      >
        <ArticleListHeader
          search={search}
          onSearchChange={setSearch}
          categories={categories}
          selectedCategory={categoryId}
          onCategoryChange={setCategoryId}
          sortBy={sortBy}
          onSortChange={setSortBy}
          selectedTags={selectedTags}
          onTagsChange={setSelectedTags}
          allTags={allTags}
        />
      </div>

      {/* 主内容区域 */}
      <div
        style={{
          maxWidth: 1400,
          margin: '0 auto',
          padding: '0 24px',
          marginBottom: 60,
        }}
      >
        <Row gutter={[32, 32]}>
          {/* 左侧：文章列表 */}
          <Col xs={24} lg={16}>
            {/* 文章网格 */}
            <ArticleGridList
              articles={paginatedArticles}
              loading={articlesLoading}
            />

            {/* 分页 */}
            {paginatedArticles.length > 0 && (
              <motion.div
                initial={{ opacity: 0 }}
                whileInView={{ opacity: 1 }}
                transition={{ duration: 0.5, delay: 0.2 }}
                viewport={{ once: true }}
                style={{
                  textAlign: 'center',
                  marginTop: 40,
                }}
              >
                <Pagination
                  current={currentPage}
                  pageSize={PAGE_SIZE}
                  total={displayTotal}
                  onChange={setCurrentPage}
                  showSizeChanger={false}
                  showQuickJumper
                  showTotal={(total) => `共 ${total} 篇文章`}
                  style={{
                    color: theme === 'dark' ? '#fff' : '#000',
                  }}
                />
              </motion.div>
            )}
          </Col>

          {/* 右侧：侧边栏 */}
          <Col xs={24} lg={8}>
            <Space vertical size={24} style={{ width: '100%' }}>
              {/* 标签云 */}
              <TagCloud
                tags={allTags}
                selectedTags={selectedTags}
                onTagClick={(tagName) => {
                  const newTags = selectedTags.includes(tagName)
                    ? selectedTags.filter((t) => t !== tagName)
                    : [...selectedTags, tagName]
                  setSelectedTags(newTags)
                }}
                loading={false}
              />

              {/* Newsletter 小卡片 */}
              <NewsletterMini />

              {/* 热门文章和分类（原 Sidebar） */}
              <motion.div
                initial={{ opacity: 0, x: 20 }}
                whileInView={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.6 }}
                viewport={{ once: true }}
              >
                <Sidebar />
              </motion.div>
            </Space>
          </Col>
        </Row>
      </div>

      {/* 底部间隔 */}
      <div style={{ height: 40 }} />
    </motion.div>
  )
}

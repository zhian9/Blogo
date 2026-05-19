import { Input, Select, Space, Button, Popover, Badge, Tag } from 'antd'
import { SearchOutlined, FilterOutlined, SortAscendingOutlined } from '@ant-design/icons'
import { motion } from 'framer-motion'
import { useAppStore } from '../store/appStore'
import type { Category, Tag as TagType } from '../types'

interface Props {
  search: string
  onSearchChange: (value: string) => void
  categories: Category[]
  selectedCategory: string | undefined
  onCategoryChange: (value: string | undefined) => void
  sortBy: 'latest' | 'hot'
  onSortChange: (value: 'latest' | 'hot') => void
  selectedTags: string[]
  onTagsChange: (tags: string[]) => void
  allTags: TagType[]
}

export default function ArticleListHeader({
  search,
  onSearchChange,
  categories,
  selectedCategory,
  onCategoryChange,
  sortBy,
  onSortChange,
  selectedTags,
  onTagsChange,
  allTags,
}: Props) {
  const theme = useAppStore((s) => s.theme)

  const headerVariants = {
    hidden: { opacity: 0, y: -20 },
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: 0.6, ease: 'easeOut' },
    },
  }

  const searchVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: 0.6, delay: 0.1, ease: 'easeOut' },
    },
  }

  const filterVariants = {
    hidden: { opacity: 0, x: -20 },
    visible: {
      opacity: 1,
      x: 0,
      transition: { duration: 0.6, delay: 0.2, ease: 'easeOut' },
    },
  }

  const activeFilters = (selectedCategory ? 1 : 0) + (selectedTags.length || 0) + (sortBy === 'hot' ? 1 : 0)

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={headerVariants}
      style={{
        marginBottom: 40,
      }}
    >
      {/* 大标题 */}
      <motion.h1
        style={{
          fontSize: 'clamp(28px, 6vw, 42px)',
          fontWeight: 800,
          marginBottom: 24,
          color: theme === 'dark' ? '#fff' : '#000',
          textAlign: 'center',
        }}
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.6 }}
      >
        全部文章
      </motion.h1>

      {/* 搜索框 */}
      <motion.div variants={searchVariants}>
        <Input
          size="large"
          placeholder="搜索文章标题..."
          prefix={<SearchOutlined />}
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          style={{
            borderRadius: 8,
            fontSize: 15,
            height: 48,
            marginBottom: 24,
            background: theme === 'dark' ? '#1a1a1a' : '#fff',
            border: `2px solid ${theme === 'dark' ? '#303030' : '#e6e6e6'}`,
            color: theme === 'dark' ? '#fff' : '#000',
          }}
          onFocus={(e) => {
            e.currentTarget.style.borderColor = '#1890ff'
            e.currentTarget.style.boxShadow = '0 0 0 3px rgba(24, 144, 255, 0.15)'
          }}
          onBlur={(e) => {
            e.currentTarget.style.borderColor = theme === 'dark' ? '#303030' : '#e6e6e6'
            e.currentTarget.style.boxShadow = 'none'
          }}
        />
      </motion.div>

      {/* 筛选和排序 */}
      <motion.div variants={filterVariants}>
        <Space
          wrap
          style={{
            width: '100%',
            gap: '16px',
            marginBottom: 20,
          }}
        >
          {/* 分类选择 */}
          <Select
            placeholder="选择分类"
            allowClear
            value={selectedCategory}
            onChange={onCategoryChange}
            style={{ minWidth: 140, borderRadius: 6 }}
            options={[
              { label: '全部分类', value: '' },
              ...categories.map((cat) => ({
                label: cat.name,
                value: cat.id,
              })),
            ]}
          />

          {/* 排序 */}
          <Select
            value={sortBy}
            onChange={onSortChange}
            style={{ minWidth: 120, borderRadius: 6 }}
            options={[
              { label: '最新', value: 'latest', icon: <SortAscendingOutlined /> },
              { label: '最热', value: 'hot', icon: <SortAscendingOutlined /> },
            ]}
          />

          {/* 标签过滤 */}
          <Popover
            title="选择标签"
            content={
              <div style={{ maxWidth: 300 }}>
                <Space>
                  {allTags.map((tag) => (
                    <Badge
                      key={tag.id}
                      count={selectedTags.includes(tag.name) ? '✓' : ''}
                      style={{
                        backgroundColor: selectedTags.includes(tag.name) ? '#1890ff' : '#d9d9d9',
                      }}
                    >
                      <Tag
                        onClick={() => {
                          const newTags = selectedTags.includes(tag.name)
                            ? selectedTags.filter((t) => t !== tag.name)
                            : [...selectedTags, tag.name]
                          onTagsChange(newTags)
                        }}
                        style={{
                          cursor: 'pointer',
                          borderRadius: 4,
                          padding: '4px 8px',
                          fontSize: 12,
                          userSelect: 'none',
                          border: `1px solid ${selectedTags.includes(tag.name) ? '#1890ff' : '#d9d9d9'}`,
                          color: selectedTags.includes(tag.name) ? '#1890ff' : theme === 'dark' ? '#aaa' : '#666',
                        }}
                      >
                        {tag.name}
                      </Tag>
                    </Badge>
                  ))}
                </Space>
              </div>
            }
            trigger="click"
          >
            <Button
              icon={<FilterOutlined />}
              style={{
                borderRadius: 6,
                height: 36,
              }}
            >
              标签
              {selectedTags.length > 0 && (
                <Badge
                  count={selectedTags.length}
                  style={{
                    backgroundColor: '#1890ff',
                    marginLeft: 4,
                  }}
                />
              )}
            </Button>
          </Popover>

          {/* 活跃筛选指示 */}
          {activeFilters > 0 && (
            <Button
              type="text"
              size="small"
              onClick={() => {
                onCategoryChange(undefined)
                onTagsChange([])
                onSortChange('latest')
              }}
              style={{
                color: '#ff7875',
                marginLeft: 'auto',
              }}
            >
              清除筛选 ({activeFilters})
            </Button>
          )}
        </Space>
      </motion.div>

      {/* 活跃标签展示 */}
      {(selectedCategory || selectedTags.length > 0 || sortBy === 'hot') && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.3 }}
          style={{
            display: 'flex',
            gap: 8,
            flexWrap: 'wrap',
            marginTop: 16,
          }}
        >
          {selectedCategory && (
            <Tag closable onClose={() => onCategoryChange(undefined)}>
              {categories.find((c) => c.id === selectedCategory)?.name}
            </Tag>
          )}
          {selectedTags.map((tag) => (
            <Tag
              key={tag}
              closable
              onClose={() => onTagsChange(selectedTags.filter((t) => t !== tag))}
            >
              {tag}
            </Tag>
          ))}
          {sortBy === 'hot' && (
            <Tag closable onClose={() => onSortChange('latest')}>
              最热排序
            </Tag>
          )}
        </motion.div>
      )}
    </motion.div>
  )
}

import { useState } from 'react'
import { Pagination, Empty, Spin } from 'antd'
import { useArticles } from '../hooks/useArticles'
import ArticleCard from './ArticleCard'
import type { ArticleListParams } from '../api/articles'

interface Props {
  params: ArticleListParams
  showPagination?: boolean
  emptyDescription?: string
}

export default function ArticleList({ params, showPagination = true, emptyDescription }: Props) {
  const [current, setCurrent] = useState(params.current || 1)
  const [pageSize] = useState(params.pageSize || 10)

  const queryParams = { ...params, current, pageSize }
  const { data, isLoading } = useArticles(queryParams)

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 80 }}>
        <Spin size="large" />
      </div>
    )
  }

  const articles = data?.data || []
  const total = data?.total || 0

  if (articles.length === 0) {
    return <Empty description={emptyDescription || 'No articles yet'} style={{ padding: 80 }} />
  }

  return (
    <div>
      {articles.map((article) => (
        <ArticleCard key={article.id} article={article} />
      ))}

      {showPagination && total > pageSize && (
        <div style={{ textAlign: 'center', marginTop: 24 }}>
          <Pagination
            current={current}
            pageSize={pageSize}
            total={total}
            onChange={(page) => setCurrent(page)}
            showSizeChanger={false}
          />
        </div>
      )}
    </div>
  )
}

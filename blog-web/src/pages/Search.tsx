import { useSearchParams } from 'react-router-dom'
import { Typography, Spin, Empty } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import ArticleList from '../components/ArticleList'

const { Title, Text } = Typography

export default function Search() {
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q') || ''

  return (
    <div style={{ maxWidth: 900, margin: '0 auto' }}>
      <Title level={2} style={{ textAlign: 'center', marginBottom: 8 }}>
        <SearchOutlined /> Search Results
      </Title>
      {query ? (
        <Text type="secondary" style={{ display: 'block', textAlign: 'center', marginBottom: 32, fontSize: 15 }}>
          Results for: <Text strong>"{query}"</Text>
        </Text>
      ) : (
        <Text type="secondary" style={{ display: 'block', textAlign: 'center', marginBottom: 32, fontSize: 15 }}>
          Enter a search term in the search bar above
        </Text>
      )}

      {query ? (
        <ArticleList
          params={{
            current: 1,
            pageSize: 10,
            title: query,
            status: 'published',
          }}
        />
      ) : (
        <Empty description="No search query entered" style={{ padding: 80 }}>
          <Text type="secondary">Use the search bar in the navigation to find articles</Text>
        </Empty>
      )}
    </div>
  )
}

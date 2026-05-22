import { Typography, Timeline, Card, Space, Spin, Empty, Tag } from 'antd'
import { CalendarOutlined } from '@ant-design/icons'
import { Link } from 'react-router-dom'
import { useArchives } from '../hooks/useArticles'
import { useAppStore } from '../store/appStore'

const { Title, Text } = Typography

const monthNames = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

export default function Archives() {
  const { data, isLoading } = useArchives()
  const theme = useAppStore((s) => s.theme)
  const archives = data?.data || []

  const groupedByYear = archives.reduce((acc, item) => {
    if (!acc[item.year]) acc[item.year] = []
    acc[item.year].push(item)
    return acc
  }, {} as Record<number, typeof archives>)

  const years = Object.keys(groupedByYear)
    .map(Number)
    .sort((a, b) => b - a)

  return (
    <div style={{ maxWidth: 800, margin: '0 auto' }}>
      <Title level={2} style={{ textAlign: 'center', marginBottom: 8 }}>
        <CalendarOutlined /> Article Archives
      </Title>
      <Text type="secondary" style={{ display: 'block', textAlign: 'center', marginBottom: 40, fontSize: 15 }}>
        Browse all articles by date
      </Text>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 80 }}><Spin size="large" /></div>
      ) : archives.length === 0 ? (
        <Empty description="No articles yet" style={{ padding: 80 }} />
      ) : (
        <Space vertical size={32} style={{ width: '100%' }}>
          {years.map((year) => (
            <div key={year}>
              <Title
                level={3}
                style={{
                  marginBottom: 16,
                  paddingBottom: 8,
                  borderBottom: `2px solid ${theme === 'dark' ? '#303030' : '#f0f0f0'}`,
                }}
              >
                {year}
              </Title>
              <Timeline
                items={groupedByYear[year]
                  .sort((a, b) => b.month - a.month)
                  .map((item) => ({
                    color: 'blue',
                    children: (
                      <Link to={`/articles?year=${item.year}&month=${item.month}`}>
                        <Card
                          size="small"
                          hoverable
                          style={{
                            borderRadius: 8,
                            background: theme === 'dark' ? '#141414' : '#fff',
                          }}
                        >
                          <Space>
                            <Text strong style={{ fontSize: 15 }}>
                              {monthNames[item.month - 1]} {item.year}
                            </Text>
                            <Tag color="blue">{item.count} article{item.count > 1 ? 's' : ''}</Tag>
                          </Space>
                        </Card>
                      </Link>
                    ),
                  }))}
              />
            </div>
          ))}
        </Space>
      )}
    </div>
  )
}

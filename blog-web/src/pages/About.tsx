import { Typography, Spin } from 'antd'
import { InfoCircleOutlined } from '@ant-design/icons'
import MarkdownRenderer from '../components/MarkdownRenderer'
import { usePage } from '../hooks/useSettings'

const { Title, Text } = Typography

export default function About() {
  const { data, isLoading } = usePage('about')
  const page = data?.data

  if (isLoading) {
    return <div style={{ textAlign: 'center', padding: 120 }}><Spin size="large" /></div>
  }

  if (!page) {
    return (
      <div style={{ maxWidth: 800, margin: '0 auto', textAlign: 'center', padding: 80 }}>
        <InfoCircleOutlined style={{ fontSize: 48, color: '#bbb', marginBottom: 16 }} />
        <Title level={3}>Page not found</Title>
        <Text type="secondary">The about page has not been created yet.</Text>
      </div>
    )
  }

  return (
    <div style={{ maxWidth: 800, margin: '0 auto' }}>
      <Title level={2} style={{ textAlign: 'center', marginBottom: 8 }}>
        {page.title}
      </Title>
      <Text type="secondary" style={{ display: 'block', textAlign: 'center', marginBottom: 32, fontSize: 15 }}>
        Learn more about this site and its author
      </Text>
      <div
        style={{
          background: '#fff',
          borderRadius: 12,
          padding: '32px 40px',
          boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
        }}
        className="dark-card"
      >
        <MarkdownRenderer content={page.content} />
      </div>
    </div>
  )
}

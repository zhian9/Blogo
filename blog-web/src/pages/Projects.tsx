import { Typography } from 'antd'
import { CodeOutlined } from '@ant-design/icons'

const { Title, Paragraph } = Typography

export default function Projects() {
  return (
    <div style={{ textAlign: 'center', padding: '80px 20px', maxWidth: 600, margin: '0 auto' }}>
      <CodeOutlined style={{ fontSize: 64, color: '#1890ff', marginBottom: 20 }} />
      <Title level={2}>项目库</Title>
      <Paragraph type="secondary" style={{ fontSize: 16 }}>
        这个页面即将推出。我正在整理一些有趣的开源项目和作品。
      </Paragraph>
      <Paragraph type="secondary" style={{ marginTop: 20 }}>
        敬请期待...
      </Paragraph>
    </div>
  )
}

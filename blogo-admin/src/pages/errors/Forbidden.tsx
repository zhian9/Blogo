import { useNavigate } from 'react-router-dom'
import { Button, Result } from 'antd'
import { LockOutlined } from '@ant-design/icons'

export default function Forbidden() {
  const navigate = useNavigate()
  return (
    <div style={{
      display: 'flex', justifyContent: 'center', alignItems: 'center',
      minHeight: '60vh',
    }}>
      <Result
        status="403"
        icon={<LockOutlined style={{ color: '#f87171' }} />}
        title={<span style={{ color: 'rgba(255,255,255,0.85)' }}>403</span>}
        subTitle={<span style={{ color: 'rgba(255,255,255,0.45)' }}>抱歉，您无权访问此页面</span>}
        extra={
          <Button type="primary" onClick={() => navigate('/', { replace: true })}>
            返回首页
          </Button>
        }
      />
    </div>
  )
}

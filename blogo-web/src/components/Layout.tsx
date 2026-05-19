import { Outlet } from 'react-router-dom'
import { Layout as AntLayout, FloatButton } from 'antd'
import Header from './Header'
import Footer from './Footer'

const { Content } = AntLayout

export default function Layout() {
  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Header />
      <Content style={{ width: '100%', margin: '0 auto', padding: '24px 16px' }}>
        <Outlet />
      </Content>
      <Footer />
      <FloatButton.BackTop />
    </AntLayout>
  )
}

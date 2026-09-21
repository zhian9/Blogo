import { useEffect } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import { Layout as AntLayout, FloatButton } from 'antd'
import Header from './Header'
import Footer from './Footer'
import { recordVisit } from '../api/statistics'

const { Content } = AntLayout

export default function Layout() {
  const location = useLocation()

  // 每次页面跳转上报一次访问量（PV/UV），失败静默
  useEffect(() => {
    recordVisit()
  }, [location.pathname])

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

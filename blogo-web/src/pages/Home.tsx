import { Row, Col } from 'antd'
import { motion } from 'framer-motion'
import HeroSection from '../components/HeroSection'
import FeaturedArticles from '../components/FeaturedArticles'
import CategoriesSection from '../components/CategoriesSection'
import NewsletterSection from '../components/NewsletterSection'
import Sidebar from '../components/Sidebar'
import { useAppStore } from '../store/appStore'

export default function Home() {
  const theme = useAppStore((s) => s.theme)

  const pageVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: { duration: 0.5, ease: 'easeOut' },
    },
  }

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={pageVariants}
      style={{
        padding: '0',
        margin: '-24px -16px 0 -16px', // offset Layout Content padding — Hero goes edge-to-edge
      }}
    >
      {/* Hero Section */}
      <HeroSection />

      {/* Main Content Area with Sidebar */}
      <Row
        gutter={[32, 32]}
        style={{
          maxWidth: 1400,
          margin: '0 auto',
          padding: '0 24px',
          marginBottom: 60,
        }}
      >
        <Col xs={24} lg={16}>
          {/* Featured Articles */}
          <FeaturedArticles />

          {/* Categories Section */}
          <CategoriesSection />
        </Col>

        <Col xs={24} lg={8}>
          {/* Sidebar - Hot articles, tags, categories */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            whileInView={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.6 }}
            viewport={{ once: true }}
          >
            <Sidebar />
          </motion.div>
        </Col>
      </Row>

      {/* Newsletter Section */}
      <div
        style={{
          padding: '0 24px',
          maxWidth: 1400,
          margin: '0 auto',
          marginBottom: 60,
        }}
      >
        <NewsletterSection />
      </div>

      {/* Footer Spacer */}
      <div style={{ height: 40 }} />
    </motion.div>
  )
}


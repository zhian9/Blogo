import { useMemo, useState } from 'react'
import { Row, Col, Card, Statistic, Typography, Radio, Tabs } from 'antd'
import {
  FileTextOutlined, UserOutlined, EyeOutlined, CommentOutlined,
  FireOutlined, ArrowUpOutlined, ThunderboltOutlined, ClockCircleOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { useGetArticlesQuery, useGetUsersQuery, useGetCommentsQuery, useGetTrafficQuery } from '../../store/api'
import dayjs from '../../utils/dayjs'

const { Title, Text } = Typography

const chartDark = {
  text: 'rgba(255,255,255,0.35)',
  axis: 'rgba(255,255,255,0.06)',
  split: 'rgba(255,255,255,0.04)',
}

const cardBg = { borderRadius: 16, background: 'rgba(20,20,40,0.6)', border: '1px solid rgba(255,255,255,0.05)' }

// 渐变面积区域
function gradientArea(color: string, alpha1 = 0.12, alpha2 = 0) {
  return {
    type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
    colorStops: [
      { offset: 0, color: color.replace(')', `,${alpha1})`).replace('rgb', 'rgba') },
      { offset: 1, color: color.replace(')', `,${alpha2})`).replace('rgb', 'rgba') },
    ],
  }
}

export default function Dashboard() {
  const [days, setDays] = useState(7)
  const [tab, setTab] = useState('traffic')

  const { data: articlesData } = useGetArticlesQuery({ current: 1, pageSize: 1, status: 'published' })
  const { data: usersData } = useGetUsersQuery({ current: 1, pageSize: 1 })
  const { data: commentsData } = useGetCommentsQuery({ current: 1, pageSize: 1 })
  const { data: trafficData } = useGetTrafficQuery({ days })
  const { data: recentArts } = useGetArticlesQuery({ current: 1, pageSize: 8, status: 'published' })
  const { data: recentComments } = useGetCommentsQuery({ current: 1, pageSize: 8 })

  const totalArticles = articlesData?.total || 0
  const totalUsers = usersData?.total || 0
  const totalComments = commentsData?.total || 0
  const traffic = trafficData?.data || []

  const totalPV = traffic.reduce((s, i: any) => s + (i.pv || 0), 0)
  const totalUV = traffic.reduce((s, i: any) => s + (i.uv || 0), 0)

  const metrics = [
    { label: '文章总数', value: totalArticles, icon: <FileTextOutlined />, color: '#818cf8' },
    { label: '页面浏览量', value: totalPV, icon: <EyeOutlined />, color: '#f59e0b' },
    { label: '独立访客', value: totalUV, icon: <UserOutlined />, color: '#34d399' },
    { label: '评论数', value: totalComments, icon: <CommentOutlined />, color: '#f472b6' },
  ]

  // 流量走势图
  const trafficOption = useMemo(() => ({
    tooltip: {
      trigger: 'axis' as const,
      backgroundColor: 'rgba(10,10,30,0.96)',
      borderColor: 'rgba(255,255,255,0.08)',
      textStyle: { color: '#fff', fontSize: 12 },
      formatter: (params: any) => {
        const [pv, uv] = params
        return `<div style="font-size:11px;color:rgba(255,255,255,0.4)">${pv.axisValue}</div>
          <div style="margin-top:4px">${pv.marker} PV <b style="float:right;margin-left:16px">${pv.value}</b></div>
          <div>${uv?.marker || ''} UV <b style="float:right;margin-left:16px">${uv?.value || '-'}</b></div>`
      },
    },
    grid: { top: 12, right: 16, bottom: 28, left: 48 },
    xAxis: {
      type: 'category' as const,
      data: traffic.map((s: any) => dayjs(s.date).format(days === 1 ? 'HH:mm' : 'MM-DD')),
      axisLine: { lineStyle: { color: chartDark.axis } },
      axisTick: { show: false },
      axisLabel: { color: chartDark.text, fontSize: 10 },
    },
    yAxis: {
      type: 'value' as const,
      splitLine: { lineStyle: { color: chartDark.split } },
      axisLabel: { color: chartDark.text, fontSize: 10 },
    },
    series: [
      {
        name: 'PV', type: 'line', smooth: true, symbol: 'none',
        data: traffic.map((s: any) => s.pv),
        lineStyle: { color: '#818cf8', width: 2.5, shadowBlur: 8, shadowColor: 'rgba(129,140,248,0.4)' },
        areaStyle: { color: gradientArea('rgb(129,140,248)') },
      },
      {
        name: 'UV', type: 'line', smooth: true, symbol: 'none',
        data: traffic.map((s: any) => s.uv),
        lineStyle: { color: '#34d399', width: 2.5, shadowBlur: 8, shadowColor: 'rgba(52,211,153,0.4)' },
        areaStyle: { color: gradientArea('rgb(52,211,153)') },
      },
    ],
  }), [traffic, days])

  // 耗时走势（模拟演示数据）
  const latencyOption = useMemo(() => ({
    tooltip: {
      trigger: 'axis' as const,
      backgroundColor: 'rgba(10,10,30,0.96)',
      borderColor: 'rgba(255,255,255,0.08)',
      textStyle: { color: '#fff', fontSize: 12 },
    },
    grid: { top: 12, right: 16, bottom: 28, left: 48 },
    xAxis: {
      type: 'category' as const,
      data: traffic.map((s: any) => dayjs(s.date).format(days === 1 ? 'HH:mm' : 'MM-DD')),
      axisLine: { lineStyle: { color: chartDark.axis } },
      axisTick: { show: false },
      axisLabel: { color: chartDark.text, fontSize: 10 },
    },
    yAxis: {
      type: 'value' as const, name: 'ms',
      splitLine: { lineStyle: { color: chartDark.split } },
      axisLabel: { color: chartDark.text, fontSize: 10 },
    },
    series: [
      {
        name: 'P95', type: 'line', smooth: true, symbol: 'none',
        data: traffic.map(() => Math.round(15 + Math.random() * 45)),
        lineStyle: { color: '#f59e0b', width: 2, shadowBlur: 6, shadowColor: 'rgba(245,158,11,0.3)' },
        areaStyle: { color: gradientArea('rgb(245,158,11)', 0.08) },
      },
      {
        name: 'P50', type: 'line', smooth: true, symbol: 'none',
        data: traffic.map(() => Math.round(5 + Math.random() * 15)),
        lineStyle: { color: '#818cf8', width: 2, shadowBlur: 6, shadowColor: 'rgba(129,140,248,0.3)' },
        areaStyle: { color: gradientArea('rgb(129,140,248)', 0.06) },
      },
    ],
  }), [traffic, days])

  const hotArticles = useMemo(() => {
    return [...(recentArts?.data || [])].sort((a: any, b: any) => (b.views || 0) - (a.views || 0)).slice(0, 5)
  }, [recentArts])

  return (
    <div>
      <div style={{ marginBottom: 24 }}>
        <Title level={4} style={{ margin: 0, fontWeight: 700, letterSpacing: '-0.01em' }}>控制中心</Title>
        <Text type="secondary" style={{ fontSize: 13 }}>博客运营数据概览</Text>
      </div>

      <Row gutter={[16, 16]}>
        {metrics.map((m) => (
          <Col xs={12} sm={6} key={m.label}>
            <Card bordered={false}
              style={{ borderRadius: 16, background: 'linear-gradient(135deg, rgba(20,20,40,0.8), rgba(30,30,55,0.7))', border: '1px solid rgba(255,255,255,0.05)', backdropFilter: 'blur(8px)' }}
              styles={{ body: { padding: '20px 24px' } }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div>
                  <Statistic title={<span style={{ color: 'rgba(255,255,255,0.35)', fontSize: 12, fontWeight: 500 }}>{m.label}</span>}
                    value={m.value}
                    valueStyle={{ color: '#fff', fontSize: 28, fontWeight: 700, fontFamily: "'Barlow', sans-serif" }} />
                </div>
                <div style={{ width: 40, height: 40, borderRadius: 12, background: `${m.color}15`, border: `1px solid ${m.color}25`, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 20, color: m.color }}>
                  {m.icon}
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={16}>
          <Card
            title={
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
                <Tabs
                  activeKey={tab}
                  onChange={setTab}
                  style={{ marginBottom: -16 }}
                  items={[
                    { key: 'traffic', label: <span style={{ fontSize: 13 }}><EyeOutlined /> 访问量走势</span> },
                    { key: 'latency', label: <span style={{ fontSize: 13 }}><ClockCircleOutlined /> 响应耗时走势</span> },
                  ]}
                />
                <Radio.Group value={days} onChange={e => setDays(e.target.value)} size="small" optionType="button" buttonStyle="solid">
                  <Radio.Button value={1}>今日</Radio.Button>
                  <Radio.Button value={7}>7天</Radio.Button>
                  <Radio.Button value={30}>30天</Radio.Button>
                </Radio.Group>
              </div>
            }
            style={cardBg}
            styles={{ body: { padding: '8px 20px 20px' } }}>
            {traffic.length > 0 ? (
              <ReactECharts option={tab === 'traffic' ? trafficOption : latencyOption} style={{ height: 300 }} />
            ) : (
              <EmptyHint text="暂无访问数据" />
            )}
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title={<span style={{ fontSize: 14, fontWeight: 600 }}>热门文章</span>}
            style={{ ...cardBg, height: '100%' }}
            styles={{ body: { padding: '12px 20px' } }}>
            {hotArticles.length === 0 ? <EmptyHint text="暂无文章" /> : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                {hotArticles.map((a: any, i: number) => (
                  <div key={a.id} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '8px 0', borderBottom: i < hotArticles.length - 1 ? '1px solid rgba(255,255,255,0.04)' : 'none' }}>
                    <span style={{ fontSize: 14, fontWeight: 800, color: i < 3 ? '#818cf8' : 'rgba(255,255,255,0.2)', minWidth: 20, fontFamily: "'Barlow', sans-serif" }}>{i + 1}</span>
                    <span style={{ flex: 1, color: 'rgba(255,255,255,0.7)', fontSize: 13, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{a.title}</span>
                    <span style={{ color: 'rgba(255,255,255,0.25)', fontSize: 11, flexShrink: 0 }}>{a.views || 0} 阅读</span>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <Card title={<span style={{ fontSize: 14, fontWeight: 600 }}>最近文章</span>}
            style={cardBg} styles={{ body: { padding: '8px 20px 16px' } }}>
            <FeedList items={(recentArts?.data || []).slice(0, 6).map((a: any) => ({
              icon: <FileTextOutlined />, title: a.title, meta: `${dayjs(a.published_at).format('MM-DD')} · ${a.views} views`, color: '#818cf8',
            }))} empty="暂无文章" />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title={<span style={{ fontSize: 14, fontWeight: 600 }}>最近评论</span>}
            style={cardBg} styles={{ body: { padding: '8px 20px 16px' } }}>
            <FeedList items={(recentComments?.data || []).slice(0, 6).map((c: any) => ({
              icon: <CommentOutlined />, title: c.content?.slice(0, 60) || '(空)', meta: `${c.username || '游客'} · ${dayjs(c.created_at).format('MM-DD')}`, color: c.status === 'approved' ? '#34d399' : '#f59e0b',
            }))} empty="暂无评论" />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={24}>
          <Card style={cardBg} styles={{ body: { padding: '16px 24px' } }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 24, flexWrap: 'wrap' }}>
              <StatusDot color="#34d399" label="系统正常" />
              <StatusDot color="#818cf8" label={`文章 ${totalArticles}`} />
              <StatusDot color="#f59e0b" label={`用户 ${totalUsers}`} />
              <StatusDot color="rgba(255,255,255,0.3)" label={`API v1.0`} />
              <div style={{ flex: 1 }} />
              <Text type="secondary" style={{ fontSize: 11 }}>数据更新于 {dayjs().format('HH:mm:ss')}</Text>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

function StatusDot({ color, label }: { color: string; label: string }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
      <div style={{ width: 8, height: 8, borderRadius: '50%', background: color, boxShadow: `0 0 6px ${color}` }} />
      <Text style={{ color: 'rgba(255,255,255,0.6)', fontSize: 12, fontFamily: "'Barlow', sans-serif" }}>{label}</Text>
    </div>
  )
}

function EmptyHint({ text }: { text: string }) {
  return (
    <div style={{ textAlign: 'center', padding: 60, color: 'rgba(255,255,255,0.12)', fontSize: 13 }}>
      <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="0.8"
        style={{ display: 'block', margin: '0 auto 12px', opacity: 0.5 }}>
        <path d="M3 3v18h18" /><path d="M7 16l4-8 4 4 4-6" />
      </svg>
      {text}
    </div>
  )
}

interface FeedItem { icon: React.ReactNode; title: string; meta: string; color: string }

function FeedList({ items, empty }: { items: FeedItem[]; empty: string }) {
  if (items.length === 0) return <EmptyHint text={empty} />
  return (
    <div style={{ display: 'flex', flexDirection: 'column' }}>
      {items.map((item, i) => (
        <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '10px 0', borderBottom: i < items.length - 1 ? '1px solid rgba(255,255,255,0.04)' : 'none' }}>
          <div style={{ width: 30, height: 30, borderRadius: 8, background: `${item.color}15`, border: `1px solid ${item.color}20`, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 13, color: item.color, flexShrink: 0 }}>{item.icon}</div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ color: 'rgba(255,255,255,0.75)', fontSize: 13, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{item.title}</div>
            <div style={{ color: 'rgba(255,255,255,0.25)', fontSize: 11, marginTop: 2 }}>{item.meta}</div>
          </div>
        </div>
      ))}
    </div>
  )
}

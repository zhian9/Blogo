import { useState } from 'react'
import {
  Table, Tag, Input, Select, DatePicker, Button, Space, Typography,
  Modal, Descriptions, Tooltip, Empty,
} from 'antd'
import { SearchOutlined, ReloadOutlined, EyeOutlined } from '@ant-design/icons'
import { useGetOperationLogsQuery } from '../store/api'
import dayjs from '../utils/dayjs'
import type { OperationLog } from '../types'

const { Text, Paragraph } = Typography
const { RangePicker } = DatePicker

const cardStyle = { borderRadius: 14, background: 'rgba(20,20,40,0.6)', border: '1px solid rgba(255,255,255,0.05)', backdropFilter: 'blur(8px)' }
const muted = { color: 'rgba(255,255,255,0.35)', fontSize: 12 }

const METHOD_COLORS: Record<string, string> = {
  GET: 'blue', POST: 'green', PUT: 'orange', PATCH: 'gold', DELETE: 'red',
}

export interface OperationLogTableProps {
  /** 卡片标题 */
  title: string
  /** 副标题说明 */
  subtitle?: string
  /** 固定过滤的模块（如"认证模块"） */
  fixedModule?: string
  /** 固定过滤的结果状态（true=成功，false=失败） */
  fixedStatus?: boolean
  /** 是否显示"模块"筛选框 */
  showModuleFilter?: boolean
}

/**
 * 操作日志表格（操作日志 / 登录日志 / 安全日志 / 审计日志共用）。
 *
 * 数据来自后端 /api/v1/operation-logs（GORM 分页），
 * 支持模块、操作人、描述关键字、结果状态、时间范围五个维度筛选。
 */
export default function OperationLogTable({
  title,
  subtitle,
  fixedModule,
  fixedStatus,
  showModuleFilter = true,
}: OperationLogTableProps) {
  const [page, setPage] = useState({ current: 1, pageSize: 20 })
  // 输入框里的值（点"查询"后才生效，避免每敲一个字就请求一次）
  const [draft, setDraft] = useState({ module: '', operator: '', description: '' })
  const [filters, setFilters] = useState({ module: '', operator: '', description: '' })
  const [status, setStatus] = useState<boolean | undefined>(fixedStatus)
  const [range, setRange] = useState<any>(null)
  const [detail, setDetail] = useState<OperationLog | null>(null)

  const params = {
    current: page.current,
    pageSize: page.pageSize,
    module: fixedModule || filters.module || undefined,
    operator: filters.operator || undefined,
    description: filters.description || undefined,
    status: fixedStatus === undefined ? status : fixedStatus,
    startTime: range?.[0] ? range[0].format('YYYY-MM-DD HH:mm:ss') : undefined,
    endTime: range?.[1] ? range[1].format('YYYY-MM-DD HH:mm:ss') : undefined,
  }

  const { data, isLoading, isFetching, refetch } = useGetOperationLogsQuery(params)
  const logs = (data?.data || []) as OperationLog[]
  const total = data?.total || 0

  const handleSearch = () => {
    setFilters(draft)
    setPage((p) => ({ ...p, current: 1 }))
  }

  const handleReset = () => {
    setDraft({ module: '', operator: '', description: '' })
    setFilters({ module: '', operator: '', description: '' })
    setStatus(fixedStatus)
    setRange(null)
    setPage((p) => ({ ...p, current: 1 }))
  }

  const columns = [
    {
      title: '时间', dataIndex: 'created_at', width: 165,
      render: (v: string) => <Text style={{ fontSize: 12 }}>{v ? dayjs(v).format('YYYY-MM-DD HH:mm:ss') : '-'}</Text>,
    },
    {
      title: '操作人', dataIndex: 'operator', width: 130,
      render: (v: string, r: OperationLog) => (
        <Space direction="vertical" size={0}>
          <Text style={{ fontSize: 12 }}>{v || '匿名'}</Text>
          <Text style={muted}>{r.operator_ip || '-'}</Text>
        </Space>
      ),
    },
    {
      title: '模块', dataIndex: 'module', width: 110,
      render: (v: string) => (v ? <Tag color="geekblue" style={{ borderRadius: 6, fontSize: 11 }}>{v}</Tag> : <Text style={muted}>-</Text>),
    },
    {
      title: '动作', dataIndex: 'action_type', width: 90,
      render: (v: string) => (v ? <Tag style={{ borderRadius: 6, fontSize: 11 }}>{v}</Tag> : <Text style={muted}>-</Text>),
    },
    {
      title: '描述', dataIndex: 'description', ellipsis: true,
      render: (v: string) => (
        <Tooltip title={v}><Text style={{ fontSize: 12 }}>{v || '-'}</Text></Tooltip>
      ),
    },
    {
      title: '请求', dataIndex: 'request_path', width: 200, ellipsis: true,
      render: (v: string, r: OperationLog) => (
        <Space size={6}>
          <Tag color={METHOD_COLORS[r.request_method] || 'default'} style={{ borderRadius: 6, fontSize: 11, marginInlineEnd: 0 }}>
            {r.request_method || '-'}
          </Tag>
          <Text style={{ fontSize: 12 }} ellipsis>{v || '-'}</Text>
        </Space>
      ),
    },
    {
      title: '结果', dataIndex: 'status', width: 110,
      render: (v: boolean, r: OperationLog) => (
        <Tag color={v ? 'success' : 'error'} style={{ borderRadius: 6, fontSize: 11 }}>
          {v ? '成功' : '失败'} {r.status_code || ''}
        </Tag>
      ),
    },
    {
      title: '操作', key: 'action', width: 80, fixed: 'right' as const,
      render: (_: unknown, r: OperationLog) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setDetail(r)}>详情</Button>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Typography.Title level={3} style={{ marginBottom: 4 }}>{title}</Typography.Title>
        {subtitle ? <Text style={muted}>{subtitle}</Text> : null}
      </div>

      <div style={{ ...cardStyle, padding: 16, marginBottom: 16 }}>
        <Space wrap size={12}>
          {showModuleFilter && !fixedModule && (
            <Input
              allowClear
              placeholder="模块，如 认证模块"
              style={{ width: 170 }}
              value={draft.module}
              onChange={(e) => setDraft({ ...draft, module: e.target.value })}
              onPressEnter={handleSearch}
            />
          )}
          <Input
            allowClear
            placeholder="操作人"
            style={{ width: 150 }}
            value={draft.operator}
            onChange={(e) => setDraft({ ...draft, operator: e.target.value })}
            onPressEnter={handleSearch}
          />
          <Input
            allowClear
            placeholder="描述关键字"
            style={{ width: 200 }}
            value={draft.description}
            onChange={(e) => setDraft({ ...draft, description: e.target.value })}
            onPressEnter={handleSearch}
          />
          {fixedStatus === undefined && (
            <Select
              allowClear
              placeholder="结果状态"
              style={{ width: 130 }}
              value={status}
              onChange={(v) => { setStatus(v); setPage((p) => ({ ...p, current: 1 })) }}
              options={[
                { value: true, label: '成功' },
                { value: false, label: '失败' },
              ]}
            />
          )}
          <RangePicker
            showTime
            placeholder={['开始时间', '结束时间']}
            value={range}
            onChange={(v) => setRange(v)}
          />
          <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>查询</Button>
          <Button icon={<ReloadOutlined />} onClick={handleReset}>重置</Button>
          <Button type="text" loading={isFetching} onClick={() => refetch()}>刷新</Button>
          <Text style={muted}>共 {total} 条</Text>
        </Space>
      </div>

      <div style={{ ...cardStyle, padding: 8 }}>
        <Table<OperationLog>
          rowKey="id"
          size="small"
          loading={isLoading}
          columns={columns}
          dataSource={logs}
          scroll={{ x: 1100 }}
          locale={{ emptyText: <Empty description="暂无操作日志" /> }}
          pagination={{
            current: page.current,
            pageSize: page.pageSize,
            total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (current, pageSize) => setPage({ current, pageSize }),
          }}
        />
      </div>

      <Modal
        open={!!detail}
        title="操作日志详情"
        footer={null}
        width={640}
        onCancel={() => setDetail(null)}
      >
        {detail && (
          <Descriptions column={1} size="small" bordered>
            <Descriptions.Item label="时间">{detail.created_at ? dayjs(detail.created_at).format('YYYY-MM-DD HH:mm:ss') : '-'}</Descriptions.Item>
            <Descriptions.Item label="操作人">{detail.operator || '匿名'}（{detail.operator_id || '未登录'}）</Descriptions.Item>
            <Descriptions.Item label="来源 IP">{detail.operator_ip || '-'}</Descriptions.Item>
            <Descriptions.Item label="模块 / 动作">{detail.module || '-'} / {detail.action_type || '-'}</Descriptions.Item>
            <Descriptions.Item label="描述">{detail.description || '-'}</Descriptions.Item>
            <Descriptions.Item label="请求">{detail.request_method} {detail.request_path || '-'}</Descriptions.Item>
            <Descriptions.Item label="目标资源">{detail.resource_name || '-'} {detail.resource_id ? `(${detail.resource_id})` : ''}</Descriptions.Item>
            <Descriptions.Item label="结果">
              <Tag color={detail.status ? 'success' : 'error'}>{detail.status ? '成功' : '失败'} {detail.status_code || ''}</Tag>
              {detail.error_msg ? <Paragraph type="danger" style={{ marginTop: 8, marginBottom: 0 }}>{detail.error_msg}</Paragraph> : null}
            </Descriptions.Item>
            <Descriptions.Item label="User-Agent">
              <Text style={{ fontSize: 12 }}>{detail.user_agent || '-'}</Text>
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  )
}

import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Table, Button, Space, Tag, Popconfirm, message, Typography } from 'antd'
import { PlusOutlined, DeleteOutlined, EditOutlined } from '@ant-design/icons'
import { useGetPagesQuery, useDeletePageMutation } from '../../store/api'
import dayjs from '../../utils/dayjs'

const { Title } = Typography

export default function PageList() {
  const navigate = useNavigate()
  const [params, setParams] = useState({ current: 1, pageSize: 10 })
  const { data, isLoading } = useGetPagesQuery(params)
  const [del] = useDeletePageMutation()

  const pages = data?.data || []
  const total = data?.total || 0

  return (
    <div>
      <Title level={4}>Pages</Title>
      <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/pages/new')} style={{ marginBottom: 16 }}>New Page</Button>
      <Table rowKey="id" loading={isLoading} dataSource={pages}
        pagination={{ current: params.current, pageSize: params.pageSize, total }}
        onChange={(p) => setParams({ current: p.current || 1, pageSize: p.pageSize || 10 })}
        columns={[
          { title: 'Title', dataIndex: 'title', key: 'title' },
          { title: 'Slug', dataIndex: 'slug', key: 'slug' },
          {
            title: 'Published', dataIndex: 'is_published', key: 'is_published', width: 100,
            render: (v: boolean) => v ? <Tag color="green">Yes</Tag> : <Tag color="orange">No</Tag>,
          },
          { title: 'Updated', dataIndex: 'updated_at', key: 'updated_at', width: 120, render: (v: string) => dayjs(v).format('YYYY-MM-DD') },
          {
            title: 'Actions', key: 'actions', width: 150,
            render: (_: any, r: any) => (
              <Space>
                <Button size="small" icon={<EditOutlined />} onClick={() => navigate(`/pages/${r.id}`)}>Edit</Button>
                <Popconfirm title="Delete?" onConfirm={async () => { try { await del(r.id); message.success('Deleted') } catch {} }}>
                  <Button size="small" danger icon={<DeleteOutlined />} />
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />
    </div>
  )
}

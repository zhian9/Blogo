import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Table, Button, Space, Popconfirm, Modal, Form, Input, message, Typography, Tag as AntTag, Empty } from 'antd'
import { PlusOutlined, DeleteOutlined, EditOutlined, EyeOutlined } from '@ant-design/icons'
import { useGetTagsQuery, useCreateTagMutation, useUpdateTagMutation, useDeleteTagMutation, useGetTagReferencesQuery } from '../../store/api'
import dayjs from '../../utils/dayjs'

const { Title, Text } = Typography

export default function TagManage() {
  const [params, setParams] = useState({ current: 1, pageSize: 20 })
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<any>(null)
  const [refTag, setRefTag] = useState<any>(null) // 正在查看引用详情的标签
  const [form] = Form.useForm()

  const { data, isLoading, refetch } = useGetTagsQuery(params)
  const [create] = useCreateTagMutation()
  const [update] = useUpdateTagMutation()
  const [del] = useDeleteTagMutation()
  // 引用详情：只有点开某个标签时才请求
  const { data: refData, isFetching: refLoading } = useGetTagReferencesQuery(refTag?.id ?? '', { skip: !refTag })

  const items = data?.data || []
  const total = data?.total || 0

  const openCreate = () => { setEditing(null); form.resetFields(); setModalOpen(true) }
  const openEdit = (item: any) => { setEditing(item); form.setFieldsValue(item); setModalOpen(true) }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      if (editing) { await update({ id: editing.id, body: values }).unwrap(); message.success('标签已更新') }
      else { await create(values).unwrap(); message.success('标签已创建') }
      setModalOpen(false)
      refetch()
    } catch (err: any) {
      message.error(err?.data?.error?.detail || err?.message || '保存失败')
    }
  }

  // 删除失败时要显示后端的原因（例如：该标签被 N 篇文章引用，请先解除关联后再删除）
  const handleDelete = async (id: string) => {
    try {
      await del(id).unwrap()
      message.success('标签已删除')
      refetch()
    } catch (err: any) {
      message.error(err?.data?.error?.detail || '删除失败')
    }
  }

  return (
    <div>
      <Title level={4}>Tags</Title>
      <Button type="primary" icon={<PlusOutlined />} onClick={openCreate} style={{ marginBottom: 16 }}>Add Tag</Button>
      <Table rowKey="id" loading={isLoading} dataSource={items}
        pagination={{ current: params.current, pageSize: params.pageSize, total }}
        onChange={(p) => setParams({ current: p.current || 1, pageSize: p.pageSize || 20 })}
        columns={[
          { title: 'Name', dataIndex: 'name', key: 'name' },
          {
            title: '引用', dataIndex: 'article_count', key: 'article_count', width: 110,
            render: (count: number, r: any) => (
              count > 0
                ? <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setRefTag(r)}>{count} 篇</Button>
                : <Text type="secondary" style={{ fontSize: 12 }}>未引用</Text>
            ),
          },
          { title: 'Created', dataIndex: 'created_at', key: 'created_at', width: 120, render: (v: string) => dayjs(v).format('YYYY-MM-DD') },
          {
            title: 'Actions', key: 'actions', width: 150,
            render: (_: any, r: any) => (
              <Space>
                <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)}>Edit</Button>
                <Popconfirm
                  title="删除标签？"
                  description={r.article_count > 0 ? `该标签被 ${r.article_count} 篇文章引用，需要先解除关联` : undefined}
                  onConfirm={() => handleDelete(r.id)}
                >
                  <Button size="small" danger icon={<DeleteOutlined />} />
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />
      <Modal title={editing ? 'Edit Tag' : 'Create Tag'} open={modalOpen} onOk={handleSave} onCancel={() => setModalOpen(false)}>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="Name" rules={[{ required: true, max: 100 }]}><Input /></Form.Item>
        </Form>
      </Modal>

      {/* 引用详情：被哪些文章使用（有引用时后端会拒绝删除，这里方便你去文章里解除关联） */}
      <Modal
        title={`「${refTag?.name ?? ''}」引用的文章`}
        open={!!refTag}
        footer={<Button onClick={() => setRefTag(null)}>关闭</Button>}
        onCancel={() => setRefTag(null)}
      >
        {refLoading ? (
          <div style={{ textAlign: 'center', padding: 24 }}>加载中…</div>
        ) : (refData?.data?.articles?.length ?? 0) === 0 ? (
          <Empty description="该标签没有被任何文章引用，可以直接删除" />
        ) : (
          <div>
            <Text type="secondary" style={{ fontSize: 12 }}>
              共 {refData?.data?.total ?? 0} 篇文章引用，需先在文章编辑页移除该标签后才能删除
            </Text>
            <div style={{ marginTop: 12, display: 'flex', flexDirection: 'column', gap: 8 }}>
              {(refData?.data?.articles ?? []).map((a) => (
                <div key={a.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}>
                  <Link to={`/articles/${a.id}`} onClick={() => setRefTag(null)}>{a.title}</Link>
                  <AntTag color={a.status === 'published' ? 'green' : 'default'} style={{ marginInlineEnd: 0 }}>
                    {a.status === 'published' ? '已发布' : '草稿'}
                  </AntTag>
                </div>
              ))}
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}

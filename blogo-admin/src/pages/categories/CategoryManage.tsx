import { useState } from 'react'
import { Table, Button, Space, Popconfirm, Modal, Form, Input, InputNumber, message, Typography } from 'antd'
import { PlusOutlined, DeleteOutlined, EditOutlined } from '@ant-design/icons'
import { useGetCategoriesQuery, useCreateCategoryMutation, useUpdateCategoryMutation, useDeleteCategoryMutation } from '../../store/api'
import dayjs from '../../utils/dayjs'

const { Title } = Typography

export default function CategoryManage() {
  const [params, setParams] = useState({ current: 1, pageSize: 20 })
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<any>(null)
  const [form] = Form.useForm()

  const { data, isLoading } = useGetCategoriesQuery(params)
  const [create] = useCreateCategoryMutation()
  const [update] = useUpdateCategoryMutation()
  const [del] = useDeleteCategoryMutation()

  const items = data?.data || []
  const total = data?.total || 0

  const openCreate = () => { setEditing(null); form.resetFields(); setModalOpen(true) }
  const openEdit = (item: any) => { setEditing(item); form.setFieldsValue(item); setModalOpen(true) }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      if (editing) { await update({ id: editing.id, body: values }); message.success('Updated') }
      else { await create(values); message.success('Created') }
      setModalOpen(false)
    } catch (err: any) { if (err.message) message.error(err.message) }
  }

  return (
    <div>
      <Title level={4}>Categories</Title>
      <Button type="primary" icon={<PlusOutlined />} onClick={openCreate} style={{ marginBottom: 16 }}>Add Category</Button>
      <Table rowKey="id" loading={isLoading} dataSource={items}
        pagination={{ current: params.current, pageSize: params.pageSize, total }}
        onChange={(p) => setParams({ current: p.current || 1, pageSize: p.pageSize || 20 })}
        columns={[
          { title: 'Name', dataIndex: 'name', key: 'name' },
          { title: 'Sort', dataIndex: 'sort', key: 'sort', width: 80 },
          { title: 'Created', dataIndex: 'created_at', key: 'created_at', width: 120, render: (v: string) => dayjs(v).format('YYYY-MM-DD') },
          {
            title: 'Actions', key: 'actions', width: 150,
            render: (_: any, r: any) => (
              <Space>
                <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)}>Edit</Button>
                <Popconfirm title="Delete?" onConfirm={async () => { try { await del(r.id) } catch {} }}>
                  <Button size="small" danger icon={<DeleteOutlined />} />
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />
      <Modal title={editing ? 'Edit Category' : 'Create Category'} open={modalOpen} onOk={handleSave} onCancel={() => setModalOpen(false)}>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="Name" rules={[{ required: true, max: 100 }]}><Input /></Form.Item>
          <Form.Item name="sort" label="Sort"><InputNumber min={0} /></Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

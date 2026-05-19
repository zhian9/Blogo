import { useState } from 'react'
import { Table, Button, Space, Tag, Popconfirm, Modal, Form, Input, Select, InputNumber, message, Typography } from 'antd'
import { PlusOutlined, DeleteOutlined, EditOutlined } from '@ant-design/icons'
import { useGetMenusQuery, useCreateMenuMutation, useUpdateMenuMutation, useDeleteMenuMutation } from '../../store/api'
import dayjs from '../../utils/dayjs'

const { Title } = Typography

// Build tree from flat menu list
function buildTree(menus: any[]): any[] {
  const map = new Map<string, any>()
  const roots: any[] = []
  for (const m of menus) {
    map.set(m.id, { ...m, children: [] })
  }
  for (const m of menus) {
    const node = map.get(m.id)!
    if (m.parent_id && map.has(m.parent_id)) {
      map.get(m.parent_id)!.children.push(node)
    } else {
      roots.push(node)
    }
  }
  // Clean up empty children
  const clean = (nodes: any[]) => {
    for (const n of nodes) {
      if (n.children && n.children.length === 0) delete n.children
      else if (n.children) clean(n.children)
    }
  }
  clean(roots)
  return roots
}

export default function MenuList() {
  const [params] = useState({ current: 1, pageSize: 500 })
  const [modalOpen, setModalOpen] = useState(false)
  const [editingMenu, setEditingMenu] = useState<any>(null)
  const [form] = Form.useForm()

  const { data, isLoading } = useGetMenusQuery(params)
  const [createMenu] = useCreateMenuMutation()
  const [updateMenu] = useUpdateMenuMutation()
  const [deleteMenu] = useDeleteMenuMutation()

  const flatMenus = data?.data || []
  const treeMenus = buildTree(flatMenus)

  const openCreate = (parentId?: string) => {
    setEditingMenu(null)
    form.resetFields()
    if (parentId) form.setFieldsValue({ parent_id: parentId, type: 'menu' })
    else form.setFieldsValue({ type: 'directory' })
    setModalOpen(true)
  }

  const openEdit = (menu: any) => {
    setEditingMenu(menu)
    form.setFieldsValue(menu)
    setModalOpen(true)
  }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      if (editingMenu) {
        await updateMenu({ id: editingMenu.id, body: values }).unwrap()
        message.success('已更新')
      } else {
        await createMenu(values).unwrap()
        message.success('已创建')
      }
      setModalOpen(false)
    } catch (err: any) { if (err.message) message.error(err.message) }
  }

  const handleDelete = async (id: string) => {
    try { await deleteMenu(id).unwrap(); message.success('已删除') } catch (err: any) { message.error(err.message) }
  }

  const columns: any[] = [
    { title: '名称', dataIndex: 'name', key: 'name', width: 180 },
    { title: 'Code', dataIndex: 'code', key: 'code', width: 130 },
    {
      title: '类型', dataIndex: 'type', key: 'type', width: 80,
      render: (t: string) => {
        const colors: Record<string, string> = { directory: 'blue', menu: 'green', page: 'green', button: 'orange' }
        const labels: Record<string, string> = { directory: '目录', menu: '菜单', page: '页面', button: '按钮' }
        return <Tag color={colors[t] || 'default'}>{labels[t] || t}</Tag>
      },
    },
    { title: '路径', dataIndex: 'path', key: 'path', width: 160, render: (v: string) => v || <span style={{ color: 'rgba(255,255,255,0.2)' }}>-</span> },
    { title: '排序', dataIndex: 'sequence', key: 'sequence', width: 60 },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 80,
      render: (s: string) => <Tag color={s === 'enabled' ? 'green' : 'red'}>{s === 'enabled' ? '启用' : '禁用'}</Tag>,
    },
    {
      title: '操作', key: 'actions', width: 200,
      render: (_: any, record: any) => (
        <Space>
          <Button size="small" icon={<PlusOutlined />} onClick={() => openCreate(record.id)}>子项</Button>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(record)}>编辑</Button>
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(record.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Title level={4} style={{ fontWeight: 700 }}>菜单管理</Title>
      <Button type="primary" icon={<PlusOutlined />} onClick={() => openCreate()} style={{ marginBottom: 16 }}>添加根菜单</Button>

      <Table
        rowKey="id"
        loading={isLoading}
        dataSource={treeMenus}
        columns={columns}
        pagination={false}
        defaultExpandAllRows
      />

      <Modal
        title={editingMenu ? '编辑菜单' : '创建菜单'}
        open={modalOpen}
        onOk={handleSave}
        onCancel={() => setModalOpen(false)}
        width={640}
      >
        <Form form={form} layout="vertical">
          <Space size="large" wrap>
            <Form.Item name="code" label="Code" rules={[{ required: true }]}>
              <Input disabled={!!editingMenu} style={{ width: 160 }} />
            </Form.Item>
            <Form.Item name="name" label="名称" rules={[{ required: true }]}>
              <Input style={{ width: 160 }} />
            </Form.Item>
            <Form.Item name="type" label="类型" rules={[{ required: true }]}>
              <Select style={{ width: 120 }} options={[
                { label: '目录', value: 'directory' },
                { label: '菜单', value: 'menu' },
                { label: '按钮', value: 'button' },
              ]} />
            </Form.Item>
          </Space>
          <Space size="large" wrap>
            <Form.Item name="path" label="路径">
              <Input placeholder="/articles" style={{ width: 200 }} />
            </Form.Item>
            <Form.Item name="sequence" label="排序">
              <InputNumber min={0} style={{ width: 100 }} />
            </Form.Item>
            <Form.Item name="status" label="状态" rules={[{ required: true }]}>
              <Select style={{ width: 100 }} options={[
                { label: '启用', value: 'enabled' },
                { label: '禁用', value: 'disabled' },
              ]} />
            </Form.Item>
          </Space>
          <Form.Item name="properties" label="图标 (AntD Icon)">
            <Input placeholder='{"icon":"FileTextOutlined"}' style={{ width: 300 }} />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="parent_id" label="父级 ID">
            <Input placeholder="留空为顶级菜单" style={{ width: 260 }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

import { useState } from 'react'
import { Table, Button, Space, Tag, Popconfirm, Modal, Form, Input, Select, InputNumber, message, Typography, Tree } from 'antd'
import { PlusOutlined, DeleteOutlined, EditOutlined } from '@ant-design/icons'
import { useGetRolesQuery, useCreateRoleMutation, useUpdateRoleMutation, useDeleteRoleMutation, useGetMenusQuery } from '../../store/api'
import dayjs from '../../utils/dayjs'

const { Title } = Typography

function buildMenuTree(menus: any[]): any[] {
  const map = new Map<string, any>()
  const roots: any[] = []
  for (const m of menus) {
    map.set(m.id, { key: m.id, title: m.name, children: [] })
  }
  for (const m of menus) {
    const node = map.get(m.id)!
    if (m.parent_id && map.has(m.parent_id)) {
      map.get(m.parent_id)!.children.push(node)
    } else {
      roots.push(node)
    }
  }
  const clean = (nodes: any[]) => {
    for (const n of nodes) {
      if (n.children.length === 0) delete n.children
      else clean(n.children)
    }
  }
  clean(roots)
  return roots
}

export default function RoleList() {
  const [params, setParams] = useState({ current: 1, pageSize: 10 })
  const [modalOpen, setModalOpen] = useState(false)
  const [editingRole, setEditingRole] = useState<any>(null)
  const [selectedMenuKeys, setSelectedMenuKeys] = useState<string[]>([])
  const [form] = Form.useForm()

  const { data, isLoading } = useGetRolesQuery(params)
  const { data: menusData } = useGetMenusQuery({ current: 1, pageSize: 500 })
  const [createRole] = useCreateRoleMutation()
  const [updateRole] = useUpdateRoleMutation()
  const [deleteRole] = useDeleteRoleMutation()

  const roles = data?.data || []
  const total = data?.total || 0
  const menus = menusData?.data || []

  const openCreate = () => {
    setEditingRole(null)
    form.resetFields()
    setSelectedMenuKeys([])
    setModalOpen(true)
  }

  const openEdit = (role: any) => {
    setEditingRole(role)
    form.setFieldsValue(role)
    setSelectedMenuKeys(role.menus?.map((m: any) => m.menu_id || m.id) || [])
    setModalOpen(true)
  }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      const body = {
        ...values,
        menus: selectedMenuKeys.map((id) => ({ menu_id: id })),
      }
      if (editingRole) {
        await updateRole({ id: editingRole.id, body }).unwrap()
        message.success('Updated')
      } else {
        await createRole(body).unwrap()
        message.success('Created')
      }
      setModalOpen(false)
    } catch (err: any) { if (err.message) message.error(err.message) }
  }

  const handleDelete = async (id: string) => {
    try { await deleteRole(id).unwrap(); message.success('Deleted') } catch (err: any) { message.error(err.message) }
  }

  return (
    <div>
      <Title level={4}>Roles</Title>
      <Button type="primary" icon={<PlusOutlined />} onClick={openCreate} style={{ marginBottom: 16 }}>Add Role</Button>

      <Table
        rowKey="id"
        loading={isLoading}
        dataSource={roles}
        pagination={{ current: params.current, pageSize: params.pageSize, total, showTotal: (t) => `Total ${t}` }}
        onChange={(p) => setParams({ current: p.current || 1, pageSize: p.pageSize || 10 })}
        columns={[
          { title: 'Code', dataIndex: 'code', key: 'code' },
          { title: 'Name', dataIndex: 'name', key: 'name' },
          { title: 'Sequence', dataIndex: 'sequence', key: 'sequence', width: 80 },
          {
            title: 'Status', dataIndex: 'status', key: 'status', width: 100,
            render: (s: string) => <Tag color={s === 'enabled' ? 'green' : 'red'}>{s}</Tag>,
          },
          { title: 'Created', dataIndex: 'created_at', key: 'created_at', width: 120, render: (v: string) => dayjs(v).format('YYYY-MM-DD') },
          {
            title: 'Actions', key: 'actions', width: 150,
            render: (_: any, record: any) => (
              <Space>
                <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(record)}>Edit</Button>
                <Popconfirm title="Delete?" onConfirm={() => handleDelete(record.id)}>
                  <Button size="small" danger icon={<DeleteOutlined />} />
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      <Modal
        title={editingRole ? 'Edit Role' : 'Create Role'}
        open={modalOpen}
        onOk={handleSave}
        onCancel={() => setModalOpen(false)}
        width={700}
      >
        <Form form={form} layout="vertical">
          <Space size="large">
            <Form.Item name="code" label="Code" rules={[{ required: true, max: 32 }]}>
              <Input disabled={!!editingRole} />
            </Form.Item>
            <Form.Item name="name" label="Name" rules={[{ required: true, max: 128 }]}>
              <Input />
            </Form.Item>
          </Space>
          <Form.Item name="description" label="Description">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Space size="large">
            <Form.Item name="sequence" label="Sequence">
              <InputNumber min={0} />
            </Form.Item>
            <Form.Item name="status" label="Status" rules={[{ required: true }]}>
              <Select options={[{ label: 'Enabled', value: 'enabled' }, { label: 'Disabled', value: 'disabled' }]} />
            </Form.Item>
          </Space>
          <Form.Item label="菜单权限">
            <Tree
              checkable
              defaultExpandAll
              checkedKeys={selectedMenuKeys}
              onCheck={(keys: any) => setSelectedMenuKeys(keys as string[])}
              treeData={buildMenuTree(menus)}
              style={{ maxHeight: 360, overflow: 'auto' }}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

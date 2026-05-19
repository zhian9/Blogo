import { useState } from 'react'
import { Table, Button, Modal, Form, Input, message, Typography } from 'antd'
import { EditOutlined } from '@ant-design/icons'
import { useGetSettingsQuery, useUpdateSettingMutation } from '../../store/api'
import dayjs from '../../utils/dayjs'

const { Title } = Typography

export default function SettingsManage() {
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<any>(null)
  const [form] = Form.useForm()

  const { data, isLoading } = useGetSettingsQuery()
  const [update] = useUpdateSettingMutation()

  const settings = data?.data || []

  const openEdit = (item: any) => {
    setEditing(item)
    form.setFieldsValue(item)
    setModalOpen(true)
  }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      await update({ key: editing.key, body: values }).unwrap()
      message.success('Updated')
      setModalOpen(false)
    } catch (err: any) { if (err.message) message.error(err.message) }
  }

  return (
    <div>
      <Title level={4}>System Settings</Title>

      <Table rowKey="key" loading={isLoading} dataSource={settings} pagination={false}
        columns={[
          { title: 'Key', dataIndex: 'key', key: 'key', width: 200 },
          { title: 'Value', dataIndex: 'value', key: 'value', ellipsis: true },
          { title: 'Description', dataIndex: 'description', key: 'description' },
          { title: 'Updated', dataIndex: 'updated_at', key: 'updated_at', width: 120, render: (v: string) => dayjs(v).format('YYYY-MM-DD') },
          {
            title: 'Actions', key: 'actions', width: 100,
            render: (_: any, r: any) => (
              <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)}>Edit</Button>
            ),
          },
        ]}
      />

      <Modal title="Edit Setting" open={modalOpen} onOk={handleSave} onCancel={() => setModalOpen(false)}>
        <Form form={form} layout="vertical">
          <Form.Item label="Key"><Input value={editing?.key} disabled /></Form.Item>
          <Form.Item name="value" label="Value" rules={[{ required: true }]}><Input.TextArea rows={4} /></Form.Item>
          <Form.Item name="description" label="Description"><Input /></Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

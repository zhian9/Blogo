import { useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Form, Input, Switch, Button, Card, Space, message, Typography, Spin } from 'antd'
import { useGetPagesQuery, useCreatePageMutation, useUpdatePageMutation } from '../../store/api'

const { Title } = Typography
const { TextArea } = Input

export default function PageEdit() {
  const { id } = useParams<{ id: string }>()
  const isEdit = !!id
  const navigate = useNavigate()
  const [form] = Form.useForm()

  const { data: pageData, isLoading } = useGetPagesQuery({ current: 1, pageSize: 1 }, { skip: !isEdit })
  const [create] = useCreatePageMutation()
  const [update] = useUpdatePageMutation()

  useEffect(() => {
    if (isEdit && pageData?.data?.[0]) {
      form.setFieldsValue(pageData.data[0])
    }
  }, [isEdit, pageData, form])

  const onFinish = async (values: any) => {
    try {
      // .unwrap() 不能省：RTK Query 的 mutation 默认不抛错，不 unwrap 时保存失败也会提示成功
      if (isEdit) { await update({ id: id!, body: values }).unwrap(); message.success('页面已更新') }
      else { await create(values).unwrap(); message.success('页面已创建') }
      navigate('/pages')
    } catch (err: any) { message.error(err?.data?.error?.detail || err?.message || '保存失败') }
  }

  if (isEdit && isLoading) return <div style={{ textAlign: 'center', padding: 80 }}><Spin size="large" /></div>

  return (
    <div>
      <Title level={4}>{isEdit ? 'Edit Page' : 'New Page'}</Title>
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish} initialValues={{ is_published: true }}>
          <Form.Item name="title" label="Title" rules={[{ required: true, max: 255 }]}><Input /></Form.Item>
          <Form.Item name="slug" label="Slug" rules={[{ required: true, max: 255 }]}><Input disabled={isEdit} /></Form.Item>
          <Form.Item name="content" label="Content (Markdown)" rules={[{ required: true }]}>
            <TextArea rows={16} />
          </Form.Item>
          <Form.Item name="is_published" label="Published" valuePropName="checked"><Switch /></Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">{isEdit ? 'Update' : 'Create'}</Button>
              <Button onClick={() => navigate('/pages')}>Cancel</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

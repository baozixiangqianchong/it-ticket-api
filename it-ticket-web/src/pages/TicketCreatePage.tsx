import { ArrowLeftOutlined } from '@ant-design/icons'
import { App, Button, Card, Form, Input, Select, Space, Typography } from 'antd'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { api } from '../api/client'
import type { TicketCategory } from '../api/types'
import { categoryLabel } from '../lib/labels'
import { errText } from '../lib/toast'

const categories: TicketCategory[] = ['hardware', 'software', 'network', 'other']

type CreateForm = {
  title: string
  category: TicketCategory
  description: string
}

export function TicketCreatePage() {
  const navigate = useNavigate()
  const { message } = App.useApp()
  const [submitting, setSubmitting] = useState(false)

  async function onFinish(values: CreateForm) {
    setSubmitting(true)
    try {
      const ticket = await api.createTicket(
        values.title.trim(),
        values.description.trim(),
        values.category,
      )
      navigate(`/tickets/${ticket.id}`, { replace: true })
    } catch (err) {
      message.error(errText(err, '创建失败'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Space direction="vertical" size={16} style={{ width: '100%', maxWidth: 720 }}>
      <div>
        <Button type="link" icon={<ArrowLeftOutlined />} onClick={() => navigate('/tickets')}>
          返回列表
        </Button>
        <Typography.Title level={3} style={{ margin: '8px 0 4px' }}>
          新建工单
        </Typography.Title>
        <Typography.Text type="secondary">提交后进入待派池，管理员可指派，IT 也可以自己领取</Typography.Text>
      </div>
      <Card>
        <Form
          layout="vertical"
          requiredMark={false}
          initialValues={{ category: 'hardware' }}
          onFinish={onFinish}
        >
          <Form.Item
            name="title"
            label="标题"
            rules={[{ required: true, message: '请填写标题' }]}
          >
            <Input maxLength={120} showCount placeholder="例如：会议室投屏连不上" size="large" />
          </Form.Item>
          <Form.Item name="category" label="分类" rules={[{ required: true, message: '请选择分类' }]}>
            <Select
              size="large"
              options={categories.map((c) => ({ value: c, label: categoryLabel[c] }))}
            />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
            rules={[{ required: true, message: '请填写描述' }]}
          >
            <Input.TextArea rows={6} placeholder="发生了什么、已经试过哪些办法" />
          </Form.Item>
          <Button type="primary" htmlType="submit" size="large" loading={submitting}>
            提交工单
          </Button>
        </Form>
      </Card>
    </Space>
  )
}

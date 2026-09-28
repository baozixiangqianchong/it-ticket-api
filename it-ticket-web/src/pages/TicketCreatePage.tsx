import { ArrowLeftOutlined, EditOutlined, MailOutlined, PrinterOutlined, WifiOutlined } from '@ant-design/icons'
import { App, Button, Card, Col, Form, Input, Row, Select, Space } from 'antd'
import { useEffect, useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'

import { api } from '../api/client'
import type { TemplateIcon, TicketCategory, TicketPriority, TicketTemplate } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { PageHeader } from '../components/PageHeader'
import { categoryLabel, priorityLabel } from '../lib/labels'
import { errText } from '../lib/toast'

const categories: TicketCategory[] = ['hardware', 'software', 'network', 'other']
const allPriorities: TicketPriority[] = ['p1', 'p2', 'p3']
const staffPriorities: TicketPriority[] = ['p2', 'p3']
const FREE = 'free'

function templateIcon(icon?: TemplateIcon | 'free'): ReactNode {
  if (icon === 'printer') return <PrinterOutlined className="template-icon" />
  if (icon === 'email') return <MailOutlined className="template-icon" />
  if (icon === 'network') return <WifiOutlined className="template-icon" />
  return <EditOutlined className="template-icon" />
}

type CreateForm = {
  title: string
  category: TicketCategory
  priority: TicketPriority
  description: string
  extra: string
  fields: Record<string, string>
}

function composeDescription(template: TicketTemplate, fields: Record<string, string>, extra: string): string {
  const lines = [`【${template.name}】`]
  for (const field of template.fields) {
    lines.push(`${field.label}：${(fields[field.key] ?? '').trim() || '（未填）'}`)
  }
  const note = extra.trim()
  if (note) {
    lines.push(`补充：${note}`)
  }
  return lines.join('\n')
}

export function TicketCreatePage() {
  const navigate = useNavigate()
  const { message } = App.useApp()
  const { user } = useAuth()
  const [form] = Form.useForm<CreateForm>()
  const [templates, setTemplates] = useState<TicketTemplate[]>([])
  const [templateId, setTemplateId] = useState(FREE)
  const [submitting, setSubmitting] = useState(false)
  const priorities = user?.role === 'admin' ? allPriorities : staffPriorities
  const template = templates.find((item) => String(item.id) === templateId)

  useEffect(() => {
    api
      .listTicketTemplates()
      .then((out) => setTemplates(out.items ?? []))
      .catch(() => undefined)
  }, [])

  function applyTemplate(id: string) {
    setTemplateId(id)
    const next = templates.find((item) => String(item.id) === id)
    const fields: Record<string, string> = {}
    for (const item of templates) {
      for (const field of item.fields) {
        fields[field.key] = ''
      }
    }
    form.resetFields(['title', 'category', 'description', 'extra', 'fields'])
    form.setFieldsValue({
      title: next?.title_hint ?? '',
      category: next?.category ?? 'hardware',
      description: '',
      extra: '',
      fields,
    })
  }

  async function onFinish(values: CreateForm) {
    const description = template
      ? composeDescription(template, values.fields ?? {}, values.extra ?? '')
      : values.description.trim()
    if (!description) {
      message.error('请填写描述')
      return
    }
    if (template) {
      const missing = template.fields.find((field) => field.required && !(values.fields?.[field.key] ?? '').trim())
      if (missing) {
        message.error(`请填写${missing.label}`)
        return
      }
    }
    setSubmitting(true)
    try {
      const title = values.title.trim() || template?.title_hint || ''
      const ticket = await api.createTicket(title, description, values.category, values.priority)
      message.success('工单已提交')
      navigate(`/tickets/${ticket.id}`, { replace: true })
    } catch (err) {
      message.error(errText(err, '创建失败'))
    } finally {
      setSubmitting(false)
    }
  }

  const options = [
    { id: FREE, name: '自由填写', hint: '自己写标题和描述', icon: 'other' as const },
    ...templates.map((item) => ({
      id: String(item.id),
      name: item.name,
      hint: item.hint || item.title_hint,
      icon: item.icon,
    })),
  ]

  return (
    <div className="page-stack">
      <PageHeader
        title="新建工单"
        extra={
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/tickets')}>
            返回列表
          </Button>
        }
      >
        用模板填会把关键信息写进描述。紧急优先级只有管理员能标。
      </PageHeader>
      <Card>
        <Form
          form={form}
          layout="vertical"
          requiredMark={false}
          initialValues={{ category: 'hardware', priority: 'p2', fields: {}, extra: '' }}
          onFinish={onFinish}
        >
          <Form.Item label="报修模板">
            <div className="template-grid">
              {options.map((item) => (
                <button
                  key={item.id}
                  type="button"
                  className={`template-card${templateId === item.id ? ' active' : ''}`}
                  onClick={() => applyTemplate(item.id)}
                >
                  {templateIcon(item.icon)}
                  <strong>{item.name}</strong>
                  <span>{item.hint}</span>
                </button>
              ))}
            </div>
          </Form.Item>
          <Form.Item name="title" label="标题" rules={[{ required: true, message: '请填写标题' }]}>
            <Input
              maxLength={120}
              showCount
              placeholder={template?.title_hint || '例如：会议室投屏连不上'}
              size="large"
            />
          </Form.Item>
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item name="category" label="分类" rules={[{ required: true, message: '请选择分类' }]}>
                <Select
                  size="large"
                  disabled={Boolean(template)}
                  options={categories.map((c) => ({ value: c, label: categoryLabel[c] }))}
                />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item
                name="priority"
                label="优先级"
                extra={user?.role === 'admin' ? undefined : '紧急单请让管理员标记'}
                rules={[{ required: true, message: '请选择优先级' }]}
              >
                <Select
                  size="large"
                  options={priorities.map((item) => ({ value: item, label: priorityLabel[item] }))}
                />
              </Form.Item>
            </Col>
          </Row>
          {template ? (
            <div key={template.id}>
              {template.fields.map((field) => (
                <Form.Item
                  key={field.key}
                  name={['fields', field.key]}
                  label={field.label}
                  rules={field.required ? [{ required: true, message: `请填写${field.label}` }] : undefined}
                >
                  <Input.TextArea
                    rows={field.key === 'tried' || field.key === 'error' || field.key === 'symptom' ? 3 : 2}
                    placeholder={field.placeholder}
                  />
                </Form.Item>
              ))}
              <Form.Item name="extra" label="补充说明">
                <Input.TextArea rows={3} placeholder="还有什么要补充的，写这里" />
              </Form.Item>
            </div>
          ) : (
            <Form.Item name="description" label="描述" rules={[{ required: true, message: '请填写描述' }]}>
              <Input.TextArea rows={6} placeholder="写清机器型号、已试过的步骤、报错原文" />
            </Form.Item>
          )}
          <Space>
            <Button type="primary" htmlType="submit" size="large" loading={submitting}>
              提交工单
            </Button>
            <Button size="large" onClick={() => navigate('/tickets')}>
              取消
            </Button>
          </Space>
        </Form>
      </Card>
    </div>
  )
}

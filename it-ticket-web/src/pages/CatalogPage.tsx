import { DeleteOutlined, PlusOutlined } from '@ant-design/icons'
import { App, Button, Card, Checkbox, Col, Form, Input, InputNumber, Modal, Popconfirm, Row, Select, Space, Switch, Table, Tabs, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useCallback, useEffect, useState } from 'react'

import { api } from '../api/client'
import type { CannedReply, TemplateIcon, TicketCategory, TicketTemplate, TicketTemplateField } from '../api/types'
import { PageHeader } from '../components/PageHeader'
import { categoryLabel } from '../lib/labels'
import { errText } from '../lib/toast'

const categories: TicketCategory[] = ['hardware', 'software', 'network', 'other']
const icons: TemplateIcon[] = ['printer', 'email', 'network', 'other']
const iconLabel: Record<TemplateIcon, string> = {
  printer: '打印机',
  email: '邮箱',
  network: '网络',
  other: '其他',
}

type TemplateForm = {
  name: string
  category: TicketCategory
  title_hint: string
  hint: string
  icon: TemplateIcon
  sort_order: number
  enabled: boolean
  fields: TicketTemplateField[]
}

type ReplyForm = {
  title: string
  body: string
  sort_order: number
  enabled: boolean
}

export function CatalogPage() {
  const { message } = App.useApp()
  const [tab, setTab] = useState('templates')
  const [templates, setTemplates] = useState<TicketTemplate[]>([])
  const [replies, setReplies] = useState<CannedReply[]>([])
  const [loading, setLoading] = useState(true)
  const [templateOpen, setTemplateOpen] = useState(false)
  const [replyOpen, setReplyOpen] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<TicketTemplate | null>(null)
  const [editingReply, setEditingReply] = useState<CannedReply | null>(null)
  const [saving, setSaving] = useState(false)
  const [templateForm] = Form.useForm<TemplateForm>()
  const [replyForm] = Form.useForm<ReplyForm>()
  const templateEnabled = Form.useWatch('enabled', templateForm)
  const replyEnabled = Form.useWatch('enabled', replyForm)

  const reloadTemplates = useCallback(async () => {
    const out = await api.adminListTicketTemplates()
    setTemplates(out.items ?? [])
  }, [])

  const reloadReplies = useCallback(async () => {
    const out = await api.adminListCannedReplies()
    setReplies(out.items ?? [])
  }, [])

  useEffect(() => {
    setLoading(true)
    Promise.all([reloadTemplates(), reloadReplies()])
      .catch((err) => message.error(errText(err, '加载失败')))
      .finally(() => setLoading(false))
  }, [reloadTemplates, reloadReplies, message])

  function openCreateTemplate() {
    setEditingTemplate(null)
    templateForm.setFieldsValue({
      name: '',
      category: 'hardware',
      title_hint: '',
      hint: '',
      icon: 'other',
      sort_order: (templates.at(-1)?.sort_order ?? 0) + 10,
      enabled: true,
      fields: [{ key: '', label: '', placeholder: '', required: true }],
    })
    setTemplateOpen(true)
  }

  function openEditTemplate(row: TicketTemplate) {
    setEditingTemplate(row)
    templateForm.setFieldsValue({
      name: row.name,
      category: row.category,
      title_hint: row.title_hint,
      hint: row.hint,
      icon: row.icon,
      sort_order: row.sort_order,
      enabled: row.enabled,
      fields: row.fields.map((field) => ({ ...field })),
    })
    setTemplateOpen(true)
  }

  async function saveTemplate() {
    const values = await templateForm.validateFields()
    const body = {
      ...values,
      name: values.name.trim(),
      title_hint: values.title_hint.trim(),
      hint: (values.hint ?? '').trim(),
      fields: (values.fields ?? []).map((field) => ({
        key: (field.key ?? '').trim(),
        label: (field.label ?? '').trim(),
        placeholder: (field.placeholder ?? '').trim(),
        required: Boolean(field.required),
      })),
    }
    setSaving(true)
    try {
      if (editingTemplate) {
        await api.updateTicketTemplate(editingTemplate.id, body)
        message.success('模板已保存')
      } else {
        await api.createTicketTemplate(body)
        message.success('模板已添加')
      }
      setTemplateOpen(false)
      await reloadTemplates()
    } catch (err) {
      message.error(errText(err, '保存失败'))
    } finally {
      setSaving(false)
    }
  }

  async function toggleTemplate(row: TicketTemplate, enabled: boolean) {
    try {
      await api.updateTicketTemplate(row.id, {
        name: row.name,
        category: row.category,
        title_hint: row.title_hint,
        hint: row.hint,
        icon: row.icon,
        sort_order: row.sort_order,
        enabled,
        fields: row.fields,
      })
      message.success(enabled ? '已启用' : '已停用，提单页不再展示')
      await reloadTemplates()
    } catch (err) {
      message.error(errText(err, '更新失败'))
    }
  }

  async function removeTemplate(id: number) {
    try {
      await api.deleteTicketTemplate(id)
      message.success('模板已删除')
      await reloadTemplates()
    } catch (err) {
      message.error(errText(err, '删除失败'))
    }
  }

  function openCreateReply() {
    setEditingReply(null)
    replyForm.setFieldsValue({
      title: '',
      body: '',
      sort_order: (replies.at(-1)?.sort_order ?? 0) + 10,
      enabled: true,
    })
    setReplyOpen(true)
  }

  function openEditReply(row: CannedReply) {
    setEditingReply(row)
    replyForm.setFieldsValue({
      title: row.title,
      body: row.body,
      sort_order: row.sort_order,
      enabled: row.enabled,
    })
    setReplyOpen(true)
  }

  async function saveReply() {
    const values = await replyForm.validateFields()
    const body = {
      title: values.title.trim(),
      body: values.body.trim(),
      sort_order: values.sort_order,
      enabled: values.enabled,
    }
    setSaving(true)
    try {
      if (editingReply) {
        await api.updateCannedReply(editingReply.id, body)
        message.success('常用回复已保存')
      } else {
        await api.createCannedReply(body)
        message.success('常用回复已添加')
      }
      setReplyOpen(false)
      await reloadReplies()
    } catch (err) {
      message.error(errText(err, '保存失败'))
    } finally {
      setSaving(false)
    }
  }

  async function toggleReply(row: CannedReply, enabled: boolean) {
    try {
      await api.updateCannedReply(row.id, {
        title: row.title,
        body: row.body,
        sort_order: row.sort_order,
        enabled,
      })
      message.success(enabled ? '已启用' : '已停用，评论框不再展示')
      await reloadReplies()
    } catch (err) {
      message.error(errText(err, '更新失败'))
    }
  }

  async function removeReply(id: number) {
    try {
      await api.deleteCannedReply(id)
      message.success('常用回复已删除')
      await reloadReplies()
    } catch (err) {
      message.error(errText(err, '删除失败'))
    }
  }

  const templateColumns: ColumnsType<TicketTemplate> = [
    { title: '名称', dataIndex: 'name', width: 120 },
    {
      title: '分类',
      dataIndex: 'category',
      width: 88,
      render: (value: TicketCategory) => categoryLabel[value],
    },
    { title: '默认标题', dataIndex: 'title_hint', ellipsis: true },
    {
      title: '字段',
      width: 72,
      render: (_, row) => row.fields.length,
    },
    {
      title: '排序',
      dataIndex: 'sort_order',
      width: 72,
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 88,
      render: (_, row) => (
        <Switch size="small" checked={row.enabled} onChange={(checked) => toggleTemplate(row, checked)} />
      ),
    },
    {
      title: '操作',
      width: 148,
      render: (_, row) => (
        <Space>
          <Button type="link" size="small" onClick={() => openEditTemplate(row)}>
            编辑
          </Button>
          <Popconfirm
            title={`删除「${row.name}」？`}
            description="已提交的工单描述不会改。提单页不再出现这个模板。"
            okText="删除"
            okButtonProps={{ danger: true }}
            onConfirm={() => removeTemplate(row.id)}
          >
            <Button type="link" size="small" danger>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const replyColumns: ColumnsType<CannedReply> = [
    { title: '标题', dataIndex: 'title', width: 180 },
    { title: '内容', dataIndex: 'body', ellipsis: true },
    { title: '排序', dataIndex: 'sort_order', width: 72 },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 88,
      render: (_, row) => (
        <Switch size="small" checked={row.enabled} onChange={(checked) => toggleReply(row, checked)} />
      ),
    },
    {
      title: '操作',
      width: 148,
      render: (_, row) => (
        <Space>
          <Button type="link" size="small" onClick={() => openEditReply(row)}>
            编辑
          </Button>
          <Popconfirm
            title={`删除「${row.title}」？`}
            okText="删除"
            okButtonProps={{ danger: true }}
            onConfirm={() => removeReply(row.id)}
          >
            <Button type="link" size="small" danger>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div className="page-stack">
      <PageHeader title="报修模板">
        只有管理员能改。停用后员工提单看不到，IT 评论框也看不到对应常用回复。最多 20 个模板、30 条常用回复。
      </PageHeader>
      <Card styles={{ body: { paddingTop: 8 } }}>
        <Tabs
          activeKey={tab}
          onChange={setTab}
          tabBarExtraContent={
            tab === 'templates' ? (
              <Button type="primary" disabled={templates.length >= 20} onClick={openCreateTemplate}>
                新增模板
              </Button>
            ) : (
              <Button type="primary" disabled={replies.length >= 30} onClick={openCreateReply}>
                新增回复
              </Button>
            )
          }
          items={[
            {
              key: 'templates',
              label: `提单模板 ${templates.length}/20`,
              children: (
                <Table
                  rowKey="id"
                  loading={loading}
                  columns={templateColumns}
                  dataSource={templates}
                  pagination={false}
                />
              ),
            },
            {
              key: 'replies',
              label: `常用回复 ${replies.length}/30`,
              children: (
                <Table
                  rowKey="id"
                  loading={loading}
                  columns={replyColumns}
                  dataSource={replies}
                  pagination={false}
                />
              ),
            },
          ]}
        />
      </Card>

      <Modal
        className="app-modal"
        rootClassName="app-modal-wrap"
        centered
        title={
          <div className="catalog-modal-title">
            <span>{editingTemplate ? '编辑模板' : '新增模板'}</span>
            <Tag color={templateEnabled ? 'success' : 'default'}>{templateEnabled ? '启用中' : '已停用'}</Tag>
          </div>
        }
        open={templateOpen}
        onCancel={() => setTemplateOpen(false)}
        onOk={saveTemplate}
        confirmLoading={saving}
        okText="保存"
        destroyOnHidden
        width={640}
        styles={{
          header: { marginBottom: 0, paddingInlineEnd: 48 },
          body: { maxHeight: 'min(68vh, 640px)', overflowX: 'hidden', overflowY: 'auto', paddingBlock: 16 },
          footer: { marginTop: 0, borderTop: '1px solid #e4eeea', paddingTop: 16 },
        }}
      >
        <Form form={templateForm} layout="vertical" requiredMark={false} className="catalog-form">
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请填写名称' }]}>
            <Input maxLength={32} showCount placeholder="例如：打印机" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="category" label="分类" rules={[{ required: true, message: '请选择分类' }]}>
                <Select options={categories.map((item) => ({ value: item, label: categoryLabel[item] }))} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="icon" label="图标">
                <Select options={icons.map((item) => ({ value: item, label: iconLabel[item] }))} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="title_hint" label="默认标题" rules={[{ required: true, message: '请填写默认标题' }]}>
            <Input maxLength={120} showCount placeholder="点选模板后填进标题" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={16}>
              <Form.Item name="hint" label="卡片说明">
                <Input maxLength={80} showCount placeholder="显示在提单页模板卡片上" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="sort_order" label="排序" extra="越小越靠前">
                <InputNumber min={0} max={9999} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="enabled" label="在提单页显示" valuePropName="checked">
            <Switch />
          </Form.Item>
          <div className="catalog-fields-label">员工要填的内容</div>
          <Form.List
            name="fields"
            rules={[
              {
                validator: async (_, fields: TicketTemplateField[]) => {
                  if (!fields || fields.length < 1) {
                    return Promise.reject(new Error('至少要有一个字段'))
                  }
                },
              },
            ]}
          >
            {(fields, { add, remove }, { errors }) => (
              <div className="catalog-field-list">
                {fields.map((field, index) => (
                  <div key={field.key} className="catalog-field-card">
                    <div className="catalog-field-head">
                      <span>字段 {index + 1}</span>
                      <Space size={8}>
                        <Form.Item name={[field.name, 'required']} valuePropName="checked" noStyle>
                          <Checkbox>必填</Checkbox>
                        </Form.Item>
                        {fields.length > 1 ? (
                          <Button
                            type="text"
                            size="small"
                            danger
                            icon={<DeleteOutlined />}
                            onClick={() => remove(field.name)}
                          />
                        ) : null}
                      </Space>
                    </div>
                    <Form.Item name={[field.name, 'key']} hidden>
                      <Input />
                    </Form.Item>
                    <Row gutter={12}>
                      <Col span={10}>
                        <Form.Item
                          name={[field.name, 'label']}
                          label="字段名"
                          rules={[{ required: true, message: '请填写字段名' }]}
                        >
                          <Input placeholder="例如：型号" maxLength={32} />
                        </Form.Item>
                      </Col>
                      <Col span={14}>
                        <Form.Item name={[field.name, 'placeholder']} label="填写提示">
                          <Input placeholder="显示在输入框里" maxLength={80} />
                        </Form.Item>
                      </Col>
                    </Row>
                  </div>
                ))}
                <Button
                  type="dashed"
                  block
                  onClick={() => add({ key: '', label: '', placeholder: '', required: true })}
                  icon={<PlusOutlined />}
                  disabled={fields.length >= 8}
                >
                  添加字段
                </Button>
                <Form.ErrorList errors={errors} />
              </div>
            )}
          </Form.List>
        </Form>
      </Modal>

      <Modal
        className="app-modal"
        rootClassName="app-modal-wrap"
        centered
        title={
          <div className="catalog-modal-title">
            <span>{editingReply ? '编辑常用回复' : '新增常用回复'}</span>
            <Tag color={replyEnabled ? 'success' : 'default'}>{replyEnabled ? '启用中' : '已停用'}</Tag>
          </div>
        }
        open={replyOpen}
        onCancel={() => setReplyOpen(false)}
        onOk={saveReply}
        confirmLoading={saving}
        okText="保存"
        destroyOnHidden
        width={520}
        styles={{
          header: { marginBottom: 0, paddingInlineEnd: 48 },
          body: { maxHeight: 'min(68vh, 560px)', overflowX: 'hidden', overflowY: 'auto', paddingBlock: 16 },
          footer: { marginTop: 0, borderTop: '1px solid #e4eeea', paddingTop: 16 },
        }}
      >
        <Form form={replyForm} layout="vertical" requiredMark={false} className="catalog-form">
          <Form.Item name="title" label="标题" rules={[{ required: true, message: '请填写标题' }]}>
            <Input maxLength={80} showCount />
          </Form.Item>
          <Form.Item name="body" label="内容" rules={[{ required: true, message: '请填写内容' }]}>
            <Input.TextArea rows={5} maxLength={2000} showCount />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="sort_order" label="排序" extra="越小越靠前">
                <InputNumber min={0} max={9999} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="enabled" label="在评论框显示" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  )
}

import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  SendOutlined,
  StopOutlined,
} from '@ant-design/icons'
import {
  Alert,
  App,
  Button,
  Card,
  Col,
  Descriptions,
  Empty,
  Form,
  Input,
  List,
  Row,
  Select,
  Space,
  Spin,
  Typography,
} from 'antd'
import { useEffect, useState, type ReactNode } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { ApiError, api } from '../api/client'
import type { AdminUser, TicketAction, TicketDetail } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { CategoryTag, StatusTag } from '../components/StatusTag'
import { useUserNames, userLabel } from '../hooks/useUserNames'
import { actionLabel, availableActions, canAssign, formatTime } from '../lib/labels'

const actionIcon: Record<TicketAction, ReactNode> = {
  start: <PlayCircleOutlined />,
  resolve: <CheckCircleOutlined />,
  close: <StopOutlined />,
  reopen: <ReloadOutlined />,
}

export function TicketDetailPage() {
  const { id } = useParams()
  const ticketId = Number(id)
  const navigate = useNavigate()
  const { message } = App.useApp()
  const { user } = useAuth()
  const names = useUserNames()
  const [ticket, setTicket] = useState<TicketDetail | null>(null)
  const [agents, setAgents] = useState<AdminUser[]>([])
  const [assigneeId, setAssigneeId] = useState<number | undefined>()
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const [commentForm] = Form.useForm<{ body: string }>()

  useEffect(() => {
    if (!Number.isInteger(ticketId) || ticketId <= 0) {
      setTicket(null)
      setError('工单 id 不合法')
      return
    }
    let cancelled = false
    setTicket(null)
    setAssigneeId(undefined)
    setError('')
    commentForm.resetFields()
    api
      .getTicket(ticketId)
      .then((detail) => {
        if (cancelled) return
        setTicket(detail)
        setAssigneeId(detail.assignee_id ?? undefined)
      })
      .catch((err) => {
        if (cancelled) return
        setError(err instanceof ApiError ? err.message : '加载详情失败')
      })
    return () => {
      cancelled = true
    }
  }, [ticketId, commentForm])

  useEffect(() => {
    if (user?.role !== 'admin') return
    api
      .listUsers()
      .then((page) => setAgents((page.items ?? []).filter((item) => item.role === 'agent')))
      .catch(() => undefined)
  }, [user?.role])

  async function reload() {
    const detail = await api.getTicket(ticketId)
    setTicket(detail)
    setAssigneeId(detail.assignee_id ?? undefined)
  }

  async function runAction(action: TicketAction) {
    if (!ticket) return
    setBusy(true)
    setError('')
    try {
      await api.updateTicket(ticket.id, action)
      await reload()
      message.success(`已${actionLabel[action]}`)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '操作失败')
    } finally {
      setBusy(false)
    }
  }

  async function onAssign() {
    if (!ticket || !assigneeId) {
      setError('请选择 IT 处理人')
      return
    }
    setBusy(true)
    setError('')
    try {
      await api.assignTicket(ticket.id, assigneeId)
      await reload()
      message.success('已指派处理人')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '指派失败')
    } finally {
      setBusy(false)
    }
  }

  async function onComment(values: { body: string }) {
    if (!ticket) return
    setBusy(true)
    setError('')
    try {
      await api.addComment(ticket.id, values.body.trim())
      commentForm.resetFields()
      await reload()
      message.success('评论已发送')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '评论失败')
    } finally {
      setBusy(false)
    }
  }

  if (!ticket) {
    return (
      <Space direction="vertical" size={16}>
        {error ? <Alert type="error" showIcon message={error} /> : <Spin tip="加载工单中…" />}
        <Button type="link" icon={<ArrowLeftOutlined />} onClick={() => navigate('/tickets')}>
          返回列表
        </Button>
      </Space>
    )
  }

  const actions = user ? availableActions(ticket, user) : []
  const assignable = user ? canAssign(ticket, user) : false
  const comments = (ticket.comments ?? []).filter((item) => item.ticket_id === ticket.id)

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <div className="page-head">
        <div>
          <Button type="link" icon={<ArrowLeftOutlined />} onClick={() => navigate('/tickets')}>
            返回列表
          </Button>
          <Typography.Title level={3} style={{ margin: '8px 0 8px' }}>
            {ticket.title}
          </Typography.Title>
          <Space wrap>
            <Typography.Text type="secondary">#{ticket.id}</Typography.Text>
            <CategoryTag category={ticket.category} />
            <StatusTag status={ticket.status} />
          </Space>
        </div>
        {actions.length > 0 ? (
          <Space wrap>
            {actions.map((action) => (
              <Button
                key={action}
                type={action === 'close' ? 'default' : 'primary'}
                icon={actionIcon[action]}
                loading={busy}
                onClick={() => void runAction(action)}
              >
                {actionLabel[action]}
              </Button>
            ))}
          </Space>
        ) : null}
      </div>

      {error ? <Alert type="error" showIcon message={error} /> : null}

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={assignable ? 16 : 24}>
          <Card title="工单信息">
            <Descriptions column={{ xs: 1, sm: 2 }} size="middle">
              <Descriptions.Item label="创建人">
                {userLabel(ticket.creator_id, names)}
              </Descriptions.Item>
              <Descriptions.Item label="处理人">
                {userLabel(ticket.assignee_id, names)}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">{formatTime(ticket.created_at)}</Descriptions.Item>
              <Descriptions.Item label="更新时间">{formatTime(ticket.updated_at)}</Descriptions.Item>
            </Descriptions>
            <Typography.Paragraph style={{ marginTop: 16, marginBottom: 0, whiteSpace: 'pre-wrap' }}>
              {ticket.description}
            </Typography.Paragraph>
          </Card>
        </Col>
        {assignable ? (
          <Col xs={24} lg={8}>
            <Card title="指派处理人">
              <Typography.Paragraph type="secondary">只能派给角色为 IT 的用户</Typography.Paragraph>
              <Space direction="vertical" style={{ width: '100%' }} size={12}>
                <Select
                  size="large"
                  placeholder="请选择处理人"
                  value={assigneeId}
                  onChange={setAssigneeId}
                  options={agents.map((agent) => ({
                    value: agent.id,
                    label: `${agent.display_name}（${agent.email}）`,
                  }))}
                  notFoundContent="当前没有 IT 账号"
                />
                <Button type="primary" block loading={busy} disabled={agents.length === 0} onClick={() => void onAssign()}>
                  确认指派
                </Button>
              </Space>
            </Card>
          </Col>
        ) : null}
      </Row>

      <Card title="评论">
        {comments.length === 0 ? (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有评论" />
        ) : (
          <List
            itemLayout="vertical"
            dataSource={comments}
            renderItem={(item) => (
              <List.Item key={item.id}>
                <List.Item.Meta
                  title={userLabel(item.author_id, names)}
                  description={formatTime(item.created_at)}
                />
                <Typography.Paragraph style={{ marginBottom: 0, whiteSpace: 'pre-wrap' }}>
                  {item.body}
                </Typography.Paragraph>
              </List.Item>
            )}
          />
        )}
        {ticket.status === 'closed' ? (
          <Alert type="info" showIcon message="工单已关闭，不能再评论" style={{ marginTop: 16 }} />
        ) : (
          <Form form={commentForm} layout="vertical" style={{ marginTop: 16 }} onFinish={onComment}>
            <Form.Item
              name="body"
              label="追加评论"
              rules={[{ required: true, message: '请填写评论' }]}
            >
              <Input.TextArea rows={4} maxLength={2000} showCount placeholder="补充进展或需要对方配合的事项" />
            </Form.Item>
            <Button type="primary" htmlType="submit" icon={<SendOutlined />} loading={busy}>
              发送
            </Button>
          </Form>
        )}
      </Card>
    </Space>
  )
}

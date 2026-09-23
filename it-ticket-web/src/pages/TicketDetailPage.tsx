import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  CloseOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  SendOutlined,
  StopOutlined,
  UserAddOutlined,
} from '@ant-design/icons'
import {
  App,
  Button,
  Card,
  Col,
  Descriptions,
  Empty,
  Form,
  Input,
  List,
  Modal,
  Row,
  Select,
  Space,
  Spin,
  Timeline,
  Typography,
} from 'antd'
import { useEffect, useState, type ReactNode } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { api } from '../api/client'
import type { AdminUser, TicketAction, TicketDetail } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { CategoryTag, StatusTag } from '../components/StatusTag'
import { actionLabel, auditLabel, formatTime, personName, statusLabel } from '../lib/labels'
import { errText } from '../lib/toast'

const actionIcon: Partial<Record<TicketAction, ReactNode>> = {
  start: <PlayCircleOutlined />,
  resolve: <CheckCircleOutlined />,
  close: <StopOutlined />,
  reopen: <ReloadOutlined />,
  cancel: <CloseOutlined />,
  claim: <UserAddOutlined />,
}

export function TicketDetailPage() {
  const { id } = useParams()
  const ticketId = Number(id)
  const navigate = useNavigate()
  const { message, modal } = App.useApp()
  const { user } = useAuth()
  const [ticket, setTicket] = useState<TicketDetail | null>(null)
  const [agents, setAgents] = useState<AdminUser[]>([])
  const [assigneeId, setAssigneeId] = useState<number | undefined>()
  const [loadFailed, setLoadFailed] = useState(false)
  const [busy, setBusy] = useState(false)
  const [reasonOpen, setReasonOpen] = useState(false)
  const [pendingAction, setPendingAction] = useState<TicketAction | null>(null)
  const [reason, setReason] = useState('')
  const [commentForm] = Form.useForm<{ body: string }>()

  useEffect(() => {
    if (!Number.isInteger(ticketId) || ticketId <= 0) {
      setTicket(null)
      setLoadFailed(true)
      message.error('工单 id 不合法')
      return
    }
    let cancelled = false
    setTicket(null)
    setAssigneeId(undefined)
    setLoadFailed(false)
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
        setLoadFailed(true)
        message.error(errText(err, '加载详情失败'))
      })
    return () => {
      cancelled = true
    }
  }, [ticketId, commentForm, message])

  useEffect(() => {
    if (user?.role !== 'admin') return
    api
      .listUsers({ role: 'agent', pageSize: 50 })
      .then((page) => setAgents(page.items ?? []))
      .catch(() => undefined)
  }, [user?.role])

  async function reload() {
    const detail = await api.getTicket(ticketId)
    setTicket(detail)
    setAssigneeId(detail.assignee_id ?? undefined)
  }

  async function runUpdate(action: TicketAction, actionReason = '') {
    if (!ticket) return
    setBusy(true)
    try {
      await api.updateTicket(ticket.id, action, actionReason)
      await reload()
      message.success(`已${actionLabel[action]}`)
    } catch (err) {
      message.error(errText(err, '操作失败'))
    } finally {
      setBusy(false)
    }
  }

  async function onClaim() {
    if (!ticket) return
    setBusy(true)
    try {
      await api.claimTicket(ticket.id)
      await reload()
      message.success('已领取工单')
    } catch (err) {
      message.error(errText(err, '领取失败'))
    } finally {
      setBusy(false)
    }
  }

  function needsReason(action: TicketAction): boolean {
    if (action === 'reopen') return true
    if (action === 'close' && user?.role === 'admin' && user.id !== ticket?.creator_id) return true
    return false
  }

  function onAction(action: TicketAction) {
    if (action === 'claim') {
      void onClaim()
      return
    }
    if (action === 'cancel') {
      modal.confirm({
        title: '撤回这张工单？',
        content: '撤回后工单将关闭，不能再改。',
        okText: '撤回',
        onOk: () => runUpdate('cancel'),
      })
      return
    }
    if (needsReason(action)) {
      setPendingAction(action)
      setReason('')
      setReasonOpen(true)
      return
    }
    void runUpdate(action)
  }

  async function submitReason() {
    if (!pendingAction) return
    const text = reason.trim()
    if (!text) {
      message.error(pendingAction === 'reopen' ? '重开必须填写原因' : '管理员代关必须填写原因')
      return
    }
    setReasonOpen(false)
    await runUpdate(pendingAction, text)
    setPendingAction(null)
  }

  async function onAssign() {
    if (!ticket || !assigneeId) {
      message.error('请选择 IT 处理人')
      return
    }
    setBusy(true)
    try {
      await api.assignTicket(ticket.id, assigneeId)
      await reload()
      message.success('已指派处理人')
    } catch (err) {
      message.error(errText(err, '指派失败'))
    } finally {
      setBusy(false)
    }
  }

  async function onComment(values: { body: string }) {
    if (!ticket) return
    setBusy(true)
    try {
      await api.addComment(ticket.id, values.body.trim())
      commentForm.resetFields()
      await reload()
      message.success('评论已发送')
    } catch (err) {
      message.error(errText(err, '评论失败'))
    } finally {
      setBusy(false)
    }
  }

  if (!ticket) {
    return (
      <Space direction="vertical" size={16}>
        {loadFailed ? null : <Spin tip="加载工单中…" />}
        <Button type="link" icon={<ArrowLeftOutlined />} onClick={() => navigate('/tickets')}>
          返回列表
        </Button>
      </Space>
    )
  }

  const actions = (ticket.available_actions ?? []).filter((item) => item !== 'assign')
  const assignable = (ticket.available_actions ?? []).includes('assign')
  const comments = ticket.comments ?? []
  const audits = ticket.audits ?? []

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
                type={action === 'close' || action === 'cancel' ? 'default' : 'primary'}
                danger={action === 'cancel'}
                icon={actionIcon[action]}
                loading={busy}
                onClick={() => onAction(action)}
              >
                {actionLabel[action]}
              </Button>
            ))}
          </Space>
        ) : null}
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={assignable ? 16 : 24}>
          <Card title="工单信息">
            <Descriptions column={{ xs: 1, sm: 2 }} size="middle">
              <Descriptions.Item label="创建人">{personName(ticket.creator)}</Descriptions.Item>
              <Descriptions.Item label="处理人">{personName(ticket.assignee)}</Descriptions.Item>
              <Descriptions.Item label="创建时间">{formatTime(ticket.created_at)}</Descriptions.Item>
              <Descriptions.Item label="更新时间">{formatTime(ticket.updated_at)}</Descriptions.Item>
              {ticket.closed_at ? (
                <Descriptions.Item label="关闭时间">{formatTime(ticket.closed_at)}</Descriptions.Item>
              ) : null}
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

      <Card title="处理记录">
        {audits.length === 0 ? (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有记录" />
        ) : (
          <Timeline
            items={audits.map((item) => ({
              children: (
                <div>
                  <Typography.Text>
                    {personName(item.actor)} {auditLabel[item.action] ?? item.action}
                    {item.from_status && item.to_status
                      ? `（${statusLabel[item.from_status]} → ${statusLabel[item.to_status]}）`
                      : ''}
                  </Typography.Text>
                  {item.reason ? (
                    <div>
                      <Typography.Text type="secondary">{item.reason}</Typography.Text>
                    </div>
                  ) : null}
                  <div>
                    <Typography.Text type="secondary">{formatTime(item.created_at)}</Typography.Text>
                  </div>
                </div>
              ),
            }))}
          />
        )}
      </Card>

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
                  title={personName(item.author, `用户 #${item.author_id}`)}
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
          <Typography.Text type="secondary" style={{ display: 'block', marginTop: 16 }}>
            工单已关闭，不能再评论
          </Typography.Text>
        ) : (
          <Form form={commentForm} layout="vertical" style={{ marginTop: 16 }} onFinish={onComment}>
            <Form.Item name="body" label="追加评论" rules={[{ required: true, message: '请填写评论' }]}>
              <Input.TextArea rows={4} maxLength={2000} showCount placeholder="补充进展或需要对方配合的事项" />
            </Form.Item>
            <Button type="primary" htmlType="submit" icon={<SendOutlined />} loading={busy}>
              发送
            </Button>
          </Form>
        )}
      </Card>

      <Modal
        title={pendingAction ? actionLabel[pendingAction] : '填写原因'}
        open={reasonOpen}
        okText="确认"
        confirmLoading={busy}
        onOk={() => void submitReason()}
        onCancel={() => {
          setReasonOpen(false)
          setPendingAction(null)
        }}
      >
        <Input.TextArea
          rows={4}
          maxLength={500}
          showCount
          value={reason}
          placeholder="请填写原因"
          onChange={(e) => setReason(e.target.value)}
        />
      </Modal>
    </Space>
  )
}

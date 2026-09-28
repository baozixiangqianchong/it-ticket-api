import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  EditOutlined,
  CloseOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  SendOutlined,
  StopOutlined,
  SwapOutlined,
  UserAddOutlined,
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
  Modal,
  Row,
  Select,
  Space,
  Spin,
  Tag,
  Timeline,
  Typography,
} from 'antd'
import { useEffect, useState, type ReactNode } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { api } from '../api/client'
import type { AgentRef, CannedReply, TicketAction, TicketCategory, TicketDetail, TicketPriority } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { PageHeader } from '../components/PageHeader'
import { CategoryTag, PriorityTag, StatusTag } from '../components/StatusTag'
import { actionLabel, auditLabel, categoryLabel, formatTime, personName, priorityLabel, statusLabel } from '../lib/labels'
import { errText } from '../lib/toast'

const actionIcon: Partial<Record<TicketAction, ReactNode>> = {
  start: <PlayCircleOutlined />,
  resolve: <CheckCircleOutlined />,
  close: <StopOutlined />,
  reopen: <ReloadOutlined />,
  cancel: <CloseOutlined />,
  claim: <UserAddOutlined />,
  wait: <PauseCircleOutlined />,
  resume: <PlayCircleOutlined />,
  transfer: <SwapOutlined />,
  edit: <EditOutlined />,
}

export function TicketDetailPage() {
  const { id } = useParams()
  const ticketId = Number(id)
  const navigate = useNavigate()
  const { message, modal } = App.useApp()
  const { user, refreshMe } = useAuth()
  const [ticket, setTicket] = useState<TicketDetail | null>(null)
  const [agents, setAgents] = useState<AgentRef[]>([])
  const [agentsReady, setAgentsReady] = useState(false)
  const [assigneeId, setAssigneeId] = useState<number | undefined>()
  const [loadFailed, setLoadFailed] = useState(false)
  const [busy, setBusy] = useState(false)
  const [reasonOpen, setReasonOpen] = useState(false)
  const [pendingAction, setPendingAction] = useState<TicketAction | null>(null)
  const [reason, setReason] = useState('')
  const [transferOpen, setTransferOpen] = useState(false)
  const [transferId, setTransferId] = useState<number | undefined>()
  const [transferReason, setTransferReason] = useState('')
  const [commentForm] = Form.useForm<{ body: string }>()
  const [editForm] = Form.useForm<{
    title: string
    description: string
    category: TicketCategory
    priority: TicketPriority
  }>()
  const [editOpen, setEditOpen] = useState(false)
  const [cannedReplies, setCannedReplies] = useState<CannedReply[]>([])

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
    if (user?.role !== 'admin' && user?.role !== 'agent') return
    api
      .listAgents()
      .then((out) => setAgents(out.items ?? []))
      .catch((err) => message.error(errText(err, '加载 IT 列表失败')))
      .finally(() => setAgentsReady(true))
    api
      .listCannedReplies()
      .then((out) => setCannedReplies(out.items ?? []))
      .catch(() => undefined)
  }, [user?.role, message])

  useEffect(() => {
    if (!agentsReady) return
    const currentId = ticket?.assignee_id
    setAssigneeId(currentId && agents.some((agent) => agent.id === currentId) ? currentId : undefined)
  }, [agentsReady, agents, ticket?.assignee_id])

  async function reload() {
    const detail = await api.getTicket(ticketId)
    setTicket(detail)
  }

  async function runUpdate(action: TicketAction, actionReason = '') {
    if (!ticket) return
    setBusy(true)
    try {
      await api.updateTicket(ticket.id, action, actionReason)
      await reload()
      await refreshMe()
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
      await refreshMe()
      message.success('已领取工单')
    } catch (err) {
      message.error(errText(err, '领取失败'))
    } finally {
      setBusy(false)
    }
  }

  function needsReason(action: TicketAction): boolean {
    if (action === 'reopen' || action === 'wait') return true
    if (action === 'close' && user?.role === 'admin' && user.id !== ticket?.creator_id) return true
    return false
  }

  function onAction(action: TicketAction) {
    if (!ticket) return
    if (action === 'claim') {
      void onClaim()
      return
    }
    if (action === 'edit') {
      editForm.setFieldsValue({
        title: ticket.title,
        description: ticket.description,
        category: ticket.category,
        priority: ticket.priority,
      })
      setEditOpen(true)
      return
    }
    if (action === 'transfer') {
      setTransferId(undefined)
      setTransferReason('')
      setTransferOpen(true)
      return
    }
    if (action === 'cancel') {
      modal.confirm({
        title: '撤回这张工单？',
        content:
          ticket.status === 'in_progress'
            ? '处理人已开始处理。撤回后工单将关闭，7 天内可以重开。'
            : '撤回后工单将关闭。关闭后 7 天内可以重开。',
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
      const hint =
        pendingAction === 'wait'
          ? '等用户必须写明需要补充什么'
          : pendingAction === 'reopen'
            ? '重开必须填写原因'
            : '管理员代关必须填写原因'
      message.error(hint)
      return
    }
    setReasonOpen(false)
    await runUpdate(pendingAction, text)
    setPendingAction(null)
  }

  async function onTransfer() {
    if (!ticket || !transferId) {
      message.error('请选择要转给的 IT')
      return
    }
    setBusy(true)
    try {
      await api.transferTicket(ticket.id, transferId, transferReason.trim())
      setTransferOpen(false)
      await reload()
      await refreshMe()
      message.success('已转派工单')
    } catch (err) {
      message.error(errText(err, '转派失败'))
    } finally {
      setBusy(false)
    }
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
      await refreshMe()
      message.success('已指派处理人')
    } catch (err) {
      message.error(errText(err, '指派失败'))
    } finally {
      setBusy(false)
    }
  }

  async function onEdit() {
    if (!ticket) return
    try {
      const values = await editForm.validateFields()
      setBusy(true)
      await api.editTicket(ticket.id, values)
      setEditOpen(false)
      await reload()
      await refreshMe()
      message.success('工单已修改')
    } catch (err) {
      if (err && typeof err === 'object' && 'errorFields' in err) return
      message.error(errText(err, '修改失败'))
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
      await refreshMe()
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
  const selectedAssignee = agents.some((agent) => agent.id === assigneeId) ? assigneeId : undefined
  const assigneeLeftIT = agentsReady && Boolean(ticket.assignee) && !agents.some((agent) => agent.id === ticket.assignee_id)
  const comments = ticket.comments ?? []
  const audits = ticket.audits ?? []
  const hint = nextStepHint(ticket, user?.id, user?.role)

  const primaryAction = actions.find((item) => ['claim', 'start', 'resume', 'resolve', 'close'].includes(item))

  return (
    <div className="page-stack">
      <PageHeader
        title={
          <span>
            <Typography.Text type="secondary" style={{ marginRight: 8 }}>
              #{ticket.id}
            </Typography.Text>
            {ticket.title}
          </span>
        }
        extra={
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/tickets')}>
            返回列表
          </Button>
        }
      >
        <Space wrap size={6}>
          <CategoryTag category={ticket.category} />
          <PriorityTag priority={ticket.priority} />
          <StatusTag status={ticket.status} />
          {ticket.stale ? <Tag color="red">积压</Tag> : null}
        </Space>
      </PageHeader>

      {actions.length > 0 ? (
        <div className="ticket-action-bar">
          <Space wrap>
            {actions.map((action) => (
              <Button
                key={action}
                type={action === primaryAction ? 'primary' : 'default'}
                danger={action === 'cancel'}
                icon={actionIcon[action]}
                loading={busy}
                onClick={() => onAction(action)}
              >
                {actionLabel[action]}
              </Button>
            ))}
          </Space>
        </div>
      ) : null}

      {hint ? <Alert type={hint.type} showIcon message={hint.text} /> : null}

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={assignable ? 16 : 24}>
          <Card title="工单信息">
            <Descriptions column={{ xs: 1, sm: 2 }} size="middle">
              <Descriptions.Item label="创建人">{personName(ticket.creator)}</Descriptions.Item>
              <Descriptions.Item label="处理人">{personName(ticket.assignee)}</Descriptions.Item>
              <Descriptions.Item label="优先级">
                <PriorityTag priority={ticket.priority} />
              </Descriptions.Item>
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
              <Typography.Paragraph type="secondary">
                可以派给在职 IT 或管理员。派单后直接进入处理中；若当前在等用户，状态保持不变。
              </Typography.Paragraph>
              {assigneeLeftIT ? (
                <Typography.Paragraph type="warning">
                  当前处理人是{personName(ticket.assignee)}，但已经不是 IT，需要改派。
                </Typography.Paragraph>
              ) : null}
              <div className="assign-box">
                <Select
                  className="assign-select"
                  placeholder="请选择处理人"
                  value={selectedAssignee}
                  onChange={setAssigneeId}
                  optionFilterProp="label"
                  options={agents.map((agent) => ({
                    value: agent.id,
                    label: agent.display_name,
                    email: agent.email,
                  }))}
                  optionRender={(option) => (
                    <div className="assign-option">
                      <span>{option.label}</span>
                      {option.data.email ? <span className="assign-option-email">{option.data.email}</span> : null}
                    </div>
                  )}
                  notFoundContent="当前没有 IT 账号"
                />
                <Button type="primary" block loading={busy} disabled={agents.length === 0} onClick={() => void onAssign()}>
                  确认指派
                </Button>
              </div>
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
          <div>
            {comments.map((item) => (
              <div className="comment-item" key={item.id}>
                <Space size={8}>
                  <Typography.Text strong>{personName(item.author, `用户 #${item.author_id}`)}</Typography.Text>
                  <Typography.Text type="secondary">{formatTime(item.created_at)}</Typography.Text>
                </Space>
                <Typography.Paragraph style={{ margin: '8px 0 0', whiteSpace: 'pre-wrap' }}>
                  {item.body}
                </Typography.Paragraph>
              </div>
            ))}
          </div>
        )}
        {ticket.status === 'closed' ? (
          <Typography.Text type="secondary" style={{ display: 'block', marginTop: 16 }}>
            工单已关闭，不能再评论
          </Typography.Text>
        ) : (
          <Form form={commentForm} layout="vertical" style={{ marginTop: comments.length ? 8 : 0 }} onFinish={onComment}>
            {cannedReplies.length > 0 ? (
              <Form.Item label="常用回复">
                <Select
                  allowClear
                  placeholder="插入一条常用回复，仍可再改"
                  options={cannedReplies.map((item) => ({ value: item.id, label: item.title }))}
                  onChange={(value) => {
                    const reply = cannedReplies.find((item) => item.id === value)
                    if (reply) {
                      commentForm.setFieldsValue({ body: reply.body })
                    }
                  }}
                />
              </Form.Item>
            ) : null}
            <Form.Item
              name="body"
              label="追加评论"
              extra={
                ticket.status === 'pending' && user?.id === ticket.creator_id
                  ? '发送后工单会回到处理中'
                  : undefined
              }
              rules={[{ required: true, message: '请填写评论' }]}
            >
              <Input.TextArea
                rows={4}
                maxLength={2000}
                showCount
                placeholder={
                  ticket.status === 'pending' && user?.id === ticket.creator_id
                    ? '补充处理人要的信息，发送后工单回到处理中'
                    : '补充进展或需要对方配合的事项'
                }
              />
            </Form.Item>
            <Button type="primary" htmlType="submit" icon={<SendOutlined />} loading={busy}>
              发送
            </Button>
          </Form>
        )}
      </Card>

      <Modal
        className="app-modal"
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
        <Form className="modal-stack" layout="vertical">
          <Form.Item label={pendingAction === 'wait' ? '需要对方补充什么' : '原因'} required>
            <Input.TextArea
              rows={4}
              maxLength={500}
              showCount
              value={reason}
              placeholder={pendingAction === 'wait' ? '请写明需要对方补充什么' : '请填写原因'}
              onChange={(e) => setReason(e.target.value)}
            />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        className="app-modal"
        title="转派工单"
        open={transferOpen}
        okText="确认转派"
        confirmLoading={busy}
        onOk={() => void onTransfer()}
        onCancel={() => setTransferOpen(false)}
      >
        <Form className="modal-stack" layout="vertical">
          <Form.Item label="转给" required>
            <Select
              placeholder="选择一位在职 IT"
              value={transferId}
              onChange={setTransferId}
              optionFilterProp="label"
              options={agents
                .filter((agent) => agent.id !== ticket.assignee_id)
                .map((agent) => ({
                  value: agent.id,
                  label: agent.display_name,
                  email: agent.email,
                }))}
              optionRender={(option) => (
                <div className="assign-option">
                  <span>{option.label}</span>
                  {option.data.email ? <span className="assign-option-email">{option.data.email}</span> : null}
                </div>
              )}
              notFoundContent="没有可转派的 IT"
            />
          </Form.Item>
          <Form.Item label="转派原因">
            <Input.TextArea
              rows={4}
              maxLength={500}
              showCount
              value={transferReason}
              placeholder="可以不填"
              onChange={(e) => setTransferReason(e.target.value)}
            />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        className="app-modal"
        title={ticket.status === 'open' ? '修改工单' : '调整优先级'}
        open={editOpen}
        okText="保存"
        confirmLoading={busy}
        onOk={() => void onEdit()}
        onCancel={() => setEditOpen(false)}
      >
        <Form form={editForm} layout="vertical">
          {ticket.status === 'open' ? (
            <>
              <Form.Item name="title" label="标题" rules={[{ required: true, message: '请填写标题' }]}>
                <Input maxLength={120} showCount />
              </Form.Item>
              <Form.Item name="category" label="分类" rules={[{ required: true, message: '请选择分类' }]}>
                <Select
                  options={(['hardware', 'software', 'network', 'other'] as TicketCategory[]).map((item) => ({
                    value: item,
                    label: categoryLabel[item],
                  }))}
                />
              </Form.Item>
              <Form.Item name="description" label="描述" rules={[{ required: true, message: '请填写描述' }]}>
                <Input.TextArea rows={5} />
              </Form.Item>
            </>
          ) : null}
          <Form.Item
            name="priority"
            label="优先级"
            extra={user?.role === 'admin' ? undefined : '紧急只能由管理员标记'}
            rules={[{ required: true, message: '请选择优先级' }]}
          >
            <Select
              options={(user?.role === 'admin' || ticket.priority === 'p1'
                ? (['p1', 'p2', 'p3'] as TicketPriority[])
                : (['p2', 'p3'] as TicketPriority[])
              ).map((item) => ({ value: item, label: priorityLabel[item] }))}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

function nextStepHint(
  ticket: TicketDetail,
  userId?: number,
  role?: string,
): { type: 'info' | 'warning' | 'success'; text: string } | null {
  const isCreator = userId === ticket.creator_id
  const lastWait = [...(ticket.audits ?? [])].reverse().find((item) => item.action === 'wait' && item.reason)

  switch (ticket.status) {
    case 'open':
      return isCreator
        ? { type: 'info', text: '已进入待派池。写错了可以修改或撤回；关闭后 7 天内还能重开。' }
        : { type: 'info', text: '待领取。领取后直接进入处理中。' }
    case 'assigned':
      return { type: 'info', text: '工单已回到处理人名下（通常是重开）。点「开始处理」继续修。' }
    case 'pending':
      if (isCreator) {
        return {
          type: 'warning',
          text: lastWait?.reason
            ? `处理人在等你补充：${lastWait.reason}。在下方评论回复后，工单会回到处理中。`
            : '处理人在等你补充信息。在下方评论回复后，工单会回到处理中。',
        }
      }
      return { type: 'info', text: '正在等提单人补充。对方评论后会自动回到处理中，你也可以手动继续处理。' }
    case 'resolved':
      if (isCreator) {
        return { type: 'success', text: '处理人已标记修好。确认没问题就关闭工单；还没好请重开并说明原因。' }
      }
      if (role === 'admin') {
        return { type: 'warning', text: '等待提单人确认关闭。对方长期不关时，管理员可代关并填写原因。' }
      }
      return { type: 'info', text: '已交给提单人确认关闭。确认前请不要再改状态。' }
    default:
      return null
  }
}

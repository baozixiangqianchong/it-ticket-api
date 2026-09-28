import { PlusOutlined } from '@ant-design/icons'
import { App, Button, Card, Empty, Input, Segmented, Select, Space, Table, Tag, Tooltip, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useCallback, useEffect, useState, type MouseEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'

import { api } from '../api/client'
import type { AgentRef, PublicTicket, TicketCategory, TicketPriority, TicketScope, TicketStats, TicketStatus, TicketTimeField } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { DateRangeFilter } from '../components/DateRangeFilter'
import { PageHeader } from '../components/PageHeader'
import { CategoryTag, PriorityTag, StatusTag } from '../components/StatusTag'
import { categoryLabel, formatRelative, formatTime, personName, priorityLabel, scopeLabel, statusLabel } from '../lib/labels'
import { listPagination } from '../lib/pagination'
import { errText } from '../lib/toast'

const statuses: TicketStatus[] = ['open', 'assigned', 'in_progress', 'pending', 'resolved', 'closed']
const categories: TicketCategory[] = ['hardware', 'software', 'network', 'other']
const priorities: TicketPriority[] = ['p1', 'p2', 'p3']

export function TicketListPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { message } = App.useApp()
  const { user, refreshMe } = useAuth()
  const showScopes = user?.role === 'agent' || user?.role === 'admin'
  const isAdmin = user?.role === 'admin'

  const [tickets, setTickets] = useState<PublicTicket[]>([])
  const [total, setTotal] = useState(0)
  const [stats, setStats] = useState<TicketStats | null>(null)
  const [agents, setAgents] = useState<AgentRef[]>([])
  const [loading, setLoading] = useState(true)
  const [scope, setScope] = useState<TicketScope>(() => {
    const raw = searchParams.get('scope')
    return raw && raw in scopeLabel ? (raw as TicketScope) : 'all'
  })
  const [claimingId, setClaimingId] = useState<number | null>(null)
  const [status, setStatus] = useState<TicketStatus | ''>(() => {
    const raw = searchParams.get('status')
    return statuses.includes(raw as TicketStatus) ? (raw as TicketStatus) : ''
  })
  const [category, setCategory] = useState<TicketCategory | ''>('')
  const [priority, setPriority] = useState<TicketPriority | ''>('')
  const [assigneeId, setAssigneeId] = useState<number | undefined>(() => {
    const n = Number(searchParams.get('assignee'))
    return Number.isInteger(n) && n > 0 ? n : undefined
  })
  const [keyword, setKeyword] = useState('')
  const [keywordDraft, setKeywordDraft] = useState('')
  const [timeField, setTimeField] = useState<TicketTimeField>('created_at')
  const [fromDay, setFromDay] = useState('')
  const [toDay, setToDay] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const hasFilter = Boolean(status || category || priority || keyword || fromDay || toDay || assigneeId)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [list, counts] = await Promise.all([
        api.listTickets({
          page,
          pageSize,
          scope: showScopes ? scope : 'all',
          status,
          category,
          priority,
          q: keyword,
          assigneeId: isAdmin ? assigneeId : undefined,
          timeField,
          from: fromDay,
          to: toDay,
        }),
        api.ticketStats(),
      ])
      setTickets(list.items ?? [])
      setTotal(list.total)
      setStats(counts)
    } catch (err) {
      message.error(errText(err, '加载工单失败'))
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, scope, status, category, priority, keyword, assigneeId, timeField, fromDay, toDay, showScopes, isAdmin, message])

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    if (!isAdmin) return
    api
      .listAgents()
      .then((out) => setAgents(out.items ?? []))
      .catch(() => undefined)
  }, [isAdmin])

  function resetFilters() {
    setStatus('')
    setCategory('')
    setPriority('')
    setAssigneeId(undefined)
    setKeyword('')
    setKeywordDraft('')
    setTimeField('created_at')
    setFromDay('')
    setToDay('')
    setPage(1)
  }

  const columns: ColumnsType<PublicTicket> = [
    {
      title: '工单',
      key: 'ticket',
      render: (_, row) => (
        <div>
          <div className="ticket-cell-title">
            <span className="ticket-id">#{row.id}</span>
            <Link to={`/tickets/${row.id}`} onClick={(e) => e.stopPropagation()}>
              {row.title}
            </Link>
          </div>
          <div className="ticket-cell-meta">
            <Space size={4} wrap>
              <CategoryTag category={row.category} />
              <PriorityTag priority={row.priority} />
              {row.stale ? <Tag color="red">积压</Tag> : null}
            </Space>
          </div>
        </div>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 110,
      render: (value) => <StatusTag status={value} />,
    },
    {
      title: '最新评论',
      key: 'last_comment',
      width: 280,
      render: (_, row) => {
        if (!row.last_comment) {
          return <Typography.Text type="secondary">暂无评论</Typography.Text>
        }
        return (
          <div className="last-comment">
            <Typography.Paragraph ellipsis={{ rows: 2 }} style={{ marginBottom: 2 }}>
              {row.last_comment.body}
            </Typography.Paragraph>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              {personName(row.last_comment.author)} · {formatRelative(row.last_comment.created_at)}
            </Typography.Text>
          </div>
        )
      },
    },
    {
      title: '处理人',
      dataIndex: 'assignee',
      width: 120,
      render: (_, row) => personName(row.assignee),
    },
    {
      title: '更新',
      dataIndex: 'updated_at',
      width: 140,
      render: (value: string) => <Tooltip title={formatTime(value)}>{formatRelative(value)}</Tooltip>,
    },
    {
      title: '',
      key: 'quick',
      width: 88,
      render: (_, row) =>
        row.available_actions?.includes('claim') ? (
          <Button
            size="small"
            type="primary"
            loading={claimingId === row.id}
            onClick={(e) => void onClaim(row, e)}
          >
            领取
          </Button>
        ) : null,
    },
  ]

  const emptyText =
    scope === 'pool'
      ? '待派池是空的'
      : scope === 'assigned'
        ? '没有待你处理的单'
        : scope === 'waiting'
          ? '没有等对方的单'
          : hasFilter
            ? '没有符合筛选的工单'
            : '还没有工单，先建一张试试'

  async function onClaim(row: PublicTicket, e: MouseEvent) {
    e.stopPropagation()
    setClaimingId(row.id)
    try {
      await api.claimTicket(row.id)
      message.success('已领取，开始处理')
      await refreshMe()
      await load()
    } catch (err) {
      message.error(errText(err, '领取失败'))
    } finally {
      setClaimingId(null)
    }
  }

  return (
    <div className="page-stack">
      <PageHeader
        title="工单列表"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/tickets/new')}>
            新建工单
          </Button>
        }
      >
        {showScopes
          ? '待我处理是进行中的单；等对方是等用户补充或确认关单。'
          : '只显示你提交的工单。处理人修好后请确认关闭。'}
      </PageHeader>
      {showScopes ? (
        <Segmented
          block
          value={scope}
          onChange={(value) => {
            setScope(value as TicketScope)
            setPage(1)
          }}
          options={(Object.keys(scopeLabel) as TicketScope[]).map((key) => ({
            value: key,
            label: stats ? `${scopeLabel[key]} ${stats[key]}` : scopeLabel[key],
          }))}
        />
      ) : null}
      <Card size="small">
        <div className="filter-bar">
          <div className="filter-row">
            <Select
              allowClear
              placeholder="状态"
              style={{ width: 132 }}
              value={status || undefined}
              options={statuses.map((item) => ({ value: item, label: statusLabel[item] }))}
              onChange={(value) => {
                setStatus(value ?? '')
                setPage(1)
              }}
            />
            <Select
              allowClear
              placeholder="分类"
              style={{ width: 132 }}
              value={category || undefined}
              options={categories.map((item) => ({ value: item, label: categoryLabel[item] }))}
              onChange={(value) => {
                setCategory(value ?? '')
                setPage(1)
              }}
            />
            <Select
              allowClear
              placeholder="优先级"
              style={{ width: 132 }}
              value={priority || undefined}
              options={priorities.map((item) => ({ value: item, label: priorityLabel[item] }))}
              onChange={(value) => {
                setPriority(value ?? '')
                setPage(1)
              }}
            />
            {isAdmin ? (
              <Select
                allowClear
                placeholder="处理人"
                style={{ width: 160 }}
                value={assigneeId}
                options={agents.map((agent) => ({
                  value: agent.id,
                  label: agent.display_name,
                }))}
                onChange={(value) => {
                  setAssigneeId(value)
                  setPage(1)
                }}
              />
            ) : null}
            <Input.Search
              allowClear
              value={keywordDraft}
              placeholder="单号、标题、描述或评论"
              style={{ width: 280 }}
              onChange={(e) => setKeywordDraft(e.target.value)}
              onSearch={(value) => {
                setKeyword(value.trim())
                setPage(1)
              }}
            />
            {hasFilter ? (
              <Button type="link" onClick={resetFilters}>
                重置筛选
              </Button>
            ) : null}
          </div>
          <div className="filter-row">
            <Select
              value={timeField}
              style={{ width: 132 }}
              options={[
                { value: 'created_at', label: '创建时间' },
                { value: 'updated_at', label: '更新时间' },
                { value: 'closed_at', label: '关闭时间' },
              ]}
              onChange={(value) => {
                setTimeField(value)
                setPage(1)
              }}
            />
            <DateRangeFilter
              from={fromDay}
              to={toDay}
              onChange={(from, to) => {
                setFromDay(from)
                setToDay(to)
                setPage(1)
              }}
            />
          </div>
        </div>
      </Card>
      <Card className="table-card">
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={tickets}
          scroll={{ x: 980 }}
          pagination={listPagination(page, pageSize, total, (nextPage, nextSize) => {
            setPage(nextPage)
            setPageSize(nextSize)
          })}
          locale={{
            emptyText: (
              <Empty description={emptyText}>
                {!hasFilter && scope === 'all' ? (
                  <Button type="primary" onClick={() => navigate('/tickets/new')}>
                    新建工单
                  </Button>
                ) : null}
              </Empty>
            ),
          }}
          onRow={(row) => ({
            onClick: () => navigate(`/tickets/${row.id}`),
            style: { cursor: 'pointer', background: row.stale ? '#fff7f6' : undefined },
          })}
        />
      </Card>
    </div>
  )
}

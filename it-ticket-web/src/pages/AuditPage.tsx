import { App, Card, Input, Segmented, Select, Space, Table, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { api } from '../api/client'
import type { ActivityKind, PublicActivity } from '../api/types'
import { DateRangeFilter } from '../components/DateRangeFilter'
import { PageHeader } from '../components/PageHeader'
import { accountStatusLabel, auditLabel, formatTime, personName, roleLabel, statusLabel } from '../lib/labels'
import { listPagination } from '../lib/pagination'
import { errText } from '../lib/toast'

const ticketActions = [
  'create',
  'assign',
  'claim',
  'start',
  'wait',
  'resume',
  'resolve',
  'close',
  'reopen',
  'cancel',
  'transfer',
  'edit',
  'release',
]
const accountActions = ['role', 'status']

function valueLabel(kind: ActivityKind, action: string, value: string | null | undefined): string {
  if (!value) return '—'
  if (kind === 'ticket') return statusLabel[value as keyof typeof statusLabel] ?? value
  if (action === 'role') return roleLabel[value as keyof typeof roleLabel] ?? value
  if (action === 'status') return accountStatusLabel[value as keyof typeof accountStatusLabel] ?? value
  return value
}

export function AuditPage() {
  const { message } = App.useApp()
  const [items, setItems] = useState<PublicActivity[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [kind, setKind] = useState<ActivityKind | ''>('')
  const [action, setAction] = useState('')
  const [keyword, setKeyword] = useState('')
  const [fromDay, setFromDay] = useState('')
  const [toDay, setToDay] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const out = await api.listAudits({
        page,
        pageSize,
        kind,
        action,
        q: keyword,
        from: fromDay,
        to: toDay,
      })
      setItems(out.items ?? [])
      setTotal(out.total)
    } catch (err) {
      message.error(errText(err, '加载审计失败'))
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, kind, action, keyword, fromDay, toDay, message])

  useEffect(() => {
    void load()
  }, [load])

  const actionOptions = (kind === 'account' ? accountActions : kind === 'ticket' ? ticketActions : [...ticketActions, ...accountActions]).map(
    (item) => ({ value: item, label: auditLabel[item] ?? item }),
  )

  const columns: ColumnsType<PublicActivity> = [
    {
      title: '时间',
      dataIndex: 'created_at',
      width: 180,
      render: (value: string) => formatTime(value),
    },
    {
      title: '类型',
      dataIndex: 'kind',
      width: 90,
      render: (value: ActivityKind) => (value === 'ticket' ? <Tag color="blue">工单</Tag> : <Tag color="purple">账号</Tag>),
    },
    {
      title: '动作',
      dataIndex: 'action',
      width: 130,
      render: (value: string) => auditLabel[value] ?? value,
    },
    {
      title: '操作人',
      dataIndex: 'actor',
      width: 140,
      render: (_, row) => personName(row.actor),
    },
    {
      title: '对象',
      key: 'target',
      render: (_, row) => {
        if (row.kind === 'ticket' && row.ticket_id) {
          return <Link to={`/tickets/${row.ticket_id}`}>#{row.ticket_id} {row.ticket_title}</Link>
        }
        return row.user_name || (row.user_id ? `用户 #${row.user_id}` : '—')
      },
    },
    {
      title: '变更',
      key: 'change',
      width: 200,
      render: (_, row) => `${valueLabel(row.kind, row.action, row.from_value)} → ${valueLabel(row.kind, row.action, row.to_value)}`,
    },
    {
      title: '原因',
      dataIndex: 'reason',
      width: 220,
      render: (value: string | null) => value || '—',
    },
  ]

  return (
    <div className="page-stack">
      <PageHeader title="全局审计">最近谁关了单、谁改了角色，不用一张张翻详情。</PageHeader>
      <Segmented
        value={kind || 'all'}
        onChange={(value) => {
          setKind(value === 'all' ? '' : (value as ActivityKind))
          setAction('')
          setPage(1)
        }}
        options={[
          { value: 'all', label: '全部' },
          { value: 'ticket', label: '工单' },
          { value: 'account', label: '账号' },
        ]}
      />
      <Card size="small">
        <Space wrap>
          <Select
            allowClear
            placeholder="动作"
            style={{ width: 180 }}
            value={action || undefined}
            options={actionOptions}
            onChange={(value) => {
              setAction(value ?? '')
              setPage(1)
            }}
          />
          <Input.Search
            allowClear
            placeholder="单号、工单标题或人名"
            style={{ width: 280 }}
            onSearch={(value) => {
              setKeyword(value.trim())
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
        </Space>
      </Card>
      <Card className="table-card">
        <Table
          rowKey={(row) => `${row.kind}-${row.id}`}
          loading={loading}
          columns={columns}
          dataSource={items}
          pagination={listPagination(page, pageSize, total, (nextPage, nextSize) => {
            setPage(nextPage)
            setPageSize(nextSize)
          })}
        />
      </Card>
    </div>
  )
}

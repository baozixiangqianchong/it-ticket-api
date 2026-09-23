import { PlusOutlined } from '@ant-design/icons'
import { App, Button, Card, Empty, Input, Segmented, Select, Space, Table, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { api } from '../api/client'
import type { PublicTicket, TicketCategory, TicketScope, TicketStats, TicketStatus } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { CategoryTag, StatusTag } from '../components/StatusTag'
import { categoryLabel, formatTime, personName, scopeLabel, statusLabel } from '../lib/labels'
import { errText } from '../lib/toast'

const PAGE_SIZE = 10

const statuses: TicketStatus[] = ['open', 'assigned', 'in_progress', 'resolved', 'closed']
const categories: TicketCategory[] = ['hardware', 'software', 'network', 'other']

export function TicketListPage() {
  const navigate = useNavigate()
  const { message } = App.useApp()
  const { user } = useAuth()
  const showScopes = user?.role === 'agent' || user?.role === 'admin'

  const [tickets, setTickets] = useState<PublicTicket[]>([])
  const [total, setTotal] = useState(0)
  const [stats, setStats] = useState<TicketStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [scope, setScope] = useState<TicketScope>('all')
  const [status, setStatus] = useState<TicketStatus | ''>('')
  const [category, setCategory] = useState<TicketCategory | ''>('')
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [list, counts] = await Promise.all([
        api.listTickets({
          page,
          pageSize: PAGE_SIZE,
          scope: showScopes ? scope : 'all',
          status,
          category,
          q: keyword,
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
  }, [page, scope, status, category, keyword, showScopes, message])

  useEffect(() => {
    void load()
  }, [load])

  const columns: ColumnsType<PublicTicket> = [
    {
      title: '编号',
      dataIndex: 'id',
      width: 80,
      render: (id: number) => <Typography.Text type="secondary">#{id}</Typography.Text>,
    },
    {
      title: '标题',
      dataIndex: 'title',
      render: (title: string, row) => <Link to={`/tickets/${row.id}`}>{title}</Link>,
    },
    {
      title: '分类',
      dataIndex: 'category',
      width: 100,
      render: (value) => <CategoryTag category={value} />,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 110,
      render: (value) => <StatusTag status={value} />,
    },
    {
      title: '创建人',
      dataIndex: 'creator',
      width: 140,
      render: (_, row) => personName(row.creator, `用户 #${row.creator_id}`),
    },
    {
      title: '处理人',
      dataIndex: 'assignee',
      width: 140,
      render: (_, row) => personName(row.assignee),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      width: 180,
      render: (value: string) => formatTime(value),
    },
  ]

  const emptyText =
    scope === 'pool' ? '待派池是空的' : scope === 'assigned' ? '没有待你处理的单' : '还没有工单，先建一张试试'

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <div className="page-head">
        <div>
          <Typography.Title level={3} style={{ margin: 0 }}>
            工单列表
          </Typography.Title>
          <Typography.Text type="secondary">
            {showScopes ? '待领可由 IT 自己领取，管理员仍可指定处理人' : '只显示你提交的工单'}
          </Typography.Text>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/tickets/new')}>
          新建工单
        </Button>
      </div>
      {showScopes ? (
        <Segmented
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
        <Space wrap>
          <Select
            allowClear
            placeholder="状态"
            style={{ width: 140 }}
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
            style={{ width: 140 }}
            value={category || undefined}
            options={categories.map((item) => ({ value: item, label: categoryLabel[item] }))}
            onChange={(value) => {
              setCategory(value ?? '')
              setPage(1)
            }}
          />
          <Input.Search
            allowClear
            placeholder="按标题搜索"
            style={{ width: 240 }}
            onSearch={(value) => {
              setKeyword(value.trim())
              setPage(1)
            }}
          />
        </Space>
      </Card>
      <Card styles={{ body: { padding: tickets.length ? 0 : 24 } }}>
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={tickets}
          pagination={{
            current: page,
            pageSize: PAGE_SIZE,
            total,
            hideOnSinglePage: true,
            onChange: setPage,
          }}
          locale={{ emptyText: <Empty description={emptyText} /> }}
          onRow={(row) => ({
            onClick: () => navigate(`/tickets/${row.id}`),
            style: { cursor: 'pointer' },
          })}
        />
      </Card>
    </Space>
  )
}

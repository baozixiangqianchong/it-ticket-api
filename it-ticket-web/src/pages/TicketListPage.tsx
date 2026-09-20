import { PlusOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Empty, Space, Table, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { ApiError, api } from '../api/client'
import type { PublicTicket } from '../api/types'
import { CategoryTag, StatusTag } from '../components/StatusTag'
import { useUserNames, userLabel } from '../hooks/useUserNames'
import { formatTime } from '../lib/labels'

export function TicketListPage() {
  const navigate = useNavigate()
  const names = useUserNames()
  const [tickets, setTickets] = useState<PublicTicket[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api
      .listTickets()
      .then((list) => setTickets(list ?? []))
      .catch((err) => setError(err instanceof ApiError ? err.message : '加载工单失败'))
      .finally(() => setLoading(false))
  }, [])

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
      render: (category) => <CategoryTag category={category} />,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 110,
      render: (status) => <StatusTag status={status} />,
    },
    {
      title: '创建人',
      dataIndex: 'creator_id',
      width: 140,
      render: (id: number) => userLabel(id, names),
    },
    {
      title: '处理人',
      dataIndex: 'assignee_id',
      width: 140,
      render: (id: number | null) => userLabel(id, names),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      width: 180,
      render: (value: string) => formatTime(value),
    },
  ]

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <div className="page-head">
        <div>
          <Typography.Title level={3} style={{ margin: 0 }}>
            工单列表
          </Typography.Title>
          <Typography.Text type="secondary">可见范围由当前角色权限决定</Typography.Text>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/tickets/new')}>
          新建工单
        </Button>
      </div>
      {error ? <Alert type="error" showIcon message={error} /> : null}
      <Card styles={{ body: { padding: tickets.length ? 0 : 24 } }}>
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={tickets}
          pagination={{ pageSize: 10, hideOnSinglePage: true }}
          locale={{ emptyText: <Empty description="还没有工单，先建一张试试" /> }}
          onRow={(row) => ({
            onClick: () => navigate(`/tickets/${row.id}`),
            style: { cursor: 'pointer' },
          })}
        />
      </Card>
    </Space>
  )
}

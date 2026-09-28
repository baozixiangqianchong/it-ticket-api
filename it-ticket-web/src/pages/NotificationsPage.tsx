import { App, Button, Card, Empty, Segmented, Table, Tag, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { api } from '../api/client'
import type { PublicNotification } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { PageHeader } from '../components/PageHeader'
import { formatRelative, notifyLabel } from '../lib/labels'
import { listPagination } from '../lib/pagination'
import { errText } from '../lib/toast'

export function NotificationsPage() {
  const { message } = App.useApp()
  const { refreshMe } = useAuth()
  const navigate = useNavigate()
  const [items, setItems] = useState<PublicNotification[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [unreadOnly, setUnreadOnly] = useState(false)
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const out = await api.listNotifications({
        page,
        pageSize,
        unread: unreadOnly,
      })
      setItems(out.items ?? [])
      setTotal(out.total)
    } catch (err) {
      message.error(errText(err, '加载通知失败'))
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, unreadOnly, message])

  useEffect(() => {
    void load()
  }, [load])

  async function openNotice(item: PublicNotification) {
    try {
      if (!item.read) {
        await api.readNotification(item.id)
        await refreshMe()
        await load()
      }
      if (item.ticket_id) {
        navigate(`/tickets/${item.ticket_id}`)
      }
    } catch (err) {
      message.error(errText(err, '打开通知失败'))
    }
  }

  async function readAll() {
    setBusy(true)
    try {
      await api.readAllNotifications()
      await refreshMe()
      await load()
      message.success('已全部标为已读')
    } catch (err) {
      message.error(errText(err, '操作失败'))
    } finally {
      setBusy(false)
    }
  }

  const columns: ColumnsType<PublicNotification> = [
    {
      title: '类型',
      dataIndex: 'type',
      width: 100,
      render: (type: string) => notifyLabel[type] ?? type,
    },
    {
      title: '对象',
      dataIndex: 'title',
      render: (title: string, row) =>
        row.ticket_id ? (
          <Link to={`/tickets/${row.ticket_id}`} onClick={(e) => e.stopPropagation()}>
            <Typography.Text strong={!row.read}>{title}</Typography.Text>
          </Link>
        ) : (
          <Typography.Text strong={!row.read}>{title}</Typography.Text>
        ),
    },
    {
      title: '内容',
      dataIndex: 'body',
      render: (body: string, row) => (
        <Typography.Text type={row.read ? 'secondary' : undefined}>{body}</Typography.Text>
      ),
    },
    {
      title: '状态',
      dataIndex: 'read',
      width: 90,
      render: (read: boolean) =>
        read ? <Tag>已读</Tag> : <Tag color="blue">未读</Tag>,
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      width: 180,
      render: (value: string) => formatRelative(value),
    },
  ]

  return (
    <div className="page-stack">
      <PageHeader
        title="通知中心"
        extra={
          <Button loading={busy} onClick={() => void readAll()}>
            全部已读
          </Button>
        }
      >
        共 {total} 条。工单通知点一行打开对应工单；权限和账号通知只标已读。
      </PageHeader>
      <Segmented
        value={unreadOnly ? 'unread' : 'all'}
        onChange={(value) => {
          setUnreadOnly(value === 'unread')
          setPage(1)
        }}
        options={[
          { value: 'all', label: '全部' },
          { value: 'unread', label: '未读' },
        ]}
      />
      <Card className="table-card">
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={items}
          pagination={listPagination(page, pageSize, total, (nextPage, nextSize) => {
            setPage(nextPage)
            setPageSize(nextSize)
          })}
          locale={{ emptyText: <Empty description={unreadOnly ? '没有未读通知' : '还没有通知'} /> }}
          onRow={(row) => ({
            onClick: () => void openNotice(row),
            style: { cursor: 'pointer' },
          })}
        />
      </Card>
    </div>
  )
}

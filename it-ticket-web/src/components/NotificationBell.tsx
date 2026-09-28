import { BellOutlined } from '@ant-design/icons'
import { App, Badge, Button, Empty, List, Popover, Typography } from 'antd'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { api } from '../api/client'
import type { PublicNotification } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { formatTime } from '../lib/labels'
import { errText } from '../lib/toast'

const PREVIEW = 5

export function NotificationBell() {
  const { user, refreshMe } = useAuth()
  const { message } = App.useApp()
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const [items, setItems] = useState<PublicNotification[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)

  async function load() {
    setLoading(true)
    try {
      const page = await api.listNotifications({ pageSize: PREVIEW })
      setItems(page.items ?? [])
      setTotal(page.total)
    } catch (err) {
      message.error(errText(err, '加载通知失败'))
    } finally {
      setLoading(false)
    }
  }

  async function onOpenChange(next: boolean) {
    setOpen(next)
    if (next) await load()
  }

  function goAll() {
    setOpen(false)
    navigate('/notifications')
  }

  async function onItem(item: PublicNotification) {
    if (!item.read) {
      await api.readNotification(item.id)
      await refreshMe()
    }
    setOpen(false)
    if (item.ticket_id) {
      navigate(`/tickets/${item.ticket_id}`)
      return
    }
    navigate('/notifications')
  }

  return (
    <Popover
      trigger="click"
      open={open}
      onOpenChange={(next) => void onOpenChange(next)}
      placement="bottomRight"
      content={
        <div style={{ width: 360 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
            <Typography.Text strong>最近通知</Typography.Text>
            <Button type="link" size="small" onClick={goAll}>
              查看全部
            </Button>
          </div>
          <List
            size="small"
            loading={loading}
            locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无通知" /> }}
            dataSource={items}
            renderItem={(item) => (
              <List.Item
                style={{ cursor: 'pointer', opacity: item.read ? 0.65 : 1, padding: '8px 0' }}
                onClick={() => void onItem(item).catch((err) => message.error(errText(err, '打开通知失败')))}
              >
                <List.Item.Meta
                  title={
                    <Typography.Text strong={!item.read} ellipsis>
                      {item.title}
                    </Typography.Text>
                  }
                  description={
                    <div>
                      {item.body ? (
                        <Typography.Paragraph ellipsis={{ rows: 2 }} style={{ marginBottom: 0 }}>
                          {item.body}
                        </Typography.Paragraph>
                      ) : null}
                      <Typography.Text type="secondary">{formatTime(item.created_at)}</Typography.Text>
                    </div>
                  }
                />
              </List.Item>
            )}
          />
          {total > PREVIEW ? (
            <Button type="link" block onClick={goAll}>
              还有 {total - PREVIEW} 条，去通知中心
            </Button>
          ) : null}
        </div>
      }
    >
      <Badge count={user?.unread_count ?? 0} size="small">
        <Button type="text" icon={<BellOutlined />} />
      </Badge>
    </Popover>
  )
}

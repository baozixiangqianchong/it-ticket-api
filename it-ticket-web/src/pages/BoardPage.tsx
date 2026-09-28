import { App, Button, Card, Col, Empty, Row, Table } from 'antd'
import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { api } from '../api/client'
import type { AgentLoad, TicketBoard } from '../api/types'
import { PageHeader } from '../components/PageHeader'
import { errText } from '../lib/toast'

function BoardStat({
  title,
  hint,
  value,
  alert,
  loading,
  onClick,
}: {
  title: string
  hint: string
  value: number
  alert?: boolean
  loading: boolean
  onClick: () => void
}) {
  return (
    <Card hoverable loading={loading} onClick={onClick} className="board-stat">
      <div className="board-stat-title">{title}</div>
      <div className={`board-stat-value${alert && value ? ' is-alert' : ''}`}>{value}</div>
      <div className="board-stat-hint">{hint}</div>
    </Card>
  )
}

export function BoardPage() {
  const { message, modal } = App.useApp()
  const navigate = useNavigate()
  const [board, setBoard] = useState<TicketBoard | null>(null)
  const [loading, setLoading] = useState(true)
  const [closing, setClosing] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setBoard(await api.ticketBoard())
    } catch (err) {
      message.error(errText(err, '加载工作台失败'))
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => {
    void load()
  }, [load])

  function goTickets(query: string) {
    navigate(`/tickets?${query}`)
  }

  async function closeStale() {
    modal.confirm({
      title: '关闭超期未确认的工单？',
      content: `将关闭 ${board?.resolved_stale ?? 0} 单超过 7 天仍待确认的工单，并通知提单人。`,
      okText: '批量代关',
      onOk: async () => {
        setClosing(true)
        try {
          const out = await api.closeStaleTickets()
          message.success(out.closed ? `已关闭 ${out.closed} 单` : '没有超期未确认的工单')
          await load()
        } catch (err) {
          message.error(errText(err, '批量代关失败'))
        } finally {
          setClosing(false)
        }
      },
    })
  }

  return (
    <div className="page-stack">
      <PageHeader
        title="工作台"
        extra={
          <Button danger loading={closing} disabled={!board?.resolved_stale} onClick={() => void closeStale()}>
            关闭超期未确认
          </Button>
        }
      >
        点卡片看对应工单。下面这列是每位 IT 手上还有多少单。
      </PageHeader>
      <Row gutter={[16, 16]}>
        <Col xs={12} md={8} lg={4}>
          <BoardStat
            loading={loading}
            title="待领取"
            hint="还没人领"
            value={board?.pool ?? 0}
            onClick={() => goTickets('scope=pool')}
          />
        </Col>
        <Col xs={12} md={8} lg={4}>
          <BoardStat
            loading={loading}
            title="领太久"
            hint="超过 2 天没人领"
            value={board?.pool_stale ?? 0}
            alert
            onClick={() => goTickets('scope=pool')}
          />
        </Col>
        <Col xs={12} md={8} lg={4}>
          <BoardStat
            loading={loading}
            title="处理中"
            hint="IT 正在修"
            value={board?.in_progress ?? 0}
            onClick={() => goTickets('scope=all&status=in_progress')}
          />
        </Col>
        <Col xs={12} md={8} lg={4}>
          <BoardStat
            loading={loading}
            title="等用户补充"
            hint="停下来等人回"
            value={board?.pending ?? 0}
            onClick={() => goTickets('scope=all&status=pending')}
          />
        </Col>
        <Col xs={12} md={8} lg={4}>
          <BoardStat
            loading={loading}
            title="待用户确认"
            hint="修好了等人关"
            value={board?.resolved ?? 0}
            onClick={() => goTickets('scope=all&status=resolved')}
          />
        </Col>
        <Col xs={12} md={8} lg={4}>
          <BoardStat
            loading={loading}
            title="确认太久"
            hint="超过 7 天没关"
            value={board?.resolved_stale ?? 0}
            alert
            onClick={() => goTickets('scope=all&status=resolved')}
          />
        </Col>
      </Row>
      <Card title="每位处理人手上的单" styles={{ body: { padding: board?.agents.length ? 0 : 24 } }}>
        <Table<AgentLoad>
          rowKey="id"
          loading={loading}
          dataSource={board?.agents ?? []}
          pagination={false}
          locale={{ emptyText: <Empty description="还没有可接单的账号" /> }}
          onRow={(row) => ({
            onClick: () => goTickets(`scope=all&assignee=${row.id}`),
            style: { cursor: 'pointer' },
          })}
          columns={[
            { title: '处理人', dataIndex: 'display_name' },
            { title: '邮箱', dataIndex: 'email' },
            { title: '正在处理', dataIndex: 'active', width: 120 },
            { title: '等用户 / 待确认', dataIndex: 'waiting', width: 160 },
          ]}
        />
      </Card>
    </div>
  )
}

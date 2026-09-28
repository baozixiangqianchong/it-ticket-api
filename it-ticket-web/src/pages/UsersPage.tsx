import { App, Avatar, Button, Card, Input, Modal, Popconfirm, Select, Space, Switch, Table, Tag, Tooltip, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useCallback, useEffect, useState } from 'react'

import { api } from '../api/client'
import type { AccountStatus, AdminUser, PublicInvite, Role } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { PageHeader } from '../components/PageHeader'
import { RoleTag } from '../components/RoleTag'
import { accountStatusLabel, formatTime, roleLabel } from '../lib/labels'
import { listPagination } from '../lib/pagination'
import { errText } from '../lib/toast'

const roles: Role[] = ['user', 'agent', 'admin']

function canHandle(role: Role) {
  return role === 'agent' || role === 'admin'
}

function inviteStatusLabel(status: PublicInvite['status']) {
  if (status === 'used') return '已使用'
  if (status === 'expired') return '已过期'
  return '未使用'
}

export function UsersPage() {
  const { message } = App.useApp()
  const { user: me } = useAuth()
  const [users, setUsers] = useState<AdminUser[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [busyId, setBusyId] = useState<number | null>(null)
  const [roleFilter, setRoleFilter] = useState<Role | ''>('')
  const [statusFilter, setStatusFilter] = useState<AccountStatus | ''>('')
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [invites, setInvites] = useState<PublicInvite[]>([])
  const [inviteBusy, setInviteBusy] = useState(false)
  const [roleConfirm, setRoleConfirm] = useState<{ userId: number; next: Role } | null>(null)

  const reload = useCallback(async () => {
    const out = await api.listUsers({
      page,
      pageSize,
      role: roleFilter,
      status: statusFilter,
      q: keyword,
    })
    setUsers(out.items ?? [])
    setTotal(out.total)
  }, [page, pageSize, roleFilter, statusFilter, keyword])

  const reloadInvites = useCallback(async () => {
    const out = await api.listInvites()
    setInvites(out.items ?? [])
  }, [])

  useEffect(() => {
    setLoading(true)
    reload()
      .catch((err) => message.error(errText(err, '加载用户失败')))
      .finally(() => setLoading(false))
  }, [reload, message])

  useEffect(() => {
    reloadInvites().catch(() => undefined)
  }, [reloadInvites])

  function releasedToast(count: number | undefined, fallback: string) {
    if (count && count > 0) {
      message.success(`${fallback}，${count} 张未关单已退回待派池`)
      return
    }
    message.success(fallback)
  }

  function rolePopconfirm(row: AdminUser, next: Role) {
    const n = row.open_ticket_count ?? 0
    const bits: string[] = []
    if (canHandle(row.role) && canHandle(next)) {
      bits.push('管理员和 IT 都能接单，手上的工单会留下。')
    } else if (canHandle(row.role) && !canHandle(next) && n > 0) {
      bits.push(`${n} 张未关单会退回待派池。`)
    }
    if (row.role === 'admin' && next !== 'admin') {
      bits.push('最后一位管理员不能取消。')
    }
    return {
      title: `改为${roleLabel[next]}？`,
      description: bits.join(''),
    }
  }

  function statusConfirm(row: AdminUser) {
    if (row.status === 'disabled') {
      return {
        title: `启用「${row.display_name}」？`,
        description: '启用后即可重新登录。',
        okText: '启用',
        next: 'active' as AccountStatus,
        danger: false,
      }
    }
    const n = row.open_ticket_count ?? 0
    const bits = ['停用后对方无法登录。']
    if (canHandle(row.role) && n > 0) {
      bits.push(`${n} 张未关单会退回待派池。`)
    }
    if (row.role === 'admin') {
      bits.push('最后一位管理员不能停用。')
    }
    return {
      title: `停用「${row.display_name}」？`,
      description: bits.join(''),
      okText: '停用',
      next: 'disabled' as AccountStatus,
      danger: true,
    }
  }

  async function changeRole(row: AdminUser, role: Role) {
    if (row.role === role) return
    if (me?.id === row.id) {
      message.error('不能取消自己的管理员身份')
      return
    }
    setBusyId(row.id)
    try {
      const out = await api.updateUserRole(row.id, role)
      await reload()
      setRoleConfirm(null)
      releasedToast(out.released_count, `已改为${roleLabel[role]}`)
    } catch (err) {
      message.error(errText(err, '改角色失败'))
    } finally {
      setBusyId(null)
    }
  }

  async function changeStatus(row: AdminUser, status: AccountStatus) {
    setBusyId(row.id)
    try {
      const out = await api.updateUserStatus(row.id, status)
      await reload()
      releasedToast(out.released_count, status === 'disabled' ? '已停用' : '已启用')
    } catch (err) {
      message.error(errText(err, '改状态失败'))
    } finally {
      setBusyId(null)
    }
  }

  async function createInvite() {
    setInviteBusy(true)
    try {
      const invite = await api.createInvite()
      await reloadInvites()
      message.success('邀请码已生成，24 小时内有效，只能注册一次')
      Modal.info({
        title: '邀请码',
        content: (
          <div>
            <Typography.Paragraph copyable style={{ fontSize: 20, fontWeight: 600, marginBottom: 8 }}>
              {invite.code}
            </Typography.Paragraph>
            <Typography.Text type="secondary">有效期至 {formatTime(invite.expires_at)}，只能注册 1 个账号。</Typography.Text>
          </div>
        ),
        okText: '知道了',
      })
    } catch (err) {
      message.error(errText(err, '生成邀请码失败'))
    } finally {
      setInviteBusy(false)
    }
  }

  const columns: ColumnsType<AdminUser> = [
    {
      title: '用户',
      render: (_, row) => (
        <Space>
          <Avatar style={{ background: '#0f766e' }}>{row.display_name.slice(0, 1)}</Avatar>
          <div>
            <div>
              {row.display_name}
              {me?.id === row.id ? (
                <Typography.Text type="secondary">（我）</Typography.Text>
              ) : null}
            </div>
            <Typography.Text type="secondary">#{row.id}</Typography.Text>
          </div>
        </Space>
      ),
    },
    { title: '邮箱', dataIndex: 'email' },
    {
      title: '当前角色',
      dataIndex: 'role',
      width: 120,
      render: (role: Role) => <RoleTag role={role} />,
    },
    {
      title: '调整角色',
      dataIndex: 'role',
      width: 160,
      render: (role: Role, row) => {
        const isSelf = me?.id === row.id
        const pending = roleConfirm?.userId === row.id ? roleConfirm.next : null
        const copy = pending ? rolePopconfirm(row, pending) : { title: '确认改角色？', description: '' }
        const select = (
          <Select
            value={role}
            disabled={isSelf}
            style={{ width: 128 }}
            loading={busyId === row.id}
            options={roles.map((item) => ({ value: item, label: roleLabel[item] }))}
            onChange={(next) => {
              if (next === role) return
              setRoleConfirm({ userId: row.id, next })
            }}
          />
        )
        if (isSelf) {
          return (
            <Tooltip title="不能取消自己的管理员身份">
              <span>{select}</span>
            </Tooltip>
          )
        }
        return (
          <Popconfirm
            open={pending !== null}
            title={copy.title}
            description={copy.description || undefined}
            okText="确认"
            cancelText="取消"
            placement="left"
            okButtonProps={{ loading: busyId === row.id }}
            onConfirm={() => pending && void changeRole(row, pending)}
            onCancel={() => setRoleConfirm(null)}
            onOpenChange={(visible) => {
              if (!visible) setRoleConfirm((cur) => (cur?.userId === row.id ? null : cur))
            }}
          >
            <span>{select}</span>
          </Popconfirm>
        )
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 140,
      render: (status: AccountStatus, row) => {
        const isSelf = me?.id === row.id
        const confirm = statusConfirm(row)
        const toggle = (
          <Switch size="small" checked={status === 'active'} disabled={isSelf} loading={busyId === row.id} />
        )
        return (
          <Space>
            <Tag color={status === 'disabled' ? 'default' : 'green'}>{accountStatusLabel[status]}</Tag>
            {isSelf ? (
              <Tooltip title="不能停用自己的账号">
                <span>{toggle}</span>
              </Tooltip>
            ) : (
              <Popconfirm
                title={confirm.title}
                description={confirm.description}
                okText={confirm.okText}
                cancelText="取消"
                placement="left"
                okButtonProps={{ danger: confirm.danger, loading: busyId === row.id }}
                onConfirm={() => void changeStatus(row, confirm.next)}
              >
                <span>{toggle}</span>
              </Popconfirm>
            )}
          </Space>
        )
      },
    },
    {
      title: '未关单',
      dataIndex: 'open_ticket_count',
      width: 88,
      render: (count: number) => count || 0,
    },
    {
      title: '注册时间',
      dataIndex: 'created_at',
      width: 180,
      render: (value: string) => formatTime(value),
    },
  ]

  const inviteColumns: ColumnsType<PublicInvite> = [
    {
      title: '邀请码',
      dataIndex: 'code',
      render: (code: string) => <Typography.Text copyable>{code}</Typography.Text>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (status: PublicInvite['status']) => (
        <Tag color={status === 'unused' ? 'green' : status === 'used' ? 'blue' : 'default'}>{inviteStatusLabel(status)}</Tag>
      ),
    },
    { title: '创建人', dataIndex: 'creator_name', width: 120 },
    {
      title: '有效期至',
      dataIndex: 'expires_at',
      width: 180,
      render: (value: string) => formatTime(value),
    },
    {
      title: '使用者',
      dataIndex: 'used_by_name',
      width: 120,
      render: (name: string | null) => name || '—',
    },
  ]

  return (
    <div className="page-stack">
      <PageHeader
        title="用户管理"
        extra={
          <Button type="primary" loading={inviteBusy} onClick={() => void createInvite()}>
            生成邀请码
          </Button>
        }
      >
        新员工必须用邀请码注册。改角色、停用会先确认；取消接单资格或停用时，未关单退回待派池。管理员仍可接单。
      </PageHeader>
      <Card size="small">
        <Space wrap>
          <Select
            allowClear
            placeholder="角色"
            style={{ width: 140 }}
            value={roleFilter || undefined}
            options={roles.map((item) => ({ value: item, label: roleLabel[item] }))}
            onChange={(value) => {
              setRoleFilter(value ?? '')
              setPage(1)
            }}
          />
          <Select
            allowClear
            placeholder="账号状态"
            style={{ width: 140 }}
            value={statusFilter || undefined}
            options={[
              { value: 'active', label: '正常' },
              { value: 'disabled', label: '已停用' },
            ]}
            onChange={(value) => {
              setStatusFilter(value ?? '')
              setPage(1)
            }}
          />
          <Input.Search
            allowClear
            placeholder="邮箱或显示名"
            style={{ width: 240 }}
            onSearch={(value) => {
              setKeyword(value.trim())
              setPage(1)
            }}
          />
        </Space>
      </Card>
      <Card className="table-card">
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={users}
          pagination={listPagination(page, pageSize, total, (nextPage, nextSize) => {
            setPage(nextPage)
            setPageSize(nextSize)
          })}
        />
      </Card>
      <Card title="邀请码" className="table-card">
        <Table
          rowKey="id"
          columns={inviteColumns}
          dataSource={invites}
          pagination={false}
          locale={{ emptyText: '还没有邀请码，点右上角生成' }}
        />
      </Card>
    </div>
  )
}
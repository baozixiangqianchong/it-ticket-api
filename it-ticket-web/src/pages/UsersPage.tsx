import { App, Avatar, Card, Input, Select, Space, Table, Tooltip, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useCallback, useEffect, useState } from 'react'

import { api } from '../api/client'
import type { AdminUser, Role } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { RoleTag } from '../components/RoleTag'
import { formatTime, roleLabel } from '../lib/labels'
import { errText } from '../lib/toast'

const roles: Role[] = ['user', 'agent', 'admin']
const PAGE_SIZE = 10

export function UsersPage() {
  const { message } = App.useApp()
  const { user: me } = useAuth()
  const [users, setUsers] = useState<AdminUser[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [busyId, setBusyId] = useState<number | null>(null)
  const [roleFilter, setRoleFilter] = useState<Role | ''>('')
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)

  const reload = useCallback(async () => {
    const out = await api.listUsers({
      page,
      pageSize: PAGE_SIZE,
      role: roleFilter,
      q: keyword,
    })
    setUsers(out.items ?? [])
    setTotal(out.total)
  }, [page, roleFilter, keyword])

  useEffect(() => {
    setLoading(true)
    reload()
      .catch((err) => message.error(errText(err, '加载用户失败')))
      .finally(() => setLoading(false))
  }, [reload, message])

  async function changeRole(id: number, role: Role) {
    if (me?.id === id) {
      message.error('不能取消自己的管理员身份')
      return
    }
    setBusyId(id)
    try {
      await api.updateUserRole(id, role)
      await reload()
      message.success('角色已更新')
    } catch (err) {
      message.error(errText(err, '改角色失败'))
    } finally {
      setBusyId(null)
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
        return (
          <Tooltip title={isSelf ? '不能取消自己的管理员身份' : undefined}>
            <span>
              <Select
                value={role}
                disabled={isSelf}
                style={{ width: 128 }}
                loading={busyId === row.id}
                options={roles.map((item) => ({ value: item, label: roleLabel[item] }))}
                onChange={(next) => void changeRole(row.id, next)}
              />
            </span>
          </Tooltip>
        )
      },
    },
    {
      title: '注册时间',
      dataIndex: 'created_at',
      width: 180,
      render: (value: string) => formatTime(value),
    },
  ]

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <div>
        <Typography.Title level={3} style={{ margin: 0 }}>
          用户管理
        </Typography.Title>
        <Typography.Text type="secondary">共 {total} 人。改角色立即生效，对方不用重新登录。</Typography.Text>
      </div>
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
      <Card styles={{ body: { padding: 0 } }}>
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={users}
          pagination={{
            current: page,
            pageSize: PAGE_SIZE,
            total,
            hideOnSinglePage: true,
            onChange: setPage,
          }}
        />
      </Card>
    </Space>
  )
}

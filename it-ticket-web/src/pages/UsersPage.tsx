import { Alert, Avatar, Card, Select, Space, Table, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useEffect, useState } from 'react'

import { ApiError, api } from '../api/client'
import type { AdminUser, Role } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { RoleTag } from '../components/RoleTag'
import { formatTime, roleLabel } from '../lib/labels'

const roles: Role[] = ['user', 'agent', 'admin']

export function UsersPage() {
  const { user: me } = useAuth()
  const [users, setUsers] = useState<AdminUser[]>([])
  const [total, setTotal] = useState(0)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [busyId, setBusyId] = useState<number | null>(null)

  async function reload() {
    const page = await api.listUsers()
    setUsers(page.items ?? [])
    setTotal(page.total)
  }

  useEffect(() => {
    reload()
      .catch((err) => setError(err instanceof ApiError ? err.message : '加载用户失败'))
      .finally(() => setLoading(false))
  }, [])

  async function changeRole(id: number, role: Role) {
    setBusyId(id)
    setError('')
    try {
      await api.updateUserRole(id, role)
      await reload()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '改角色失败')
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
      render: (role: Role, row) => (
        <Select
          value={role}
          style={{ width: 128 }}
          loading={busyId === row.id}
          options={roles.map((item) => ({ value: item, label: roleLabel[item] }))}
          onChange={(next) => void changeRole(row.id, next)}
        />
      ),
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
        <Typography.Text type="secondary">
          共 {total} 人
        </Typography.Text>
      </div>
      {error ? <Alert type="error" showIcon message={error} /> : null}
      <Card styles={{ body: { padding: 0 } }}>
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={users}
          pagination={{ pageSize: 10, hideOnSinglePage: true }}
        />
      </Card>
    </Space>
  )
}

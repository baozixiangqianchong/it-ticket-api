import {
  CustomerServiceOutlined,
  FileAddOutlined,
  LogoutOutlined,
  TeamOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons'
import { Avatar, Button, Layout, Menu, Space, Typography } from 'antd'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'

import { useAuth } from '../auth/AuthContext'
import { RoleTag } from '../components/RoleTag'

const { Header, Sider, Content } = Layout

export function AppShell() {
  const { user, logout } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()

  const selected = location.pathname.startsWith('/users')
    ? '/users'
    : location.pathname.startsWith('/tickets/new')
      ? '/tickets/new'
      : '/tickets'

  return (
    <Layout className="app-layout">
      <Sider breakpoint="lg" collapsedWidth={72} width={232}>
        <div className="brand">
          <CustomerServiceOutlined className="brand-icon" />
          <div>
            <strong>IT 工单台</strong>
            <span>内部服务台</span>
          </div>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selected]}
          onClick={({ key }) => navigate(key)}
          items={[
            { key: '/tickets', icon: <UnorderedListOutlined />, label: '工单列表' },
            { key: '/tickets/new', icon: <FileAddOutlined />, label: '新建工单' },
            ...(user?.role === 'admin'
              ? [{ key: '/users', icon: <TeamOutlined />, label: '用户管理' }]
              : []),
          ]}
        />
      </Sider>
      <Layout>
        <Header className="app-header">
          <Typography.Text type="secondary">工单处理工作台</Typography.Text>
          <Space size={12}>
            <Avatar style={{ background: '#0f766e' }}>
              {user?.display_name?.slice(0, 1) || 'U'}
            </Avatar>
            <div className="header-user">
              <div className="header-name">{user?.display_name}</div>
              <Space size={6}>
                {user ? <RoleTag role={user.role} /> : null}
                <Typography.Text type="secondary">{user?.email}</Typography.Text>
              </Space>
            </div>
            <Button icon={<LogoutOutlined />} onClick={logout}>
              退出
            </Button>
          </Space>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}

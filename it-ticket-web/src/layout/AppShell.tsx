import {
  BellOutlined,
  CustomerServiceOutlined,
  DashboardOutlined,
  FileAddOutlined,
  FormOutlined,
  HistoryOutlined,
  LogoutOutlined,
  TeamOutlined,
  UnorderedListOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { Avatar, Dropdown, Layout, Menu, Space, Typography } from 'antd'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'

import { useAuth } from '../auth/AuthContext'
import { NotificationBell } from '../components/NotificationBell'
import { RoleTag } from '../components/RoleTag'

const { Header, Sider, Content } = Layout

function pageTitle(pathname: string): string {
  if (pathname.startsWith('/tickets/new')) return '新建工单'
  if (pathname.startsWith('/tickets/')) return '工单详情'
  if (pathname.startsWith('/tickets')) return '工单列表'
  if (pathname.startsWith('/board')) return '工作台'
  if (pathname.startsWith('/users')) return '用户管理'
  if (pathname.startsWith('/catalog')) return '报修模板'
  if (pathname.startsWith('/audits')) return '全局审计'
  if (pathname.startsWith('/notifications')) return '通知中心'
  if (pathname.startsWith('/profile')) return '个人资料'
  return 'IT 工单台'
}

export function AppShell() {
  const { user, logout } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()

  const selected = location.pathname.startsWith('/users')
    ? '/users'
    : location.pathname.startsWith('/board')
      ? '/board'
      : location.pathname.startsWith('/catalog')
        ? '/catalog'
        : location.pathname.startsWith('/audits')
          ? '/audits'
        : location.pathname.startsWith('/profile')
          ? '/profile'
          : location.pathname.startsWith('/notifications')
            ? '/notifications'
            : location.pathname.startsWith('/tickets/new')
              ? '/tickets/new'
              : '/tickets'

  return (
    <Layout className="app-layout">
      <Sider className="app-sider" breakpoint="lg" collapsedWidth={72} width={232}>
        <div className="brand">
          <CustomerServiceOutlined className="brand-icon" />
          <div>
            <strong>IT 工单台</strong>
            <span>内部服务台</span>
          </div>
        </div>
        <div className="sider-nav">
          <Menu
            theme="dark"
            mode="inline"
            selectedKeys={[selected]}
            onClick={({ key }) => navigate(key)}
            items={[
              { key: '/tickets', icon: <UnorderedListOutlined />, label: '工单列表' },
              { key: '/tickets/new', icon: <FileAddOutlined />, label: '新建工单' },
              ...(user?.role === 'admin'
                ? [{ key: '/board', icon: <DashboardOutlined />, label: '工作台' }]
                : []),
              { key: '/notifications', icon: <BellOutlined />, label: '通知中心' },
              ...(user?.role === 'admin'
                ? [
                    { type: 'divider' as const },
                    { key: '/users', icon: <TeamOutlined />, label: '用户管理' },
                    { key: '/catalog', icon: <FormOutlined />, label: '报修模板' },
                    { key: '/audits', icon: <HistoryOutlined />, label: '全局审计' },
                  ]
                : []),
            ]}
          />
        </div>
        <Menu
          className="sider-account"
          theme="dark"
          mode="inline"
          selectedKeys={[selected]}
          onClick={({ key }) => navigate(key)}
          items={[{ key: '/profile', icon: <UserOutlined />, label: '个人资料' }]}
        />
      </Sider>
      <Layout className="app-main">
        <Header className="app-header">
          <Typography.Title level={5} className="app-header-title">
            {pageTitle(location.pathname)}
          </Typography.Title>
          <Space size={8}>
            <NotificationBell />
            <Dropdown
              trigger={['click']}
              menu={{
                items: [
                  { key: 'profile', icon: <UserOutlined />, label: '个人资料' },
                  { type: 'divider' },
                  { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true },
                ],
                onClick: ({ key }) => {
                  if (key === 'logout') {
                    logout()
                    return
                  }
                  navigate('/profile')
                },
              }}
            >
              <button type="button" className="header-user-btn">
                <Avatar style={{ background: '#0f766e' }}>
                  {user?.display_name?.slice(0, 1) || 'U'}
                </Avatar>
                <div className="header-user">
                  <div className="header-name">{user?.display_name}</div>
                  <Space size={6}>
                    {user ? <RoleTag role={user.role} /> : null}
                    <Typography.Text type="secondary" className="header-user-email">
                      {user?.email}
                    </Typography.Text>
                  </Space>
                </div>
              </button>
            </Dropdown>
          </Space>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}

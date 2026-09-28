import { CustomerServiceOutlined } from '@ant-design/icons'
import { Typography } from 'antd'
import type { ReactNode } from 'react'

const { Title, Paragraph } = Typography

export function AuthShell({
  title,
  subtitle,
  children,
}: {
  title: string
  subtitle: string
  children: ReactNode
}) {
  return (
    <div className="auth-shell">
      <section className="auth-hero">
        <div className="auth-hero-mark">
          <CustomerServiceOutlined />
        </div>
        <Title level={2} style={{ color: '#ecfdf8', marginBottom: 8 }}>
          IT 工单台
        </Title>
        <Paragraph style={{ color: '#99f6e4', fontSize: 16 }}>
          提单、指派、跟进、关闭，一套流程走完。
        </Paragraph>
        <ul className="auth-points">
          <li>员工提单，IT 领取或管理员指派</li>
          <li>等用户、转派、重开都留在同一张单上</li>
          <li>通知、审计、角色变更都可追溯</li>
        </ul>
      </section>
      <section className="auth-panel">
        <div className="auth-form">
          <Title level={3} style={{ marginBottom: 4 }}>
            {title}
          </Title>
          <Paragraph type="secondary">{subtitle}</Paragraph>
          {children}
        </div>
      </section>
    </div>
  )
}

import { Space, Typography } from 'antd'
import type { ReactNode } from 'react'

export function PageHeader({
  title,
  extra,
  children,
}: {
  title: ReactNode
  extra?: ReactNode
  children?: ReactNode
}) {
  return (
    <div className="page-head">
      <div>
        <Typography.Title level={3} className="page-title">
          {title}
        </Typography.Title>
        {children ? <div className="page-desc">{children}</div> : null}
      </div>
      {extra ? <Space wrap>{extra}</Space> : null}
    </div>
  )
}

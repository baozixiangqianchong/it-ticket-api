import { Spin } from 'antd'
import { Navigate, Outlet, useLocation } from 'react-router-dom'

import { useAuth } from './AuthContext'

export function RequireAuth() {
  const { user, loading } = useAuth()
  const location = useLocation()

  if (loading) {
    return (
      <div className="boot-spin">
        <Spin size="large" tip="正在确认登录状态…" />
      </div>
    )
  }
  if (!user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }
  return <Outlet />
}

export function RequireAdmin() {
  const { user } = useAuth()
  if (user?.role !== 'admin') {
    return <Navigate to="/tickets" replace />
  }
  return <Outlet />
}

export function GuestOnly() {
  const { user, loading } = useAuth()
  if (loading) {
    return (
      <div className="boot-spin">
        <Spin size="large" tip="正在确认登录状态…" />
      </div>
    )
  }
  if (user) {
    return <Navigate to="/tickets" replace />
  }
  return <Outlet />
}

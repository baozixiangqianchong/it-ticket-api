import { LockOutlined, MailOutlined } from '@ant-design/icons'
import { App, Button, Form, Input } from 'antd'
import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'

import { useAuth } from '../auth/AuthContext'
import { AuthShell } from '../layout/AuthShell'
import { errText } from '../lib/toast'

type LoginForm = {
  email: string
  password: string
}

export function LoginPage() {
  const { login } = useAuth()
  const { message } = App.useApp()
  const navigate = useNavigate()
  const location = useLocation()
  const from =
    (location.state as { from?: string } | null)?.from &&
    (location.state as { from?: string }).from !== '/login'
      ? (location.state as { from: string }).from
      : '/tickets'
  const [submitting, setSubmitting] = useState(false)

  async function onFinish(values: LoginForm) {
    setSubmitting(true)
    try {
      await login(values.email.trim(), values.password)
      navigate(from, { replace: true })
    } catch (err) {
      message.error(errText(err, '登录失败'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthShell title="欢迎回来" subtitle="使用已开通账号进入管理端">
      <Form layout="vertical" requiredMark={false} onFinish={onFinish}>
        <Form.Item name="email" label="邮箱" rules={[{ required: true, message: '请填写邮箱' }]}>
          <Input size="large" prefix={<MailOutlined />} placeholder="name@company.com" />
        </Form.Item>
        <Form.Item name="password" label="密码" rules={[{ required: true, message: '请填写密码' }]}>
          <Input.Password size="large" prefix={<LockOutlined />} placeholder="请输入密码" />
        </Form.Item>
        <Button type="primary" htmlType="submit" size="large" block loading={submitting}>
          登录
        </Button>
      </Form>
      <div className="auth-switch">
        还没有账号？<Link to="/register">去注册</Link>
      </div>
    </AuthShell>
  )
}

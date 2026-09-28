import { KeyOutlined, LockOutlined, MailOutlined, UserOutlined } from '@ant-design/icons'
import { App, Button, Form, Input } from 'antd'
import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'

import { useAuth } from '../auth/AuthContext'
import { AuthShell } from '../layout/AuthShell'
import { errText } from '../lib/toast'

type RegisterForm = {
  inviteCode: string
  displayName: string
  email: string
  password: string
}

export function RegisterPage() {
  const { register } = useAuth()
  const { message } = App.useApp()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [submitting, setSubmitting] = useState(false)

  async function onFinish(values: RegisterForm) {
    setSubmitting(true)
    try {
      await register(values.email.trim(), values.password, values.displayName.trim(), values.inviteCode.trim())
      navigate('/tickets', { replace: true })
    } catch (err) {
      message.error(errText(err, '注册失败'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthShell title="创建账号" subtitle="需要管理员发放的邀请码，新账号默认是员工">
      <Form
        layout="vertical"
        requiredMark={false}
        initialValues={{ inviteCode: searchParams.get('code') ?? searchParams.get('invite') ?? '' }}
        onFinish={onFinish}
      >
        <Form.Item
          name="inviteCode"
          label="邀请码"
          rules={[{ required: true, message: '请填写邀请码' }]}
        >
          <Input size="large" prefix={<KeyOutlined />} placeholder="管理员提供的 8 位邀请码" />
        </Form.Item>
        <Form.Item
          name="displayName"
          label="显示名"
          rules={[{ required: true, message: '请填写显示名' }]}
        >
          <Input size="large" prefix={<UserOutlined />} placeholder="怎么称呼你" />
        </Form.Item>
        <Form.Item name="email" label="邮箱" rules={[{ required: true, message: '请填写邮箱' }]}>
          <Input size="large" prefix={<MailOutlined />} placeholder="name@company.com" />
        </Form.Item>
        <Form.Item
          name="password"
          label="密码"
          rules={[
            { required: true, message: '请填写密码' },
            { min: 8, message: '密码至少 8 位' },
          ]}
        >
          <Input.Password size="large" prefix={<LockOutlined />} placeholder="8～72 位" />
        </Form.Item>
        <Button type="primary" htmlType="submit" size="large" block loading={submitting}>
          注册并登录
        </Button>
      </Form>
      <div className="auth-switch">
        已有账号？<Link to="/login">去登录</Link>
      </div>
    </AuthShell>
  )
}
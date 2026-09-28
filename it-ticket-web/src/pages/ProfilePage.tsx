import { App, Button, Card, Col, Form, Input, Row, Space, Typography } from 'antd'
import { useEffect, useState } from 'react'

import { api } from '../api/client'
import { useAuth } from '../auth/AuthContext'
import { PageHeader } from '../components/PageHeader'
import { RoleTag } from '../components/RoleTag'
import { errText } from '../lib/toast'

type NameForm = { displayName: string }
type PasswordForm = { currentPassword: string; newPassword: string; confirm: string }

export function ProfilePage() {
  const { user, refreshMe } = useAuth()
  const { message } = App.useApp()
  const [nameForm] = Form.useForm<NameForm>()
  const [passwordForm] = Form.useForm<PasswordForm>()
  const [savingName, setSavingName] = useState(false)
  const [savingPassword, setSavingPassword] = useState(false)

  useEffect(() => {
    if (user) {
      nameForm.setFieldsValue({ displayName: user.display_name })
    }
  }, [user, nameForm])

  async function saveName(values: NameForm) {
    setSavingName(true)
    try {
      await api.updateProfile(values.displayName.trim())
      await refreshMe()
      message.success('显示名已更新')
    } catch (err) {
      message.error(errText(err, '改显示名失败'))
    } finally {
      setSavingName(false)
    }
  }

  async function savePassword(values: PasswordForm) {
    setSavingPassword(true)
    try {
      await api.updatePassword(values.currentPassword, values.newPassword)
      passwordForm.resetFields()
      message.success('密码已更新')
    } catch (err) {
      message.error(errText(err, '改密码失败'))
    } finally {
      setSavingPassword(false)
    }
  }

  return (
    <div className="page-stack">
      <PageHeader title="个人资料">自己改花名和密码，不用找管理员。</PageHeader>
      <Card size="small">
        <Space>
          <Typography.Text>{user?.email}</Typography.Text>
          {user ? <RoleTag role={user.role} /> : null}
        </Space>
      </Card>
      <Row gutter={16}>
        <Col xs={24} md={12}>
          <Card title="显示名">
            <Form
              form={nameForm}
              layout="vertical"
              requiredMark={false}
              initialValues={{ displayName: user?.display_name ?? '' }}
              onFinish={saveName}
            >
              <Form.Item
                name="displayName"
                label="怎么称呼你"
                rules={[
                  { required: true, message: '请填写显示名' },
                  { max: 64, message: '最多 64 个字' },
                ]}
              >
                <Input placeholder="花名或真名" maxLength={64} />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={savingName}>
                保存显示名
              </Button>
            </Form>
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title="登录密码">
            <Form form={passwordForm} layout="vertical" requiredMark={false} onFinish={savePassword}>
              <Form.Item
                name="currentPassword"
                label="当前密码"
                rules={[{ required: true, message: '请填写当前密码' }]}
              >
                <Input.Password autoComplete="current-password" />
              </Form.Item>
              <Form.Item
                name="newPassword"
                label="新密码"
                rules={[
                  { required: true, message: '请填写新密码' },
                  { min: 8, message: '密码至少 8 位' },
                ]}
              >
                <Input.Password autoComplete="new-password" placeholder="8～72 位" />
              </Form.Item>
              <Form.Item
                name="confirm"
                label="确认新密码"
                dependencies={['newPassword']}
                rules={[
                  { required: true, message: '请再输入一次新密码' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('newPassword') === value) {
                        return Promise.resolve()
                      }
                      return Promise.reject(new Error('两次输入的新密码不一致'))
                    },
                  }),
                ]}
              >
                <Input.Password autoComplete="new-password" />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={savingPassword}>
                更新密码
              </Button>
            </Form>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

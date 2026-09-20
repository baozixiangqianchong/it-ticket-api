import { Tag } from 'antd'

import type { Role } from '../api/types'
import { roleColor, roleLabel } from '../lib/labels'

export function RoleTag({ role }: { role: Role }) {
  return <Tag color={roleColor[role]}>{roleLabel[role]}</Tag>
}

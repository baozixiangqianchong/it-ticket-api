import { Tag } from 'antd'

import type { TicketCategory, TicketPriority, TicketStatus } from '../api/types'
import { categoryColor, categoryLabel, priorityColor, priorityLabel, statusColor, statusLabel } from '../lib/labels'

export function StatusTag({ status }: { status: TicketStatus }) {
  return <Tag color={statusColor[status]}>{statusLabel[status] ?? status}</Tag>
}

export function CategoryTag({ category }: { category: TicketCategory }) {
  return <Tag color={categoryColor[category]}>{categoryLabel[category] ?? category}</Tag>
}

export function PriorityTag({ priority }: { priority: TicketPriority }) {
  return <Tag color={priorityColor[priority]}>{priorityLabel[priority] ?? priority}</Tag>
}

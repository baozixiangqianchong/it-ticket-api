import type {
  PublicTicket,
  PublicUser,
  Role,
  TicketAction,
  TicketCategory,
  TicketStatus,
} from '../api/types'

export const roleLabel: Record<Role, string> = {
  user: '员工',
  agent: 'IT',
  admin: '管理员',
}

export const statusLabel: Record<TicketStatus, string> = {
  open: '待处理',
  assigned: '已指派',
  in_progress: '处理中',
  resolved: '已解决',
  closed: '已关闭',
}

export const categoryLabel: Record<TicketCategory, string> = {
  hardware: '硬件',
  software: '软件',
  network: '网络',
  other: '其他',
}

export const actionLabel: Record<TicketAction, string> = {
  start: '开始处理',
  resolve: '标记已解决',
  close: '关闭工单',
  reopen: '重开工单',
}

export const statusColor: Record<TicketStatus, 'blue' | 'purple' | 'gold' | 'green' | 'default'> = {
  open: 'blue',
  assigned: 'purple',
  in_progress: 'gold',
  resolved: 'green',
  closed: 'default',
}

export const roleColor: Record<Role, 'default' | 'cyan' | 'geekblue'> = {
  user: 'default',
  agent: 'cyan',
  admin: 'geekblue',
}

export const categoryColor: Record<TicketCategory, 'orange' | 'processing' | 'geekblue' | 'default'> = {
  hardware: 'orange',
  software: 'processing',
  network: 'geekblue',
  other: 'default',
}

export function formatTime(value: string): string {
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  return d.toLocaleString('zh-CN', { hour12: false })
}

// 按钮按状态机和角色裁剪，和后台 resolveTicketAction 对齐。
export function availableActions(
  ticket: PublicTicket,
  user: PublicUser,
): TicketAction[] {
  const isAdmin = user.role === 'admin'
  const isAssignee = ticket.assignee_id === user.id
  const isCreator = ticket.creator_id === user.id

  if (ticket.status === 'assigned' && (isAdmin || isAssignee)) {
    return ['start']
  }
  if (ticket.status === 'in_progress' && (isAdmin || isAssignee)) {
    return ['resolve']
  }
  if (ticket.status === 'resolved' && (isAdmin || isCreator)) {
    return ['close', 'reopen']
  }
  return []
}

export function canAssign(ticket: PublicTicket, user: PublicUser): boolean {
  if (user.role !== 'admin') return false
  return (
    ticket.status === 'open' ||
    ticket.status === 'assigned' ||
    ticket.status === 'in_progress'
  )
}

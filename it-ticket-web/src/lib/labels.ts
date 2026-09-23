import type {
  Role,
  TicketAction,
  TicketCategory,
  TicketScope,
  TicketStatus,
  UserRef,
} from '../api/types'

export const roleLabel: Record<Role, string> = {
  user: '员工',
  agent: 'IT',
  admin: '管理员',
}

export const statusLabel: Record<TicketStatus, string> = {
  open: '待派单',
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

export const scopeLabel: Record<TicketScope, string> = {
  all: '全部',
  pool: '待领',
  assigned: '待我处理',
  created: '我提交的',
}

export const actionLabel: Record<TicketAction, string> = {
  start: '开始处理',
  resolve: '标记已解决',
  close: '关闭工单',
  reopen: '重开工单',
  cancel: '撤回工单',
  claim: '领取工单',
  assign: '指派处理人',
}

export const auditLabel: Record<string, string> = {
  create: '创建工单',
  assign: '指派处理人',
  claim: '领取工单',
  start: '开始处理',
  resolve: '标记已解决',
  close: '关闭工单',
  reopen: '重开工单',
  cancel: '撤回工单',
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

export function personName(ref: UserRef | null | undefined, empty = '未指派'): string {
  if (!ref) return empty
  return ref.display_name || `用户 #${ref.id}`
}

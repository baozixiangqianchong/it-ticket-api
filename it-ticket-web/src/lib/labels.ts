import type {
  AccountStatus,
  Role,
  TicketAction,
  TicketCategory,
  TicketPriority,
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
  assigned: '待开始',
  in_progress: '处理中',
  pending: '等用户',
  resolved: '待确认',
  closed: '已关闭',
}

export const priorityLabel: Record<TicketPriority, string> = {
  p1: '紧急',
  p2: '普通',
  p3: '低',
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
  waiting: '等对方',
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
  wait: '等用户补充',
  resume: '继续处理',
  transfer: '转派工单',
  edit: '修改工单',
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
  wait: '标记等用户',
  resume: '继续处理',
  transfer: '转派工单',
  edit: '修改工单',
  release: '退回待派池',
  role: '改角色',
  status: '改账号状态',
}

export const notifyLabel: Record<string, string> = {
  assign: '指派',
  claim: '领取',
  start: '开始处理',
  wait: '等用户',
  resume: '继续处理',
  resolve: '已解决',
  close: '关闭',
  reopen: '重开',
  comment: '评论',
  transfer: '转派',
  cancel: '撤回',
  role: '权限',
  account: '账号',
  release: '退回待派',
}

export const statusColor: Record<TicketStatus, 'blue' | 'purple' | 'gold' | 'orange' | 'cyan' | 'default'> = {
  open: 'blue',
  assigned: 'purple',
  in_progress: 'gold',
  pending: 'orange',
  resolved: 'cyan',
  closed: 'default',
}

export const priorityColor: Record<TicketPriority, 'red' | 'orange' | 'default'> = {
  p1: 'red',
  p2: 'orange',
  p3: 'default',
}

export const accountStatusLabel: Record<AccountStatus, string> = {
  active: '正常',
  disabled: '已停用',
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

export function formatRelative(value: string): string {
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  const diff = Date.now() - d.getTime()
  const minutes = Math.round(diff / 60000)
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  const hours = Math.round(minutes / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.round(hours / 24)
  if (days === 1) {
    return `昨天 ${d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false })}`
  }
  if (days < 7) return `${days} 天前`
  return formatTime(value)
}

export function personName(ref: UserRef | null | undefined, empty = '未指派'): string {
  if (!ref) return empty
  return ref.display_name || `用户 #${ref.id}`
}

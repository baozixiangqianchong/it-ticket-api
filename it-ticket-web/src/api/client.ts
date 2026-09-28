import type {
  AccountStatus,
  AdminUser,
  AgentRef,
  LoginResult,
  NotificationPage,
  PublicComment,
  PublicTicket,
  PublicUser,
  Role,
  TicketAction,
  TicketCategory,
  TicketDetail,
  TicketListQuery,
  TicketBoard,
  TicketPage,
  TicketPriority,
  TicketStats,
  UserPage,
  ActivityKind,
  ActivityPage,
  CannedReply,
  CannedReplyInput,
  PublicInvite,
  TicketTemplate,
  TicketTemplateInput,
} from './types'

export const API_BASE = 'http://127.0.0.1:8080'

const TOKEN_KEY = 'it-ticket-token'

export class ApiError extends Error {
  code: number
  status: number
  error: string

  constructor(message: string, code: number, status: number, error = '') {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
    this.error = error
  }
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  const token = getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers })
  let body: { code?: number; error?: string; message?: string; data?: T }
  try {
    body = await res.json()
  } catch {
    throw new ApiError('服务器响应无法解析', 500, res.status)
  }

  const code = body.code ?? res.status
  if (code !== 0) {
    if (code === 401 && token && !path.startsWith('/api/v1/auth/login')) {
      clearToken()
      window.dispatchEvent(new Event('auth:unauthorized'))
    }
    throw new ApiError(body.message || '请求失败', code, res.status, body.error ?? '')
  }
  return body.data as T
}

export const api = {
  register(email: string, password: string, displayName: string, inviteCode: string) {
    return request<PublicUser>('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify({
        invite_code: inviteCode,
        email,
        password,
        display_name: displayName,
      }),
    })
  },

  login(email: string, password: string) {
    return request<LoginResult>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
  },

  me() {
    return request<PublicUser>('/api/v1/me')
  },

  updateProfile(displayName: string) {
    return request<PublicUser>('/api/v1/me/profile', {
      method: 'POST',
      body: JSON.stringify({ display_name: displayName }),
    })
  },

  updatePassword(currentPassword: string, newPassword: string) {
    return request<Record<string, never>>('/api/v1/me/password', {
      method: 'POST',
      body: JSON.stringify({
        current_password: currentPassword,
        new_password: newPassword,
      }),
    })
  },

  listTickets(query: TicketListQuery = {}) {
    const q = new URLSearchParams({
      page: String(query.page ?? 1),
      page_size: String(query.pageSize ?? 20),
    })
    if (query.status) q.set('status', query.status)
    if (query.category) q.set('category', query.category)
    if (query.priority) q.set('priority', query.priority)
    if (query.q?.trim()) q.set('q', query.q.trim())
    if (query.scope) q.set('scope', query.scope)
    if (query.assigneeId) q.set('assignee_id', String(query.assigneeId))
    if (query.timeField) q.set('time_field', query.timeField)
    if (query.from) q.set('from', query.from)
    if (query.to) q.set('to', query.to)
    return request<TicketPage>(`/api/v1/tickets/list?${q}`)
  },

  listTicketTemplates() {
    return request<{ items: TicketTemplate[] }>('/api/v1/ticket-templates')
  },

  listCannedReplies() {
    return request<{ items: CannedReply[] }>('/api/v1/canned-replies')
  },

  adminListTicketTemplates() {
    return request<{ items: TicketTemplate[] }>('/api/v1/admin/ticket-templates')
  },

  createTicketTemplate(body: TicketTemplateInput) {
    return request<TicketTemplate>('/api/v1/admin/ticket-templates', {
      method: 'POST',
      body: JSON.stringify(body),
    })
  },

  updateTicketTemplate(id: number, body: TicketTemplateInput) {
    return request<TicketTemplate>(`/api/v1/admin/ticket-templates/${id}`, {
      method: 'POST',
      body: JSON.stringify(body),
    })
  },

  deleteTicketTemplate(id: number) {
    return request<Record<string, never>>(`/api/v1/admin/ticket-templates/${id}/delete`, { method: 'POST' })
  },

  adminListCannedReplies() {
    return request<{ items: CannedReply[] }>('/api/v1/admin/canned-replies')
  },

  createCannedReply(body: CannedReplyInput) {
    return request<CannedReply>('/api/v1/admin/canned-replies', {
      method: 'POST',
      body: JSON.stringify(body),
    })
  },

  updateCannedReply(id: number, body: CannedReplyInput) {
    return request<CannedReply>(`/api/v1/admin/canned-replies/${id}`, {
      method: 'POST',
      body: JSON.stringify(body),
    })
  },

  deleteCannedReply(id: number) {
    return request<Record<string, never>>(`/api/v1/admin/canned-replies/${id}/delete`, { method: 'POST' })
  },

  ticketBoard() {
    return request<TicketBoard>('/api/v1/tickets/board')
  },

  closeStaleTickets() {
    return request<{ closed: number }>('/api/v1/tickets/close-stale', { method: 'POST' })
  },

  ticketStats() {
    return request<TicketStats>('/api/v1/tickets/stats')
  },

  createTicket(title: string, description: string, category: TicketCategory, priority: TicketPriority) {
    return request<PublicTicket>('/api/v1/tickets/create', {
      method: 'POST',
      body: JSON.stringify({ title, description, category, priority }),
    })
  },

  getTicket(id: number) {
    return request<TicketDetail>(`/api/v1/tickets/${id}`)
  },

  assignTicket(id: number, assigneeId: number) {
    return request<PublicTicket>(`/api/v1/tickets/${id}/assign`, {
      method: 'POST',
      body: JSON.stringify({ assignee_id: assigneeId }),
    })
  },

  claimTicket(id: number) {
    return request<PublicTicket>(`/api/v1/tickets/${id}/claim`, {
      method: 'POST',
      body: JSON.stringify({}),
    })
  },

  transferTicket(id: number, assigneeId: number, reason = '') {
    return request<PublicTicket>(`/api/v1/tickets/${id}/transfer`, {
      method: 'POST',
      body: JSON.stringify({ assignee_id: assigneeId, reason }),
    })
  },

  listAgents() {
    return request<{ items: AgentRef[] }>('/api/v1/agents')
  },

  listNotifications(opts: { page?: number; pageSize?: number; unread?: boolean } = {}) {
    const q = new URLSearchParams({
      page: String(opts.page ?? 1),
      page_size: String(opts.pageSize ?? 20),
    })
    if (opts.unread) q.set('unread', '1')
    return request<NotificationPage>(`/api/v1/notifications?${q}`)
  },

  readNotification(id: number) {
    return request<Record<string, never>>(`/api/v1/notifications/${id}/read`, { method: 'POST' })
  },

  readAllNotifications() {
    return request<Record<string, never>>('/api/v1/notifications/read-all', { method: 'POST' })
  },

  editTicket(
    id: number,
    body: { title: string; description: string; category: TicketCategory; priority: TicketPriority },
  ) {
    return request<PublicTicket>(`/api/v1/tickets/${id}/edit`, {
      method: 'POST',
      body: JSON.stringify(body),
    })
  },

  updateTicket(id: number, action: TicketAction, reason?: string) {
    return request<PublicTicket>(`/api/v1/tickets/${id}/update`, {
      method: 'POST',
      body: JSON.stringify({ action, reason: reason ?? '' }),
    })
  },

  addComment(id: number, body: string) {
    return request<PublicComment>(`/api/v1/tickets/${id}/comments`, {
      method: 'POST',
      body: JSON.stringify({ body }),
    })
  },

  listUsers(opts: { page?: number; pageSize?: number; role?: Role | ''; status?: AccountStatus | ''; q?: string } = {}) {
    const q = new URLSearchParams({
      page: String(opts.page ?? 1),
      page_size: String(opts.pageSize ?? 50),
    })
    if (opts.role) q.set('role', opts.role)
    if (opts.status) q.set('status', opts.status)
    if (opts.q?.trim()) q.set('q', opts.q.trim())
    return request<UserPage>(`/api/v1/admin/users?${q}`)
  },

  updateUserRole(id: number, role: Role) {
    return request<AdminUser>(`/api/v1/admin/users/${id}/role`, {
      method: 'POST',
      body: JSON.stringify({ role }),
    })
  },

  updateUserStatus(id: number, status: AccountStatus) {
    return request<AdminUser>(`/api/v1/admin/users/${id}/status`, {
      method: 'POST',
      body: JSON.stringify({ status }),
    })
  },

  listInvites() {
    return request<{ items: PublicInvite[] }>('/api/v1/admin/invites')
  },

  createInvite() {
    return request<PublicInvite>('/api/v1/admin/invites', { method: 'POST' })
  },

  listAudits(opts: { page?: number; pageSize?: number; kind?: ActivityKind | ''; action?: string; q?: string; from?: string; to?: string } = {}) {
    const q = new URLSearchParams({
      page: String(opts.page ?? 1),
      page_size: String(opts.pageSize ?? 20),
    })
    if (opts.kind) q.set('kind', opts.kind)
    if (opts.action) q.set('action', opts.action)
    if (opts.q?.trim()) q.set('q', opts.q.trim())
    if (opts.from) q.set('from', opts.from)
    if (opts.to) q.set('to', opts.to)
    return request<ActivityPage>(`/api/v1/admin/audits?${q}`)
  },
}

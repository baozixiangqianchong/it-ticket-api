import type {
  AdminUser,
  LoginResult,
  PublicComment,
  PublicTicket,
  PublicUser,
  TicketAction,
  TicketCategory,
  TicketDetail,
  UserPage,
} from './types'

// 开发阶段直连本机后台，不走 Vite 代理。
export const API_BASE = 'http://127.0.0.1:8080'

const TOKEN_KEY = 'it-ticket-token'

export class ApiError extends Error {
  code: number
  status: number

  constructor(message: string, code: number, status: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
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
  let body: { code?: number; message?: string; data?: T }
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
    throw new ApiError(body.message || '请求失败', code, res.status)
  }
  return body.data as T
}

export const api = {
  register(email: string, password: string, displayName: string) {
    return request<PublicUser>('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify({
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

  listTickets() {
    return request<PublicTicket[]>('/api/v1/tickets/list')
  },

  createTicket(title: string, description: string, category: TicketCategory) {
    return request<PublicTicket>('/api/v1/tickets/create', {
      method: 'POST',
      body: JSON.stringify({ title, description, category }),
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

  updateTicket(id: number, action: TicketAction) {
    return request<PublicTicket>(`/api/v1/tickets/${id}/update`, {
      method: 'POST',
      body: JSON.stringify({ action }),
    })
  },

  addComment(id: number, body: string) {
    return request<PublicComment>(`/api/v1/tickets/${id}/comments`, {
      method: 'POST',
      body: JSON.stringify({ body }),
    })
  },

  listUsers(page = 1, pageSize = 50) {
    const q = new URLSearchParams({
      page: String(page),
      page_size: String(pageSize),
    })
    return request<UserPage>(`/api/v1/admin/users?${q}`)
  },

  updateUserRole(id: number, role: PublicUser['role']) {
    return request<AdminUser>(`/api/v1/admin/users/${id}/role`, {
      method: 'POST',
      body: JSON.stringify({ role }),
    })
  },
}

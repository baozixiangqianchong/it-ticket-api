export type Role = 'user' | 'agent' | 'admin'

export type TicketStatus =
  | 'open'
  | 'assigned'
  | 'in_progress'
  | 'resolved'
  | 'closed'

export type TicketCategory = 'hardware' | 'software' | 'network' | 'other'

export type TicketScope = 'all' | 'created' | 'assigned' | 'pool'

/** 详情按钮以后台 available_actions 为准。assign / claim 走单独接口。 */
export type TicketAction = 'start' | 'resolve' | 'close' | 'reopen' | 'cancel' | 'claim' | 'assign'

export type UserRef = {
  id: number
  display_name: string
}

export type PublicUser = {
  id: number
  email: string
  display_name: string
  role: Role
}

export type LoginResult = {
  token: string
  user: PublicUser
}

export type PublicTicket = {
  id: number
  title: string
  description: string
  category: TicketCategory
  status: TicketStatus
  creator_id: number
  assignee_id: number | null
  creator: UserRef
  assignee: UserRef | null
  created_at: string
  updated_at: string
  closed_at: string | null
  available_actions?: TicketAction[]
}

export type PublicComment = {
  id: number
  ticket_id: number
  author_id: number
  author: UserRef
  body: string
  created_at: string
}

export type PublicAudit = {
  id: number
  action: string
  from_status: TicketStatus | null
  to_status: TicketStatus
  reason: string | null
  actor: UserRef
  created_at: string
}

export type TicketDetail = PublicTicket & {
  comments: PublicComment[]
  audits: PublicAudit[]
}

export type TicketPage = {
  items: PublicTicket[]
  total: number
  page: number
  page_size: number
}

export type TicketStats = {
  created: number
  assigned: number
  pool: number
  all: number
}

export type TicketListQuery = {
  page?: number
  pageSize?: number
  status?: TicketStatus | ''
  category?: TicketCategory | ''
  q?: string
  scope?: TicketScope
}

export type AdminUser = PublicUser & {
  created_at: string
}

export type UserPage = {
  items: AdminUser[]
  total: number
  page: number
  page_size: number
}

export type Role = 'user' | 'agent' | 'admin'

export type TicketStatus =
  | 'open'
  | 'assigned'
  | 'in_progress'
  | 'resolved'
  | 'closed'

export type TicketCategory = 'hardware' | 'software' | 'network' | 'other'

export type TicketAction = 'start' | 'resolve' | 'close' | 'reopen'

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
  created_at: string
  updated_at: string
}

export type PublicComment = {
  id: number
  ticket_id: number
  author_id: number
  body: string
  created_at: string
}

export type TicketDetail = PublicTicket & {
  comments: PublicComment[]
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

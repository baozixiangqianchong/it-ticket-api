export type Role = 'user' | 'agent' | 'admin'

export type TicketStatus =
  | 'open'
  | 'assigned'
  | 'in_progress'
  | 'pending'
  | 'resolved'
  | 'closed'

export type TicketPriority = 'p1' | 'p2' | 'p3'

export type AccountStatus = 'active' | 'disabled'

export type TicketCategory = 'hardware' | 'software' | 'network' | 'other'

export type TicketScope = 'all' | 'created' | 'assigned' | 'waiting' | 'pool'

/** 详情按钮以后台 available_actions 为准。assign / claim 走单独接口。 */
export type TicketAction =
  | 'start'
  | 'resolve'
  | 'close'
  | 'reopen'
  | 'cancel'
  | 'claim'
  | 'assign'
  | 'wait'
  | 'resume'
  | 'transfer'
  | 'edit'

export type UserRef = {
  id: number
  display_name: string
}

export type PublicUser = {
  id: number
  email: string
  display_name: string
  role: Role
  status: AccountStatus
  unread_count: number
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
  priority: TicketPriority
  status: TicketStatus
  creator_id: number
  assignee_id: number | null
  creator: UserRef
  assignee: UserRef | null
  created_at: string
  updated_at: string
  closed_at: string | null
  stale: boolean
  last_comment?: LastCommentPreview | null
  available_actions?: TicketAction[]
}

export type LastCommentPreview = {
  body: string
  author: UserRef
  created_at: string
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
  waiting: number
  pool: number
  all: number
}

export type TicketTimeField = 'created_at' | 'updated_at' | 'closed_at'

export type TicketListQuery = {
  page?: number
  pageSize?: number
  status?: TicketStatus | ''
  category?: TicketCategory | ''
  priority?: TicketPriority | ''
  q?: string
  scope?: TicketScope
  assigneeId?: number
  timeField?: TicketTimeField
  from?: string
  to?: string
}

export type TicketTemplateField = {
  key: string
  label: string
  placeholder: string
  required: boolean
}

export type TemplateIcon = 'printer' | 'email' | 'network' | 'other'

export type TicketTemplate = {
  id: number
  name: string
  category: TicketCategory
  title_hint: string
  hint: string
  icon: TemplateIcon
  sort_order: number
  enabled: boolean
  fields: TicketTemplateField[]
}

export type CannedReply = {
  id: number
  title: string
  body: string
  sort_order: number
  enabled: boolean
}

export type TicketTemplateInput = {
  name: string
  category: TicketCategory
  title_hint: string
  hint: string
  icon: TemplateIcon
  sort_order: number
  enabled: boolean
  fields: TicketTemplateField[]
}

export type CannedReplyInput = {
  title: string
  body: string
  sort_order: number
  enabled: boolean
}

export type AdminUser = Omit<PublicUser, 'unread_count'> & {
  created_at: string
  open_ticket_count: number
  released_count?: number
}

export type InviteStatus = 'unused' | 'used' | 'expired'

export type PublicInvite = {
  id: number
  code: string
  expires_at: string
  used_at: string | null
  used_by_name: string | null
  creator_name: string
  status: InviteStatus
  created_at: string
}

export type PublicNotification = {
  id: number
  ticket_id: number | null
  type: string
  title: string
  body: string
  read: boolean
  created_at: string
}

export type NotificationPage = {
  items: PublicNotification[]
  total: number
  page: number
  page_size: number
}

export type AgentLoad = {
  id: number
  email: string
  display_name: string
  active: number
  waiting: number
}

export type TicketBoard = {
  pool: number
  pool_stale: number
  in_progress: number
  pending: number
  resolved: number
  resolved_stale: number
  agents: AgentLoad[]
}

export type CloseStaleResult = {
  closed: number
}

export type AgentRef = {
  id: number
  email: string
  display_name: string
}

export type UserPage = {
  items: AdminUser[]
  total: number
  page: number
  page_size: number
}

export type ActivityKind = 'ticket' | 'account'

export type PublicActivity = {
  id: number
  kind: ActivityKind
  action: string
  actor: UserRef
  ticket_id: number | null
  ticket_title: string | null
  user_id: number | null
  user_name: string | null
  from_value: string | null
  to_value: string
  reason: string | null
  created_at: string
}

export type ActivityPage = {
  items: PublicActivity[]
  total: number
  page: number
  page_size: number
}

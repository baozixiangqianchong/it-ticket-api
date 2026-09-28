# V2 接口说明

关联：[V2 产品](./04-v2-product.md) · [V1 接口](./03-architecture-and-api.md)

本文只写 **相对当前线上实现** 的接口差异：哪些路径改契约，哪些路径是新的。产品规则以 04 为准。

对照基准是现在仓库里已挂上的路由，不是 V1 文档里那份尚未落地的 REST 路径。

| 现在实际路径 | V1 文档曾写 | V2 怎么办 |
|---|---|---|
| `POST /api/v1/tickets/create` | `POST /api/v1/tickets` | **保持现网路径**，只改请求 / 响应 |
| `GET /api/v1/tickets/list` | `GET /api/v1/tickets` | 保持路径，改查询参数和 `data` 形状 |
| `POST /api/v1/tickets/:id/update` | `POST /api/v1/tickets/:id/actions` | 保持路径，扩大 `action` |
| 其余 | 与实现一致 | 见下文 |

V2 **不改 URL 前缀**，仍是 `/api/v1`。不为此开 `/api/v2`。前端和后台按批次一起切契约。

---

## 1. 现状接口（对照表）

当前实现一共 12 个接口（V1 文档写 11 个，角色接口已经提前做了）。

| # | 方法 | 路径 | V2 | 说明 |
|---|---|---|---|---|
| 1 | GET | `/healthz` | 不变 | |
| 2 | POST | `/api/v1/auth/register` | 小改 | P1 才加邮箱后缀限制 |
| 3 | POST | `/api/v1/auth/login` | 小改 | 停用账号后 401（P1） |
| 4 | GET | `/api/v1/me` | 改 | 角色以库为准；P1 带未读数 |
| 5 | POST | `/api/v1/tickets/create` | 改 | 响应带显示名；P1 加优先级 |
| 6 | GET | `/api/v1/tickets/list` | **改（破坏性）** | 分页信封 + 筛选 + `scope` |
| 7 | GET | `/api/v1/tickets/:id` | **改** | 显示名、审计、可做动作 |
| 8 | POST | `/api/v1/tickets/:id/assign` | 小改 | 冲突码更明确 |
| 9 | POST | `/api/v1/tickets/:id/update` | **改** | 新动作、`reason`、重开留人 |
| 10 | POST | `/api/v1/tickets/:id/comments` | 改 | 作者显示名；P1 可拉回 pending |
| 11 | GET | `/api/v1/admin/users` | 改 | 增加 `role` / `q` |
| 12 | POST | `/api/v1/admin/users/:id/role` | 小改 | 立即生效（鉴权改动，路径不变） |

P0 新增 2 个；P1 再新增一批。见第 5、第 6 节。

---

## 2. 所有接口都要遵守的改动

### 2.1 鉴权：JWT 只认人，角色查库

现在：中间件把 token 里的 `role` 放进 context，列表 / 指派 / 改状态都信它。有效期 24 小时。

V2：

1. 验签成功后，用 `uid` 读 `users` 一行。
2. 人不存在 → 401。
3. P1 若已停用 → 401，文案「账号已停用」。
4. context 里的 `role` 用**这一行的当前角色**，忽略 token 里的 `role`。
5. `GET /me` 继续以库为准（现在已经是）。

`RequireAdmin` 不用改写法，但它读到的 role 已经是新的。管理员刚把某人降为员工，对方用旧 token 再调 `/admin/*` 或 `/assign` 必须 403。

登录签发的 JWT 仍可带 `role`，仅供调试；服务端不采信。

### 2.2 统一响应

成功不变：`code` 仍是数字 `0`，前端现在的 `code !== 0` 判断继续有效。

失败时**多一个** `error` 字段，用来区分同为 409 的几种业务冲突。旧字段都保留。

```json
{
  "code": 409,
  "error": "TICKET_CLOSED",
  "message": "工单已关闭",
  "data": null
}
```

| HTTP | `code` | `error` | 何时 |
|---|---|---|---|
| 400 | 400 | `INVALID_ARGUMENT` | 缺字段、非法分类、管理员代关却不写原因、重开不写原因 |
| 401 | 401 | `UNAUTHENTICATED` | 未登录、token 无效、登录失败、账号停用 |
| 403 | 403 | `PERMISSION_DENIED` | 已登录但这个动作轮不到他（员工去指派、非处理人去 resolve） |
| 404 | 404 | `NOT_FOUND` | 没有这张单，或有但对当前用户不可见 |
| 409 | 409 | `EMAIL_TAKEN` | 注册邮箱已存在 |
| 409 | 409 | `TICKET_INVALID_TRANSITION` | 状态不允许这个动作 |
| 409 | 409 | `TICKET_CLOSED` | 已关闭且不能再改 / 再评 / 超期重开 |
| 409 | 409 | `TICKET_ALREADY_ASSIGNED` | 自领时已经有处理人 |
| 409 | 409 | `LAST_ADMIN` | 取消最后一个管理员（现有逻辑，补上 error） |
| 500 | 500 | `INTERNAL` | 未预期错误 |
| 503 | 503 | `UNAVAILABLE` | 健康检查发现 MySQL 不通 |

`error` 缺失时前端按 `code` + `message` 兜底，便于分批上线。

### 2.3 工单上的人怎么返回

现在只回 `creator_id` / `assignee_id`。员工端会显示「用户 #2」。

V2 **保留这两个 id**，并加上对象。评论、审计同样带作者。

```json
{
  "id": 12,
  "title": "显示器不亮",
  "description": "开机后面板灯是暗的",
  "category": "hardware",
  "status": "assigned",
  "creator_id": 3,
  "assignee_id": 8,
  "creator": { "id": 3, "display_name": "Alice" },
  "assignee": { "id": 8, "display_name": "Bob" },
  "created_at": "2026-09-23T10:00:00+08:00",
  "updated_at": "2026-09-23T10:05:00+08:00",
  "closed_at": null
}
```

未指派时 `assignee_id` 和 `assignee` 都是 `null`。不要再让前端调用户列表拼名字。

---

## 3. 改动的接口（P0）

### 3.1 `GET /api/v1/tickets/list` — 破坏性

现在：`data` 是工单数组，无筛选、无分页，一次全表。

V2：`data` 改成分页信封，和用户列表同一形状。

查询参数：

| 参数 | 默认 | 说明 |
|---|---|---|
| `page` | 1 | 从 1 开始 |
| `page_size` | 20 | 最大 50 |
| `status` | — | 单个状态 |
| `category` | — | `hardware` / `software` / `network` / `other` |
| `q` | — | 标题模糊 |
| `scope` | `all` | `all` / `created` / `assigned` / `waiting` / `pool`，见产品文档第 7 节 |
| `assignee_id` | — | 仅 `admin` 有效；其他人传入则忽略 |

`data`：

```json
{
  "items": [],
  "total": 0,
  "page": 1,
  "page_size": 20
}
```

`items[]` 用第 2.3 节的工单结构。列表带 `last_comment`（最后一条评论预览），不带完整评论列表和审计。`q` 搜标题、描述、评论和单号。

前端：`listTickets()` 不能再把 `data` 当成数组。

---

### 3.2 `GET /api/v1/tickets/:id`

现在：工单字段 + `comments`（只有 author_id）。没有审计。

V2 在现有字段上增加：

| 字段 | 类型 | 说明 |
|---|---|---|
| `creator` / `assignee` | 对象 / null | 见 2.3 |
| `closed_at` | 时间 / null | 关闭或撤回时写入；重开清掉 |
| `comments[].author` | 对象 | `{ id, display_name }` |
| `audits` | 数组 | 按 `created_at` 升序 |
| `available_actions` | 字符串数组 | 当前用户此刻能做的动作，供按钮用 |

`audits[]`：

```json
{
  "id": 41,
  "action": "claim",
  "from_status": "open",
  "to_status": "assigned",
  "reason": null,
  "actor": { "id": 8, "display_name": "Bob" },
  "created_at": "2026-09-23T10:05:00+08:00"
}
```

`available_actions` 取值是下一节 `update` / `claim` 用的动作名，外加前端要单独调的 `assign`（仅管理员且状态允许时出现）。

看不见 → 404 `NOT_FOUND`。这一点不变。

---

### 3.3 `POST /api/v1/tickets/:id/update`

现在：`{ "action": "start" | "resolve" | "close" | "reopen" }`。重开回到 `open` 并清空处理人。没有原因。

V2 请求：

```json
{
  "action": "reopen",
  "reason": "换完电源还是不亮"
}
```

| `action` | 从 | 到 | 谁 | `reason` |
|---|---|---|---|---|
| `start` | `assigned` | `in_progress` | 当前处理人或 admin | 不需要 |
| `resolve` | `in_progress` | `resolved` | 当前处理人或 admin | 不需要 |
| `close` | `resolved` | `closed` | 创建人或 admin | 管理员且不是创建人时**必填** |
| `reopen` | `resolved` 或 `closed` | `assigned`（保留处理人） | 创建人或 admin | **必填** |
| `cancel` | `open` | `closed` | 创建人 | 可选 |

`claim` **不要**走这个接口，走新增的自领接口，避免和「改状态」混在一个 handler 里抢锁语义。

`reopen` 规则：

- `resolved`：随时可重开
- `closed`：仅 `closed_at + 7 天` 内；超时 409 `TICKET_CLOSED`
- 原处理人仍是 `agent` 且未停用：状态 `assigned`，`assignee_id` 不变
- 否则：状态 `open`，`assignee_id` 清空，回到待派池
- `closed_at` 清空

成功响应：第 2.3 节的工单（不强制带回评论）。写审计，`reason` 进审计备注。

---

### 3.4 `POST /api/v1/tickets/:id/assign`

路径和入参不变：`{ "assignee_id": 8 }`。

补齐：

- 目标不是 `agent`：400 `INVALID_ARGUMENT`（现在已是）
- 状态不是 `open` / `assigned` / `in_progress`：409 `TICKET_INVALID_TRANSITION`（现在是笼统 409）
- 成功后响应带显示名
- 写审计 `assign`；P1 起给新处理人写站内通知

---

### 3.5 `POST /api/v1/tickets/create`

入参不变：`title`、`description`、`category`。

响应改成 2.3 节结构（创建时 `assignee` 为 null，`status` 为 `open`）。审计 `create` 照旧。

P1 才允许可选字段 `priority`。P0 不要收这个字段，收了就忽略。

---

### 3.6 `POST /api/v1/tickets/:id/comments`

入参不变：`{ "body": "..." }`。

响应增加 `author`。`closed` 仍是 409 `TICKET_CLOSED`。

P1：若工单是 `pending` 且评论人是创建人，同一事务把状态拉回 `in_progress` 并写审计 `resume`。P0 不要做这条副作用。

---

### 3.7 `GET /api/v1/admin/users`

现有分页参数保留。增加：

| 参数 | 说明 |
|---|---|
| `role` | `user` / `agent` / `admin`，不传则全部 |
| `q` | 邮箱或显示名模糊 |

派单下拉应调 `?role=agent&page_size=50`，不要再拉全表再在浏览器里 filter。超过 50 个 IT 就翻页，或继续加大请求直到 `items.length < page_size`（管理端自己拼）。P0 不做「一次返回全部 agent」的新接口。

---

### 3.8 `GET /api/v1/me`

路径不变。`data` 仍是 `id` / `email` / `display_name` / `role`，但 role 必须是库里的当前值。

P1 增加 `unread_count`。P0 不要加，避免前端空字段分支。

---

### 3.9 `POST /api/v1/auth/register` / `login`

P0 契约不变，但注册必须带 `invite_code`（管理员生成，24 小时、一次性）。登录时账号已停用 → 401 `UNAUTHENTICATED`，文案不要写成「用户不存在」。

---

## 4. 不变的接口

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/healthz` | 继续 ping MySQL |
| POST | `/api/v1/admin/users/:id/role` | 入参仍是 `{ "role": "agent" }`。不能取消自己的管理员；最后一个 admin 也不能降级。**生效方式**变了：对方不必重新登录 |

角色接口已经存在，V2 不要再把它当成新接口。

---

## 5. 新增接口（P0）

只加两个。其它新能力先复用 `update` / `list`。

### 5.1 `POST /api/v1/tickets/:id/claim` — 自领

登录：是。谁能成功：`agent` 或 `admin`。

请求体：空对象 `{}`，或不传 body。处理人就是当前登录用户，客户端不能指定别人（指定别人走 assign）。

规则：

- 看不见这张单 → 404（员工看不见别人的 open 单，所以员工会 404，不要 403）
- `status` 必须是 `open` 且 `assignee_id` 为空，否则 409 `TICKET_INVALID_TRANSITION` 或 `TICKET_ALREADY_ASSIGNED`
- 成功：`assignee_id = 当前用户`，`status = in_progress`，审计 `claim`
- 并发：更新条件带上 `status = open AND assignee_id IS NULL`，影响行数为 0 则 409 `TICKET_ALREADY_ASSIGNED`

响应：2.3 节工单。

员工调用：404（他看不见别人的单）。若领的是自己提的 open 单：员工角色仍不能 claim，403 `PERMISSION_DENIED`（自己的单对员工可见，但自领是 IT 动作）。

### 5.2 `GET /api/v1/tickets/stats` — 列表页数字

登录：是。给「待领 / 待我处理 / 等对方 / 我提交的」盒子提供计数，避免前端为了角标把列表拉三遍。

无查询参数。按当前角色可见范围统计。

`data`：

```json
{
  "created": 3,
  "assigned": 5,
  "waiting": 1,
  "pool": 2,
  "all": 8
}
```

| 字段 | 含义 |
|---|---|
| `created` | `creator_id = 自己` |
| `assigned` | `assignee_id = 自己` 且状态是 `assigned` / `in_progress` |
| `waiting` | `assignee_id = 自己` 且状态是 `pending` / `resolved` |
| `pool` | 当前用户能看见的 `open` 且无处理人。员工恒为 0 |
| `all` | 角色默认全集条数，和 `scope=all` 的 `total` 一致 |

注意路由登记顺序：`/tickets/stats` 必须写在 `/tickets/:id` **前面**，否则 `stats` 会被当成 id。

---

## 6. 新增接口（P1，已落地；附件不做）

### 6.1 转派

`POST /api/v1/tickets/:id/transfer`

```json
{ "assignee_id": 9, "reason": "这类网络单归你" }
```

当前处理人或 admin；目标必须是另一位 `agent`；状态 `assigned` 或 `in_progress` 或 `pending`。成功后：`pending` 保持等待，其余进入 `in_progress`。审计 `transfer`。

### 6.2 站内通知

| 方法 | 路径 | 作用 |
|---|---|---|
| GET | `/api/v1/notifications` | 当前用户的通知，分页。`unread=1` 只看未读 |
| POST | `/api/v1/notifications/:id/read` | 标已读 |
| POST | `/api/v1/notifications/read-all` | 全部标已读 |

触发点（服务端写，不另开接口）。工单类 `title` 是 `#id 工单标题`，`body` 是「谁做了什么」；账号类 `ticket_id` 为 `null`。

| type | 写给谁 | 什么时候 |
|---|---|---|
| `assign` | 新处理人、原处理人（改派时）、提单人 | 管理员指派 |
| `claim` | 提单人 | IT 自领 |
| `start` | 提单人 | 开始处理 |
| `wait` | 提单人 | 等用户补充 |
| `resume` | 提单人 | IT 点继续处理（提单人评论拉回只走 `comment`） |
| `resolve` | 提单人 | 标记已解决 |
| `close` | 提单人、处理人 | 关单 |
| `reopen` | 提单人、留下的处理人 | 重开 |
| `comment` | 提单人、处理人 | 新评论 |
| `transfer` | 新处理人、原处理人、提单人 | 转派 |
| `cancel` | 处理人 | 提单人撤回已有处理人的单 |
| `role` | 被改角色的人 | 管理员改角色 |
| `account` | 被启用的人 | 管理员重新启用账号 |

不写：待派单撤回（没有处理人）、停用账号、提单广播给全部 IT。操作人自己不给自己发。

`GET /me` 此时增加 `unread_count`。

### 6.3 附件

不做。没有对象存储。路径预留，不要占用 `/attachments`。

### 6.4 停用账号

`POST /api/v1/admin/users/:id/status`

```json
{ "status": "disabled" }
```

`status`：`active` / `disabled`。不能停用最后一个 admin。被停用的人下一请求 401。

### 6.5 等用户（不新开路径）

`wait` / `resume` 加进 `POST /tickets/:id/update` 的 `action`。`wait` 必须写 `reason`。提单人在 `pending` 下评论会拉回处理中。

### 6.6 自助资料、邀请码与模板

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/v1/me/profile` | `{ display_name }`，1～64 字 |
| POST | `/api/v1/me/password` | `{ current_password, new_password }`，新密码 8～72 位 |
| GET | `/api/v1/admin/audits` | `kind=ticket\|account`、`action`、`q`、`from` / `to`、分页。工单时间线与账号变更拼在一起 |
| GET | `/api/v1/admin/invites` | 邀请码列表 |
| POST | `/api/v1/admin/invites` | 生成一条，24 小时后失效 |
| GET | `/api/v1/ticket-templates` | 登录用户，只返回 `enabled=true` 的模板 |
| GET | `/api/v1/canned-replies` | IT / 管理员，只返回启用的常用回复 |
| GET | `/api/v1/admin/ticket-templates` | 管理员，含停用。最多 20 个 |
| POST | `/api/v1/admin/ticket-templates` | `{ name, category, title_hint, hint, icon, sort_order, enabled, fields }`。`icon`：`printer` / `email` / `network` / `other`。`fields` 至少 1 个、最多 8 个，`key` 可空（后台生成 `field_1`） |
| POST | `/api/v1/admin/ticket-templates/:id` | 同上，整份覆盖 |
| POST | `/api/v1/admin/ticket-templates/:id/delete` | 删模板，已提交工单文本不变 |
| GET | `/api/v1/admin/canned-replies` | 管理员，含停用。最多 30 条 |
| POST | `/api/v1/admin/canned-replies` | `{ title, body, sort_order, enabled }` |
| POST | `/api/v1/admin/canned-replies/:id` | 整份覆盖 |
| POST | `/api/v1/admin/canned-replies/:id/delete` | 删除 |

---

## 7. 前端必须一起改的点

P0 接口一落地，现网管理端会坏在这些地方，不要只发后台：

| 文件 / 行为 | 要改什么 |
|---|---|
| `api.listTickets` | `data` 从数组改成 `{ items, total, page, page_size }` |
| 工单列表页 | 筛选、服务端分页、`scope` 三个盒子、调 `stats` |
| 工单详情 | 渲染 `assignee.display_name`、`audits`；按钮看 `available_actions` |
| 待派池 | 调 `POST /tickets/:id/claim` |
| 关闭 / 重开 | 带 `reason`；管理员代关必填 |
| 指派下拉 | `GET /admin/users?role=agent` |
| `useUserNames` | 可以删。名字走工单字段 |
| 改角色后 | 对方不用重新登录。管理员不能取消自己的管理员身份 |

---

## 8. 表结构增量（接口依赖，不是本期重点）

P0 仍是这 4 张表。只加列，不加表。

| 表 | 增加 | 用途 |
|---|---|---|
| `tickets` | `closed_at DATETIME NULL` | 重开 7 天窗口 |
| `audit_logs` | `reason VARCHAR(500) NULL` | 重开 / 代关 / 撤回原因 |
| `tickets.status` 等枚举 | 不动 | `pending` 留给 P1 |

P1 才新建 `notifications`，以及 `users.status`、`tickets.priority`。后续增量：`account_audits`（`sql/005_activity.sql`）、`invite_codes`（`sql/006_invite.sql`）、`ticket_templates` / `canned_replies`（`sql/007_catalog.sql`）。附件表不做。

迁移放 `sql/002_v2.sql`，启动时仍然不要自动建表。具体列定义实现时再写进 `02-database.md` 的续节，不要在本文展开。

---

## 9. 批次和接口对照

和产品文档第 11 节对齐，避免一次改完所有 handler。

| 批次 | 改哪些现有接口 | 新接口 |
|---|---|---|
| A. 读得全 | `list`（分页筛选）、`GET :id`（姓名 + 审计）、鉴权改查库、`admin/users?role=` | `GET /tickets/stats` |
| B. 领得走 | `list` 的 `scope=pool` | `POST /tickets/:id/claim` |
| C. 回得去 | `update` 增加 `cancel` / `reason` / 重开留人 | 无 |
| D. P1 | `update` 增加 `wait` / `resume`；`create` 加 priority；`comments` 拉回 pending；`/me` 未读数 | transfer、notifications、user status、`GET /agents`。附件不做 |

A 可以单独上线（`list` 的 `data` 形状是破坏性的，必须和前端同一天发）。B、C 改状态机，建议同一发布。

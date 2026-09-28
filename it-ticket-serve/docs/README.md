# 后台文档

企业内部 IT 报修：员工提单，管理员派给 IT，IT 修好后由员工确认关闭。三种角色共用同一套工单接口，按角色缩小可见范围。

按主题拆开，不要把库表、接口、角色混在一份长文里。先看下面三块能力对照，再按顺序读分册。


| 阶段       | 状态                              | 读哪份                                                             |
| -------- | ------------------------------- | --------------------------------------------------------------- |
| **当前代码** | 已实现 V2 P0 + P1（不做附件），`go run ./cmd/server` 就是这一套 | 以 `internal/router/router.go` 为准 |
| **一期** | 第一期合同，已被现网路径和 V2 覆盖一部分 | [00](./00-roles-overview.md)～[03](./03-architecture-and-api.md) |
| **二期** | P0 / P1 已落地（附件明确不做）；P2 还没做 | [04](./04-v2-product.md)、[05](./05-v2-api.md) |


---

## 当前代码能做什么

跑起来之后，实际具备这些能力。路径以现网路由为准，不是 03 文档里那份 REST 表。


| 模块 | 已实现 |
| --- | --- |
| 账号 | 管理员发邀请码（24 小时有效、一次性）；注册必须填邀请码，新账号固定 `user`。登录发 JWT、`GET /me`、自己改显示名 / 密码。角色每次查库，改角色下一请求生效 |
| 角色 | `user` / `agent` / `admin`。管理员可改他人角色，不能取消自己的管理员，也不能拿掉最后一个 admin。改成员工或停用处理人时，未关的单退回待派池。用户列表可按 `role` / `q` 筛 |
| 工单 | 提单可用启用中的模板（管理员在「报修模板」配置）；紧急只有管理员能标。`open` 可改标题描述；派出去后只能改优先级。列表可按创建 / 更新 / 关闭时间筛，搜索覆盖标题 / 描述 / 评论 / 单号，并带最后一条评论。IT 有常用回复。管理员有工作台 |
| 流转 | 管理员指派或 IT / 管理员自领 / 转派后直接处理中（等用户的单保持等待）；`start` 主要用于重开；`wait` 必填原因；`resume` / `resolve` / `close` / `reopen` / `cancel`。重开留处理人 |
| 评论 | 对可见且未关闭的单留言。提单人在 `pending` 下评论会拉回处理中 |
| 审计 | 写入 `audit_logs`（含 reason）。详情返回时间线和人名、`available_actions`。改角色 / 停用写入 `account_audits`，管理员可看全局审计 |
| 通知 | 工单类标题是 `#编号 工单标题`，正文是「谁做了什么」。覆盖指派、自领、开始、等用户、继续处理、解决、关闭、重开、评论、转派；另有角色变更、账号启用、未关单退回池。操作人自己不收。`/me` 带未读数 |
| 管理 | 管理员看用户、改角色、停用 / 启用账号、发邀请码、配置提单模板和常用回复 |


现网接口（相对 P0 多了转派、通知、停用、IT 名单）：


| 方法 | 路径 | 谁能成功 |
| --- | --- | --- |
| GET | `/healthz` | 任何人 |
| POST | `/api/v1/auth/register` | 任何人（必须带有效邀请码；新账号是员工） |
| POST | `/api/v1/auth/login` | 已注册用户 |
| GET | `/api/v1/me` | 登录用户（含 `unread_count`） |
| POST | `/api/v1/me/profile` | 登录用户改显示名 |
| POST | `/api/v1/me/password` | 登录用户改密码 |
| GET | `/api/v1/agents` | IT 或管理员 |
| GET | `/api/v1/ticket-templates` | 登录用户（只返回启用的） |
| GET | `/api/v1/canned-replies` | IT 或管理员（只返回启用的） |
| GET | `/api/v1/notifications` | 登录用户 |
| POST | `/api/v1/notifications/read-all` | 登录用户 |
| POST | `/api/v1/notifications/:id/read` | 登录用户 |
| POST | `/api/v1/tickets/create` | 登录用户 |
| GET | `/api/v1/tickets/list` | 登录用户（`scope`；`from` / `to` / `time_field`） |
| GET | `/api/v1/tickets/stats` | 登录用户 |
| GET | `/api/v1/tickets/board` | 仅管理员 |
| POST | `/api/v1/tickets/close-stale` | 仅管理员 |
| POST | `/api/v1/tickets/:id/edit` | 按状态：提单人改待派单，处理人改优先级 |
| GET | `/api/v1/tickets/:id` | 对该单可见的人 |
| POST | `/api/v1/tickets/:id/assign` | 仅管理员 |
| POST | `/api/v1/tickets/:id/claim` | IT 或管理员 |
| POST | `/api/v1/tickets/:id/transfer` | 当前处理人或管理员 |
| POST | `/api/v1/tickets/:id/update` | 按动作：`start` / `wait` / `resume` / `resolve` / `close` / `reopen` / `cancel` |
| POST | `/api/v1/tickets/:id/comments` | 对该单可见的人 |
| GET | `/api/v1/admin/users` | 仅管理员 |
| GET | `/api/v1/admin/audits` | 仅管理员 |
| POST | `/api/v1/admin/users/:id/role` | 仅管理员 |
| POST | `/api/v1/admin/users/:id/status` | 仅管理员 |
| GET | `/api/v1/admin/invites` | 仅管理员 |
| POST | `/api/v1/admin/invites` | 仅管理员 |
| GET | `/api/v1/admin/ticket-templates` | 仅管理员（含停用） |
| POST | `/api/v1/admin/ticket-templates` | 仅管理员，最多 20 个 |
| POST | `/api/v1/admin/ticket-templates/:id` | 仅管理员 |
| POST | `/api/v1/admin/ticket-templates/:id/delete` | 仅管理员 |
| GET | `/api/v1/admin/canned-replies` | 仅管理员（含停用） |
| POST | `/api/v1/admin/canned-replies` | 仅管理员，最多 30 条 |
| POST | `/api/v1/admin/canned-replies/:id` | 仅管理员 |
| POST | `/api/v1/admin/canned-replies/:id/delete` | 仅管理员 |


主路径：员工提单（`open`）→ **管理员指派或 IT 自领后直接进入处理中** → 标记已解决（待确认）→ 员工关闭。重开回到 `assigned`（待开始）并留下原处理人。「待我处理」只含进行中的单；`pending` / `resolved` 在「等对方」。提单人在待派、待开始、处理中可撤回，关闭后 7 天内可重开。等用户必须写原因。

已有库要执行 `sql/002_v2.sql`，再执行 `sql/003_p1.sql`（优先级、`pending`、账号状态、通知表），再执行 `sql/004_notify.sql`（`notifications.ticket_id` 可空，给角色 / 账号通知用），再执行 `sql/005_activity.sql`（账号审计表），再执行 `sql/006_invite.sql`（邀请码），再执行 `sql/007_catalog.sql`（提单模板、常用回复）。附件明确不做。

还没有（V2 P2）：邮件 / IM、SLA、知识库、自动分派、完整 RBAC。

---



## 一期做什么

一期目标：把「鉴权、角色可见范围、状态机、事务审计」跑通。合同在 00～03。


| 项   | 一期合同                                                                          |
| --- | ----------------------------------------------------------------------------- |
| 角色  | 员工只看自己的单；IT 只看派给自己的（加上自己提的）；管理员看全部、负责派单                                       |
| 状态  | `open` → `assigned` → `in_progress` → `resolved` → `closed`；重开回 `open` 并清空处理人 |
| 库表  | 1 个库 `it_ticket`，4 张表：`users`、`tickets`、`ticket_comments`、`audit_logs`        |
| 接口  | 文档写 11 个；路径写成 `POST /tickets`、`GET /tickets`、`POST /tickets/:id/actions`      |


和当前代码的差别（读 03 时不要拿去对现网）：


| 一期文档                                                       | 当前代码                                                    |
| ---------------------------------------------------------- | ------------------------------------------------------- |
| 11 个接口；改角色标成 P1、先不做                                        | 角色接口已经做了，现网是 12 个                                       |
| `POST /tickets`、`GET /tickets`、`POST /tickets/:id/actions` | `/tickets/create`、`/tickets/list`、`/tickets/:id/update` |
| 列表分页、详情带审计和人名 | 现网已按 V2 补上 |
| 错误码是字符串（如 `TICKET_CLOSED`） | `code` 仍是数字，另加 `error` 字符串 |


一期明确不做：附件、邮件、SLA、知识库、自动分派、前端曾标成非目标（管理端后来已经有了）。

---



## 二期做什么

P0 已写进后台。只改闭环，不换角色模型，不开 `/api/v2`。合同在 04、05。


| 批次       | 产品                                                                                      | 接口                                                                                 |
| -------- | --------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| P0 必须    | IT 看待派池并自领；重开留下原处理人；员工可撤回错单；关单 7 天内可重开；管理员代关写原因；详情能看姓名和时间线；列表筛选分页；「我报的 / 我修的」分开；改角色立即生效 | 改现有 `list` / 详情 / `update` 等；**新加** `POST /tickets/:id/claim`、`GET /tickets/stats` |
| P1 已做（附件除外） | 优先级、等用户（`pending`）、站内通知、IT 转派、停用账号 | `transfer`、`notifications`、用户 `status`；附件不做 |
| P2 禁止开工  | 邮件 / IM、SLA、知识库、自动分派、完整 RBAC                                                            | —                                                                                  |


二期主路径：员工提单 → **管理员派单或 IT 自领** → IT 处理 → 员工确认；重开后处理人还在。

---



## 分册目录


| 顺序  | 文档                                      | 对应阶段 | 内容                   |
| --- | --------------------------------------- | ---- | -------------------- |
| 0   | [三个角色各干什么](./00-roles-overview.md)      | 一期   | 一张单怎么走、三种人能做什么       |
| 1   | [需求与角色](./01-requirements-and-roles.md) | 一期   | 做什么、状态机细则、范围、验收      |
| 2   | [数据库](./02-database.md)                 | 一期   | 库名、4 张表、字段、关系、索引     |
| 3   | [项目结构与接口](./03-architecture-and-api.md) | 一期   | 分层、接口合同、错误码（路径以现网为准） |
| 4   | [V2 产品](./04-v2-product.md)             | 二期   | 自领、重开留人、撤回、可见范围、验收   |
| 5   | [V2 接口](./05-v2-api.md)                 | 二期   | 相对现网：哪些接口改契约、哪些是新的   |


Cursor 里的 Ticket Roles 画布只是编辑器预览，进不了这个仓库。角色说明以 `00-roles-overview.md` 为准。

后台与前端同在一个仓库：接口在 `it-ticket-serve/`，管理端在 `it-ticket-web/`。
# 项目结构与接口

关联：[需求与角色](./01-requirements-and-roles.md) · [数据库](./02-database.md)

第一期：**11 个 HTTP 接口**。列表和详情按角色自动缩小可见范围，不要给员工、IT、管理员各做一套列表接口。

改状态不提供 `PUT /tickets/:id` 去随便改字段，避免绕过状态机。

---

## 1. 仓库与目录

Go 模块名 `it-ticket-api`。后台代码在仓库的 `it-ticket-serve/` 目录。

```text
it-ticket-serve/
  cmd/server/main.go          # 进程入口：读配置、连库、挂路由、听端口
  internal/
    config/                   # 端口、DSN、JWT secret、bootstrap admin
    middleware/               # Recovery、RequestID、AccessLog、JWT、角色
    router/                   # 登记 URL → handler
    handler/                  # 解析参数、调 service、写 JSON
    service/                  # 状态机、权限、事务（业务规则只放这里）
    store/                    # 只谈 SQL
    model/                    # 结构体，对应表或接口入参出参
    response/                 # 统一 JSON 包一层
    logger/                   # 日志，禁止打印密码和 JWT 原文
  sql/001_init.sql            # 建表，与 docs/02-database.md 一致
  docs/
  README.md
```

请求怎么走：

```text
HTTP → middleware（鉴权 / 打日志）
     → handler（绑定 JSON、取当前用户）
     → service（能不能看、能不能改状态、开事务）
     → store（SQL）
     → MySQL
```

分层约定：

| 层 | 负责 | 不负责 |
|---|---|---|
| handler | 读参数、写响应、HTTP 状态码 | 状态是否允许跳、开事务 |
| service | 角色数据范围、状态机、审计与工单一事务 | 拼 SQL |
| store | CRUD SQL | 业务 if/else（除「按角色拼 WHERE」这类查询条件） |

配置全部来自环境变量：HTTP 端口、MySQL DSN、JWT secret、bootstrap 管理员邮箱和密码。代码里不写死数据库密码。

---

## 2. 接口总览（11 个）

| # | 方法 | 路径 | 登录 | 谁能成功 | 作用 |
|---|---|---|---|---|---|
| 1 | GET | `/healthz` | 否 | 任何人 | 健康检查，含 MySQL ping |
| 2 | POST | `/api/v1/auth/register` | 否 | 任何人 | 注册，默认 `user` |
| 3 | POST | `/api/v1/auth/login` | 否 | 已注册用户 | 登录，返回 JWT |
| 4 | GET | `/api/v1/me` | 是 | 登录用户 | 当前用户信息 |
| 5 | POST | `/api/v1/tickets` | 是 | 登录用户 | 创建工单 |
| 6 | GET | `/api/v1/tickets` | 是 | 登录用户 | 列表（按角色过滤） |
| 7 | GET | `/api/v1/tickets/:id` | 是 | 对该单可见的人 | 详情（含评论、审计） |
| 8 | POST | `/api/v1/tickets/:id/assign` | 是 | 仅 `admin` | 指派 / 重新指派 |
| 9 | POST | `/api/v1/tickets/:id/actions` | 是 | 见状态机 | 开始处理 / 已解决 / 关闭 / 重开 |
| 10 | POST | `/api/v1/tickets/:id/comments` | 是 | 对该单可见的人 | 追加评论 |
| 11 | GET | `/api/v1/admin/users` | 是 | 仅 `admin` | 用户列表 |

登录方式：`Authorization: Bearer <token>`。

P1 若做「设为 agent」，再加 1 个接口即可，例如 `POST /api/v1/admin/users/:id/role`。第一期不要做。

---

## 3. 统一响应

成功：

```json
{
  "code": "OK",
  "message": "ok",
  "data": {}
}
```

失败时 `data` 为 `null`，`code` 为业务错误码。HTTP 状态码与错误码同时用，不靠 200 包一切错误。

分页列表的 `data`：

```json
{
  "items": [],
  "total": 0,
  "page": 1,
  "page_size": 20
}
```

`page` 从 1 开始，默认 1；`page_size` 默认 20，最大 50。

---

## 4. 公开接口

### 4.1 `GET /healthz`

作用：进程是否活着，MySQL 能否 ping。

不走 JWT。MySQL 不通时返回 503。

### 4.2 `POST /api/v1/auth/register`

作用：注册普通员工。

请求：

```json
{
  "email": "alice@example.com",
  "password": "password1",
  "display_name": "Alice"
}
```

规则：忽略客户端传来的 `role`；新用户固定 `user`。email 已存在返回 409 `EMAIL_TAKEN`。

`data`：用户信息（无密码哈希），第一期可不直接发 JWT，让客户端再调登录。

写表：`users`。

### 4.3 `POST /api/v1/auth/login`

作用：校验密码，发 JWT。

请求：

```json
{
  "email": "alice@example.com",
  "password": "password1"
}
```

`data` 含 `token` 和用户信息（id、email、display_name、role）。邮箱不存在或密码错误统一 401 `UNAUTHENTICATED`，不要区分「用户不存在」。

读表：`users`。不写会话表。

---

## 5. 登录后的工单接口

### 5.1 `GET /api/v1/me`

作用：用 token 换当前用户，方便前端或 curl 确认身份。

`data`：`id`、`email`、`display_name`、`role`。

### 5.2 `POST /api/v1/tickets`

作用：员工（以及 agent / admin 以员工身份）提单。

请求：

```json
{
  "title": "显示器不亮",
  "description": "开机后面板灯是暗的",
  "category": "hardware"
}
```

服务端写入：`status = open`，`creator_id = 当前用户`，`assignee_id = NULL`，并在同一事务插入 `audit_logs.action = create`。

写表：`tickets`、`audit_logs`。

### 5.3 `GET /api/v1/tickets`

作用：列表。同一条 URL，三种角色看到的集合不同。

查询参数：`status`、`category`、`q`（标题模糊）、`page`、`page_size`。`admin` 额外支持 `assignee_id`。

| 角色 | 默认看到 |
|---|---|
| `user` | 自己创建的 |
| `agent` | 自己创建的，或指派给自己的 |
| `admin` | 全部 |

读表：`tickets`（可 join `users` 带出创建人 / 处理人显示名）。

### 5.4 `GET /api/v1/tickets/:id`

作用：详情。不可见时 **404** `NOT_FOUND`。

`data` 含工单字段、创建人、处理人、评论列表、审计时间线。

读表：`tickets`、`ticket_comments`、`audit_logs`、`users`。

### 5.5 `POST /api/v1/tickets/:id/assign`

作用：管理员把工单派给某位 IT。

请求：

```json
{
  "assignee_id": 2
}
```

规则：

- 非 `admin` → 403 `PERMISSION_DENIED`
- 目标用户必须是 `agent`，否则 400 `INVALID_ARGUMENT`
- 当前状态须为 `open`、`assigned` 或 `in_progress`（重新指派），否则 409
- 成功后 `status = assigned`，写审计 `assign`

写表：`tickets`、`audit_logs`（同一事务）。

### 5.6 `POST /api/v1/tickets/:id/actions`

作用：按动作推进状态，而不是客户端直接改 `status`。

请求：

```json
{
  "action": "start"
}
```

`action` 取值：`start` / `resolve` / `close` / `reopen`。指派走上一节，不放在这里。

| action | 从 | 到 | 谁可以 |
|---|---|---|---|
| `start` | `assigned` | `in_progress` | 当前处理人或 admin |
| `resolve` | `in_progress` | `resolved` | 当前处理人或 admin |
| `close` | `resolved` | `closed` | 创建人或 admin |
| `reopen` | `resolved` | `open`（清空处理人） | 创建人或 admin |

非法跳转：409 `TICKET_INVALID_TRANSITION`。角色不对：403。看不见这张单：404。

写表：`tickets`、`audit_logs`（同一事务）。

### 5.7 `POST /api/v1/tickets/:id/comments`

作用：对可见工单留言。

请求：

```json
{
  "body": "已经换过电源线，还是不亮"
}
```

`closed` → 409 `TICKET_CLOSED`。看不见 → 404。

写表：`ticket_comments`。不写审计。

---

## 6. 管理员接口

### 6.1 `GET /api/v1/admin/users`

作用：查看系统里有哪些人，方便派单前确认谁是 `agent`。

非 admin → 403。分页参数同列表。

`data.items` 字段：`id`、`email`、`display_name`、`role`、`created_at`。禁止返回 `password_hash`。

读表：`users`。

---

## 7. 错误码

| HTTP | code | 何时 |
|---|---|---|
| 400 | `INVALID_ARGUMENT` | 参数校验失败（缺字段、非法 category、把 user 当 agent 派单等） |
| 401 | `UNAUTHENTICATED` | 未登录、token 无效、登录失败 |
| 403 | `PERMISSION_DENIED` | 已登录但角色不够（例如员工去指派） |
| 404 | `NOT_FOUND` | 资源不存在，或存在但对当前用户不可见 |
| 409 | `EMAIL_TAKEN` | 注册邮箱已存在 |
| 409 | `TICKET_INVALID_TRANSITION` | 非法状态流转 |
| 409 | `TICKET_CLOSED` | 已关闭仍评论或改状态 |
| 500 | `INTERNAL` | 未预期错误；日志记详细原因，响应不暴露 SQL |
| 503 | `UNAVAILABLE` | 健康检查发现 MySQL 不通 |

中间件：

- 全局：Recovery、RequestID（`X-Request-ID`）、AccessLog（方法、路径、状态码、耗时）
- `/api/v1` 下除 register/login 外：JWT
- `/api/v1/admin/*`：额外校验 `role == admin`

---

## 8. 接口与表的对应

| 接口 | 读 | 写 |
|---|---|---|
| 注册 | — | `users` |
| 登录 | `users` | — |
| 当前用户 | `users` | — |
| 创建工单 | — | `tickets` + `audit_logs` |
| 工单列表 | `tickets` | — |
| 工单详情 | `tickets` + `ticket_comments` + `audit_logs` + `users` | — |
| 指派 | `users`（校验 agent） | `tickets` + `audit_logs` |
| 状态动作 | `tickets` | `tickets` + `audit_logs` |
| 评论 | `tickets`（是否可见、是否关闭） | `ticket_comments` |
| 用户列表 | `users` | — |
| 健康检查 | ping | — |

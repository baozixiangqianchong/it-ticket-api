# it-ticket-api 文档

按主题拆开，不要把库表、接口、角色混在一份长文里。建议按下面顺序读。

| 顺序 | 文档 | 内容 |
|---|---|---|
| 0 | [三个角色各干什么](./00-roles-overview.md) | 角色总览图：一张单怎么走、三种人能做什么 |
| 1 | [需求与角色](./01-requirements-and-roles.md) | 做什么、状态机细则、第一期范围、验收 |
| 2 | [数据库](./02-database.md) | 库名、4 张表、字段、关系、索引 |
| 3 | [项目结构与接口](./03-architecture-and-api.md) | 目录分层、11 个接口、错误码 |

Cursor 里的 Ticket Roles 画布只是编辑器预览，文件在本机 `canvases/`，进不了这个仓库。仓库里以 `00-roles-overview.md` 为准，内容和画布相同。 |

第一期规模（读完应对得上这三个数）：

- **1 个库**：`it_ticket`
- **4 张表**：`users`、`tickets`、`ticket_comments`、`audit_logs`
- **11 个接口**：3 个公开 + 7 个登录后 + 1 个管理员

本项目是独立仓库，不改造 `gin-demo`。

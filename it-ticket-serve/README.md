# it-ticket-serve

企业内部 IT 报修工单后端。员工提交报修，管理员派给 IT，IT 处理完成后由员工确认关闭。

Go 模块名仍是 `it-ticket-api`。在本目录下编译和运行。

## 功能

- 账号：邀请码注册、登录（JWT）、查看当前用户、自己改显示名和密码
- 工单：创建、列表、详情；按角色限制可见范围
- 流转：管理员或 IT 指派 / 自领；IT 开始处理 / 标记已解决；员工关闭或重开
- 报修模板与常用回复：登录可见启用中的模板；IT 用常用回复；仅管理员可配置
- 评论与审计：可见工单可留言；状态变更写入审计日志；管理员可看工单 + 账号全局审计
- CORS：允许本机管理端直连

三种角色：`user`（员工）、`agent`（IT）、`admin`（管理员）。管理员也能接单。

## 技术

Go、Gin、MySQL。

## 目录

```text
it-ticket-serve/
├── cmd/server/          程序入口
├── internal/
│   ├── config/          配置
│   ├── router/          路由
│   ├── handler/         接口处理
│   ├── service/         业务逻辑
│   ├── store/           数据库访问
│   ├── middleware/      中间件
│   ├── model/           数据模型
│   ├── response/        统一响应
│   └── logger/          日志
├── sql/                 数据库脚本
└── docs/                需求与接口文档
```

## 运行

先执行 `sql/001_init.sql` 建库建表。已经跑过一期库的，再按顺序执行 `sql/002_v2.sql`、`sql/003_p1.sql`、`sql/004_notify.sql`、`sql/005_activity.sql`、`sql/006_invite.sql`（邀请码）、`sql/007_catalog.sql`（提单模板和常用回复）。然后：

```bash
go run ./cmd/server
```

默认 `http://127.0.0.1:8080`。健康检查：`GET /healthz`。

配置来自环境变量：`HTTP_ADDR`、`MYSQL_*` / `MYSQL_DSN`、`JWT_SECRET`。未设置时默认连本机 MySQL 的 `it_ticket` 库。

文档从 [docs/README.md](./docs/README.md) 读起。

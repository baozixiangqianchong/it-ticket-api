# IT 工单系统

企业内部报修工单：员工提单，管理员派给 IT，处理完成后由员工关闭。本仓库同时包含后台 API 和 Web 管理端。

| 目录 | 说明 |
|---|---|
| [it-ticket-serve](./it-ticket-serve) | Go + Gin + MySQL，HTTP 接口 |
| [it-ticket-web](./it-ticket-web) | React + Vite + Ant Design，管理端页面 |

三种角色：`user`（员工）、`agent`（IT）、`admin`（管理员）。同一套工单接口按角色过滤可见范围。

## 目录

```text
.
├── it-ticket-serve/      后台 API
│   ├── cmd/server/       进程入口
│   ├── internal/         业务代码
│   ├── sql/              建表脚本
│   └── docs/             需求、库表、接口说明
├── it-ticket-web/        前端管理端
└── README.md             本文件
```

## 运行

先准备 MySQL，执行 `it-ticket-serve/sql/001_init.sql` 建库建表。

后台：

```bash
cd it-ticket-serve
go run ./cmd/server
```

默认 `http://127.0.0.1:8080`。健康检查：`GET /healthz`。

前端：

```bash
cd it-ticket-web
npm install
npm run dev
```

浏览器打开 `http://127.0.0.1:5173`。页面直连 `http://127.0.0.1:8080`。

更细的接口和表结构见 [it-ticket-serve/docs](./it-ticket-serve/docs)。页面说明见 [it-ticket-web/README.md](./it-ticket-web/README.md)。

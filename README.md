# it-ticket-api

企业内部 IT 报修工单后端 API。员工提交报修，管理员派给 IT，IT 处理完成后由员工确认关闭。

第一期只做 HTTP 接口，不包含前端页面。

## 功能

- 账号：注册、登录（JWT）、查看当前用户
- 工单：创建、列表、详情；按角色限制可见范围
- 流转：管理员指派；IT 开始处理 / 标记已解决；员工关闭或重开
- 评论与审计：可见工单可留言；状态变更写入审计日志

三种角色：`user`（员工）、`agent`（IT）、`admin`（管理员）。

## 技术

Go、Gin、MySQL。

## 目录

```text
it-ticket-api/
├── cmd/
│   └── server/          程序入口
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

```bash
go run ./cmd/server
```

默认 `http://127.0.0.1:8080`。健康检查：`GET /healthz`。

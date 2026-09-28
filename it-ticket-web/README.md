# IT 工单管理端

后台工单 API 的 Web 管理端。登录后按当前角色看工单、提单、跟进状态、留言；管理员还可以指派处理人、改用户角色、发邀请码、配置提单模板。

## 技术

React 19、TypeScript、Vite 8、React Router 7、Ant Design 6。

开发时页面直连本机后台 `http://127.0.0.1:8080`，请求头带 `Authorization: Bearer <token>`。

## 页面

| 路径 | 谁能进 | 作用 |
|---|---|---|
| `/login`、`/register` | 未登录 | 登录；注册必须填管理员发的邀请码（新账号默认员工） |
| `/tickets` | 登录用户 | 工单列表。IT / 管理员有待领、待我处理、我提交的；支持筛选、分页、时间范围、描述/评论搜索、最新评论预览 |
| `/tickets/new` | 登录用户 | 新建工单；模板来自后台启用中的配置 |
| `/notifications` | 登录用户 | 通知中心，分页、未读筛选 |
| `/tickets/:id` | 对该单可见的人 | 详情、后台下发的动作、领取 / 指派 / 转派、等用户、评论、处理记录 |
| `/users` | 仅管理员 | 用户列表、邀请码、改角色、停用账号 |
| `/catalog` | 仅管理员 | 提单模板和常用回复 |
| `/profile` | 登录用户 | 自己改显示名和密码 |
| `/audits` | 仅管理员 | 工单动作与账号变更总览 |
| `/board` | 仅管理员 | 工作台 |

按钮以后台返回的 `available_actions` 为准，页面不直接改 `status`。重开和管理员代关要填原因。顶栏有站内通知。不上传附件。

## 目录

```text
it-ticket-web/
├── src/
│   ├── api/             请求封装和类型，基址在 client.ts
│   ├── auth/            登录状态、路由守卫
│   ├── components/      状态 / 角色标签
│   ├── layout/          登录页外壳、后台侧栏
│   ├── pages/           登录、注册、工单、用户
│   ├── lib/labels.ts    中文文案
│   ├── theme.tsx        Ant Design 主题
│   └── App.tsx          路由
├── index.html
├── package.json
└── vite.config.ts
```

## 运行

先在仓库的 `it-ticket-serve/` 目录启动 API，再：

```bash
npm install
npm run dev
```

浏览器打开 `http://127.0.0.1:5173`。

其它命令：`npm run build` 打包，`npm run preview` 预览产物，`npm run lint` 检查。

`node_modules`、`dist` 等依赖和构建产物不提交。

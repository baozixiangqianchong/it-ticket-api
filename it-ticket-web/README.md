# IT 工单管理端

后台工单 API 的 Web 管理端。登录后按当前角色看工单、提单、跟进状态、留言；管理员还可以指派处理人和改用户角色。

## 技术

React 19、TypeScript、Vite 8、React Router 7、Ant Design 6。

开发时页面直连本机后台 `http://127.0.0.1:8080`，请求头带 `Authorization: Bearer <token>`。

## 页面

| 路径 | 谁能进 | 作用 |
|---|---|---|
| `/login`、`/register` | 未登录 | 登录 / 注册（新账号默认员工） |
| `/tickets` | 登录用户 | 工单列表（可见范围由角色决定） |
| `/tickets/new` | 登录用户 | 新建工单 |
| `/tickets/:id` | 对该单可见的人 | 详情、状态动作、评论；管理员可指派 |
| `/users` | 仅管理员 | 用户列表、改角色 |

状态动作：`start` / `resolve` / `close` / `reopen`，和后台状态机一致，页面不直接改 `status`。

## 目录

```text
it-ticket-web/
├── src/
│   ├── api/             请求封装和类型，基址在 client.ts
│   ├── auth/            登录状态、路由守卫
│   ├── components/      状态 / 角色标签
│   ├── hooks/           用户显示名
│   ├── layout/          登录页外壳、后台侧栏
│   ├── pages/           登录、注册、工单、用户
│   ├── lib/labels.ts    中文文案、可执行动作
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

`node_modules`、`dist` 等依赖和构建产物不提交。改角色后对方需要重新登录，JWT 里的角色才会更新。

# Web 模块设计文档

## 概述

为 KnowSync 项目新增 `web` 前端模块，基于 Vite + React + TypeScript + Tailwind CSS，使用 Caddy 作为反向代理和静态文件服务器，Docker 部署。

## 架构

```
[浏览器] ←→ Caddy(:80/:443)
                ├── /*           → SPA 静态文件 (dist/)
                ├── /api/v1/*    → gateway:8080 (内部网络)
                └── /files/*     → gateway:8080 (内部网络)
```

- Caddy 容器作为唯一外部入口，暴露 80/443 端口
- gateway/ai-server 等后端服务不暴露端口到宿主机，仅通过 Docker 内部网络通信
- 前端代码通过同源路径 `/api/v1/*` 请求后端，由 Caddy 反向代理

## 技术栈

| 层面 | 选择 |
|------|------|
| 框架 | React 19 + TypeScript |
| 构建工具 | Vite |
| 样式 | Tailwind CSS |
| 路由 | react-router-dom v7 |
| HTTP 客户端 | 原生 fetch |
| Markdown 编辑器 | Monaco Editor (VS Code 内核) |
| 流式渲染 | react-markdown + react-syntax-highlighter |
| 头像裁切 | Ant Design Upload + ImageCropper (仅此功能局部引入) |

## 目录结构

```
web/
├── Taskfile.yml              # 构建命令
├── scripts/
│   ├── build.py              # 前端编译脚本 (npm run build)
│   └── build_image.py        # Docker 镜像构建脚本
├── Dockerfile                # Caddy 镜像
├── Caddyfile                 # Caddy 配置
├── package.json
├── vite.config.ts
├── tsconfig.json
├── tailwind.config.ts
├── index.html
└── src/
    ├── main.tsx               # 入口
    ├── App.tsx                # 路由定义
    ├── api/                   # HTTP 请求封装
    │   ├── client.ts          # fetch 封装 + token 拦截
    │   ├── auth.ts            # 认证接口 (登录/注册/刷新)
    │   ├── repos.ts           # 知识库 CRUD 接口
    │   ├── nodes.ts           # 文章节点 CRUD + 上传
    │   └── chat.ts            # AI 对话 + SSE 流
    ├── components/            # 公共组件
    │   ├── Layout/
    │   │   ├── PublicLayout.tsx      # 未登录布局（Login/Register）
    │   │   ├── AppLayout.tsx         # 登录后主框架（顶部栏 + 侧边栏 + 内容）
    │   │   ├── TopBar.tsx            # 顶部栏：logo、搜索框、用户信息
    │   │   ├── Sidebar.tsx           # 左侧仓库列表（可折叠）
    │   │   └── ChatLayout.tsx        # 聊天三栏布局（左侧+中间+右侧）
    │   ├── Chat/
    │   │   ├── ChatWindow.tsx        # 中间聊天窗口（流式消息展示）
    │   │   ├── MessageBubble.tsx     # 单条消息气泡
    │   │   ├── Streamdown.tsx        # 流式 Markdown 渲染组件
    │   │   ├── ChatHistory.tsx       # 右侧会话消息记录
    │   │   └── SessionList.tsx       # 会话列表侧栏
    │   ├── Repo/
    │   │   ├── RepoTree.tsx          # 左侧文章树（文件夹+文章）
    │   │   ├── RepoCard.tsx          # 知识库卡片
    │   │   └── CollabList.tsx        # 协作者列表
    │   ├── Editor/
    │   │   └── MarkdownEditor.tsx    # Monaco Editor 封装
    │   ├── Settings/
    │   │   └── SettingsModal.tsx     # 设置弹窗（含头像裁切）
    │   └── ui/                    # 通用 UI 组件（按钮、输入框、弹窗等）
    ├── pages/                 # 页面组件
    │   ├── Login.tsx               # 登录页
    │   ├── Register.tsx            # 注册页
    │   ├── RepoList.tsx            # 知识库列表页（两栏布局）
    │   ├── RepoDetail.tsx          # 仓库详情页（左侧文章树 + 右侧内容）
    │   ├── ArticleView.tsx         # 文章阅读（渲染后的 Markdown）
    │   ├── ArticleEditor.tsx       # 文章编辑（Monaco Editor）
    │   ├── Chat.tsx                # AI 对话页（三栏布局）
    │   ├── SearchResult.tsx        # 语义搜索结果页
    │   └── NotFound.tsx            # 404 页面
    ├── hooks/                 # 自定义 Hooks
    ├── store/                 # 状态管理 (React Context)
    └── styles/
        └── index.css          # Tailwind 入口 + 全局样式
```

## 路由设计

| 路径 | 页面 | 访问控制 |
|------|------|---------|
| `/login` | 登录 | 仅未登录 |
| `/register` | 注册 | 仅未登录 |
| `/` | 判断 token → 有效则跳 `/chat`，无效则跳 `/login` | 自动 |
| `/repos` | 知识库列表（两栏） | 需登录 |
| `/repos/:repoId` | 仓库详情（左侧文章树 + 右侧内容） | 需登录 |
| `/repos/:repoId/nodes/:nodeId` | 文章阅读（默认） | 需登录 |
| `/repos/:repoId/nodes/:nodeId/edit` | 文章编辑（Monaco Editor） | 需登录 |
| `/chat` | AI 对话（默认首页，三栏） | 需登录 |
| `/chat/:sessionId` | AI 对话（指定会话，三栏） | 需登录 |
| `/search` | 语义搜索结果页 | 需登录 |

设置页为弹窗（Modal）形式，从顶部栏用户头像位置触发，不占用路由。

## 页面布局

### 首页（聊天页）— 三栏布局

```
┌──────────────────────────────────────────────────────┐
│  TopBar: Logo(→首页) | 搜索框(回车→/search) | 头像▼    │
├──────────┬───────────────────────────┬────────────────┤
│  左侧     │     中间                  │  右侧           │
│  [AI对话] │     聊天窗口               │  会话列表       │
│  ─────── │     Streamdown 流式渲染     │  所有历史会话    │
│  仓库列表  │     + 输入框 + 发送按钮     │  点击切换会话    │
│  (可折叠)  │                           │  (可折叠)       │
└──────────┴───────────────────────────┴────────────────┘
```

- 左侧上部有"AI 对话"图标入口（当前页高亮）
- 左侧下部为仓库列表，点击仓库跳转到 `/repos/:repoId`
- 右侧为会话列表，显示所有历史会话标题，点击切换中心聊天内容

### 仓库详情（知识库内部）— 两栏布局

```
┌──────────────────────────────────────────────────────┐
│  TopBar: Logo(→首页) | 搜索框(回车→/search) | 头像▼    │
├────────────────────┬─────────────────────────────────┤
│  左侧              │  右侧                            │
│  ← 返回所有知识库   │  选中文章 → 阅读模式（默认）       │
│  ───────           │   顶部"编辑"→ 编辑模式            │
│  文件树             │   Monaco Editor (编辑/预览切换)    │
│  (文件夹可展开)      │  或空白 → 新建空白文章            │
│  5 层限制           │  或上传 .md 文件                 │
│  右键菜单/拖拽       │  协作者管理列表                   │
└────────────────────┴─────────────────────────────────┘
```

- 左侧顶部有"← 返回所有知识库"按钮，点击回到首页（仓库列表恢复）
- 左侧为文件树（代替了首页的仓库列表）
- 点击左上角 Logo 回到首页（聊天页）

### 设置 — 弹窗 Modal

从 TopBar 用户头像点击触发，覆盖在当前内容之上：
- 用户信息编辑（名称、邮箱）
- 头像上传 + Ant Design ImageCropper 裁切
- 修改密码

## 设计风格

- **整体**：清新、简约、护眼、柔和
- **配色**：柔和的灰白底色，低饱和度绿色系主色调（与 KnowSync 品牌呼应），文字用深灰而非纯黑
- **字体**：系统默认无衬线字体（`-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`）
- **卡片/区块**：轻微阴影 + 圆角，大面积留白
- **侧边栏**：左侧可折叠（点击按钮变为图标或收起），右侧消息记录也可折叠

### 搜索 — 独立全屏页

```
┌──────────────────────────────────────────────────────┐
│  TopBar: Logo(→首页) | 搜索框(回车搜索) | 头像▼        │
├──────────────────────────────────────────────────────┤
│  搜索结果 (路由: /search?q=关键词)                    │
│  ┌──────────────────────────────────────────────┐   │
│  │ 匹配的知识库卡片列表（repo_ids → repo 详情）   │   │
│  │ 每个卡片显示：仓库名、匹配摘要                 │   │
│  │ 分页: has_more + total_pages                 │   │
│  └──────────────────────────────────────────────┘   │
│  无侧边栏，全屏展示结果                              │
└─────────────────────────────────────────────────────┘
```

## 交互流程

```
用户访问 knowsync.local
  ├─ localStorage 有 token → 请求任意接口验证
  │    ├─ 有效 → 进入首页 (/chat，三栏)
  │    └─ 401 → 尝试 refresh_token → 失败则跳 /login
  └─ 无 token → 跳转登录页 (/login)

=== 首页（/chat）三栏布局 ===
  TopBar:    Logo(点击→首页) | 搜索框(回车→/search) | 头像▼(设置/退出)
  左侧栏:    [AI 对话] 图标（当前页高亮）
             ─────────
             仓库列表（可折叠）
             点击任一仓库 → 跳转 /repos/:repoId
  中间区:    聊天窗口
             空状态→ 显示欢迎语 + 示例问题
             有对话→ 显示消息流（Streamdown）
             底部→ 多行输入框（Enter发送，Shift+Enter换行）
             收到 ask_user 事件→ 弹出 Modal 展示问题/选项
  右侧栏:    会话列表（从 GET /ai/sessions 获取）
             显示所有历史会话标题，点击切换
             可折叠收起

=== 仓库内部（/repos/:repoId）两栏布局 ===
  左侧栏:    ← 返回所有知识库（点击回首页）
             ─────────
             文件树（可展开/收起，5层限制）
             操作: 右键菜单/拖拽
             新建文件夹 / 新建文章 / 重命名 / 移动到 / 删除
  右侧区:    无选中→ 空白提示
             选中文章→ 阅读模式（/nodes/:nodeId）
             点"编辑"按钮→ 编辑模式（/nodes/:nodeId/edit）
             协作者管理列表

=== 文章阅读/编辑 ===
  文章阅读（/repos/:repoId/nodes/:nodeId）:
    渲染 Markdown，顶部有"编辑"按钮
  文章编辑（/repos/:repoId/nodes/:nodeId/edit）:
    Monaco Editor，编辑/预览切换
    保存→ PUT /content 上传 .md 文件
    也支持从本地选择 .md 文件上传

=== 搜索页（/search?q=xxx）全屏 ===
  顶部栏保留，无侧边栏
  展示语义搜索结果列表
  分页浏览

=== 头像下拉菜单 ===
  点击用户头像→ 下拉菜单:
    ├─ 设置（打开设置 Modal）
    └─ 退出登录（清除 token，跳 /login）

=== 设置 Modal ===
  用户信息编辑（名称、邮箱）
  头像上传（Ant Design ImageCropper 裁切）
  修改密码
  底部: 退出登录
```

## 状态管理

使用 React Context + useReducer，不引入额外状态管理库。

- **AuthContext** — 用户认证状态、token 存储、登录/登出、token 刷新
- **ChatContext** — 当前对话状态、会话列表、流式消息缓冲
- **RepoContext** — 仓库列表缓存、当前选中的仓库/节点

Token 存储在 `localStorage`，fetch 封装中自动附加 `Authorization` header，401 时尝试用 refresh_token 刷新，刷新失败则清除 token 并跳转登录。

## 关键功能设计

### Streamdown（流式 Markdown 渲染）

- 后端通过 SSE 推送事件流，事件类型：
  - `event: thinking` + `data: { session_id, content }` — 思考过程片段
  - `event: thinking_finished` + `data: { session_id }` — 思考结束
  - `event: content` + `data: { session_id, content }` — 回答片段
  - `event: ask_user` + `data: { session_id, question, type, options, has_other }` — 反问用户
  - `event: done` + `data: { session_id, message_id, title, title_updated }` — 对话结束
  - `event: error` + `data: { session_id, message }` — 流错误
- 前端维护消息缓冲区，每次收到 chunk 追加后整体重新渲染
- 使用 `react-markdown` + `react-syntax-highlighter` 渲染
- 渲染时处理半截代码块、表格等边缘情况
- 右侧消息记录显示每条消息的摘要（截取前几句）

### 文章编辑

- 默认进入阅读模式，顶部"编辑"按钮切换到编辑模式
- 编辑模式使用 Monaco Editor (VS Code 内核)
- 支持语法高亮、行号、自动缩进
- 编辑/预览切换按钮
- 保存 → 上传 Markdown 文件到后端 `PUT /repos/:repo_id/nodes/:node_id/content`
- 支持从本地选取 `.md` / `.markdown` 文件上传作为文章内容
- 文件格式校验（仅允许 `.md` / `.markdown`）

### AI 反问交互

- AI 反问（`ask_user` 事件）以弹窗（Modal）形式展示
- Modal 展示问题和选项（单选 single / 多选 multiple）
- 用户选择/填写后自动发送回复到对话中
- `has_other` 为 true 时提供自由输入框

### 文件上传

- 仓库内图片作为节点上传（需后端支持非 Markdown 文件格式）
- 头像上传：点击头像 → 选择图片 → Ant Design ImageCropper 裁切 → 确认 → `PUT /users/:user_id/avatar`
- 文件格式校验（图片仅允许 jpg/png/webp，最大 2MB）

### 文章树

- 左侧树组件，文件夹可展开/收起
- 嵌套深度限制 5 层
- 右键菜单操作：新建文件夹、新建文章、重命名、移动到、删除
- 支持拖拽节点到目标文件夹（移动操作）
- 新建：支持新建空白节点（然后编辑内容）和从本地上传 .md 文件两种方式

## 后端待补充功能

以下前端功能当前后端 API 尚未完全支持，需要后端补充：

| 功能 | 当前状态 | 需要补充 |
|------|---------|---------|
| 图片上传为仓库节点 | `PUT /repos/:repo_id/nodes/:node_id/content` 仅接受 `.md`/`.markdown` | 后端需放宽/增加图片文件格式支持，或新增图片上传接口 |
| 全局语义搜索 | `POST /ai/search` 仅搜索公开仓库 | 后端需支持搜索用户有权限的私有仓库 |
| 搜索结果详情 | `POST /ai/search` 返回 `repo_ids`（仅知识库 ID 列表） | 后端需返回更详细结果（匹配文章片段、页码等），或前端根据 repo_ids 查详情后组合 |
| 空文章创建 | 无"创建空白节点"的独立 API | `POST /repos/:repo_id/nodes` 创建节点后，前端可进入编辑模式，保存时上传内容即可，不需要额外 API |

## 构建流水线

参考已有服务（user-server、gateway 等）的模式：

| Task 命令 | 功能 |
|-----------|------|
| `task build` | 编译 React 前端 (npm run build → dist/) |
| `task build:image` | 编译 → 构建 Docker 镜像 |
| `task dev` | 启动 Vite 开发服务器 |
| `task clean` | 清除 dist/ |
| `task help` | 显示帮助 |

构建镜像流程：
1. `scripts/build.py` — 执行 `npm run build`
2. `scripts/build_image.py` — 先调用 build.py，再 `docker build`
3. Dockerfile 使用 `caddy:alpine` 基础镜像，复制 `dist/` 和 `Caddyfile`

## Caddy 配置

```
knowsync.local {
    root * /usr/share/caddy
    encode gzip

    # SPA fallback — 所有非文件请求返回 index.html
    try_files {path} /index.html

    # API 反向代理到 gateway
    reverse_proxy /api/* gateway:8080
    reverse_proxy /files/* gateway:8080
}
```

## 后端 API 接口映射

前端 API 请求路径均经过 Caddy 反向代理到 gateway，所有路径以 `/api/v1` 为前缀。

详细接口定义参见：
- `gateway/docs/user.jsonc` — 用户认证、用户信息、密码管理
- `gateway/docs/repo.jsonc` — 知识库 CRUD、节点管理、协作者管理、文章上传
- `gateway/docs/ai.jsonc` — AI 对话（SSE 流式）、语义搜索、会话管理

### 页面 → API 映射

| 页面 | 调用的 API |
|------|-----------|
| Login | `POST /auth/login` |
| Register | `POST /auth/register`、`POST /verify-codes` |
| RepoList | `GET /repos`（列表）、`POST /repos`（创建） |
|  | `PUT /repos/:repo_id`（重命名/切换可见性） |
|  | `DELETE /repos/:repo_id`（删除） |
| RepoDetail | `GET /repos/:repo_id`（详情）、`GET /repos/:repo_id/nodes`（节点树） |
|  | `POST /repos/:repo_id/nodes`（创建文件夹/文章） |
|  | `PUT /repos/:repo_id/nodes/:node_id`（重命名/移动） |
|  | `DELETE /repos/:repo_id/nodes/:node_id`（删除） |
|  | `GET /repos/:repo_id/collaborators`（协作者列表） |
|  | `POST /repos/:repo_id/collaborators`（添加协作者） |
|  | `PUT /repos/:repo_id/collaborators/:user_id`（修改角色） |
|  | `DELETE /repos/:repo_id/collaborators/:user_id`（移除） |
| ArticleView | `GET /repos/:repo_id/nodes/:node_id/signed-url`（获取文章链接） |
| ArticleEditor | `PUT /repos/:repo_id/nodes/:node_id/content`（上传文章） |
| Chat | `POST /ai/chat`（SSE 流式）、`GET /ai/sessions`（会话列表） |
|  | `GET /ai/sessions/:session_id/messages`（消息历史） |
|  | `DELETE /ai/sessions/:session_id`（删除会话） |
| Search | `POST /ai/search`（语义搜索） |
| Settings | `GET /users/:user_id`、`PUT /users/:user_id` |
|  | `PUT /users/:user_id/avatar`（头像上传裁切） |
|  | `PUT /users/:user_id/password`（修改密码） |

### 认证方式

所有需登录的接口在 Header 中携带 `Authorization: Bearer {access_token}`。
access_token 过期后使用 `POST /auth/refresh` 获取新的令牌对。

## Docker 部署

docker-compose.yaml 新增 web 服务：

```yaml
web:
  image: knowsync/web:latest
  container_name: knowsync-web
  restart: unless-stopped
  ports:
    - "80:80"
  depends_on:
    gateway:
      condition: service_started
  networks:
    - knowsync-net
```

gateway 不再需要在 `ports` 中暴露 8080。

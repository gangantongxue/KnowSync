# 社交聊天功能设计文档

> 创建日期: 2026-06-04

## 一、概述

在现有 KnowSync 知识库管理平台中，新增用户间即时通讯（IM）功能，支持好友管理、一对一私聊和群聊。作为独立功能模块，与现有的 AI 对话（`/chat`）完全分离。

### 核心决策

| 维度 | 决策 |
|------|------|
| 前端页面 | 独立新页面 `/messages` |
| 实时通信 | WebSocket 长连接（混合模式：HTTP CRUD + WS 推送） |
| 后端架构 | 新增 `chat-server` 微服务，通过 gRPC 与 Gateway 通信 |
| 好友模型 | 双向确认制（申请 → 同意 → 好友） |
| 消息类型 | 文本/Markdown + 图片 + 文件 |
| 消息存储 | MySQL 持久化，离线消息重连时推送 |
| 附件存储 | 复用现有 storage 模块（S3/OSS） |
| 交付范围 | 一次性全部实现 |

## 二、架构设计

```
浏览器 (React)
  ├── HTTP REST ──────►  Gateway (Hertz) ─── gRPC ───►  chat-server ─── MySQL
  ├── WebSocket ──────►  Gateway (WS 升级) ── gRPC ───►  chat-server (WS 连接管理)
  └── 文件上传 ──────►  Gateway                   ───►  Storage (S3/OSS)
```

- **Gateway**：新增 `/api/v1/friends`、`/api/v1/messages`、`/api/v1/groups` REST 路由，以及 `/ws` WebSocket 升级端点；所有 API 经 `middleware.Auth` 鉴权
- **chat-server**：独立 Go 服务，负责好友关系、消息持久化、群组管理、WebSocket 连接管理；定义 gRPC 接口供 Gateway 调用
- **go.work**：新增 `./chat-server` 模块
- **前端**：新增 `pages/Messages.tsx`，`AppLayout` 下注册路由；TopBar 新增"消息"导航按钮

## 三、数据模型（MySQL）

### 好友相关

```sql
-- 好友申请表
friend_requests (
  id          CHAR(20) PRIMARY KEY,              -- xid 生成
  sender_id   VARCHAR(20) NOT NULL,              -- 用户 ID（纯数字字符串）
  receiver_id VARCHAR(20) NOT NULL,
  status      ENUM('pending', 'accepted', 'rejected') NOT NULL DEFAULT 'pending',
  remark      VARCHAR(100) DEFAULT '',
  created_at  BIGINT NOT NULL,
  updated_at  BIGINT NOT NULL,
  INDEX idx_receiver_status (receiver_id, status),
  INDEX idx_sender (sender_id)
);

-- 好友关系表（双向各存一条，user_id 和 friend_id 互为好友）
friends (
  id                CHAR(20) PRIMARY KEY,        -- xid 生成
  user_id           VARCHAR(20) NOT NULL,
  friend_id         VARCHAR(20) NOT NULL,
  remark            VARCHAR(100) DEFAULT '',
  last_message_at   BIGINT NOT NULL DEFAULT 0,
  last_read_seq_id  BIGINT UNSIGNED NOT NULL DEFAULT 0,
  pinned            TINYINT(1) NOT NULL DEFAULT 0,
  created_at        BIGINT NOT NULL,
  UNIQUE KEY uk_user_friend (user_id, friend_id)
);

-- 好友关系表（双向各存一条，user_id 和 friend_id 互为好友）
friends (
  id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id           BIGINT UNSIGNED NOT NULL,
  friend_id         BIGINT UNSIGNED NOT NULL,
  remark            VARCHAR(100) DEFAULT '',
  last_message_at   BIGINT NOT NULL DEFAULT 0,
  last_read_seq_id  BIGINT UNSIGNED NOT NULL DEFAULT 0,  -- 该用户在此私聊中已读的最后一条消息序号
  pinned            TINYINT(1) NOT NULL DEFAULT 0,       -- 是否置顶
  created_at        BIGINT NOT NULL,
  UNIQUE KEY uk_user_friend (user_id, friend_id)
);
```

### 消息表

```sql
-- 消息表（私聊 + 群聊共用）
messages (
  id                CHAR(20) PRIMARY KEY,        -- xid 生成
  conversation_type ENUM('private', 'group') NOT NULL,
  conversation_id   VARCHAR(64) NOT NULL,   -- 私聊: "u1_u2"(排序后的用户ID拼接); 群聊: group_id
  seq_id            BIGINT UNSIGNED NOT NULL,  -- 会话内递增序号
  sender_id         VARCHAR(20) NOT NULL,     -- 用户 ID（纯数字字符串）
  content_type      ENUM('text', 'image', 'file', 'system_invitation') NOT NULL DEFAULT 'text',
  content           TEXT NOT NULL,             -- 文本内容或文件描述
  extra             JSON DEFAULT NULL,         -- { "file_url": "...", "file_name": "...", "file_size": 123, "mentions": [...] }
  reply_to_id       CHAR(20) DEFAULT NULL,     -- 引用消息 ID
  status            ENUM('normal', 'recalled') NOT NULL DEFAULT 'normal',
  created_at        BIGINT NOT NULL,
  INDEX idx_conversation (conversation_id, seq_id),
  INDEX idx_sender (sender_id)
);
```

- `conversation_id`：私聊时为双方 user_id 排序后以 `_` 拼接（如 `"1001_1002"`）；群聊时即为 `group_id` 的字符串形式
- `seq_id`：每个会话内自增序号，由 chat-server 保证严格递增，用于消息排序
- `reply_to_id`：该消息所回复的原消息 ID，为 NULL 表示普通消息
- 支持游标分页：`?cursor=<created_at>&limit=20`，以 `created_at` 时间戳为游标向前翻页

**会话列表排序规则**：前端会话列表混合展示好友私聊和群聊，按各会话的最后一条消息时间降序排列。私聊取 `friends.last_message_at`（每次发消息时更新），群聊取 `messages` 表中该 `conversation_id` 的最新 `created_at`。

### 群聊相关

```sql
-- 群组表
groups (
  id         CHAR(20) PRIMARY KEY,              -- xid 生成
  name       VARCHAR(100) NOT NULL,
  avatar     VARCHAR(500) DEFAULT '',
  owner_id   VARCHAR(20) NOT NULL,              -- 用户 ID（纯数字字符串）
  created_at BIGINT NOT NULL,
  updated_at BIGINT NOT NULL
);

-- 群成员表
group_members (
  id               CHAR(20) PRIMARY KEY,        -- xid 生成
  group_id         CHAR(20) NOT NULL,
  user_id          VARCHAR(20) NOT NULL,
  role             ENUM('owner', 'admin', 'member') NOT NULL DEFAULT 'member',
  last_read_seq_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  pinned           TINYINT(1) NOT NULL DEFAULT 0,
  joined_at        BIGINT NOT NULL,
  UNIQUE KEY uk_group_user (group_id, user_id),
  INDEX idx_user (user_id)
);

-- 群成员表
group_members (
  id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  group_id         BIGINT UNSIGNED NOT NULL,
  user_id          BIGINT UNSIGNED NOT NULL,
  role             ENUM('owner', 'admin', 'member') NOT NULL DEFAULT 'member',
  last_read_seq_id BIGINT UNSIGNED NOT NULL DEFAULT 0,  -- 已读的最后一条消息序号
  pinned           TINYINT(1) NOT NULL DEFAULT 0,       -- 是否置顶
  joined_at        BIGINT NOT NULL,
  UNIQUE KEY uk_group_user (group_id, user_id),
  INDEX idx_user (user_id)
);
```

### 在线状态

在线状态使用 Redis 管理，不存储到 MySQL：
- WebSocket 连接建立时，写入 Redis `user_online:{user_id}`，TTL 60 秒
- chat-server 定期（30s）续期 TTL，WS 断开时主动删除
- 查询好友列表时，批量从 Redis 查询各好友在线状态
- `user_server` 增加 `last_online_at` 字段记录最后在线时间，用户离线时更新

## 四、API 设计

### 4.1 好友管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/friends/requests` | 发送好友申请 |
| GET | `/api/v1/friends/requests` | 我的申请列表（收/发） |
| PUT | `/api/v1/friends/requests/:id` | 处理申请（accept/reject） |
| GET | `/api/v1/friends` | 好友列表（`?q=` 按备注搜索，按 `last_message_at DESC` 排序） |
| DELETE | `/api/v1/friends/:user_id` | 删除好友 |
| PUT | `/api/v1/friends/:user_id/remark` | 修改好友备注 |
| GET | `/api/v1/users/search?q=` | 搜索用户（加好友时搜索） |

### 4.2 消息

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/messages/:conversation_id?cursor=&limit=20` | 游标分页获取历史消息 |
| GET | `/api/v1/messages/:conversation_id/search?q=` | 会话内搜索历史消息（MySQL LIKE 匹配） |
| POST | `/api/v1/messages/send` | 发送消息（HTTP 兜底） |
| POST | `/api/v1/messages/upload` | 上传聊天图片/文件 |
| DELETE | `/api/v1/messages/:id` | 撤回消息（仅发送者本人，且 `now - created_at <= 120s`，服务端校验） |

### 4.3 会话管理

| 方法 | 路径 | 说明 |
|------|------|------|
| PUT | `/api/v1/conversations/:id/pin` | 置顶/取消置顶会话（`{ "pinned": true/false }`） |
| GET | `/api/v1/online-status/:user_id` | 查询指定用户的在线状态（用于展示好友在线状态） |

### 4.4 群聊管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/groups` | 创建群聊 |
| GET | `/api/v1/groups` | 我的群列表 |
| GET | `/api/v1/groups/:id/members` | 获取群成员 |
| POST | `/api/v1/groups/:id/members` | 添加成员 |
| DELETE | `/api/v1/groups/:id/members/:uid` | 移除成员 |
| DELETE | `/api/v1/groups/:id/leave` | 退出群聊 |
| DELETE | `/api/v1/groups/:id` | 解散群（群主） |
| PUT | `/api/v1/groups/:id/members/:uid/role` | 修改成员角色 |

### 4.5 WebSocket

端点：`GET /ws?token=<jwt_token>`（JWT 鉴权，在 URL 查询参数中传递 token）

服务端 → 客户端推送消息协议：

```json
{ "type": "new_message",       "data": { /* 完整消息对象 */ } }
{ "type": "message_recalled",  "data": { "message_id": "...", "conversation_id": "..." } }
{ "type": "friend_request",    "data": { "request_id": "...", "sender": { ... } } }
{ "type": "friend_accepted",   "data": { "user_id": "..." } }
```

消息发送优先通过 HTTPS API（`POST /api/v1/messages/send`），发送成功后服务端通过 WebSocket 向会话内其他在线用户推送。WebSocket 连接断开重连后，客户端重新拉取各会话最新消息。

## 五、群聊权限模型

| 操作 | 群主(owner) | 管理员(admin) | 普通成员(member) |
|------|:-----------:|:------------:|:---------------:|
| 发送消息 | ✅ | ✅ | ✅ |
| 添加成员 | ✅ | ✅ | ❌ |
| 移除成员 | ✅ | ✅（不能移除群主/管理员） | ❌ |
| 修改成员角色 | ✅ | ❌ | ❌ |
| 解散群 | ✅ | ❌ | ❌ |
| 退出群聊 | ❌（只能解散） | ✅ | ✅ |

## 六、前端设计

### 6.1 路由

`App.tsx` 中 `<Route element={<AppLayout />}>` 下新增：

- `/messages` — 聊天主页（无选中会话时显示占位提示）
- `/messages/:conversationId` — 选中某个会话
- `/users/:userId` — 用户首页（替换原有 SettingsModal）

### 6.2 页面布局

```
┌──────────────────────────────────────────────────────┐
│ TopBar（新增"消息"导航按钮，有未读时显示红点）         │
├──────────┬───────────────────────────────────────────┤
│ 会话列表  │         聊天区                              │
│          │                                              │
│ 🔍搜索   │  ┌─ 引用栏（回复消息时显示）────────┐       │
│          │  │  "回复 张三: 你好..."  ✕ 关闭  │        │
│ 好友A    │  └──────────────────────────────────┘       │
│ 好友B    │                                              │
│ 群聊C    │  消息气泡列表（react-markdown 渲染）         │
│ ...      │  ├ hover → [回复] [转发] [复制] [撤回]      │
│          │  └ 引用消息显示引用块（可点击跳转）          │
│ [好友申  │                                              │
│  请入口] │  发送时间（每条消息上方居中显示）           │
│          │                                              │
├──────────┼───────────────────────────────────────────┤
│          │  输入区                                      │
│          │  ├ 回复引用条（如有）                        │
│          │  ├ 文本输入框（支持 Markdown）               │
│          │  └ [图片] [文件] [发送]                     │
├──────────┴───────────────────────────────────────────┤
│ 右侧面板（可展开）：会话设置                            │
│   - 私聊：好友备注、删除好友                           │
│   - 群聊：群名称、成员列表、角色管理                    │
└──────────────────────────────────────────────────────┘
```

### 6.3 关键组件

| 组件 | 说明 |
|------|------|
| `ConversationList` | 左侧会话列表。置顶会话排在最前，其余按最后消息时间降序排列。搜索框按备注/昵称过滤。每个会话显示在线状态圆点（绿/灰）、未读红点。右键/长按可置顶/取消置顶 |
| `MessageArea` | 右侧聊天区，消息列表 + 输入区 + 引用栏 |
| `MessageBubble` | 聊天气泡，复用 `Streamdown` 渲染 Markdown，hover 显示操作栏（回复/转发/复制/撤回）。系统邀请消息以特殊卡片样式渲染 |
| `MessageInput` | 输入框，支持 Markdown、@提及、图片/文件上传，发送按钮 |
| `ReplyBar` | 回复引用条，显示被回复消息摘要，点击跳转到原消息 |
| `ConversationSettings` | 右侧面板，好友备注设置 / 群组管理 |
| `FriendRequestsModal` | 好友申请弹窗，展示收到/发出的申请列表 |
| `SearchUserModal` | 搜索用户弹窗，用于添加好友 |
| `FriendPickerModal` | 好友选择器弹窗，用于添加协作者时选择好友 |
| `UserProfile` | 用户首页页面，显示头像/昵称/ID/知识库列表，动态按钮区 |
| `InvitationCard` | 协作邀请消息卡片，接受/拒绝操作 |

### 6.4 消息气泡交互细节

| 操作 | 触发条件 | 行为 |
|------|----------|------|
| 回复 | 所有消息 | 点击后在输入框上方出现引用条，显示被回复消息摘要；发送后携带 `reply_to_id` |
| 转发 | 所有消息 | 弹窗选择目标会话，生成新消息（`content` 前加"`[转发] `"标记） |
| 复制 | 所有消息 | 复制消息纯文本到剪贴板 |
| 撤回 | 仅自己的消息 + 发送后 2 分钟内 | 调用撤回 API，气泡变为"消息已撤回"；同时 WebSocket 推送 `message_recalled` 给其他用户 |

**回复消息展示**：
- 消息气泡上方显示可点击的引用块："`回复 张三：被引用消息的摘要...`"
- 点击引用块 → `scrollIntoView` 滚动定位到被引用消息，目标消息高亮闪烁

### 6.5 状态管理

新增 `message-store.tsx`（React Context + useReducer），管理：
- 好友列表、好友申请
- 会话列表（好友 + 群聊混合）
- 当前会话的消息列表
- WebSocket 连接状态
- 未读计数

### 6.6 TopBar 修改

- 新增"消息"导航按钮（在 logo 和搜索框之间），跳转 `/messages`。当有未读消息或待处理好友申请时，显示红点角标
- 右上角用户头像 → 导航到 `/users/{当前用户ID}`（**替换** 现有 `SettingsModal`）
- 移除 `SettingsModal` 组件引用，头像下拉菜单中的"设置"改为导航到用户首页
- 轮询未读计数接口，每 30 秒一次

## 七、chat-server 模块结构

```
chat-server/
├── cmd/
│   └── main.go
├── internal/
│   ├── app/
│   │   └── app.go              # 应用启动、依赖注入
│   ├── handler/
│   │   ├── friend.go            # 好友关系 gRPC handler
│   │   ├── group.go             # 群组 gRPC handler
│   │   ├── message.go           # 消息 gRPC handler
│   │   └── websocket.go         # WebSocket 连接管理
│   ├── repository/
│   │   ├── friend.go
│   │   ├── group.go
│   │   └── message.go
│   └── service/
│       ├── friend.go
│       ├── group.go
│       ├── message.go
│       └── websocket.go
├── pkg/
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   └── database.go
│   └── logger/
│       └── logger.go
├── configs/
│   └── config.yaml
├── go.mod
└── Taskfile.yml
```

## 八、未读消息与在线状态

### 8.1 WebSocket 连接策略

- WebSocket **仅在进入 `/messages` 页面时建立连接**，离开页面时断开
- 在其他页面（如知识库、设置等）通过 **HTTP 轮询** 获取未读状态

### 8.2 轮询接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/messages/unread-count` | 返回总有未读消息的会话数量 |

- 前端在 `TopBar` 中每 30 秒轮询一次
- 未读数 > 0 时，TopBar "消息" 按钮显示红点
- 进入聊天页面后停止轮询，改为 WebSocket 实时更新

### 8.3 未读计数规则

- 每个会话记录用户已读的最后一条消息 `seq_id`（存入 `friends.last_read_seq_id` / `group_members.last_read_seq_id`）
- 未读数 = 会话中 `seq_id > last_read_seq_id` 的消息数
- 用户打开某会话时，自动将 `last_read_seq_id` 更新为该会话最大 `seq_id`（即"已读"）
- 会话列表中，有未读消息的好友/群聊显示红点角标

## 九、@提及功能

### 9.1 @规则

- 群成员可以 @其他群成员（不含自己）
- 群主可以 @所有人
- 在输入框输入 `@` 触发成员选择下拉，支持模糊搜索成员昵称/备注

### 9.2 @数据存储

消息表 `extra` JSON 字段中增加 `mentions` 数组：

```json
{
  "mentions": [101, 102]   // 被 @ 的用户 ID 列表；[0] 表示 @所有人
}
```

### 9.3 @显示规则

- 消息中的 `@用户名` 按**当前用户的备注**渲染：如果我对被 @ 的人设置了备注，则我看到的 @ 文本显示为备注名，否则显示昵称
- 自己被 @ 的消息在聊天区**高亮显示**（如背景色加深或左侧加标记条）
- @所有人 的消息显示为 "@所有人"

## 十、ID 类型规范

所有实体主键统一使用 xid 字符串（`char(20)`），引用用户的外键也使用字符串类型（纯数字字符串，如 `"1001"`）。用户 ID 保持 string 类型（纯数字，不含字母符号，不允许前导 0），由数据库自增生成后转为字符串。

- 用户 ID：纯数字字符串（如 `"1"`, `"1001"`），数据库自增 + 转为字符串，无前导零
- 其他实体 ID：xid 生成（如 `"cvn8u5h2g..."`），20 位字符
- chat-server 所有表的主键和外键均使用 string 类型

## 十一、用户首页

### 11.1 路由

新增 `/users/:userId` 路由（`AppLayout` 内）。

### 11.2 页面内容

- 用户头像（大尺寸）、昵称、数字 ID
- 用户公开知识库列表（含每个知识库的关注数）
- 操作按钮区（根据与当前登录用户的关系动态展示）：

| 关系 | 显示的按钮 |
|------|-----------|
| 自己 | [更新信息] [注销账号] |
| 非好友 | [添加好友] |
| 已是好友 | [发消息] [删除好友] |
| 有好友申请待处理 | [接受] [拒绝] |

- 点击 [发消息] 跳转到 `/messages/<conversationId>`（私聊该用户）
- 点击 [添加好友] 弹出附带备注的申请发送弹窗
- 点击 [更新信息] 进入编辑模式，可修改昵称、头像等
- 点击 [注销账号] 弹出确认框，确认后调用 `DELETE /api/v1/users/:user_id` 注销

### 11.3 TopBar 修改

- 点击右上角自身头像 → 导航到 `/users/{当前用户ID}`（**替换** 现有 `SettingsModal` 弹出）
- 移除 `TopBar` 中的 `SettingsModal` 引入和设置弹窗相关代码
- 头像下拉菜单中的"设置"按钮也改为导航到用户首页

### 11.4 入口

- 知识库搜索结果的作者区域 → 点击进入用户首页
- 协作者列表中的人名/头像 → 点击进入用户首页
- 知识库详情页的拥有者信息 → 点击进入用户首页

### 11.4 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/users/:user_id/profile` | 获取用户资料（昵称、头像、ID、公开知识库列表），含与请求者的好友关系状态 |
| GET | `/api/v1/users/:user_id/repos` | 获取用户的公开知识库列表 |

## 十二、协作者邀请机制

### 12.1 流程

```
添加协作者（知识库设置页）
  → 弹出好友选择器（搜索好友）
  → 选择好友 → 发送邀请系统消息
  → 被邀请者收到 WebSocket 推送的邀请消息卡片
  → 点击 [接受] 或 [拒绝]
  → 接受后，repo-server 执行添加协作者
  → 接受/拒绝结果通过 WebSocket 回推给邀请者
```

### 12.2 邀请消息格式

新增消息内容类型 `system_invitation`。消息 `extra` 字段存储：

```json
{
  "invitation_type": "repo_collab",
  "repo_id": "...",
  "repo_name": "知识库名称",
  "role": "DEVELOPER",
  "status": "pending",           // pending / accepted / rejected
  "inviter_id": 1001,
  "inviter_name": "邀请人昵称"
}
```

### 12.3 消息卡片渲染

聊天区中，系统邀请消息以特殊卡片样式渲染：

```
┌────────────────────────────────────────┐
│ 🔔 协作邀请                            │
│                                       │
│ 张三 邀请你成为                        │
│ [知识库名称] 的开发者                   │
│                                       │
│ [查看知识库]   [接受]   [拒绝]         │
└────────────────────────────────────────┘
```

- 点击 [查看知识库] → 跳转到 `/repos/:repoId`
- 点击 [接受] → 调用 API `PUT /api/v1/collaborators/invitations/:msgId/accept`
- 点击 [拒绝] → 调用 API `PUT /api/v1/collaborators/invitations/:msgId/reject`
- 处理完成后卡片状态更新，按钮不可用

### 12.4 好友选择器

- 弹窗列表展示所有好友（头像 + 昵称 + 备注）
- 顶部搜索框按昵称/备注过滤
- 选择目标好友后，角色选为 DEVELOPER 或 VIEWER，确认后发送邀请
- API：`POST /api/v1/collaborators/invite`

### 12.5 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/collaborators/invite` | 发送协作邀请（body: `{ friend_id, repo_id, role }`） |
| PUT | `/api/v1/collaborators/invitations/:msg_id/accept` | 接受邀请 |
| PUT | `/api/v1/collaborators/invitations/:msg_id/reject` | 拒绝邀请 |

### 12.6 影响范围

- **Gateway**：新增 `/api/v1/collaborators/invite` 等路由
- **chat-server**：消息表 `content_type` ENUM 新增 `system_invitation` 值
- **repo-server**：移除直接的 `AddCollaborator` 直接添加逻辑，改为接收邀请后由 Gateway 调用
- **前端 RepoSettings**：`添加协作者` 输入框改为好友选择器弹窗
- **前端 CollabList**：协作者显示头像 + 昵称（好友用备注），可点击跳转用户首页
- **前端 SearchResult**：作者区可点击跳转用户首页

### 12.7 协作者列表显示

协作者列表中，每个协作者显示：
- 头像（圆形）
- 昵称（如果是好友，优先显示备注）
- 角色标签
- 点击头像或昵称 → 跳转 `/users/:userId`

## 十三、技术约束

- 所有业务日志使用 `slog.Info` / `slog.Error` 等全局函数
- 数据库不使用外键和 JOIN，多表数据在应用层组合
- 所有 Go 结构体、函数、关键逻辑添加中文注释
- 前端遵循现有 Tailwind CSS + Ant Design 样式体系
- WebSocket 连接需处理断线重连、心跳保活
- 图片/文件上传大小限制：图片 10MB，文件 50MB

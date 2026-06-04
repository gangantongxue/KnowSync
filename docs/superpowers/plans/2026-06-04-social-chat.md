# 社交聊天功能实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 KnowSync 平台中新增完整的即时通讯功能，包含好友管理、私聊、群聊、协作邀请，以及用户首页。

**Architecture:** 新增 chat-server 微服务（gRPC + MySQL + Redis），Gateway 新增 REST/WebSocket 路由，前端新增聊天模块（React Context + WebSocket 客户端），并将用户 ID 从 xid 迁移为纯数字。

**Tech Stack:** Go (Hertz, gRPC, MySQL, Redis), React (TypeScript, Tailwind CSS), WebSocket

---

## 阶段一：基础设施与数据库（chat-server 搭建 + DB Schema）

### Task 1.1: 创建 chat-server 项目骨架

**Files:**
- Create: `chat-server/go.mod`
- Create: `chat-server/cmd/main.go`
- Create: `chat-server/internal/app/app.go`
- Create: `chat-server/pkg/config/config.go`
- Create: `chat-server/pkg/config/model/*.go`
- Create: `chat-server/pkg/database/database.go`
- Create: `chat-server/pkg/logger/logger.go`
- Create: `chat-server/configs/config.yaml`
- Create: `chat-server/Taskfile.yml`
- Modify: `go.work`（添加 `./chat-server`）

**目标：** 搭建 chat-server 项目骨架，参照 user-server 的目录结构和依赖注入模式（samber/do），集成 GORM + MySQL + Redis，注册到 go.work。

**步骤：**
- [x] 创建 chat-server 目录结构和 go.mod（模块路径 `github.com/gangantongxue/knowsync/chat-server`）
- [x] 编写 `pkg/config/config.go` 和配置模型（database、redis、grpc、logger）
- [x] 创建 `configs/config.yaml` 示例配置
- [x] 编写 `pkg/database/database.go`：GORM 初始化 + AutoMigrate
- [x] 编写 `pkg/logger/logger.go`：slog 初始化，设置全局默认 logger
- [x] 编写 `internal/app/app.go`：使用手动 DI 注册数据库、Redis、启动 gRPC server
- [x] 编写 `cmd/main.go`：加载配置，启动应用
- [x] 创建 `Taskfile.yml`：build、run、lint 等任务
- [x] 修改 `go.work` 添加 `./chat-server`
- [x] 验证：`go build ./chat-server/cmd/main.go` 编译通过

### Task 1.2: 数据库 Schema 设计（Migration）

**Files:**
- Create: `chat-server/pkg/database/migrations/001_create_tables.sql`

**目标：** 创建聊天模块所需的全部数据库表。

**表清单：**
- `friend_requests` — 好友申请表
- `friends` — 好友关系表（含 remark、last_message_at、last_read_seq_id、pinned）
- `messages` — 消息表（含 conversation_type、content_type 含 system_invitation、mentions JSON、reply_to_id）
- `groups` — 群组表
- `group_members` — 群成员表（含 last_read_seq_id、pinned）

**步骤：**
- [x] 编写 SQL migration 脚本（使用 GORM AutoMigrate）
- [x] 在 `database.go` 中执行 AutoMigrate
- [x] 验证：启动 chat-server 后所有表自动创建

### Task 1.3: 用户 ID 迁移（xid → 纯数字）

**Files:**
- Modify: `user-server/pkg/database/schema/user.go`
- Modify: `ks-proto/pkg/pb/*.proto`（涉及 user_id 的字段）
- Modify: `gateway/internal/handler/*.go`（适配新的数值 ID 类型）
- Modify: `web/src/lib/auth.ts`（适配数字 ID）
- Modify: `web/src/store/auth-context.tsx`（适配数字 ID）
- Modify: 所有前端使用 `user.id` 的类型声明

**目标：** 用户 ID 生成策略从 `rs/xid` 改为纯数字（雪花算法或自增）。其他实体 ID 保持 xid 不变。

**步骤：**
- [x] 在 `user-server/pkg/database/schema/user.go` 中修改 ID 字段为 `string` 类型，使用 crypto/rand 生成纯随机数字
- [x] 修改 JWT 中 user_id 类型：user-server 和 gateway 的 auth middleware
- [x] 重新生成 proto Go 代码
- [x] 修改 gateway handler 中所有涉及 `user_id` 的处理逻辑
- [x] 修改前端 `auth.ts` 类型定义：UserInfo.id 保持 string 类型（内容为纯数字字符串）
- [x] 修改 repo-server、ai-server 中所有 user_id 引用为 string
- [x] 重新编译所有服务，验证无编译错误

---

## 阶段二：chat-server 核心业务（好友 + 消息 + 群组）

### Task 2.1: 好友管理模块

**Files:**
- Create: `chat-server/pkg/database/schema/friend.go`
- Create: `chat-server/pkg/database/schema/friend_request.go`
- Create: `chat-server/internal/repository/friend.go`
- Create: `chat-server/internal/service/friend.go`
- Create: `ks-proto/chat/chat.proto`（好友相关 RPC）

**目标：** 实现好友申请的发送、处理、好友列表查询、删除、备注修改、用户搜索。

**步骤：**
- [x] 定义 GORM model：FriendRequest、Friend
- [x] 编写 FriendRepository：CRUD 操作（CreateRequest、GetRequestsByReceiver、GetRequestsBySender、UpdateRequestStatus、CreateFriendPair、DeleteFriend、UpdateRemark、GetFriendList、SearchUsers）
- [x] 编写 FriendService：业务逻辑（发送申请时检查是否已是好友/已有待处理申请、接受时创建双向好友记录、删除时双向删除）
- [x] 定义 gRPC proto（chat.proto）：SendFriendRequest、AcceptFriendRequest、RejectFriendRequest、GetFriendRequests、GetFriendList、DeleteFriend、UpdateFriendRemark、SearchUsers
- [x] 编写 gRPC handler：实现上述 RPC 接口
- [x] 好友列表支持 `?q=` 搜索备注 + `last_message_at DESC` 排序

### Task 2.2: 消息发送与获取

**Files:**
- Create: `chat-server/pkg/database/schema/message.go`
- Create: `chat-server/internal/repository/message.go`
- Create: `chat-server/internal/service/message.go`
- Modify: `ks-proto/chat/chat.proto`（消息相关 RPC）

**目标：** 实现消息发送、游标分页获取历史、关键词搜索、文件上传、撤回。

**步骤：**
- [x] 定义 GORM model：Message（含 conversation_type、conversation_id、seq_id、content_type、extra JSON、reply_to_id、status、mentions）
- [x] 编写 MessageRepository：SendMessage（生成 seq_id，保证会话内自增）、GetMessages（游标分页）、RecallMessage
- [x] 编写 MessageService：发送时自动更新 friends.last_message_at、校验撤回条件（发送者 + 5min 内）
- [x] 定义 gRPC proto：SendPrivateMessage、SendGroupMessage、GetMessages、RecallMessage、GetUnreadCount
- [x] 编写 gRPC handler（注册到 app.go）
- [ ] chat-server 集成现有 storage 模块：通过 gateway 的 storage 接口上传聊天附件，消息中存储文件 URL（延期）

### Task 2.3: 群组管理模块

**Files:**
- Create: `chat-server/pkg/database/schema/group.go`
- Create: `chat-server/pkg/database/schema/group_member.go`
- Create: `chat-server/internal/repository/group.go`
- Create: `chat-server/internal/service/group.go`
- Modify: `ks-proto/chat/chat.proto`（群组相关 RPC）

**目标：** 实现创建群聊、成员管理、角色变更、退出/解散群。

**步骤：**
- [x] 定义 GORM model：Group、GroupMember
- [x] 编写 GroupRepository：CreateGroup、GetGroupsByUser、AddMember、RemoveMember、UpdateRole、DeleteGroup、GetMembers
- [x] 编写 GroupService：创建群时自动将创建者设为 owner、权限校验（管理员不能移除群主/管理员、仅群主可解散和修改角色）
- [x] 定义 gRPC proto：CreateGroup、GetGroupInfo、UpdateGroup、AddMembers、RemoveMember、LeaveGroup、TransferOwnership、SetAdmin、RemoveAdmin、GetGroupMembers、GetUserGroups
- [x] 编写 gRPC handler

### Task 2.4: 消息附加功能（回复 + @提及 + 转发）

**Files:**
- Modify: `chat-server/internal/service/message.go`
- Modify: `chat-server/internal/repository/message.go`

**目标：** 支持回复消息、@提及、转发消息的服务端逻辑。

**步骤：**
- [x] 发送消息时支持 `reply_to_id` 字段：验证被回复消息存在于同一会话中
- [x] 发送消息时支持 `mentions` 数据：验证被 @ 的用户在群中存在、不可 @ 自己、群主可 @所有人（mentions=[0]）
- [x] 转发消息：检查消息类型，创建新消息时标记为转发

### Task 2.5: 会话列表与会话管理

**Files:**
- Create: `chat-server/internal/service/conversation.go`
- Create: `chat-server/internal/repository/conversation.go`
- Modify: `ks-proto/chat/chat.proto`

**目标：** 提供会话列表聚合接口（好友 + 群聊混合），支持置顶。

**步骤：**
- [x] 编写 GetConversations 逻辑：查询用户的所有私聊（friends 表）+ 群聊（group_members 表），按各自的最后消息时间聚合排序，置顶会话排在最前
- [x] 编写 PinConversation 逻辑：更新 friends.pinned 或 group_members.pinned
- [x] 定义 gRPC proto：GetConversations、MarkConversationRead、TogglePin、DeleteConversation

---

## 阶段三：WebSocket + 在线状态

### Task 3.1: WebSocket 连接管理

**Files:**
- Create: `chat-server/internal/ws/hub.go`
- Create: `chat-server/internal/ws/client.go`
- Create: `chat-server/internal/service/websocket.go`

**目标：** 实现 WebSocket 连接管理（连接池、消息广播、心跳保活、断线重连）。

**步骤：**
- [x] 编写 `hub.go`：Hub 结构体，维护在线用户连接映射表 `map[string][]*Client`（一个用户可能多端登录）
- [x] 编写 `client.go`：Client 结构体，封装 WebSocket 连接、读写 goroutine、心跳 ping/pong 机制
- [x] WebSocket 连接建立时：解析 `?token=` 参数验证 JWT，提取 user_id
- [x] 消息推送：SendToUser、SendToConversation（推送给会话内所有在线用户）
- [x] 断线处理：连接关闭时从 Hub 移除，更新 Redis 在线状态

### Task 3.2: 在线状态（Redis）

**Files:**
- Modify: `chat-server/internal/ws/hub.go`
- Modify: `chat-server/internal/service/websocket.go`
- Modify: `user-server/pkg/database/schema/user.go`（添加 last_online_at 字段）

**目标：** 使用 Redis 跟踪用户在线状态。

**步骤：**
- [x] WebSocket 连接建立时：`SET user_online:{user_id} 1 EX 60`
- [x] 定期续期：每 30 秒 `EXPIRE user_online:{user_id} 60`
- [x] WebSocket 断开时：`DEL user_online:{user_id}`
- [x] 实现 `GetOnlineStatus` RPC：批量查询好友在线状态
- [ ] user-server 用户表增加 `last_online_at` 字段（延期）

### Task 3.3: 消息实时推送

**Files:**
- Modify: `chat-server/internal/service/message.go`
- Modify: `chat-server/internal/service/friend.go`
- Modify: `chat-server/internal/ws/hub.go`

**目标：** 发送消息后通过 WebSocket 推送给会话内其他在线用户。

**步骤：**
- [x] SendMessage 成功后，调用 Hub 向会话内其他在线用户推送 `new_message` 事件
- [x] RecallMessage 成功后，推送 `message_recalled` 事件
- [x] 好友申请发送时，向接收者推送 `friend_request` 事件
- [x] 好友申请被接受时，向申请者推送 `friend_accepted` 事件
- [ ] 协作者邀请消息推送（与普通消息相同机制，WebSocket 推送）

---

## 阶段四：Gateway 集成

### Task 4.1: Gateway REST 路由 + gRPC 客户端

**Files:**
- Create: `gateway/internal/handler/chat_friend.go`
- Create: `gateway/internal/handler/chat_message.go`
- Create: `gateway/internal/handler/chat_group.go`
- Create: `gateway/internal/handler/chat_conversation.go`
- Create: `gateway/internal/handler/chat_websocket.go`
- Modify: `gateway/internal/router/router.go`
- Modify: `gateway/pkg/grpcclient/client.go`（添加 chat-server 连接）
- Modify: `gateway/pkg/config/model/grpc.go`（添加 chat-server 配置）

**目标：** Gateway 层添加所有聊天相关的 HTTP REST 路由和 WebSocket 端点，通过 gRPC 调用 chat-server。

**步骤：**
- [x] 在 gRPC client 中注册 chat-server 连接
- [x] 编写 `chat_friend.go`：好友申请、列表、删除、备注、用户搜索的 HTTP handler（转调 chat-server gRPC）
- [x] 编写 `chat_message.go`：发消息、获取消息、撤回、转发的 HTTP handler
- [x] 编写 `chat_group.go`：群组 CRUD、成员管理的 HTTP handler
- [x] 编写 `chat_conversation.go`：会话列表、置顶、读标记的 HTTP handler
- [x] 在 `router.go` 中注册所有新路由（`/api/v1/friends/*`、`/api/v1/messages/*`、`/api/v1/groups/*`、`/api/v1/conversations/*`）
- [x] 新增轮询接口 `GET /api/v1/messages/unread-count`

### Task 4.2: 协作者邀请接口

**Files:**
- Modify: `gateway/internal/handler/collaborator.go`
- Modify: `gateway/internal/router/router.go`

**目标：** 新增协作者邀请流程的 Gateway handler。

**步骤：**
- [x] 新增 `POST /api/v1/collaborators/invite`：接收 friend_id + repo_id + role，通过 chat-server 发送系统邀请消息
- [x] 新增 `POST /api/v1/collaborators/invitations/:message_id/accept`：验证消息类型为 invitation，调用 repo-server 添加协作者
- [x] 新增 `POST /api/v1/collaborators/invitations/:message_id/reject`：接受并返回成功
- [x] 保留原有 `POST /api/v1/repos/:repo_id/collaborators`（兼容旧版本）

### Task 4.3: 用户首页 API

**Files:**
- Create: `gateway/internal/handler/user_profile.go`
- Modify: `gateway/internal/router/router.go`

**目标：** 提供用户首页所需的聚合数据接口。

**步骤：**
- [x] 新增 `GET /api/v1/users/:user_id/profile`：聚合返回用户信息 + 公开知识库列表 + 与请求者的好友关系状态
- [x] 新增 `GET /api/v1/users/:user_id/repos`：返回用户的公开知识库列表
- [x] 在 `router.go` 注册新路由

---

## 阶段五：前端核心（聊天页面 + 好友流程）

### Task 5.1: 消息状态管理（MessageStore）

**Files:**
- Create: `web/src/store/message-store.tsx`
- Create: `web/src/lib/chat-api.ts`
- Modify: `web/src/App.tsx`（添加 MessageProvider）

**目标：** 创建前端聊天状态管理（React Context + useReducer），封装所有聊天 API 调用。

**步骤：**
- [x] 创建 `chat-api.ts`：封装所有聊天 API（好友、消息、群组、会话的 fetch 调用），复用 `client.ts` 的 `request` 函数
- [x] 创建 `message-store.tsx`：定义 MessageState 和 reducer
- [x] 实现核心 actions：loadConversations、loadMessages、sendMessage、recallMessage、pinConversation 等
- [x] 实现好友 actions：sendFriendRequest、acceptRequest、rejectRequest、deleteFriend、updateRemark
- [x] 实现群组 actions：createGroup、addMember、removeMember、updateRole、leaveGroup
- [x] 在 `App.tsx` 中包裹 `MessageProvider`

### Task 5.2: WebSocket 客户端

**Files:**
- Create: `web/src/lib/ws-client.ts`

**目标：** 实现前端 WebSocket 客户端，支持断线重连。

**步骤：**
- [x] 编写 `ws-client.ts`：封装原生 WebSocket API
- [x] 连接 URL：`/ws?token=${getToken()}`
- [x] 实现自动重连：指数退避（1s → 2s → 4s → 8s → 16s → 最大 30s）
- [x] 事件分发：new_message、message_recalled、friend_request、friend_accepted → 回调更新 MessageStore
- [x] 仅进入 `/messages` 页面时建立连接，离开时断开

### Task 5.3: 聊天页面布局 + 会话列表

**Files:**
- Create: `web/src/pages/Messages.tsx`
- Create: `web/src/components/Messages/ConversationList.tsx`
- Modify: `web/src/App.tsx`（注册路由）
- Modify: `web/src/components/Layout/AppLayout.tsx`（消息页不显示侧边栏）

**目标：** 实现聊天页面的整体布局和左侧会话列表。

**步骤：**
- [x] 创建 `Messages.tsx`：左右分栏布局（左侧 280px 会话列表 + 右侧聊天区），URL 参数控制选中会话
- [x] 在 `App.tsx` 注册路由 `/messages` 和 `/messages/:conversationType/:conversationId`
- [x] 修改 `AppLayout.tsx`：消息页不显示 Sidebar（全宽布局），类似搜索页的处理
- [x] 创建 `ConversationList.tsx`：
  - 顶部搜索框（按备注/昵称过滤会话）
  - 好友 + 群聊混合列表，按最后消息时间排序，置顶优先
  - 每个会话项显示：头像、名称/备注、最后一条消息摘要、时间、未读红点
  - 好友申请入口（顶部按钮，红点提示）
  - 右键菜单：置顶/取消置顶

### Task 5.4: 消息区 + 消息渲染

**Files:**
- Create: `web/src/components/Messages/MessageArea.tsx`
- Create: `web/src/components/Messages/MessageBubble.tsx`
- Create: `web/src/components/Messages/InvitationCard.tsx`

**目标：** 实现右侧聊天区：消息列表滚动、消息气泡渲染（文本/图片/文件/系统邀请）、hover 操作栏。

**步骤：**
- [x] 创建 `MessageArea.tsx`：消息列表、向上滚动分页、新消息自动滚动到底部、回复引用条
- [x] 创建 `MessageBubble.tsx`：文本/图片/文件/系统邀请渲染、hover 操作栏（回复/转发/撤回）、回复引用块、@高亮、撤回状态
- [x] 创建 `InvitationCard.tsx`：协作邀请卡片（接受/拒绝按钮）
- [x] 处理消息撤回状态：已撤回消息显示"消息已撤回"灰色文本

### Task 5.5: 消息输入框

**Files:**
- Create: `web/src/components/Messages/MessageInput.tsx`
- Create: `web/src/components/Messages/MentionDropdown.tsx`

**目标：** 实现底部消息输入区。

**步骤：**
- [x] 创建 `MessageInput.tsx`：多行 auto-resize 文本框、图片/文件上传、send 按钮、@触发、回复引用条
- [x] 创建 `MentionDropdown.tsx`：@输入后弹出群成员列表，支持搜索，群主显示"@所有人"

### Task 5.6: 好友申请流程

**Files:**
- Create: `web/src/components/Messages/FriendRequestsModal.tsx`
- Create: `web/src/components/Messages/SearchUserModal.tsx`

**目标：** 实现好友申请的发送和接收 UI。

**步骤：**
- [x] 创建 `FriendRequestsModal.tsx`：Tab 切换收到的/发出的申请，接受/拒绝，状态显示
- [x] 创建 `SearchUserModal.tsx`：搜索用户，添加好友（附带备注输入）
- [x] 集成到 ConversationList：顶部按钮打开好友申请弹窗

### Task 5.7: 群聊管理

**Files:**
- Create: `web/src/components/Messages/CreateGroupModal.tsx`
- Create: `web/src/components/Messages/GroupSettingsModal.tsx`

**目标：** 实现群聊创建和管理 UI。

**步骤：**
- [x] 创建 `CreateGroupModal.tsx`：输入群名称，从好友列表多选成员，创建群
- [x] 创建 `GroupSettingsModal.tsx`：群名称编辑、成员列表、角色管理、添加/移除成员、权限控制

### Task 5.8: 转发与搜索

**Files:**
- Create: `web/src/components/Messages/ForwardModal.tsx`

**目标：** 实现消息转发。

**步骤：**
- [x] 创建 `ForwardModal.tsx`：展示会话列表（好友+群聊），选择目标后发送转发消息

---

## 阶段六：前端集成（用户首页 + 协作者邀请 + UI 改进）

### Task 6.1: 用户首页

**Files:**
- Create: `web/src/pages/UserProfile.tsx`
- Modify: `web/src/App.tsx`（注册路由）

**目标：** 实现用户首页页面，替换原有 SettingsModal。

**步骤：**
- [x] 创建 `UserProfile.tsx`：展示用户头像、昵称、知识库列表、动态按钮区、编辑模式
- [x] 在 `App.tsx` 注册路由 `/users/:userId`
- [x] 创建 `web/src/lib/user-api.ts` 封装用户相关 API

### Task 6.2: TopBar 修改

**Files:**
- Modify: `web/src/components/Layout/TopBar.tsx`

**目标：** 添加"消息"按钮 + 未读轮询 + 头像点击跳转用户首页。

**步骤：**
- [x] 新增"消息"导航按钮（logo 和搜索框之间），跳转 `/messages`
- [x] 实现轮询逻辑：每 30 秒调用 `/api/v1/messages/unread-count`，未读 > 0 显示红点
- [x] 轮询也检查待处理好友申请数
- [x] 修改头像点击行为：从打开 `SettingsModal` → `navigate(\`/users/${user.id}\`)`
- [x] 移除 `SettingsModal` 的引用和显示逻辑
- [x] 头像下拉菜单"设置"改为导航到 `/users/${user.id}`

### Task 6.3: 协作者邀请改进

**Files:**
- Create: `web/src/components/Messages/FriendPickerModal.tsx`
- Modify: `web/src/components/Repo/RepoSettings.tsx`
- Modify: `web/src/components/Repo/CollabList.tsx`

**目标：** 将添加协作者改为邀请模式，协作者列表展示优化。

**步骤：**
- [x] 创建 `FriendPickerModal.tsx`：展示好友列表，搜索过滤，选择好友 + 角色 → 发送邀请
- [x] 修改 `RepoSettings.tsx`：移除直接添加 UI，改为按钮弹出 `FriendPickerModal`
- [x] 修改 `CollabList.tsx`：显示头像 + 昵称，可点击跳转 `/users/:userId`

### Task 6.4: 搜索页改进

**Files:**
- Modify: `web/src/pages/SearchResult.tsx`

**目标：** 搜索结果的作者信息可点击跳转用户首页。

**步骤：**
- [x] 每个知识库卡片中的作者区域（头像 + 昵称）包裹 `navigate(\`/users/${author.id}\`)`

### Task 6.5: TopBar 轮询集成

**Files:**
- Modify: `web/src/components/Layout/TopBar.tsx`

**目标：** TopBar 中的轮询只在非消息页面进行（消息页面使用 WebSocket）。

**步骤：**
- [x] 在 TopBar 中根据当前路由判断是否在 `/messages` 页面
- [x] 仅在非消息页面启动轮询 timer（useEffect + setInterval 30s）
- [x] 进入消息页面时清除 timer

### Task 6.6: 清理旧代码

**Files:**
- Delete: `web/src/components/Settings/SettingsModal.tsx`
- Modify: `web/src/components/Layout/TopBar.tsx`

**目标：** 移除被替换的旧 SettingsModal。

**步骤：**
- [x] 删除 `SettingsModal.tsx` 组件及文件
- [x] 清理 TopBar 中的 SettingsModal import 和相关状态
- [x] 验证无编译错误

---

## 阶段验收标准

| 阶段 | 验收目标 |
|------|---------|
| 阶段一 | chat-server 可编译运行，DB 表自动创建，用户 ID 为纯数字 |
| 阶段二 | 好友 CRUD、消息收发、群组管理全部可通过 gRPC 调用测试 |
| 阶段三 | WebSocket 连接建立/断开正常，消息实时推送验证通过，在线状态查询正确 |
| 阶段四 | 所有 REST API 通过 curl 测试通过，gateway 路由注册完整 |
| 阶段五 | 前端聊天页面可完成：发消息、收消息、加好友、建群聊、@提及、回复、转发、撤回、会话置顶、消息搜索 |
| 阶段六 | 用户首页正常展示、协作者邀请流程完整、搜索页跳转正确、TopBar 红点提示正常 |

---

## 依赖关系

```
阶段一 ──► 阶段二 ──► 阶段三 ──► 阶段四 ──► 阶段五 ──► 阶段六
                                      │
                                      └── 协作者邀请需要消息系统先有 system_invitation 类型
```

- 阶段二依赖阶段一（DB Schema + 用户 ID）
- 阶段三依赖阶段二（WebSocket 推送依赖消息/好友功能）
- 阶段四依赖阶段二、三（Gateway 路由依赖 gRPC 接口和 WebSocket）
- 阶段五依赖阶段四（前端需要 API 可用）
- 阶段六依赖阶段五（用户首页需要好友 API，协作者邀请需要消息系统）

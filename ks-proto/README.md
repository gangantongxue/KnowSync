# ks-proto — 共享 Protobuf 协议定义

KnowSync 项目服务间通信的 Protocol Buffers 协议定义与 Go 生成代码。所有 gRPC 接口契约在此定义，各服务通过 `ks-proto` 共享类型。

## 目录结构

```
ks-proto/
├── proto/              # 原始 .proto 源文件
│   ├── user.proto      # UserService 定义
│   ├── repo.proto      # RepoService 定义
│   ├── ai.proto        # AIService 定义
│   └── chat.proto      # ChatService 定义
├── pkg/pb/             # 生成的 Go 代码
│   ├── user.pb.go
│   ├── user_grpc.pb.go
│   ├── repo.pb.go
│   ├── repo_grpc.pb.go
│   ├── ai.pb.go
│   ├── ai_grpc.pb.go
│   ├── chat.pb.go
│   └── chat_grpc.pb.go
├── scripts/
│   └── build_proto.sh  # proto 生成脚本
└── go.mod
```

## 协议概览

### UserService (`user.proto`)
- `Register` — 用户注册
- `Login` — 用户登录
- `Logout` — 用户登出
- `Refresh` — 刷新令牌
- `GetUser` — 获取用户信息
- `UpdateUserInfo` — 更新用户资料
- `SetAvatar` — 设置头像
- `VerifyCode` — 发送验证码
- `ForgetPassword` — 忘记密码重置
- `ResetPassword` — 修改密码
- `Unregister` — 注销账户

### RepoService (`repo.proto`)
- `CreateRepo` / `GetRepo` / `UpdateRepo` / `DeleteRepo` — 知识库 CRUD
- `ListUserRepos` / `ListPublicRepos` — 知识库列表查询
- `AddCollaborator` / `UpdateCollaborator` / `RemoveCollaborator` / `ListCollaborators` — 协作者管理
- `FollowRepo` / `UnfollowRepo` / `ListFollowedRepos` — 关注管理
- `IncrementArticleCount` — 文章计数增减

### AIService (`ai.proto`)
- `Chat` — 流式 AI 对话（Server Streaming）
- `Search` — 语义搜索
- `GetChatSessions` / `GetChatMessages` / `DeleteChatSession` — 会话管理
- `UpdateRepoVisibility` — 可见性同步
- `VectorizeArticle` — 文章向量化
- `DeleteRepoVectors` / `DeleteFileVectors` — 向量清理

### ChatService (`chat.proto`)
- **好友**：`SendFriendRequest`、`GetFriendRequestsByReceiver` / `BySender`、`AcceptFriendRequest` / `RejectFriendRequest`、`GetFriendList`、`DeleteFriend`、`UpdateFriendRemark`、`SearchUsers`
- **消息**：`SendPrivateMessage`、`SendGroupMessage`、`GetMessages`、`GetMessageByID`、`RecallMessage`、`ForwardMessage`、`GetUnreadCount`
- **群组**：`CreateGroup`、`GetGroupInfo`、`UpdateGroup`、`AddMembers`、`RemoveMember`、`LeaveGroup`、`TransferOwnership`、`SetAdmin`、`RemoveAdmin`、`GetGroupMembers`、`GetUserGroups`
- **会话**：`GetConversationList`、`MarkConversationRead`、`TogglePin`、`DeleteConversation`
- **在线状态**：`GetOnlineStatus`

## 生成代码

修改 `.proto` 文件后需重新生成 Go 代码：

```bash
# 需要安装 protoc 和 protoc-gen-go-grpc
cd ks-proto
bash scripts/build_proto.sh
```

### 前置条件

- `protoc` — Protocol Buffers 编译器
- `protoc-gen-go` — Go 代码生成插件
- `protoc-gen-go-grpc` — gRPC Go 代码生成插件

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## 添加新服务

1. 在 `proto/` 目录下创建新的 `.proto` 文件
2. 定义 `service` 和 `message`
3. 运行 `bash scripts/build_proto.sh` 生成代码
4. 在网关和对应服务中引用新生成的 `pkg/pb` 包
5. 更新网关路由和 handler

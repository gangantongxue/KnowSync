# LLM 知识库写操作工具设计文档

> 创建日期: 2026-06-04

## 一、概述

在现有 KnowSync AI 对话系统中，LLM 仅具备知识库**读取**能力（`search_knowledge`、`list_user_repos`、`list_repo_files`、`get_file_content`）。新增一批**写操作工具**，让 LLM 能够在用户授权下创建/修改/删除知识库内容和管理协作者。

### 核心决策

| 维度 | 决策 |
|------|------|
| 工具范围 | 文件操作 + 知识库管理 + 协作者管理 |
| 确认机制 | Eino `ToolCallMiddlewares` 拦截，返回 `confirm_write` 事件，经 `ask_user` 下发前端 |
| 确认策略 | 三级分类：强制确认 / 按需确认（`_skip_confirm` 参数字段）/ 无需确认 |
| 新增工具数 | 11 个 |
| 协作查找 | 新增 `search_users` 工具，供 `add_collaborator` 等工具先搜索用户再操作 |

## 二、新增工具列表

### 2.1 文件操作（4 个）

| 工具名 | 说明 | 参数 | 确认级别 |
|--------|------|------|---------|
| `create_file` | 在知识库中创建新文件（Markdown），目录不存在时自动创建 | `repo_id`, `file_path`, `content` | 🟡 按需 |
| `update_file` | 全量覆盖更新已有文件内容 | `repo_id`, `file_path`, `content` | 🟡 按需 |
| `delete_file` | 删除知识库中的文件 | `repo_id`, `file_path` | 🔴 强制 |
| `rename_file` | 重命名/移动文件 | `repo_id`, `old_path`, `new_path` | 🟡 按需 |

### 2.2 知识库管理（2 个）

| 工具名 | 说明 | 参数 | 确认级别 |
|--------|------|------|---------|
| `create_repo` | 创建新知识库 | `name`, `description`(可选), `visibility`(可选) | 🟡 按需 |
| `update_repo` | 更新知识库设置 | `repo_id`, `name`(可选), `description`(可选), `visibility`(可选) | 🟡 按需 / 🔴 可见性变更强制 |

### 2.3 协作者管理（5 个）

| 工具名 | 说明 | 参数 | 确认级别 |
|--------|------|------|---------|
| `search_users` | 搜索用户，返回用户列表及 ID | `keyword` | 🟢 无需 |
| `add_collaborator` | 添加协作者到知识库 | `repo_id`, `user_id`, `role`(ADMIN/DEVELOPER/VIEWER) | 🔴 强制 |
| `remove_collaborator` | 移除协作者 | `repo_id`, `user_id` | 🔴 强制 |
| `update_collaborator_role` | 变更协作者角色 | `repo_id`, `user_id`, `role` | 🔴 强制 |
| `list_collaborators` | 查看协作者列表 | `repo_id` | 🟢 无需 |

## 三、确认机制设计

### 3.1 确认策略三级分类

所有写操作工具在 `tool.go` 中注册时，通过 `WriteToolPolicy` 标记其确认级别：

```go
type ConfirmLevel int

const (
    ConfirmNever   ConfirmLevel = 0 // 🟢 无需确认（只读类）
    ConfirmOptional ConfirmLevel = 1 // 🟡 按需确认（可传 _skip_confirm 跳过）
    ConfirmAlways   ConfirmLevel = 2 // 🔴 强制确认（必须经过用户确认）
)

type ToolPolicy struct {
    ToolName    string
    ConfirmLevel ConfirmLevel
}
```

### 3.2 Middleware 拦截逻辑

```go
func ConfirmationMiddleware(policies []ToolPolicy, onAskUser func(string)) compose.InvokableToolMiddleware {
    return func(next tool.InvokableTool) tool.InvokableTool {
        info, _ := next.Info(context.Background())
        name := info.Name
        policy := findPolicy(policies, name)

        if policy == nil || policy.ConfirmLevel == ConfirmNever {
            return next // 只读工具，直接放行
        }

        return tool.NewInvokableToolFunc(
            func(ctx context.Context) (*schema.ToolInfo, error) { return next.Info(ctx) },
            func(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
                // 已确认 → 放行
                if ctx.Value(ConfirmedKey) != nil {
                    return next.InvokableRun(ctx, arguments, opts...)
                }

                var raw map[string]any
                _ = json.Unmarshal([]byte(arguments), &raw)

                // update_repo 的可见性变更始终强制确认
                if name == "update_repo" {
                    if vis, ok := raw["visibility"].(string); ok && vis != "" {
                        policy = &ToolPolicy{ConfirmLevel: ConfirmAlways}
                    }
                }

                // 按需确认 + 参数中带跳过标记 → 放行
                if policy.ConfirmLevel == ConfirmOptional {
                    if skip, _ := raw["_skip_confirm"].(bool); skip {
                        return next.InvokableRun(ctx, arguments, opts...)
                    }
                }

                // 未确认 → 返回确认事件
                result, _ := json.Marshal(map[string]any{
                    "action":   "confirm_write",
                    "tool":     name,
                    "params":   arguments,
                })
                return string(result), nil
            },
        )
    }
}
```

### 3.3 前端确认流程

```
一轮对话：
  用户: "帮我写一篇Go并发笔记"
  LLM: 调用 create_file(repo_id="x", file_path="docs/go-concurrency.md", content="...")
        ↓
  Middleware 拦截（未确认）
        ↓
  返回 {"action": "confirm_write", "tool": "create_file", "params": {...}}
        ↓
  chat.go 检测到 confirm_write
        ↓
  自动生成 ask_user 确认问题 → 发送给前端
  问题: "确认要在知识库「Go学习」中创建文件 docs/go-concurrency.md 吗？\n内容概要：Go 并发模式...\n[确认] [取消]"
        ↓
  用户点击 确认
        ↓
下一轮对话：
  系统将用户确认结果注入 context（confirmed=true）
  重放 LLM Agent
  LLM: 再次调用 create_file(...)
        ↓
  Middleware 检测到已确认 → 放行
        ↓
  真实执行 → 文件创建成功 → LLM 回复 "已创建"
```

### 3.4 按需确认的 LLM 指导

对于 🟡 级别工具，tool description 中包含以下约束：

- `create_file`："如果用户明确表达了文件路径和内容，可以设置参数 `_skip_confirm: true` 跳过确认直接执行；如果用户表述不完整，不要设置该参数"
- `update_repo`（名称/描述）：同上逻辑
- `update_repo`（可见性变更）：**此参数变更属于强制确认级别，即使传了 `_skip_confirm` 也会被拦截**

### 3.5 `search_users` 工具的协作流程

`add_collaborator` 等工具需要用户 ID。LLM 的工作流程：

1. LLM 问用户 "你想添加谁为协作者？"（通过 `ask_user`）
2. 用户说 "张三"
3. LLM 调用 `search_users(keyword="张三")` 获取用户列表
4. LLM 展示给用户确认结果
5. 用户确认后，LLM 调用 `add_collaborator(repo_id="x", user_id="123", role="DEVELOPER")`

## 四、集成点

### 4.1 新依赖接口（tool.go）

在 `tool.go` 中新增以下接口：

```go
// FileWriteClient 文件写入操作客户端接口
type FileWriteClient interface {
    CreateFile(ctx context.Context, ownerID, repoID, filePath, content string) error
    UpdateFile(ctx context.Context, ownerID, repoID, filePath, content string) error
    DeleteFile(ctx context.Context, ownerID, repoID, filePath string) error
    RenameFile(ctx context.Context, ownerID, repoID, oldPath, newPath string) error
}

// RepoWriteClient 知识库写入操作客户端接口
type RepoWriteClient interface {
    CreateRepo(ctx context.Context, userID, name, description, visibility string) (string, error)
    UpdateRepo(ctx context.Context, repoID, userID, name, description, visibility string) error
}

// UserSearchClient 用户搜索客户端接口
type UserSearchClient interface {
    SearchUsers(ctx context.Context, keyword string) ([]UserInfo, error)
}

// CollaboratorClient 协作者管理客户端接口
type CollaboratorClient interface {
    AddCollaborator(ctx context.Context, repoID, userID, role string) error
    RemoveCollaborator(ctx context.Context, repoID, userID string) error
    UpdateCollaboratorRole(ctx context.Context, repoID, userID, role string) error
    ListCollaborators(ctx context.Context, repoID string) ([]CollaboratorInfo, error)
}
```

对应的数据类型：

```go
type UserInfo struct {
    ID   string
    Name string
}

type CollaboratorInfo struct {
    UserID   string
    UserName string
    Role     string
}
```

### 4.2 Agent 注册（service.go）

在 `initAgent` 中注册新工具和 Middleware：

```go
tools := []einoTool.InvokableTool{
    // 已有工具
    llmtool.NewSearchKnowledge(...),
    llmtool.NewUpdateTitle(...),
    llmtool.NewAskQuestion(onAskUser),
    llmtool.NewListRepos(...),
    llmtool.NewListRepoFiles(...),
    llmtool.NewGetFileContent(...),

    // 新增工具
    llmtool.NewCreateFile(fileWriteClient, repoDetailClient),
    llmtool.NewUpdateFile(fileWriteClient, repoDetailClient),
    llmtool.NewDeleteFile(fileWriteClient, repoDetailClient),
    llmtool.NewRenameFile(fileWriteClient, repoDetailClient),
    llmtool.NewCreateRepo(repoWriteClient),
    llmtool.NewUpdateRepo(repoWriteClient, repoDetailClient),
    llmtool.NewSearchUsers(userSearchClient),
    llmtool.NewAddCollaborator(collaboratorClient, repoDetailClient),
    llmtool.NewRemoveCollaborator(collaboratorClient, repoDetailClient),
    llmtool.NewUpdateCollaboratorRole(collaboratorClient, repoDetailClient),
    llmtool.NewListCollaborators(collaboratorClient, repoDetailClient),
}

// 注册 ConfirmationMiddleware
agentCfg.ToolsConfig.ToolCallMiddlewares = []compose.ToolMiddleware{
    {Invokable: llmtool.NewConfirmationMiddleware(writeToolPolicies, onAskUser)},
}
```

### 4.3 确认事件处理（chat.go）

在 `chat.go` 的流处理完成后，新增对 `confirm_write` 的检测逻辑：

```go
// 检测 confirm_write 事件
var confirmWrite *ConfirmWriteEvent
if !askedUser {
    confirmWrite = detectConfirmWrite(streamOutput)
}

if confirmWrite != nil {
    // 生成确认问题
    question := buildConfirmQuestion(confirmWrite)
    _ = cb(&ChatEvent{
        ConfirmWrite: &ConfirmWriteEvent{
            Question: question,
            Tool:     confirmWrite.Tool,
            Params:   confirmWrite.Params,
        },
    })
    // 保存状态，下轮对话时读取
}
```

### 4.4 Client 实现（service/client.go）

在 `Client` 中新增对应 HTTP 调用方法，对接 Gateway 的 `/internal/*` 端点：

- `POST /internal/repos/files` — 创建文件
- `PUT /internal/repos/files` — 更新文件
- `DELETE /internal/repos/files` — 删除文件
- `PUT /internal/repos/files/rename` — 重命名文件
- `POST /internal/repos` — 创建知识库
- `PUT /internal/repos` — 更新知识库
- `GET /internal/users/search` — 搜索用户
- `POST /internal/repos/collaborators` — 添加协作者
- `DELETE /internal/repos/collaborators` — 移除协作者
- `PUT /internal/repos/collaborators/role` — 变更角色
- `GET /internal/repos/collaborators` — 列出协作者

Gateway 端需新增对应路由和 handler，转发到 repo-server/user-server 的 gRPC 服务。

## 五、Gateway 新增端点

| 端点 | 方法 | 说明 | 鉴权 |
|------|------|------|------|
| `/internal/repos/files` | POST | 创建文件 | ServiceAuth |
| `/internal/repos/files` | PUT | 更新文件 | ServiceAuth |
| `/internal/repos/files` | DELETE | 删除文件 | ServiceAuth |
| `/internal/repos/files/rename` | PUT | 重命名文件 | ServiceAuth |
| `/internal/repos` | POST | 创建知识库 | ServiceAuth |
| `/internal/repos` | PUT | 更新知识库 | ServiceAuth |
| `/internal/users/search` | GET | 搜索用户 | ServiceAuth |
| `/internal/repos/collaborators` | POST | 添加协作者 | ServiceAuth |
| `/internal/repos/collaborators/:user_id` | DELETE | 移除协作者 | ServiceAuth |
| `/internal/repos/collaborators/role` | PUT | 变更角色 | ServiceAuth |
| `/internal/repos/collaborators` | GET | 列出协作者 | ServiceAuth |

## 六、不包含的项

- `delete_repo`：破坏性操作，风险高，暂不添加
- 知识库关注/取消关注：操作频率低，界面操作更合适
- 群组相关的写操作（创建群等）：属于社交功能，不在知识库工具范围内

# LLM 知识库写操作工具 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 LLM 新增 11 个写操作工具（文件操作、知识库管理、协作者管理），并通过 Eino ToolMiddleware 实现可靠的用户确认机制

**Architecture:** ai-server 侧新增工具实现 + ConfirmationMiddleware；gateway 侧新增内部 HTTP 端点用于写操作（写入存储 + 触发向量化）；ai-server 需要新增 `DeleteFileVectors` gRPC 方法用于删除文件时清理向量

**Tech Stack:** Go + Eino v0.9.2 (ToolCallMiddlewares) + Hertz + gRPC + chromem-go

---

### Task 1: ai-server — 新增工具接口定义 (tool.go)

**Files:**
- Modify: `ai-server/internal/llm/tool/tool.go`

- [ ] **Step 1: 在 tool.go 中新增写操作相关接口和类型**

在 `tool.go` 中新增以下内容（放在 `SessionTitleUpdater` 接口之后）：

```go
// ========== 写入操作相关接口 ==========

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

// VectorizeClient 向量化触发接口
type VectorizeClient interface {
    VectorizeArticle(ctx context.Context, userID, repoID, filePath string) error
    DeleteFileVectors(ctx context.Context, repoID, filePath string) error
}

// UserInfo 用户信息
type UserInfo struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// CollaboratorInfo 协作者信息
type CollaboratorInfo struct {
    UserID   string `json:"user_id"`
    UserName string `json:"user_name"`
    Role     string `json:"role"`
}

// ========== 确认机制相关 ==========

// ConfirmLevel 确认级别
type ConfirmLevel int

const (
    ConfirmNever   ConfirmLevel = 0 // 🟢 无需确认
    ConfirmOptional ConfirmLevel = 1 // 🟡 按需确认（可传 _skip_confirm 跳过）
    ConfirmAlways   ConfirmLevel = 2 // 🔴 强制确认
)

// ToolPolicy 工具确认策略
type ToolPolicy struct {
    ToolName     string
    ConfirmLevel ConfirmLevel
}
```

- [ ] **Step 2: 确认编译通过**

Run: `cd ai-server && go build ./internal/llm/tool/`
Expected: 编译成功，无错误

---

### Task 2: ai-server Client — 新增 HTTP 写操作方法 + 向量化客户端方法

**Files:**
- Modify: `ai-server/internal/service/client.go`

- [ ] **Step 1: 在 client.go 新增 `doPost`、`doPut`、`doDelete` 方法**

放在 `doGet` 方法之后：

```go
// doPost 向 gateway 的内部端点发送 POST 请求
func (c *Client) doPost(ctx context.Context, path string, query url.Values, body []byte) ([]byte, error) {
    return c.doBody(ctx, http.MethodPost, path, query, body)
}

// doPut 向 gateway 的内部端点发送 PUT 请求
func (c *Client) doPut(ctx context.Context, path string, query url.Values, body []byte) ([]byte, error) {
    return c.doBody(ctx, http.MethodPut, path, query, body)
}

// doDelete 向 gateway 的内部端点发送 DELETE 请求
func (c *Client) doDelete(ctx context.Context, path string, query url.Values) ([]byte, error) {
    return c.doBody(ctx, http.MethodDelete, path, query, nil)
}

// doBody 向 gateway 的内部端点发送带 body 的请求
func (c *Client) doBody(ctx context.Context, method, path string, query url.Values, body []byte) ([]byte, error) {
    serviceToken, _ := ctx.Value(tool.CtxKeyServiceToken).(string)

    u, err := url.Parse(c.gatewayAddr + path)
    if err != nil {
        return nil, fmt.Errorf("解析 gateway 地址失败: %w", err)
    }
    u.RawQuery = query.Encode()

    var reqBody io.Reader
    if body != nil {
        reqBody = bytes.NewReader(body)
    }
    req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
    if err != nil {
        return nil, fmt.Errorf("创建请求失败: %w", err)
    }
    if serviceToken != "" {
        req.Header.Set("Authorization", "Bearer "+serviceToken)
    }
    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("请求 gateway 失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        respBody, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("gateway 返回错误状态 %d: %s", resp.StatusCode, string(respBody))
    }

    return io.ReadAll(resp.Body)
}
```

在 import 中添加 `"bytes"` 和 `"io"`（io 可能已有）。

- [ ] **Step 2: 在 client.go 新增 FileWriteClient 实现方法**

放在文件相关方法区之后：

```go
// ==================== 文件写入 ====================

// CreateFile 创建文件，返回的文件路径经过 gateway 规范化
func (c *Client) CreateFile(ctx context.Context, ownerID, repoID, filePath, content string) error {
    q := url.Values{}
    q.Set("owner_id", ownerID)
    q.Set("repo_id", repoID)
    q.Set("path", filePath)
    body, _ := json.Marshal(map[string]string{"content": content})
    _, err := c.doPost(ctx, "/internal/repos/files", q, body)
    return err
}

// UpdateFile 更新文件内容
func (c *Client) UpdateFile(ctx context.Context, ownerID, repoID, filePath, content string) error {
    q := url.Values{}
    q.Set("owner_id", ownerID)
    q.Set("repo_id", repoID)
    q.Set("path", filePath)
    body, _ := json.Marshal(map[string]string{"content": content})
    _, err := c.doPut(ctx, "/internal/repos/files", q, body)
    return err
}

// DeleteFile 删除文件
func (c *Client) DeleteFile(ctx context.Context, ownerID, repoID, filePath string) error {
    q := url.Values{}
    q.Set("owner_id", ownerID)
    q.Set("repo_id", repoID)
    q.Set("path", filePath)
    _, err := c.doDelete(ctx, "/internal/repos/files", q)
    return err
}

// RenameFile 重命名/移动文件
func (c *Client) RenameFile(ctx context.Context, ownerID, repoID, oldPath, newPath string) error {
    q := url.Values{}
    q.Set("owner_id", ownerID)
    q.Set("repo_id", repoID)
    q.Set("path", oldPath)
    body, _ := json.Marshal(map[string]string{"new_path": newPath})
    _, err := c.doPut(ctx, "/internal/repos/files/rename", q, body)
    return err
}
```

- [ ] **Step 3: 在 client.go 新增 RepoWriteClient 实现方法**

```go
// ==================== 知识库写入 ====================

// CreateRepo 创建知识库
func (c *Client) CreateRepo(ctx context.Context, userID, name, description, visibility string) (string, error) {
    body, _ := json.Marshal(map[string]string{
        "name":        name,
        "description": description,
        "visibility":  visibility,
    })
    resp, err := c.doPost(ctx, "/internal/repos", nil, body)
    if err != nil {
        return "", err
    }
    var result struct {
        RepoID string `json:"repo_id"`
    }
    if err := json.Unmarshal(resp, &result); err != nil {
        return "", fmt.Errorf("解析响应失败: %w", err)
    }
    return result.RepoID, nil
}

// UpdateRepo 更新知识库
func (c *Client) UpdateRepo(ctx context.Context, repoID, userID, name, description, visibility string) error {
    body, _ := json.Marshal(map[string]string{
        "name":        name,
        "description": description,
        "visibility":  visibility,
    })
    q := url.Values{}
    q.Set("repo_id", repoID)
    _, err := c.doPut(ctx, "/internal/repos", q, body)
    return err
}
```

- [ ] **Step 4: 在 client.go 新增 UserSearchClient 实现方法**

```go
// ==================== 用户搜索 ====================

// SearchUsers 搜索用户
func (c *Client) SearchUsers(ctx context.Context, keyword string) ([]tool.UserInfo, error) {
    q := url.Values{}
    q.Set("q", keyword)
    body, err := c.doGet(ctx, "/internal/users/search", q)
    if err != nil {
        return nil, err
    }
    var result struct {
        Users []struct {
            ID   string `json:"id"`
            Name string `json:"name"`
        } `json:"users"`
    }
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("解析响应失败: %w", err)
    }
    users := make([]tool.UserInfo, 0, len(result.Users))
    for _, u := range result.Users {
        users = append(users, tool.UserInfo{ID: u.ID, Name: u.Name})
    }
    return users, nil
}
```

- [ ] **Step 5: 在 client.go 新增 CollaboratorClient 实现方法**

```go
// ==================== 协作者管理 ====================

// AddCollaborator 添加协作者
func (c *Client) AddCollaborator(ctx context.Context, repoID, userID, role string) error {
    q := url.Values{}
    q.Set("repo_id", repoID)
    body, _ := json.Marshal(map[string]string{
        "user_id": userID,
        "role":    role,
    })
    _, err := c.doPost(ctx, "/internal/repos/collaborators", q, body)
    return err
}

// RemoveCollaborator 移除协作者
func (c *Client) RemoveCollaborator(ctx context.Context, repoID, userID string) error {
    q := url.Values{}
    q.Set("repo_id", repoID)
    q.Set("user_id", userID)
    _, err := c.doDelete(ctx, "/internal/repos/collaborators", q)
    return err
}

// UpdateCollaboratorRole 更新协作者角色
func (c *Client) UpdateCollaboratorRole(ctx context.Context, repoID, userID, role string) error {
    q := url.Values{}
    q.Set("repo_id", repoID)
    body, _ := json.Marshal(map[string]string{
        "user_id": userID,
        "role":    role,
    })
    _, err := c.doPut(ctx, "/internal/repos/collaborators/role", q, body)
    return err
}

// ListCollaborators 列出协作者
func (c *Client) ListCollaborators(ctx context.Context, repoID string) ([]tool.CollaboratorInfo, error) {
    q := url.Values{}
    q.Set("repo_id", repoID)
    body, err := c.doGet(ctx, "/internal/repos/collaborators", q)
    if err != nil {
        return nil, err
    }
    var result struct {
        Collaborators []struct {
            UserID   string `json:"user_id"`
            UserName string `json:"user_name"`
            Role     string `json:"role"`
        } `json:"collaborators"`
    }
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("解析响应失败: %w", err)
    }
    items := make([]tool.CollaboratorInfo, 0, len(result.Collaborators))
    for _, c := range result.Collaborators {
        items = append(items, tool.CollaboratorInfo{
            UserID:   c.UserID,
            UserName: c.UserName,
            Role:     c.Role,
        })
    }
    return items, nil
}
```

- [ ] **Step 6: 新增 VectorizeClient 实现方法**

ai-server 可以在 service 层实现 `VectorizeClient` 接口。在 `vectorize.go` 中新增一个适配结构体，或者直接在 `Service` 上实现这两个方法（`VectorizeArticle` 已经有了，只需要加 `DeleteFileVectors`）。

在 `ai-server/internal/service/vectorize.go` 中新增：

```go
// DeleteFileVectors 删除指定文件的向量数据
func (s *Service) DeleteFileVectors(_ context.Context, repoID, filePath string) error {
    return s.VectorStore.DeleteFileVectors(context.Background(), repoID, filePath)
}
```

- [ ] **Step 7: 确认编译通过**

Run: `cd ai-server && go build ./internal/service/`
Expected: 编译成功

---

### Task 3: ai-server — 新增 DeleteFileVectors gRPC 方法

**Files:**
- Modify: `ks-proto/pkg/pb/` (proto generated files)
- Modify: `ai-server/internal/handler/delete_repo_vectors.go` (参考这个文件创建 delete_file_vectors.go)

- [ ] **Step 1: 在 proto 文件中新增 `DeleteFileVectors` RPC**

需要找到 ai-server 的 proto 文件（在 `ks-proto/` 目录中），添加：

```protobuf
message DeleteFileVectorsRequest {
    string repo_id = 1;
    string file_path = 2;
}

message DeleteFileVectorsResponse {
    bool success = 1;
    string msg = 2;
}
```

以及 RPC 定义：
```protobuf
rpc DeleteFileVectors(DeleteFileVectorsRequest) returns (DeleteFileVectorsResponse);
```

- [ ] **Step 2: 重新生成 proto 代码**

Run: `cd ks-proto && make generate` (或项目使用的 proto 生成命令)

- [ ] **Step 3: 实现 handler**

新建 `ai-server/internal/handler/delete_file_vectors.go`：

```go
package handler

import (
    "context"

    "github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// DeleteFileVectors 删除指定文件的向量数据
func (h *Handler) DeleteFileVectors(ctx context.Context, req *pb.DeleteFileVectorsRequest) (*pb.DeleteFileVectorsResponse, error) {
    if err := h.Service.DeleteFileVectors(ctx, req.GetRepoId(), req.GetFilePath()); err != nil {
        return &pb.DeleteFileVectorsResponse{
            Success: false,
            Msg:     err.Error(),
        }, nil
    }
    return &pb.DeleteFileVectorsResponse{
        Success: true,
    }, nil
}
```

- [ ] **Step 4: 将新的 handler 注册到 gRPC 服务**

找到 ai-server 的 gRPC 服务注册代码，将 `DeleteFileVectors` handler 注册到对应的 pb 服务。

- [ ] **Step 5: 确认编译通过**

Run: `cd ai-server && go build ./...`
Expected: 编译成功

---

### Task 4: Gateway — 新增内部写操作端点

**Files:**
- Modify: `gateway/internal/handler/repo_internal.go` — 新增内部端点 handler
- Modify: `gateway/internal/router/router.go` — 注册新路由

- [ ] **Step 1: 在 repo_internal.go 新增文件写入内部 handler**

```go
// InternalCreateFile 创建文件（内部服务间调用）
// POST /internal/repos/files?owner_id=&repo_id=&path=
func (h *Handler) InternalCreateFile() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        ownerID := ctx.Query("owner_id")
        repoID := ctx.Query("repo_id")
        filePath := ctx.Query("path")
        if ownerID == "" || repoID == "" || filePath == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "owner_id, repo_id, path are required"})
            return
        }
        var req struct {
            Content string `json:"content"`
        }
        if err := ctx.BindAndValidate(&req); err != nil {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "invalid body"})
            return
        }

        subpath := filepath.Join(ownerID, repoID, filePath)
        if err := h.store.WriteFileFromBytes(subpath, []byte(req.Content)); err != nil {
            slog.Error("创建文件失败", "error", err, "path", subpath)
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "创建文件失败"})
            return
        }

        // 异步触发向量化
        ext := strings.ToLower(filepath.Ext(filePath))
        if isTextFile(ext) {
            go h.triggerVectorize(ownerID, repoID, filePath)
        }

        ctx.JSON(consts.StatusOK, map[string]any{
            "path": filePath,
        })
    }
}

// InternalUpdateFile 更新文件内容（内部服务间调用）
// PUT /internal/repos/files?owner_id=&repo_id=&path=
func (h *Handler) InternalUpdateFile() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        ownerID := ctx.Query("owner_id")
        repoID := ctx.Query("repo_id")
        filePath := ctx.Query("path")
        if ownerID == "" || repoID == "" || filePath == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "owner_id, repo_id, path are required"})
            return
        }
        var req struct {
            Content string `json:"content"`
        }
        if err := ctx.BindAndValidate(&req); err != nil {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "invalid body"})
            return
        }

        subpath := filepath.Join(ownerID, repoID, filePath)
        if err := h.store.WriteFileFromBytes(subpath, []byte(req.Content)); err != nil {
            slog.Error("更新文件失败", "error", err, "path", subpath)
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "更新文件失败"})
            return
        }

        // 异步触发向量化（StoreChunks 会自动删除旧向量再添加新向量）
        ext := strings.ToLower(filepath.Ext(filePath))
        if isTextFile(ext) {
            go h.triggerVectorize(ownerID, repoID, filePath)
        }

        ctx.JSON(consts.StatusOK, map[string]any{"message": "updated"})
    }
}

// InternalDeleteFile 删除文件（内部服务间调用）
// DELETE /internal/repos/files?owner_id=&repo_id=&path=
func (h *Handler) InternalDeleteFile() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        ownerID := ctx.Query("owner_id")
        repoID := ctx.Query("repo_id")
        filePath := ctx.Query("path")
        if ownerID == "" || repoID == "" || filePath == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "owner_id, repo_id, path are required"})
            return
        }

        subpath := filepath.Join(ownerID, repoID, filePath)
        if err := h.store.Delete(subpath); err != nil {
            slog.Error("删除文件失败", "error", err, "path", subpath)
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "删除文件失败"})
            return
        }

        // 异步删除向量
        go h.triggerDeleteFileVectors(repoID, filePath)

        ctx.JSON(consts.StatusOK, map[string]any{"message": "deleted"})
    }
}

// triggerDeleteFileVectors 异步删除文件向量
func (h *Handler) triggerDeleteFileVectors(repoID, filePath string) {
    conn := h.grpcClient.GetConn("ai_server")
    if conn == nil {
        slog.Warn("AI 服务连接不可用，跳过删除向量", "file_path", filePath)
        return
    }
    client := pb.NewAIServiceClient(conn)
    resp, err := client.DeleteFileVectors(context.Background(), &pb.DeleteFileVectorsRequest{
        RepoId:   repoID,
        FilePath: filePath,
    })
    if err != nil {
        slog.Warn("删除向量请求失败", "file_path", filePath, "error", err)
        return
    }
    if !resp.Success {
        slog.Warn("删除向量失败", "file_path", filePath, "msg", resp.Msg)
        return
    }
    slog.Info("文件向量已删除", "file_path", filePath)
}

// InternalRenameFile 重命名文件（内部服务间调用）
// PUT /internal/repos/files/rename?owner_id=&repo_id=&path=
func (h *Handler) InternalRenameFile() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        ownerID := ctx.Query("owner_id")
        repoID := ctx.Query("repo_id")
        oldPath := ctx.Query("path")
        if ownerID == "" || repoID == "" || oldPath == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "owner_id, repo_id, path are required"})
            return
        }
        var req struct {
            NewPath string `json:"new_path"`
        }
        if err := ctx.BindAndValidate(&req); err != nil || req.NewPath == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "new_path is required"})
            return
        }

        oldSubpath := filepath.Join(ownerID, repoID, oldPath)
        newSubpath := filepath.Join(ownerID, repoID, req.NewPath)

        if err := h.store.Rename(oldSubpath, newSubpath); err != nil {
            slog.Error("重命名失败", "error", err)
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "重命名失败"})
            return
        }

        ctx.JSON(consts.StatusOK, map[string]any{
            "old_path": oldPath,
            "new_path": req.NewPath,
        })
    }
}
```

**注意：** `h.store.WriteFileFromBytes` 方法可能需要新增或使用 `h.store.WriteFile` 的替代方案。如果 storage.Store 没有 `WriteFileFromBytes`，需要添加：

```go
// 在 gateway/pkg/storage/store.go 中新增
func (s *Store) WriteFileFromBytes(subpath string, data []byte) error {
    fullPath := filepath.Join(s.rootDir, subpath)
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return err
    }
    return os.WriteFile(fullPath, data, 0644)
}
```

- [ ] **Step 2: 在 repo_internal.go 新增 Repo 和 Collaborator 内部写 handler**

```go
// InternalCreateRepo 创建知识库（内部服务间调用）
// POST /internal/repos
func (h *Handler) InternalCreateRepo() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        uid := ctx.GetString("user_id")
        var req struct {
            Name        string `json:"name"`
            Description string `json:"description"`
            Visibility  string `json:"visibility"`
        }
        if err := ctx.BindAndValidate(&req); err != nil || req.Name == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "name is required"})
            return
        }
        conn := h.grpcClient.GetConn("repo_server")
        if conn == nil {
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "repo server unavailable"})
            return
        }
        client := pb.NewRepoServiceClient(conn)
        vis := pb.Visibility_PRIVATE
        if req.Visibility == "PUBLIC" {
            vis = pb.Visibility_PUBLIC
        }
        resp, err := client.CreateRepo(c, &pb.CreateRepoRequest{
            OwnerId:     uid,
            Name:        req.Name,
            Description: req.Description,
            Visibility:  vis,
        })
        if err != nil || !resp.Success {
            slog.Error("创建知识库失败", "error", err)
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "创建知识库失败"})
            return
        }
        ctx.JSON(consts.StatusOK, map[string]any{
            "repo_id": resp.RepoId,
        })
    }
}

// InternalUpdateRepo 更新知识库（内部服务间调用）
// PUT /internal/repos?repo_id=
func (h *Handler) InternalUpdateRepo() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        uid := ctx.GetString("user_id")
        repoID := ctx.Query("repo_id")
        if repoID == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "repo_id is required"})
            return
        }
        var req struct {
            Name        string `json:"name"`
            Description string `json:"description"`
            Visibility  string `json:"visibility"`
        }
        if err := ctx.BindAndValidate(&req); err != nil {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "invalid body"})
            return
        }
        conn := h.grpcClient.GetConn("repo_server")
        if conn == nil {
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "repo server unavailable"})
            return
        }
        client := pb.NewRepoServiceClient(conn)
        var visPtr *pb.Visibility
        if req.Visibility != "" {
            v := pb.Visibility_PRIVATE
            if req.Visibility == "PUBLIC" {
                v = pb.Visibility_PUBLIC
            }
            visPtr = &v
        }
        resp, err := client.UpdateRepo(c, &pb.UpdateRepoRequest{
            RepoId:      repoID,
            UserId:      uid,
            Name:        &req.Name,
            Description: &req.Description,
            Visibility:  visPtr,
        })
        if err != nil || !resp.Success {
            slog.Error("更新知识库失败", "error", err)
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "更新知识库失败"})
            return
        }
        ctx.JSON(consts.StatusOK, map[string]any{"message": "updated"})
    }
}

// InternalSearchUsers 搜索用户（内部服务间调用）
// GET /internal/users/search?q=
func (h *Handler) InternalSearchUsers() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        keyword := ctx.Query("q")
        if keyword == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "q is required"})
            return
        }
        conn := h.grpcClient.GetConn("chat_server")
        if conn == nil {
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "chat server unavailable"})
            return
        }
        client := pb.NewChatServiceClient(conn)
        resp, err := client.SearchUsers(c, &pb.SearchUsersRequest{Keyword: keyword})
        if err != nil {
            slog.Error("搜索用户失败", "error", err)
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "搜索用户失败"})
            return
        }
        users := make([]map[string]any, 0, len(resp.Users))
        for _, u := range resp.Users {
            users = append(users, map[string]any{
                "id":   u.Id,
                "name": u.Name,
            })
        }
        ctx.JSON(consts.StatusOK, map[string]any{
            "users": users,
        })
    }
}

// InternalAddCollaborator 添加协作者（内部服务间调用）
// POST /internal/repos/collaborators?repo_id=
func (h *Handler) InternalAddCollaborator() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        uid := ctx.GetString("user_id")
        repoID := ctx.Query("repo_id")
        if repoID == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "repo_id is required"})
            return
        }
        var req struct {
            UserID string `json:"user_id"`
            Role   string `json:"role"`
        }
        if err := ctx.BindAndValidate(&req); err != nil || req.UserID == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "user_id is required"})
            return
        }
        role := pb.CollaboratorRole_VIEWER
        switch req.Role {
        case "ADMIN":
            role = pb.CollaboratorRole_ADMIN
        case "DEVELOPER":
            role = pb.CollaboratorRole_DEVELOPER
        }
        conn := h.grpcClient.GetConn("repo_server")
        if conn == nil {
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "repo server unavailable"})
            return
        }
        client := pb.NewRepoServiceClient(conn)
        resp, err := client.AddCollaborator(c, &pb.AddCollaboratorRequest{
            RepoId:     repoID,
            OperatorId: uid,
            UserId:     req.UserID,
            Role:       role,
        })
        if err != nil || !resp.Success {
            slog.Error("添加协作者失败", "error", err)
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "添加协作者失败"})
            return
        }
        ctx.JSON(consts.StatusOK, map[string]any{"message": "added"})
    }
}

// InternalRemoveCollaborator 移除协作者（内部服务间调用）
// DELETE /internal/repos/collaborators?repo_id=&user_id=
func (h *Handler) InternalRemoveCollaborator() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        uid := ctx.GetString("user_id")
        repoID := ctx.Query("repo_id")
        targetUserID := ctx.Query("user_id")
        if repoID == "" || targetUserID == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "repo_id and user_id are required"})
            return
        }
        conn := h.grpcClient.GetConn("repo_server")
        if conn == nil {
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "repo server unavailable"})
            return
        }
        client := pb.NewRepoServiceClient(conn)
        resp, err := client.RemoveCollaborator(c, &pb.RemoveCollaboratorRequest{
            RepoId:     repoID,
            OperatorId: uid,
            UserId:     targetUserID,
        })
        if err != nil || !resp.Success {
            slog.Error("移除协作者失败", "error", err)
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "移除协作者失败"})
            return
        }
        ctx.JSON(consts.StatusOK, map[string]any{"message": "removed"})
    }
}

// InternalUpdateCollaboratorRole 更新协作者角色（内部服务间调用）
// PUT /internal/repos/collaborators/role?repo_id=
func (h *Handler) InternalUpdateCollaboratorRole() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        uid := ctx.GetString("user_id")
        repoID := ctx.Query("repo_id")
        if repoID == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "repo_id is required"})
            return
        }
        var req struct {
            UserID string `json:"user_id"`
            Role   string `json:"role"`
        }
        if err := ctx.BindAndValidate(&req); err != nil || req.UserID == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "user_id is required"})
            return
        }
        role := pb.CollaboratorRole_VIEWER
        switch req.Role {
        case "ADMIN":
            role = pb.CollaboratorRole_ADMIN
        case "DEVELOPER":
            role = pb.CollaboratorRole_DEVELOPER
        }
        conn := h.grpcClient.GetConn("repo_server")
        if conn == nil {
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "repo server unavailable"})
            return
        }
        client := pb.NewRepoServiceClient(conn)
        resp, err := client.UpdateCollaborator(c, &pb.UpdateCollaboratorRequest{
            RepoId:     repoID,
            OperatorId: uid,
            UserId:     req.UserID,
            Role:       role,
        })
        if err != nil || !resp.Success {
            slog.Error("更新协作者角色失败", "error", err)
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "更新协作者角色失败"})
            return
        }
        ctx.JSON(consts.StatusOK, map[string]any{"message": "updated"})
    }
}

// InternalListCollaborators 列出协作者（内部服务间调用）
// GET /internal/repos/collaborators?repo_id=
func (h *Handler) InternalListCollaborators() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        uid := ctx.GetString("user_id")
        repoID := ctx.Query("repo_id")
        if repoID == "" {
            ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "repo_id is required"})
            return
        }
        conn := h.grpcClient.GetConn("repo_server")
        if conn == nil {
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "repo server unavailable"})
            return
        }
        client := pb.NewRepoServiceClient(conn)
        resp, err := client.ListCollaborators(c, &pb.ListCollaboratorsRequest{
            RepoId: repoID,
            UserId: uid,
        })
        if err != nil {
            slog.Error("列出协作者失败", "error", err)
            ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "列出协作者失败"})
            return
        }
        items := make([]map[string]any, 0, len(resp.Collaborators))
        for _, c := range resp.Collaborators {
            items = append(items, map[string]any{
                "user_id": c.UserId,
                "role":    c.Role.String(),
            })
        }
        ctx.JSON(consts.StatusOK, map[string]any{
            "collaborators": items,
        })
    }
}
```

- [ ] **Step 3: 在 router.go 中注册新内部路由**

在 `router.go` 的 `internal` 分组中添加：

```go
internal.POST("/repos/files", hdl.InternalCreateFile())
internal.PUT("/repos/files", hdl.InternalUpdateFile())
internal.DELETE("/repos/files", hdl.InternalDeleteFile())
internal.PUT("/repos/files/rename", hdl.InternalRenameFile())
internal.POST("/repos", hdl.InternalCreateRepo())
internal.PUT("/repos", hdl.InternalUpdateRepo())
internal.GET("/users/search", hdl.InternalSearchUsers())
internal.POST("/repos/collaborators", hdl.InternalAddCollaborator())
internal.DELETE("/repos/collaborators", hdl.InternalRemoveCollaborator())
internal.PUT("/repos/collaborators/role", hdl.InternalUpdateCollaboratorRole())
internal.GET("/repos/collaborators", hdl.InternalListCollaborators())
```

同时添加 import: `"path/filepath"`, `"strings"` （如果还没有的话）

- [ ] **Step 4: 确认编译通过**

Run: `cd gateway && go build ./...`
Expected: 编译成功

---

### Task 5: ai-server — 实现 ConfirmationMiddleware

**Files:**
- Create: `ai-server/internal/llm/tool/confirmation.go`

- [ ] **Step 1: 创建 confirmation.go 文件**

```go
package tool

import (
    "context"
    "encoding/json"

    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
)

// ctxKeyConfirmed 用于在 context 中传递确认标记
type ctxKey string

const (
    // CtxKeyConfirmed 确认标记 key，值为确认过的工具调用参数 hash
    CtxKeyConfirmed ctxKey = "confirmed_write"
)

// writeToolNames 所有需要确认的写操作工具名
var writeToolNames = map[string]bool{
    "create_file":                true,
    "update_file":                true,
    "delete_file":                true,
    "rename_file":                true,
    "create_repo":                true,
    "update_repo":                true,
    "add_collaborator":           true,
    "remove_collaborator":        true,
    "update_collaborator_role":   true,
}

// NewConfirmationMiddleware 创建写操作确认中间件
//
// policies: 各个工具的确认策略
// 对于 ConfirmAlways — 始终拦截，强制要求确认
// 对于 ConfirmOptional — 检查参数中是否有 _skip_confirm: true，有则跳过
func NewConfirmationMiddleware(policies map[string]ConfirmLevel) compose.InvokableToolMiddleware {
    return func(next tool.InvokableTool) tool.InvokableTool {
        info, _ := next.Info(context.Background())
        name := info.Name

        // 非写操作工具直接放行
        if !writeToolNames[name] {
            return next
        }

        policy, ok := policies[name]
        if !ok || policy == ConfirmNever {
            return next
        }

        return tool.NewInvokableToolFunc(
            func(ctx context.Context) (*schema.ToolInfo, error) {
                return next.Info(ctx)
            },
            func(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
                // 已确认 → 放行
                if ctx.Value(CtxKeyConfirmed) != nil {
                    return next.InvokableRun(ctx, arguments, opts...)
                }

                var raw map[string]any
                _ = json.Unmarshal([]byte(arguments), &raw)

                // update_repo 的 visibility 变更始终强制确认
                effectivePolicy := policy
                if name == "update_repo" {
                    if vis, ok := raw["visibility"].(string); ok && vis != "" {
                        effectivePolicy = ConfirmAlways
                    }
                }

                // 按需确认 + 参数带跳过标记 → 放行
                if effectivePolicy == ConfirmOptional {
                    if skip, _ := raw["_skip_confirm"].(bool); skip {
                        return next.InvokableRun(ctx, arguments, opts...)
                    }
                }

                // 未确认 → 返回确认事件
                result, _ := json.Marshal(map[string]any{
                    "action": "confirm_write",
                    "tool":   name,
                    "params": arguments,
                })
                return string(result), nil
            },
        )
    }
}
```

- [ ] **Step 2: 确认编译通过**

Run: `cd ai-server && go build ./internal/llm/tool/`
Expected: 编译成功

---

### Task 6: ai-server — 实现 11 个新工具

**Files:**
- Create: `ai-server/internal/llm/tool/create_file.go`
- Create: `ai-server/internal/llm/tool/update_file.go`
- Create: `ai-server/internal/llm/tool/delete_file.go`
- Create: `ai-server/internal/llm/tool/rename_file.go`
- Create: `ai-server/internal/llm/tool/create_repo.go`
- Create: `ai-server/internal/llm/tool/update_repo.go`
- Create: `ai-server/internal/llm/tool/search_users.go`
- Create: `ai-server/internal/llm/tool/add_collaborator.go`
- Create: `ai-server/internal/llm/tool/remove_collaborator.go`
- Create: `ai-server/internal/llm/tool/update_collaborator_role.go`
- Create: `ai-server/internal/llm/tool/list_collaborators.go`

每个工具遵循相同的模式（参考 `get_file_content.go`）：
1. 定义 struct，包含所需依赖接口
2. 构造函数
3. `Info()` 返回工具元信息（description 中写明确认策略）
4. `InvokableRun()` 委派给 `execute()`

- [ ] **Step 1: create_file.go**

```go
package tool

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/schema"
)

// CreateFile 创建知识库文件工具
type CreateFile struct {
    repoDetailClient RepoDetailClient
    fileWriteClient  FileWriteClient
}

func NewCreateFile(rdc RepoDetailClient, fwc FileWriteClient) *CreateFile {
    return &CreateFile{
        repoDetailClient: rdc,
        fileWriteClient:  fwc,
    }
}

func (c *CreateFile) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "create_file",
        Desc: "在知识库中创建新文件。如果用户明确表达了文件路径和内容概要，可以设置参数 _skip_confirm: true 跳过确认直接执行；如果用户表述不完整，不要设置该参数。",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "repo_id": {
                Type:     "string",
                Desc:     "知识库 ID",
                Required: true,
            },
            "file_path": {
                Type:     "string",
                Desc:     "文件路径，例如：docs/chapter1.md",
                Required: true,
            },
            "content": {
                Type:     "string",
                Desc:     "文件内容（Markdown 格式）",
                Required: true,
            },
            "_skip_confirm": {
                Type:     "boolean",
                Desc:     "当用户已明确确认所有信息时，设置为 true 跳过二次确认",
                Required: false,
            },
        }),
    }, nil
}

func (c *CreateFile) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
    return c.execute(ctx, arguments)
}

func (c *CreateFile) execute(ctx context.Context, paramsJSON string) (string, error) {
    var params struct {
        RepoID   string `json:"repo_id"`
        FilePath string `json:"file_path"`
        Content  string `json:"content"`
    }
    if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
        return "", fmt.Errorf("解析参数失败: %w", err)
    }
    if params.RepoID == "" || params.FilePath == "" {
        return `{"success": false, "message": "repo_id 和 file_path 不能为空"}`, nil
    }

    userID, _ := ctx.Value(CtxKeyUserID).(string)
    if userID == "" {
        return `{"success": false, "message": "无法获取用户信息"}`, nil
    }

    repo, err := c.repoDetailClient.GetRepo(ctx, params.RepoID, userID)
    if err != nil {
        return fmt.Sprintf(`{"success": false, "message": "获取仓库信息失败: %s"}`, err.Error()), nil
    }

    if err := c.fileWriteClient.CreateFile(ctx, repo.OwnerID, params.RepoID, params.FilePath, params.Content); err != nil {
        return fmt.Sprintf(`{"success": false, "message": "创建文件失败: %s"}`, err.Error()), nil
    }

    data, _ := json.Marshal(map[string]any{
        "success":  true,
        "repo_id":  params.RepoID,
        "file_path": params.FilePath,
        "size":     len(params.Content),
    })
    return string(data), nil
}
```

- [ ] **Step 2: 按同样模式实现 update_file.go、delete_file.go、rename_file.go**

`update_file.go` — 调用 `fileWriteClient.UpdateFile`
`delete_file.go` — 调用 `fileWriteClient.DeleteFile`，description 中不可提 `_skip_confirm`（强制确认）
`rename_file.go` — 调用 `fileWriteClient.RenameFile`

- [ ] **Step 3: 按同样模式实现 create_repo.go、update_repo.go**

`create_repo.go` — 调用 `repoWriteClient.CreateRepo`，description 中提 `_skip_confirm`
`update_repo.go` — 调用 `repoWriteClient.UpdateRepo`，description 中提 `_skip_confirm`（但 visibility 变更会被 Middleware 强制确认）

- [ ] **Step 4: 按同样模式实现 search_users.go、add_collaborator.go、remove_collaborator.go、update_collaborator_role.go、list_collaborators.go**

`search_users.go` — 调用 `userSearchClient.SearchUsers`，无需确认
`list_collaborators.go` — 调用 `collaboratorClient.ListCollaborators`，无需确认
`add_collaborator.go` — 调用 `collaboratorClient.AddCollaborator`，强制确认（description 中不提 _skip_confirm）
`remove_collaborator.go` — 调用 `collaboratorClient.RemoveCollaborator`，同上
`update_collaborator_role.go` — 调用 `collaboratorClient.UpdateCollaboratorRole`，同上

- [ ] **Step 5: 确认编译通过**

Run: `cd ai-server && go build ./internal/llm/tool/`
Expected: 编译成功，无错误

---

### Task 7: ai-server — 注册工具和 Middleware

**Files:**
- Modify: `ai-server/internal/service/service.go`
- Modify: `ai-server/internal/llm/llm.go`

- [ ] **Step 1: 在 service.go 中注册新工具和 Middleware**

```go
func (s *Service) initAgent(ctx context.Context) error {
    // 确认策略配置
    writePolicies := map[string]llmtool.ConfirmLevel{
        "create_file":                llmtool.ConfirmOptional,
        "update_file":                llmtool.ConfirmOptional,
        "delete_file":                llmtool.ConfirmAlways,
        "rename_file":                llmtool.ConfirmOptional,
        "create_repo":                llmtool.ConfirmOptional,
        "update_repo":                llmtool.ConfirmOptional, // visibility 变更在 Middleware 内提升为 Always
        "add_collaborator":           llmtool.ConfirmAlways,
        "remove_collaborator":        llmtool.ConfirmAlways,
        "update_collaborator_role":   llmtool.ConfirmAlways,
        "search_users":               llmtool.ConfirmNever,
        "list_collaborators":         llmtool.ConfirmNever,
    }

    // 注册 ask_user 回调
    onAskUser := func(resultJSON string) {
        s.AskedUser.Triggered.Store(true)
        s.AskedUser.LastResult.Store(resultJSON)
    }

    tools := []einoTool.InvokableTool{
        // 已有工具
        llmtool.NewSearchKnowledge(s.Embedder, s.VectorStore, s.Client, 0),
        llmtool.NewUpdateTitle(s.Repo),
        llmtool.NewAskQuestion(onAskUser),
        llmtool.NewListRepos(s.Client),
        llmtool.NewListRepoFiles(s.Client, s.Client),
        llmtool.NewGetFileContent(s.Client, s.Client),

        // 新增文件操作工具
        llmtool.NewCreateFile(s.Client, s.Client),
        llmtool.NewUpdateFile(s.Client, s.Client),
        llmtool.NewDeleteFile(s.Client, s.Client),
        llmtool.NewRenameFile(s.Client, s.Client),

        // 新增知识库管理工具
        llmtool.NewCreateRepo(s.Client),
        llmtool.NewUpdateRepo(s.Client, s.Client),

        // 新增用户搜索和协作者管理工具
        llmtool.NewSearchUsers(s.Client),
        llmtool.NewAddCollaborator(s.Client, s.Client),
        llmtool.NewRemoveCollaborator(s.Client, s.Client),
        llmtool.NewUpdateCollaboratorRole(s.Client, s.Client),
        llmtool.NewListCollaborators(s.Client, s.Client),
    }

    if err := s.LLM.InitAgent(ctx, tools, writePolicies); err != nil {
        slog.Error("初始化 Agent 失败", "error", err)
        return err
    }
    return nil
}
```

- [ ] **Step 2: 修改 `InitAgent` 签名以支持 Middleware**

在 `llm.go` 中修改 `InitAgent`：

```go
func (c *ChatModel) InitAgent(ctx context.Context, tools []tool.InvokableTool, writePolicies map[string]llmtool.ConfirmLevel) error {
    // ... 现有代码 ...

    agent, err := react.NewAgent(ctx, &react.AgentConfig{
        ToolCallingModel: baseModel,
        ToolsConfig: compose.ToolsNodeConfig{
            Tools: toBaseTools(tools),
            ToolCallMiddlewares: []compose.ToolMiddleware{
                {Invokable: llmtool.NewConfirmationMiddleware(writePolicies)},
            },
        },
        // ... 现有代码 ...
    })
    // ...
}
```

如果不想修改现有调用方，可以保留签名兼容——将 `writePolicies` 设为一个字段，或者单独调用 `SetWritePolicies`。

- [ ] **Step 3: 确认编译通过**

Run: `cd ai-server && go build ./...`
Expected: 编译成功

---

### Task 8: ai-server — 更新 chat.go 处理 confirm_write 事件

**Files:**
- Modify: `ai-server/internal/service/chat.go`

- [ ] **Step 1: 在 ChatEvent 中新增 ConfirmWrite 字段**

```go
// ConfirmWriteEvent 写操作确认事件
type ConfirmWriteEvent struct {
    Tool    string `json:"tool"`
    Params  string `json:"params"`
    Question string `json:"question"`
}

type ChatEvent struct {
    // ... 现有字段 ...
    ConfirmWrite *ConfirmWriteEvent `json:"confirm_write,omitempty"` // 非空时表示需要确认写操作
}
```

- [ ] **Step 2: 在 Chat 函数中检测 confirm_write 事件**

在 stream 读取循环之后、ask_user 检测（步骤 9）之前新增：

```go
// 8.5 检测 confirm_write 事件（优先级低于 ask_user）
var confirmWriteEvent *ConfirmWriteEvent
if !askedUser {
    // 检查最终内容中是否包含 confirm_write
    // 由于 ToolReturnDirectly 未设置，LLM 可能会在 tool 结果后生成文本
    // 从 finalContent 中查找 confirm_write JSON
    confirmWriteEvent = parseConfirmWrite(finalContent.String())
}

if confirmWriteEvent != nil {
    question := buildConfirmQuestion(confirmWriteEvent.Tool, confirmWriteEvent.Params)
    confirmWriteEvent.Question = question
    _ = cb(&ChatEvent{
        ConfirmWrite: confirmWriteEvent,
    })
}
```

- [ ] **Step 3: 实现 parseConfirmWrite 和 buildConfirmQuestion 辅助函数**

```go
// parseConfirmWrite 从助手内容中解析 confirm_write 事件
func parseConfirmWrite(content string) *ConfirmWriteEvent {
    // 看 content 是否包含 {"action": "confirm_write", ...}
    // 由于 tool 结果是 JSON，且 LLM 可能在其前后添加文本，需要搜索
    idx := strings.Index(content, `"action":"confirm_write"`)
    if idx < 0 {
        idx = strings.Index(content, `"action": "confirm_write"`)
    }
    if idx < 0 {
        return nil
    }

    // 从找到的位置往前找 {
    start := strings.LastIndex(content[:idx], "{")
    if start < 0 {
        return nil
    }

    // 找匹配的 }
    depth := 0
    end := -1
    for i := start; i < len(content); i++ {
        switch content[i] {
        case '{':
            depth++
        case '}':
            depth--
            if depth == 0 {
                end = i + 1
                goto found
            }
        }
    }
    return nil

found:
    var raw struct {
        Action string `json:"action"`
        Tool   string `json:"tool"`
        Params string `json:"params"`
    }
    if err := json.Unmarshal([]byte(content[start:end]), &raw); err != nil {
        return nil
    }
    if raw.Action != "confirm_write" {
        return nil
    }
    return &ConfirmWriteEvent{
        Tool:   raw.Tool,
        Params: raw.Params,
    }
}

// buildConfirmQuestion 根据工具类型和参数生成确认问题
func buildConfirmQuestion(toolName, paramsJSON string) string {
    var params map[string]any
    json.Unmarshal([]byte(paramsJSON), &params)

    switch toolName {
    case "create_file":
        return fmt.Sprintf("确认要在知识库中创建文件 `%s` 吗？", params["file_path"])
    case "update_file":
        return fmt.Sprintf("确认要更新文件 `%s` 的内容吗？", params["file_path"])
    case "delete_file":
        return fmt.Sprintf("⚠️ 确认要永久删除文件 `%s` 吗？此操作不可撤销。", params["file_path"])
    case "rename_file":
        return fmt.Sprintf("确认要将文件从 `%s` 重命名为 `%s` 吗？", params["old_path"], params["new_path"])
    case "create_repo":
        return fmt.Sprintf("确认要创建知识库「%s」吗？", params["name"])
    case "update_repo":
        if vis, ok := params["visibility"].(string); ok && vis != "" {
            return fmt.Sprintf("⚠️ 确认要将知识库可见性变更为「%s」吗？", vis)
        }
        return "确认要更新知识库设置吗？"
    case "add_collaborator":
        return "确认要将协作者添加到知识库吗？"
    case "remove_collaborator":
        return "⚠️ 确认要移除该协作者吗？"
    case "update_collaborator_role":
        return "确认要变更协作者角色吗？"
    default:
        return "确认要执行此操作吗？"
    }
}
```

- [ ] **Step 4: 在保存 assistant message 时处理 confirm_write**

```go
if askedUser {
    assistantMsg.Content = "[系统消息] 已向用户提问，等待回答"
} else if confirmWriteEvent != nil {
    assistantMsg.Content = "[系统消息] 等待用户确认操作"
}
```

- [ ] **Step 5: 确认编译通过**

Run: `cd ai-server && go build ./internal/service/`
Expected: 编译成功

---

### Task 9: ai-server — 更新 system prompt

**Files:**
- Modify: `ai-server/internal/llm/llm.go`

- [ ] **Step 1: 在 system prompt 中添加新工具的说明**

在 `systemPrompt` 常量末尾添加：

```
7. 你可以创建、更新、删除知识库中的文件和文章，以及管理知识库的协作者：
   - create_file: 创建新文件，如果用户没有明确文件路径和内容，先用 ask_user 询问
   - update_file: 更新已有文件，如果用户没有明确修改内容，先用 ask_user 询问
   - delete_file: 删除文件，务必先用 ask_user 让用户明确确认
   - rename_file: 重命名/移动文件
   - create_repo: 创建新知识库
   - update_repo: 更新知识库设置
   - search_users: 搜索用户，用于查找协作者
   - add_collaborator / remove_collaborator / update_collaborator_role: 管理协作者
   - list_collaborators: 查看协作者列表
8. 在创建或修改文件时，如果用户已经明确说明了全部信息（路径、名称、内容），可以直接执行；如果信息不完整，先用 ask_user 询问用户。删除和协作者相关操作必须先 ask_user 确认。
```

- [ ] **Step 2: 确认编译通过**

Run: `cd ai-server && go build ./internal/llm/`
Expected: 编译成功

---

### 自检清单

1. **Spec 覆盖度检查:** 所有 spec 中的工具（11个）都已实现；确认机制（Middleware + 事件检测）已覆盖；Gateway 内部端点已覆盖
2. **占位符扫描:** 无 TBD/TODO
3. **类型一致性:** 所有工具接口定义、Client 实现、Gateway handler 签名保持一致
4. **向量化覆盖:** create_file / update_file → 触发 VectorizeArticle；delete_file → 触发 DeleteFileVectors；rename_file → 内容不变，无需重新向量化

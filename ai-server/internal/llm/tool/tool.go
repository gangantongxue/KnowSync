package tool

import (
	"context"

	"github.com/cloudwego/eino/components/tool"

	"github.com/gangantongxue/knowsync/ai-server/internal/vectorstore"
)

// Context key 类型，用于在 context 中传递请求级参数
type ctxKey string

// CtxKeyUserID 上下文键：当前用户 ID
const CtxKeyUserID ctxKey = "user_id"

// CtxKeySessionID 上下文键：当前会话 ID
const CtxKeySessionID ctxKey = "chat_session_id"

// CtxKeyServiceToken 上下文键：service token（gateway 签发，用于内部服务间鉴权）
const CtxKeyServiceToken ctxKey = "service_token"

// ToolList 工具列表别名
type ToolList []tool.InvokableTool

// ========== 数据结构 ==========

// RepoInfo 仓库基本信息
type RepoInfo struct {
	ID           string
	OwnerID      string
	Name         string
	Visibility   string
	Description  string
	ArticleCount int64
}

// FileEntry 文件/目录项
type FileEntry struct {
	Name string
	Type string // "file" | "dir"
	Size int64
	Path string // 完整路径
}

// ========== 依赖接口定义 ==========

// Embedder 向量化接口
type Embedder interface {
	EmbedStrings(ctx context.Context, texts []string) ([][]float64, error)
}

// VectorStore 向量搜索接口
type VectorStore interface {
	SearchCrossRepos(ctx context.Context, repoIDs []string, embedding []float32, limit int) ([]vectorstore.SearchResult, error)
}

// RepoClient 仓库服务客户端接口
type RepoClient interface {
	ListUserRepos(ctx context.Context, userID string) ([]string, error)
	ListPublicRepos(ctx context.Context) ([]string, error)
}

// RepoDetailClient 仓库详情客户端接口
type RepoDetailClient interface {
	GetRepo(ctx context.Context, repoID, userID string) (*RepoInfo, error)
	ListUserReposDetail(ctx context.Context, userID string) ([]RepoInfo, error)
}

// FileClient 文件操作客户端接口
type FileClient interface {
	ListRepoFiles(ctx context.Context, ownerID, repoID, dirPath string) ([]FileEntry, error)
	GetFileContent(ctx context.Context, ownerID, repoID, filePath string) (string, error)
}

// SessionTitleUpdater 会话标题更新接口
type SessionTitleUpdater interface {
	UpdateSessionTitle(sessionID, title string) error
}

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

// RepoDetail 仓库详情（包含角色和关注信息）
type RepoDetail struct {
	RepoInfo
	FollowerCount int64
	MyRole        string
	IsFollowing   bool
}

// PublicRepoClient 公开仓库列表客户端接口
type PublicRepoClient interface {
	ListPublicReposDetail(ctx context.Context) ([]RepoInfo, error)
}

// RepoDetailGetter 仓库详情获取接口（含角色和关注状态）
type RepoDetailGetter interface {
	GetRepoDetail(ctx context.Context, repoID, userID string) (*RepoDetail, error)
}

// FollowClient 关注操作客户端接口
type FollowClient interface {
	FollowRepo(ctx context.Context, userID, repoID string) error
	UnfollowRepo(ctx context.Context, userID, repoID string) error
	ListFollowedRepos(ctx context.Context, userID string) ([]RepoInfo, error)
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
	ConfirmNever    ConfirmLevel = iota // 无需确认
	ConfirmOptional                     // 按需确认（可传 _skip_confirm 跳过）
	ConfirmAlways                       // 强制确认
)

// ToolPolicy 工具确认策略
type ToolPolicy struct {
	ToolName     string
	ConfirmLevel ConfirmLevel
}

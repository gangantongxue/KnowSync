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

// ToolList 工具列表别名
type ToolList []tool.InvokableTool

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

// SessionTitleUpdater 会话标题更新接口
type SessionTitleUpdater interface {
	UpdateSessionTitle(sessionID, title string) error
}

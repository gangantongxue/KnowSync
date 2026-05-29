package tool

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"

	"github.com/gangantongxue/knowsync/ai-server/internal/vectorstore"
)

// Tool 定义单个可调用工具
type Tool struct {
	Name        string
	Description string
	ParamsJSON  json.RawMessage
	Handler     func(ctx context.Context, paramsJSON string) (string, error)
}

// Set 工具集合
type Set struct {
	tools map[string]*Tool
}

// NewSet 创建工具集合
func NewSet() *Set {
	return &Set{tools: make(map[string]*Tool)}
}

// Register 注册工具
func (s *Set) Register(tool *Tool) {
	s.tools[tool.Name] = tool
}

// GetToolInfos 转换为 Eino ToolInfo 列表
func (s *Set) GetToolInfos() []*schema.ToolInfo {
	infos := make([]*schema.ToolInfo, 0, len(s.tools))
	for _, t := range s.tools {
		info := &schema.ToolInfo{
			Name: t.Name,
			Desc: t.Description,
		}
		if len(t.ParamsJSON) > 0 {
			var paramsSchema jsonschema.Schema
			if err := json.Unmarshal(t.ParamsJSON, &paramsSchema); err == nil {
				info.ParamsOneOf = schema.NewParamsOneOfByJSONSchema(&paramsSchema)
			} else {
				slog.Warn("解析工具参数 JSON Schema 失败", "tool", t.Name, "error", err)
			}
		}
		infos = append(infos, info)
	}
	return infos
}

// Execute 执行指定工具
func (s *Set) Execute(ctx context.Context, name, paramsJSON string) (string, error) {
	tool, ok := s.tools[name]
	if !ok {
		return "", &ErrUnknownTool{Name: name}
	}
	return tool.Handler(ctx, paramsJSON)
}

// ErrUnknownTool 未知工具错误
type ErrUnknownTool struct {
	Name string
}

func (e *ErrUnknownTool) Error() string {
	return "未知工具: " + e.Name
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

// SessionTitleUpdater 会话标题更新接口
type SessionTitleUpdater interface {
	UpdateSessionTitle(sessionID, title string) error
}

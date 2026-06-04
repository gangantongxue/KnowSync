package service

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"

	einoTool "github.com/cloudwego/eino/components/tool"

	"github.com/gangantongxue/knowsync/ai-server/internal/embedder"
	"github.com/gangantongxue/knowsync/ai-server/internal/llm"
	llmtool "github.com/gangantongxue/knowsync/ai-server/internal/llm/tool"
	"github.com/gangantongxue/knowsync/ai-server/internal/repository"
	"github.com/gangantongxue/knowsync/ai-server/internal/vectorstore"
	"github.com/gangantongxue/knowsync/ai-server/pkg/config/model"
)

// Service 业务逻辑层
type Service struct {
	Cfg         *model.Config
	RDB         *redis.Client
	Client      *Client
	Repo        *repository.Repository
	LLM         *llm.ChatModel
	Embedder    *embedder.Client
	VectorStore *vectorstore.Store
	AskedUser   *llm.AskedUser // ask_user 工具回调检测
}

// NewService 创建业务逻辑层
func NewService(cfg *model.Config, rdb *redis.Client, client *Client, repo *repository.Repository, llmModel *llm.ChatModel, emb *embedder.Client, vs *vectorstore.Store) (*Service, error) {
	svc := &Service{
		Cfg:         cfg,
		RDB:         rdb,
		Client:      client,
		Repo:        repo,
		LLM:         llmModel,
		Embedder:    emb,
		VectorStore: vs,
		AskedUser:   &llm.AskedUser{},
	}

	// 初始化 Agent 并注册工具
	if err := svc.initAgent(context.Background()); err != nil {
		return nil, err
	}

	return svc, nil
}

// initAgent 用工具列表初始化 ReAct Agent（全局初始化一次）
func (s *Service) initAgent(ctx context.Context) error {
	tools := []einoTool.InvokableTool{
		llmtool.NewSearchKnowledge(s.Embedder, s.VectorStore, s.Client, 0),
		llmtool.NewUpdateTitle(s.Repo),
		llmtool.NewAskQuestion(func(resultJSON string) {
			s.AskedUser.Triggered.Store(true)
			s.AskedUser.LastResult.Store(resultJSON)
		}),
		llmtool.NewListRepos(s.Client),
		llmtool.NewListRepoFiles(s.Client, s.Client),
		llmtool.NewGetFileContent(s.Client, s.Client),
	}

	if err := s.LLM.InitAgent(ctx, tools); err != nil {
		slog.Error("初始化 Agent 失败", "error", err)
		return err
	}
	return nil
}

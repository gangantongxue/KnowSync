package service

import (
	"github.com/redis/go-redis/v9"

	"github.com/gangantongxue/knowsync/ai-server/internal/embedder"
	"github.com/gangantongxue/knowsync/ai-server/internal/llm"
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
}

// NewService 创建业务逻辑层
func NewService(cfg *model.Config, rdb *redis.Client, client *Client, repo *repository.Repository, llmModel *llm.ChatModel, emb *embedder.Client, vs *vectorstore.Store) (*Service, error) {
	return &Service{
		Cfg:         cfg,
		RDB:         rdb,
		Client:      client,
		Repo:        repo,
		LLM:         llmModel,
		Embedder:    emb,
		VectorStore: vs,
	}, nil
}

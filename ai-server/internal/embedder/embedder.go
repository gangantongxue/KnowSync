package embedder

import (
	"context"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"

	"github.com/gangantongxue/knowsync/ai-server/pkg/config/model"
)

// Client 向量化模型客户端（基于 Eino OpenAI Embedder）
type Client struct {
	embedder embedding.Embedder
}

// NewClient 创建向量化模型客户端
func NewClient(cfg *model.EmbedderCfg) (*Client, error) {
	dim := cfg.Dimensions
	e, err := openai.NewEmbedder(context.Background(), &openai.EmbeddingConfig{
		APIKey:     cfg.APIKey,
		BaseURL:    cfg.BaseURL,
		Model:      cfg.Model,
		Dimensions: &dim,
		Timeout:    60,
	})
	if err != nil {
		return nil, err
	}
	return &Client{embedder: e}, nil
}

// EmbedStrings 将文本列表批量向量化，返回 [][]float64
func (c *Client) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	return c.embedder.EmbedStrings(ctx, texts)
}

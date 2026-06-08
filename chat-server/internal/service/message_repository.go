package service

import (
	"context"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
)

// MessageRepository 消息数据访问接口（消费者定义）.
type MessageRepository interface {
	CreateMessage(ctx context.Context, msg *schema.Message) error
	GetMessage(ctx context.Context, id string) (*schema.Message, error)
	ListMessages(ctx context.Context, conversationID string, page, pageSize int) ([]*schema.Message, int64, error)
	UpdateMessageStatus(ctx context.Context, msgID, status string) error
	DeleteMessage(ctx context.Context, id string) error
}

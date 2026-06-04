package schema

// Message 消息表（私聊 + 群聊共用）
type Message struct {
	ID               uint64  `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	ConversationType string  `gorm:"column:conversation_type;type:enum('private','group');not null" json:"conversation_type"`
	ConversationID   string  `gorm:"column:conversation_id;type:varchar(64);not null;index:idx_conversation,priority:1" json:"conversation_id"`
	SeqID            uint64  `gorm:"column:seq_id;type:bigint unsigned;not null;index:idx_conversation,priority:2" json:"seq_id"`
	SenderID         uint64  `gorm:"column:sender_id;type:bigint unsigned;not null;index:idx_sender" json:"sender_id"`
	ContentType      string  `gorm:"column:content_type;type:enum('text','image','file','system_invitation');not null;default:text" json:"content_type"`
	Content          string  `gorm:"column:content;type:text;not null" json:"content"`
	Extra            *string `gorm:"column:extra;type:json;default:null" json:"extra"`
	ReplyToID        *uint64 `gorm:"column:reply_to_id;type:bigint unsigned;default:null" json:"reply_to_id"`
	Status           string  `gorm:"column:status;type:enum('normal','recalled');not null;default:normal" json:"status"`
	CreatedAt        int64   `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
}

func (m *Message) TableName() string {
	return "messages"
}

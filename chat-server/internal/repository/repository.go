package repository

import "gorm.io/gorm"

// Repository 数据仓库，聚合所有数据访问操作
type Repository struct {
	DB           *gorm.DB
	Friend       *FriendRepository
	Message      *MessageRepository
	Group        *GroupRepository
	GroupMember  *GroupMemberRepository
	Conversation *ConversationRepository
}

// NewRepository 创建数据仓库
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		DB:           db,
		Friend:       NewFriendRepository(db),
		Message:      NewMessageRepository(db),
		Group:        NewGroupRepository(db),
		GroupMember:  NewGroupMemberRepository(db),
		Conversation: NewConversationRepository(db),
	}
}

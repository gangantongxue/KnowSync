// Package service 提供业务逻辑层.
package service

// 会话类型常量.
const (
	ConvTypePrivate = "private"
	ConvTypeGroup   = "group"
)

// 群组成员角色常量.
const (
	RoleOwner = "owner"
	RoleAdmin = "admin"
)

// WebSocket 推送常量.
const (
	PushTypeFriendRequest = "friend_request"
	PushTypeNewMessage    = "new_message"
	PushKeyType           = "type"
	PushKeyData           = "data"
)

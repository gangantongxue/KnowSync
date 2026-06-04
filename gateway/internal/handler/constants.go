// Package handler provides HTTP handlers for the gateway layer.
package handler

// Common JSON key constants used in HTTP responses.
const (
	KeyAvatar         = "avatar"
	KeyCode           = "code"
	KeyContent        = "content"
	KeyCreatedAt      = "created_at"
	KeyDescription    = "description"
	KeyEmail          = "email"
	KeyID             = "id"
	KeyMessage        = "message"
	KeyName           = "name"
	KeyOwnerID        = "owner_id"
	KeyPath           = "path"
	KeyRepos          = "repos"
	KeyRole           = "role"
	KeySessionID      = "session_id"
	KeyUpdatedAt      = "updated_at"
	KeyUser           = "user"
	KeyUserID         = "user_id"
	KeyVisibility     = "visibility"
	KeyArticleCount   = "article_count"
	KeyRepo           = "repo"
	KeyGroup          = "group"
	KeyGroups         = "groups"
	KeyMembers        = "members"
	KeyMessages       = "messages"
	KeyUsers          = "users"
	KeyFriends        = "friends"
	KeyFriendRequests = "friend_requests"
	KeyConversations  = "conversations"
	KeySessions       = "sessions"
	KeyCollaborators  = "collaborators"
	KeyHasMore        = "has_more"
	KeyTotal          = "total"
	KeyIsMember       = "is_member"
)

// Common visibility and role string constants.
const (
	KeyVisibilityPUBLIC  = "PUBLIC"
	KeyVisibilityPRIVATE = "PRIVATE"
	KeyRoleADMIN         = "ADMIN"
	KeyRoleDEVELOPER     = "DEVELOPER"
	KeyRoleVIEWER        = "VIEWER"
)

// Common error message constants used in internal endpoints.
const (
	ErrMsgRepoIDRequired    = "repo_id is required"
	ErrMsgUserIDRequired    = "user_id is required"
	ErrMsgRepoServerUnavail = "repo server unavailable"
	ErrMsgOwnerRepoPathReq  = "owner_id, repo_id, path are required"
	ErrMsgNameRequired      = "name is required"
	ErrMsgNewPathRequired   = "new_path is required"
	ErrMsgQRequired         = "q is required"
	ErrMsgInvalidBody       = "invalid body"
	ErrMsgMissingUserID     = "missing user_id"
)

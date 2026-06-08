// Package service 提供业务逻辑层实现，包含权限校验和数据组合.
package service

import (
	"context"
	"log/slog"
)

// CheckRepoPermission 校验用户在知识库中的权限
// requiredRoles 为允许的角色列表，满足任一角色即通过
// 公开知识库的 VIEWER 级别操作对非协作者也开放
// 关注者拥有 VIEWER 权限.
func CheckRepoPermission(
	ctx context.Context,
	repoRepo RepoRepository,
	collabRepo CollaboratorRepository,
	followRepo FollowRepository,
	repoID, userID string,
	requiredRoles ...string,
) error {
	role, err := collabRepo.GetUserRole(ctx, repoID, userID)
	if err != nil {
		if err.Error() != "record not found" {
			slog.Error("查询用户角色失败", "repo_id", repoID, "user_id", userID, "error", err)
			return ErrPermissionDenied
		}
		return checkPublicOrFollowPermission(ctx, repoRepo, followRepo, repoID, userID, requiredRoles)
	}

	if !isRoleSufficient(role, requiredRoles) {
		return ErrPermissionDenied
	}
	return nil
}

// checkPublicOrFollowPermission 检查非协作者是否通过公开知识库或关注关系获得权限.
func checkPublicOrFollowPermission(
	ctx context.Context,
	repoRepo RepoRepository,
	followRepo FollowRepository,
	repoID, userID string,
	requiredRoles []string,
) error {
	repo, repoErr := repoRepo.GetRepo(ctx, repoID)
	if repoErr != nil {
		return ErrNotFound
	}
	if repo.Visibility == "PUBLIC" && isRoleSufficient("VIEWER", requiredRoles) {
		return nil
	}
	if followRepo != nil {
		following, fErr := followRepo.IsFollowing(ctx, userID, repoID)
		if fErr == nil && following && isRoleSufficient("VIEWER", requiredRoles) {
			return nil
		}
	}
	return ErrPermissionDenied
}

// isRoleSufficient 判断角色是否满足所需角色中的最低级别要求.
func isRoleSufficient(role string, requiredRoles []string) bool {
	roleLevel := map[string]int{
		"ADMIN":     3,
		"DEVELOPER": 2,
		"VIEWER":    1,
	}
	requiredLevel := 0
	for _, r := range requiredRoles {
		if l, ok := roleLevel[r]; ok && l > requiredLevel {
			requiredLevel = l
		}
	}
	return roleLevel[role] >= requiredLevel
}

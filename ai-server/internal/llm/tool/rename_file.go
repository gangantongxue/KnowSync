package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// RenameFile 重命名/移动文件工具
type RenameFile struct {
	repoDetailClient RepoDetailClient
	fileWriteClient  FileWriteClient
}

// NewRenameFile 创建 RenameFile 工具
func NewRenameFile(rdc RepoDetailClient, fwc FileWriteClient) *RenameFile {
	return &RenameFile{
		repoDetailClient: rdc,
		fileWriteClient:  fwc,
	}
}

// Info 返回工具元信息
func (r *RenameFile) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "rename_file",
		Desc: "重命名或移动知识库中的文件。如果用户明确说了新路径，可以设置 _skip_confirm: true 跳过确认",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"repo_id": {
				Type:     "string",
				Desc:     "知识库 ID",
				Required: true,
			},
			"old_path": {
				Type:     "string",
				Desc:     "原文件路径，例如：docs/chapter1.md",
				Required: true,
			},
			"new_path": {
				Type:     "string",
				Desc:     "新文件路径，例如：docs/chapter2.md",
				Required: true,
			},
			"_skip_confirm": {
				Type:     "boolean",
				Desc:     "当用户已明确确认所有信息时，设置为 true 跳过二次确认",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用
func (r *RenameFile) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return r.execute(ctx, arguments)
}

func (r *RenameFile) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		RepoID  string `json:"repo_id"`
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.RepoID == "" {
		return `{"success": false, "message": "repo_id 不能为空"}`, nil
	}
	if params.OldPath == "" {
		return `{"success": false, "message": "old_path 不能为空"}`, nil
	}
	if params.NewPath == "" {
		return `{"success": false, "message": "new_path 不能为空"}`, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{"success": false, "message": "无法获取用户信息"}`, nil
	}

	repo, err := r.repoDetailClient.GetRepo(ctx, params.RepoID, userID)
	if err != nil {
		return fmt.Sprintf(`{"success": false, "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	if err := r.fileWriteClient.RenameFile(ctx, repo.OwnerID, params.RepoID, params.OldPath, params.NewPath); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "重命名文件失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		"success":  true,
		"repo_id":  params.RepoID,
		"old_path": params.OldPath,
		"new_path": params.NewPath,
	})
	return string(data), nil
}

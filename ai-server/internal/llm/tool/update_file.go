package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// UpdateFile 更新文件内容工具
type UpdateFile struct {
	repoDetailClient RepoDetailClient
	fileWriteClient  FileWriteClient
}

// NewUpdateFile 创建 UpdateFile 工具
func NewUpdateFile(rdc RepoDetailClient, fwc FileWriteClient) *UpdateFile {
	return &UpdateFile{
		repoDetailClient: rdc,
		fileWriteClient:  fwc,
	}
}

// Info 返回工具元信息
func (u *UpdateFile) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "update_file",
		Desc: "更新知识库中已有文件的内容。如果用户明确说了修改内容，可以设置 _skip_confirm: true 跳过确认直接执行",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"repo_id": {
				Type:     "string",
				Desc:     "知识库 ID",
				Required: true,
			},
			"file_path": {
				Type:     "string",
				Desc:     "文件路径，例如：docs/chapter1.md",
				Required: true,
			},
			"content": {
				Type:     "string",
				Desc:     "新的文件内容（Markdown 格式）",
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
func (u *UpdateFile) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return u.execute(ctx, arguments)
}

func (u *UpdateFile) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		RepoID   string `json:"repo_id"`
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.RepoID == "" {
		return `{"success": false, "message": "repo_id 不能为空"}`, nil
	}
	if params.FilePath == "" {
		return `{"success": false, "message": "file_path 不能为空"}`, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{"success": false, "message": "无法获取用户信息"}`, nil
	}

	repo, err := u.repoDetailClient.GetRepo(ctx, params.RepoID, userID)
	if err != nil {
		return fmt.Sprintf(`{"success": false, "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	if err := u.fileWriteClient.UpdateFile(ctx, repo.OwnerID, params.RepoID, params.FilePath, params.Content); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "更新文件失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		"success":   true,
		"repo_id":   params.RepoID,
		"file_path": params.FilePath,
		"size":      len(params.Content),
	})
	return string(data), nil
}

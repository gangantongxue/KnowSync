package tool //nolint:dupl // 与 create_file.go 结构相似但逻辑不同

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// UpdateFile 更新文件内容工具.
type UpdateFile struct {
	repoDetailClient RepoDetailClient
	fileWriteClient  FileWriteClient
}

// NewUpdateFile 创建 UpdateFile 工具.
func NewUpdateFile(rdc RepoDetailClient, fwc FileWriteClient) *UpdateFile {
	return &UpdateFile{
		repoDetailClient: rdc,
		fileWriteClient:  fwc,
	}
}

// Info 返回工具元信息.
func (u *UpdateFile) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "update_file",
		Desc: "更新知识库中已有文件的内容。如果用户明确说了修改内容，可以设置 _skip_confirm: true 跳过确认直接执行",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			ParamRepoID: {
				Type:     TypeString,
				Desc:     DescRepoID,
				Required: true,
			},
			ParamFilePath: {
				Type:     TypeString,
				Desc:     DescFilePath,
				Required: true,
			},
			ParamContent: {
				Type:     TypeString,
				Desc:     "新的文件内容（Markdown 格式）",
				Required: true,
			},
			ParamSkipCfm: {
				Type:     TypeBoolean,
				Desc:     "当用户已明确确认所有信息时，设置为 true 跳过二次确认",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用.
func (u *UpdateFile) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
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
		return ErrRespRepoIDEmpty, nil
	}
	if params.FilePath == "" {
		return `{"success": false, "message": "file_path 不能为空"}`, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return ErrRespUserInfo, nil
	}

	repo, err := u.repoDetailClient.GetRepo(ctx, params.RepoID, userID)
	if err != nil {
		return fmt.Sprintf(`{"success": false, "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	if err := u.fileWriteClient.UpdateFile(ctx, repo.OwnerID, params.RepoID, params.FilePath, params.Content); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "更新文件失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		KeySuccess:    true,
		ParamRepoID:   params.RepoID,
		ParamFilePath: params.FilePath,
		ParamSize:     len(params.Content),
	})
	return string(data), nil
}

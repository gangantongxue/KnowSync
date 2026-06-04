package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// DeleteFile 删除文件工具.
type DeleteFile struct {
	repoDetailClient RepoDetailClient
	fileWriteClient  FileWriteClient
}

// NewDeleteFile 创建 DeleteFile 工具.
func NewDeleteFile(rdc RepoDetailClient, fwc FileWriteClient) *DeleteFile {
	return &DeleteFile{
		repoDetailClient: rdc,
		fileWriteClient:  fwc,
	}
}

// Info 返回工具元信息.
func (d *DeleteFile) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "delete_file",
		Desc: "删除知识库中的文件。此操作不可撤销，务必先用 ask_user 让用户明确确认后再执行。",
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
		}),
	}, nil
}

// InvokableRun 执行工具调用.
func (d *DeleteFile) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	return d.execute(ctx, arguments)
}

func (d *DeleteFile) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		RepoID   string `json:"repo_id"`
		FilePath string `json:"file_path"`
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

	repo, err := d.repoDetailClient.GetRepo(ctx, params.RepoID, userID)
	if err != nil {
		return fmt.Sprintf(`{"success": false, "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	if err := d.fileWriteClient.DeleteFile(ctx, repo.OwnerID, params.RepoID, params.FilePath); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "删除文件失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		KeySuccess:    true,
		ParamRepoID:   params.RepoID,
		ParamFilePath: params.FilePath,
	})
	return string(data), nil
}

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type RepoStats struct {
	repoDetailClient RepoDetailClient
	fileClient       FileClient
}

func NewRepoStats(rdc RepoDetailClient, fc FileClient) *RepoStats {
	return &RepoStats{
		repoDetailClient: rdc,
		fileClient:       fc,
	}
}

func (r *RepoStats) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "repo_stats",
		Desc: "获取知识库统计数据。不传 repo_id 时返回所有知识库的概览统计；传 repo_id 时返回指定知识库的详细统计（文件数、文件类型分布、总大小等）。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"repo_id": {
				Type:     "string",
				Desc:     "知识库 ID（可选，不传时返回全部知识库概览）",
				Required: false,
			},
		}),
	}, nil
}

func (r *RepoStats) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return r.execute(ctx, arguments)
}

func (r *RepoStats) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		RepoID string `json:"repo_id"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{"success": false, "message": "无法获取用户信息"}`, nil
	}

	if params.RepoID != "" {
		return r.repoDetailStats(ctx, userID, params.RepoID)
	}
	return r.overviewStats(ctx, userID)
}

// overviewStats 所有知识库概览统计
func (r *RepoStats) overviewStats(ctx context.Context, userID string) (string, error) {
	repos, err := r.repoDetailClient.ListUserReposDetail(ctx, userID)
	if err != nil {
		return `{"success": false, "message": "获取知识库列表失败"}`, nil
	}

	if len(repos) == 0 {
		return `{"success": true, "total_repos": 0, "total_articles": 0, "repos": []}`, nil
	}

	var totalArticles int64
	type repoStat struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		Visibility    string `json:"visibility"`
		ArticleCount  int64  `json:"article_count"`
		FollowerCount int64  `json:"follower_count"`
	}
	items := make([]repoStat, 0, len(repos))

	for _, repo := range repos {
		totalArticles += repo.ArticleCount
		desc := repo.Description
		if desc == "" {
			desc = "暂无描述"
		}
		items = append(items, repoStat{
			ID:            repo.ID,
			Name:          repo.Name,
			Description:   desc,
			Visibility:    repo.Visibility,
			ArticleCount:  repo.ArticleCount,
			FollowerCount: repo.FollowerCount,
		})
	}

	// 按文章数降序排列
	sort.Slice(items, func(i, j int) bool {
		return items[i].ArticleCount > items[j].ArticleCount
	})

	data, _ := json.Marshal(map[string]any{
		"success":        true,
		"total_repos":    len(repos),
		"total_articles": totalArticles,
		"repos":          items,
	})
	return string(data), nil
}

// repoDetailStats 单个知识库详细统计
func (r *RepoStats) repoDetailStats(ctx context.Context, userID, repoID string) (string, error) {
	repo, err := r.repoDetailClient.GetRepo(ctx, repoID, userID)
	if err != nil {
		return fmt.Sprintf(`{"success": false, "message": "获取知识库信息失败: %s"}`, err.Error()), nil
	}

	// 递归遍历文件
	allFiles, dirCount, err := r.walkFiles(ctx, userID, repoID, "")
	if err != nil {
		return fmt.Sprintf(`{"success": false, "message": "遍历文件失败: %s"}`, err.Error()), nil
	}

	// 统计文件类型分布
	type extStat struct {
		Extension string `json:"extension"`
		Count     int    `json:"count"`
		TotalSize int64  `json:"total_size"`
	}
	extMap := make(map[string]*extStat)
	var totalSize int64

	for _, f := range allFiles {
		totalSize += f.Size
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext == "" {
			ext = "(无扩展名)"
		}
		if _, ok := extMap[ext]; !ok {
			extMap[ext] = &extStat{Extension: ext}
		}
		extMap[ext].Count++
		extMap[ext].TotalSize += f.Size
	}

	typeBreakdown := make([]extStat, 0, len(extMap))
	for _, s := range extMap {
		typeBreakdown = append(typeBreakdown, *s)
	}
	sort.Slice(typeBreakdown, func(i, j int) bool {
		return typeBreakdown[i].Count > typeBreakdown[j].Count
	})

	data, _ := json.Marshal(map[string]any{
		"success":        true,
		"repo_id":        repoID,
		"name":           repo.Name,
		"description":    repo.Description,
		"visibility":     repo.Visibility,
		"article_count":  repo.ArticleCount,
		"follower_count": repo.FollowerCount,
		"total_files":    len(allFiles),
		"total_dirs":     dirCount,
		"total_size":     totalSize,
		"file_types":     typeBreakdown,
	})
	return string(data), nil
}

// walkFiles 递归遍历仓库文件
func (r *RepoStats) walkFiles(ctx context.Context, ownerID, repoID, dirPath string) (files []FileEntry, dirCount int, err error) {
	entries, err := r.fileClient.ListRepoFiles(ctx, ownerID, repoID, dirPath)
	if err != nil {
		return nil, 0, err
	}

	for _, e := range entries {
		if e.Type == "dir" {
			dirCount++
			subFiles, subDirs, err := r.walkFiles(ctx, ownerID, repoID, e.Path)
			if err != nil {
				continue
			}
			files = append(files, subFiles...)
			dirCount += subDirs
		} else {
			files = append(files, e)
		}
	}
	return files, dirCount, nil
}

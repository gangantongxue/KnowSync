package handler

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// CreateNodeReq 创建节点请求体
type CreateNodeReq struct {
	ParentID string `json:"parent_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

// UpdateNodeReq 更新节点请求体
type UpdateNodeReq struct {
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
}

// CreateNode 创建节点（文件夹或文章）
func (h *Handler) CreateNode() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")

		var req CreateNodeReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		nodeType := pb.NodeType_NODE_TYPE_UNSPECIFIED
		switch req.Type {
		case "FOLDER":
			nodeType = pb.NodeType_FOLDER
		case "ARTICLE":
			nodeType = pb.NodeType_ARTICLE
		default:
			response.Error(c, ctx, 400, errcode.ErrBadReq, "节点类型错误，仅支持 FOLDER/ARTICLE")
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.CreateNode(c, &pb.CreateNodeRequest{
			RepoId:   repoID,
			UserId:   uid,
			ParentId: req.ParentID,
			Name:     req.Name,
			Type:     nodeType,
		})
		if err != nil {
			logGrpcError("repo_server", "CreateNode", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "创建节点失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, map[string]interface{}{
			"node": h.marshalNodeResponse(resp.Node),
		})
	}
}

// GetNode 获取节点信息
func (h *Handler) GetNode() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		nodeID := ctx.Param("node_id")

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.GetNode(c, &pb.GetNodeRequest{
			RepoId: repoID,
			NodeId: nodeID,
			UserId: uid,
		})
		if err != nil {
			logGrpcError("repo_server", "GetNode", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取节点失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 404, errcode.ErrNotFound, resp.Msg)
			return
		}

		response.Success(c, ctx, map[string]interface{}{
			"node": h.marshalNodeResponse(resp.Node),
		})
	}
}

// UpdateNode 更新节点（重命名或移动）
func (h *Handler) UpdateNode() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		nodeID := ctx.Param("node_id")

		var req UpdateNodeReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.UpdateNode(c, &pb.UpdateNodeRequest{
			RepoId:   repoID,
			UserId:   uid,
			NodeId:   nodeID,
			Name:     req.Name,
			ParentId: req.ParentID,
		})
		if err != nil {
			logGrpcError("repo_server", "UpdateNode", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "更新节点失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, map[string]interface{}{
			"node": h.marshalNodeResponse(resp.Node),
		})
	}
}

// DeleteNode 删除节点（递归删除子节点）
func (h *Handler) DeleteNode() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		nodeID := ctx.Param("node_id")

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.DeleteNode(c, &pb.DeleteNodeRequest{
			RepoId: repoID,
			UserId: uid,
			NodeId: nodeID,
		})
		if err != nil {
			logGrpcError("repo_server", "DeleteNode", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "删除节点失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		for _, path := range resp.DeletedFilePaths {
			if err := h.store.Delete(storage.BucketAuth, path); err != nil {
				slog.Warn("删除存储文件失败", "path", path, "error", err)
			}
		}

		response.Success(c, ctx, map[string]interface{}{
			"deleted_node_ids":   resp.DeletedNodeIds,
			"deleted_file_paths": resp.DeletedFilePaths,
		})
	}
}

// ListNodes 列出指定父目录下的子节点列表
func (h *Handler) ListNodes() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		parentID := ctx.Query("parent_id")

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.ListNodes(c, &pb.ListNodesRequest{
			RepoId:   repoID,
			UserId:   uid,
			ParentId: parentID,
		})
		if err != nil {
			logGrpcError("repo_server", "ListNodes", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取节点列表失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "获取节点列表失败")
			return
		}

		nodes := make([]map[string]interface{}, 0, len(resp.Nodes))
		for _, n := range resp.Nodes {
			nodes = append(nodes, h.marshalNodeResponse(n))
		}

		response.Success(c, ctx, map[string]interface{}{
			"nodes": nodes,
		})
	}
}

// UploadArticleContent 上传文章内容到网关存储并记录路径
func (h *Handler) UploadArticleContent() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		nodeID := ctx.Param("node_id")

		fileHeader, err := ctx.FormFile("file")
		if err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "文件上传失败")
			return
		}

		ext := filepath.Ext(fileHeader.Filename)
		if ext != ".md" && ext != ".markdown" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "仅支持 Markdown 文件")
			return
		}

		f, err := fileHeader.Open()
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "文件读取失败")
			return
		}
		defer f.Close()

		storageKey := fmt.Sprintf("repo/%s/%s/%s", repoID, nodeID, fileHeader.Filename)
		fi, err := h.store.Put(storage.BucketAuth, storageKey, f)
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "文件存储失败")
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		setResp, err := client.SetArticleContent(c, &pb.SetArticleContentRequest{
			RepoId:   repoID,
			UserId:   uid,
			NodeId:   nodeID,
			FilePath: storageKey,
			Size:     fi.Size,
		})
		if err != nil {
			logGrpcError("repo_server", "SetArticleContent", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "更新文章内容失败")
			return
		}
		if !setResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, setResp.Msg)
			return
		}

		signedURL, err := h.store.SignTempURL(storage.BucketAuth, storageKey)
		if err != nil {
			slog.Warn("生成签名 URL 失败", "error", err)
		}

		response.Success(c, ctx, map[string]interface{}{
			"node":       h.marshalNodeResponse(setResp.Node),
			"signed_url": signedURL,
		})
	}
}

// GetArticleSignedURL 获取文章内容的签名临时访问 URL
func (h *Handler) GetArticleSignedURL() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		nodeID := ctx.Param("node_id")

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		nodeResp, err := client.GetNode(c, &pb.GetNodeRequest{
			RepoId: repoID,
			NodeId: nodeID,
			UserId: uid,
		})
		if err != nil {
			logGrpcError("repo_server", "GetNode", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取节点失败")
			return
		}
		if !nodeResp.Success {
			response.Error(c, ctx, 404, errcode.ErrNotFound, nodeResp.Msg)
			return
		}

		if nodeResp.Node.FilePath == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "该节点没有关联文件")
			return
		}

		signedURL, err := h.store.SignTempURL(storage.BucketAuth, nodeResp.Node.FilePath)
		if err != nil {
			slog.Error("生成签名 URL 失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "生成访问链接失败")
			return
		}

		response.Success(c, ctx, map[string]interface{}{
			"signed_url": signedURL,
		})
	}
}

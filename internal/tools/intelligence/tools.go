package intelligence

import (
	"context"

	gogithub "github.com/google/go-github/v68/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitops-mcp-server/internal/github"
)

// RegisterIntelligenceTools 注册 Code Intelligence 相关 Tools
func RegisterIntelligenceTools(mcpServer *server.MCPServer, ghClient *github.Client) {
	gitSvc := ghClient.Git()

	// get_recent_commits
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_recent_commits",
		Description: "获取仓库最近的提交记录",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称",
				},
				"sha": map[string]any{
					"type":        "string",
					"description": "分支或 SHA（默认：默认分支）",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "限定文件路径",
				},
				"author": map[string]any{
					"type":        "string",
					"description": "限定作者",
				},
				"per_page": map[string]any{
					"type":        "integer",
					"description": "每页数量，默认 30",
				},
			},
			Required: []string{"repo"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		sha := request.GetString("sha", "")
		path := request.GetString("path", "")
		author := request.GetString("author", "")
		perPage := request.GetInt("per_page", 30)

		opts := &gogithub.CommitsListOptions{
			ListOptions: gogithub.ListOptions{PerPage: perPage},
		}
		if sha != "" {
			opts.SHA = sha
		}
		if path != "" {
			opts.Path = path
		}
		if author != "" {
			opts.Author = author
		}

		commits, err := gitSvc.ListCommits(ctx, owner, repo, opts)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(github.FormatCommitList(commits)), nil
	})

	// get_commit_detail
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_commit_detail",
		Description: "获取 commit 详细信息，包括变更统计和文件列表",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称",
				},
				"sha": map[string]any{
					"type":        "string",
					"description": "Commit SHA",
				},
			},
			Required: []string{"repo", "sha"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		sha := request.GetString("sha", "")

		detail, err := gitSvc.GetCommit(ctx, owner, repo, sha)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(github.FormatCommitDetail(detail)), nil
	})

	// compare_refs
	mcpServer.AddTool(mcp.Tool{
		Name:        "compare_refs",
		Description: "对比两个 ref（分支/tag/SHA）之间的差异",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称",
				},
				"base": map[string]any{
					"type":        "string",
					"description": "基准 ref",
				},
				"head": map[string]any{
					"type":        "string",
					"description": "对比 ref",
				},
			},
			Required: []string{"repo", "base", "head"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		base := request.GetString("base", "")
		head := request.GetString("head", "")

		result, err := gitSvc.CompareRefs(ctx, owner, repo, base, head)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(result), nil
	})

	// analyze_changes
	mcpServer.AddTool(mcp.Tool{
		Name:        "analyze_changes",
		Description: "分析两个 ref 之间的变更影响，包括文件类型分布、影响目录、风险提示",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称",
				},
				"base": map[string]any{
					"type":        "string",
					"description": "基准 ref",
				},
				"head": map[string]any{
					"type":        "string",
					"description": "对比 ref",
				},
			},
			Required: []string{"repo", "base", "head"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		base := request.GetString("base", "")
		head := request.GetString("head", "")

		result, err := gitSvc.AnalyzeChanges(ctx, owner, repo, base, head)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(result), nil
	})
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: text},
		},
	}
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: "❌ " + msg},
		},
		IsError: true,
	}
}

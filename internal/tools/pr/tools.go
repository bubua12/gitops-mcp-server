package pr

import (
	"context"
	"fmt"

	gogithub "github.com/google/go-github/v68/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitops-mcp-server/internal/github"
)

// RegisterPRTools 注册 PR 管理相关 Tools
func RegisterPRTools(mcpServer *server.MCPServer, ghClient *github.Client) {
	prSvc := ghClient.PullRequests()

	// list_pull_requests
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_pull_requests",
		Description: "列出仓库的 Pull Requests",
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
				"state": map[string]any{
					"type":        "string",
					"enum":        []string{"open", "closed", "all"},
					"description": "PR 状态，默认 open",
				},
				"base": map[string]any{
					"type":        "string",
					"description": "目标分支过滤",
				},
				"head": map[string]any{
					"type":        "string",
					"description": "源分支过滤",
				},
				"sort": map[string]any{
					"type":        "string",
					"enum":        []string{"created", "updated", "popularity", "long-running"},
					"description": "排序方式",
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
		state := request.GetString("state", "open")
		base := request.GetString("base", "")
		head := request.GetString("head", "")
		sort := request.GetString("sort", "created")
		perPage := request.GetInt("per_page", 30)

		opts := &gogithub.PullRequestListOptions{
			State:     state,
			Sort:      sort,
			Direction: "desc",
			ListOptions: gogithub.ListOptions{PerPage: perPage},
		}
		if base != "" {
			opts.Base = base
		}
		if head != "" {
			opts.Head = head
		}

		prs, err := prSvc.List(ctx, owner, repo, opts)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(github.FormatPRList(prs, owner, repo)), nil
	})

	// get_pull_request
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_pull_request",
		Description: "获取 PR 详情，包括审查状态、CI 状态、合并状态",
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
				"number": map[string]any{
					"type":        "integer",
					"description": "PR 编号",
				},
			},
			Required: []string{"repo", "number"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		number := request.GetInt("number", 0)

		pr, err := prSvc.Get(ctx, owner, repo, number)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		result := github.FormatPR(pr)

		// 获取审查历史
		reviews, err := prSvc.GetReviews(ctx, owner, repo, number)
		if err == nil && len(reviews) > 0 {
			result += "\n" + github.FormatReviewList(reviews)
		}

		return textResult(result), nil
	})

	// get_pr_diff
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_pr_diff",
		Description: "获取 PR 的代码 diff",
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
				"number": map[string]any{
					"type":        "integer",
					"description": "PR 编号",
				},
			},
			Required: []string{"repo", "number"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		number := request.GetInt("number", 0)

		diff, err := prSvc.GetDiff(ctx, owner, repo, number)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		result := fmt.Sprintf("## PR #%d Diff\n\n```\n%s\n```", number, diff)
		return textResult(result), nil
	})

	// get_pr_files
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_pr_files",
		Description: "获取 PR 中变更的文件列表及每个文件的变更统计",
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
				"number": map[string]any{
					"type":        "integer",
					"description": "PR 编号",
				},
			},
			Required: []string{"repo", "number"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		number := request.GetInt("number", 0)

		files, err := prSvc.GetFiles(ctx, owner, repo, number)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(github.FormatDiffFiles(files)), nil
	})

	// create_pull_request
	mcpServer.AddTool(mcp.Tool{
		Name:        "create_pull_request",
		Description: "创建 Pull Request",
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
				"title": map[string]any{
					"type":        "string",
					"description": "PR 标题",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "PR 描述",
				},
				"head": map[string]any{
					"type":        "string",
					"description": "源分支",
				},
				"base": map[string]any{
					"type":        "string",
					"description": "目标分支",
				},
				"draft": map[string]any{
					"type":        "boolean",
					"description": "是否为草稿，默认 false",
				},
			},
			Required: []string{"repo", "title", "head", "base"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		title := request.GetString("title", "")
		body := request.GetString("body", "")
		head := request.GetString("head", "")
		base := request.GetString("base", "")
		draft := request.GetBool("draft", false)

		req := &gogithub.NewPullRequest{
			Title: &title,
			Body:  &body,
			Head:  &head,
			Base:  &base,
			Draft: &draft,
		}

		pr, err := prSvc.Create(ctx, owner, repo, req)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 已创建 PR #%d\n\n%s", pr.Number, github.FormatPR(pr))), nil
	})

	// create_review
	mcpServer.AddTool(mcp.Tool{
		Name:        "create_review",
		Description: "对 PR 提交代码审查。支持 APPROVE / REQUEST_CHANGES / COMMENT",
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
				"number": map[string]any{
					"type":        "integer",
					"description": "PR 编号",
				},
				"event": map[string]any{
					"type":        "string",
					"enum":        []string{"APPROVE", "REQUEST_CHANGES", "COMMENT"},
					"description": "审查事件类型",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "审查总评",
				},
			},
			Required: []string{"repo", "number", "event"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		number := request.GetInt("number", 0)
		event := request.GetString("event", "COMMENT")
		body := request.GetString("body", "")

		req := &gogithub.PullRequestReviewRequest{
			Event: &event,
			Body:  &body,
		}

		review, err := prSvc.CreateReview(ctx, owner, repo, number, req)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 审查已提交 (ID: %d, 状态: %s)", review.ID, review.State)), nil
	})

	// merge_pull_request
	mcpServer.AddTool(mcp.Tool{
		Name:        "merge_pull_request",
		Description: "合并 Pull Request",
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
				"number": map[string]any{
					"type":        "integer",
					"description": "PR 编号",
				},
				"merge_method": map[string]any{
					"type":        "string",
					"enum":        []string{"merge", "squash", "rebase"},
					"description": "合并方式，默认 merge",
				},
			},
			Required: []string{"repo", "number"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		number := request.GetInt("number", 0)
		mergeMethod := request.GetString("merge_method", "merge")

		req := &gogithub.PullRequestOptions{
			MergeMethod: mergeMethod,
		}

		msg, err := prSvc.Merge(ctx, owner, repo, number, req)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ PR #%d 已合并: %s", number, msg)), nil
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

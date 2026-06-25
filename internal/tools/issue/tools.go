package issue

import (
	"context"
	"fmt"

	gogithub "github.com/google/go-github/v68/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitops-mcp-server/internal/github"
)

// RegisterIssueTools 注册 Issue 相关 Tools（读操作）
func RegisterIssueTools(mcpServer *server.MCPServer, ghClient *github.Client) {
	issueSvc := ghClient.Issues()

	// list_issues
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_issues",
		Description: "列出仓库的 Issues。支持按状态、标签、指派人过滤",
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
					"description": "Issue 状态，默认 open",
				},
				"labels": map[string]any{
					"type":        "string",
					"description": "逗号分隔的标签名",
				},
				"sort": map[string]any{
					"type":        "string",
					"enum":        []string{"created", "updated", "comments"},
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
		labels := request.GetString("labels", "")
		sort := request.GetString("sort", "created")
		perPage := request.GetInt("per_page", 30)

		opts := &gogithub.IssueListByRepoOptions{
			State:     state,
			Sort:      sort,
			Direction: "desc",
			ListOptions: gogithub.ListOptions{PerPage: perPage},
		}
		if labels != "" {
			opts.Labels = []string{labels}
		}

		issues, err := issueSvc.List(ctx, owner, repo, opts)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(github.FormatIssueList(issues, owner, repo)), nil
	})

	// get_issue
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_issue",
		Description: "获取 Issue 详情，包含全部评论",
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
					"description": "Issue 编号",
				},
			},
			Required: []string{"repo", "number"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		number := request.GetInt("number", 0)

		issue, err := issueSvc.Get(ctx, owner, repo, number)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		result := github.FormatIssue(issue)

		// 获取评论
		comments, err := issueSvc.ListComments(ctx, owner, repo, number, nil)
		if err == nil && len(comments) > 0 {
			result += "\n" + github.FormatCommentList(comments, number)
		}

		return textResult(result), nil
	})

	// search_issues
	mcpServer.AddTool(mcp.Tool{
		Name:        "search_issues",
		Description: "使用 GitHub 搜索语法搜索 Issue",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "搜索查询，支持 GitHub Issue 搜索语法，如 'is:issue is:open label:bug'",
				},
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者（可选）",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称（可选）",
				},
				"sort": map[string]any{
					"type":        "string",
					"enum":        []string{"created", "updated", "comments"},
					"description": "排序方式",
				},
				"per_page": map[string]any{
					"type":        "integer",
					"description": "每页数量，默认 30",
				},
			},
			Required: []string{"query"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := request.GetString("query", "")
		sort := request.GetString("sort", "")
		perPage := request.GetInt("per_page", 30)

		opts := &gogithub.SearchOptions{
			ListOptions: gogithub.ListOptions{PerPage: perPage},
		}
		if sort != "" {
			opts.Sort = sort
		}

		issues, total, err := issueSvc.Search(ctx, query, opts)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		var result string
		result = fmt.Sprintf("🔍 Issue 搜索结果 (共 %d 个)\n\n", total)
		for _, issue := range issues {
			state := "🟢"
			if issue.State == "closed" {
				state = "🔴"
			}
			result += fmt.Sprintf("%s #%d %s (by %s)\n", state, issue.Number, issue.Title, issue.User)
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
